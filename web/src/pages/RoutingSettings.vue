<template>
  <div>
    <!-- Error Banner -->
    <div v-if="validationError" class="mb-4 p-4 rounded-lg border border-[var(--red)] bg-[rgba(248,113,113,.1)] text-[var(--red)] text-[13px]">
      {{ validationError }}
    </div>

    <!-- Task-to-Route Mapping Table -->
    <div class="card mb-6">
      <div class="p-5 border-b border-[var(--border)] flex items-center justify-between">
        <div>
          <div class="text-[14px] font-semibold">Task-to-Route Mapping</div>
          <div class="text-[11.5px] text-[var(--text-mute)]">Atur route profile untuk setiap kategori task yang terdeteksi</div>
        </div>
        <div class="flex gap-2">
          <button class="btn btn-secondary" @click="resetAll">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="1 4 1 10 7 10"/><path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/></svg>
            Reset Default
          </button>
          <button class="btn btn-primary" :disabled="!dirty" @click="save">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"/><polyline points="17 21 17 13 7 13 7 21"/><polyline points="7 3 7 8 15 8"/></svg>
            Save Changes
          </button>
        </div>
      </div>

      <!-- Table Header -->
      <div class="route-row" style="background: var(--bg-2); color: var(--text-mute); font-size: 11px; font-weight: 600; text-transform: uppercase; letter-spacing: .05em;">
        <div>Task Category</div>
        <div>Route Profile</div>
        <div>Status</div>
        <div>Actions</div>
      </div>

      <!-- Loading -->
      <div v-if="isLoading" class="p-8 text-center text-[var(--text-mute)] text-[13px]">
        Loading routing configuration...
      </div>

      <!-- Empty State -->
      <div v-else-if="taskRoutesEmpty" class="p-8 text-center text-[var(--text-mute)] text-[13px]">
        No task mappings available. Load routing settings to view and edit the task-to-route mapping list.
      </div>

      <!-- Rows -->
      <template v-else>
        <div v-for="task in orderedTaskRoutes" :key="task" class="route-row">
          <div>
            <div class="font-medium flex items-center gap-2">
              <span class="w-2 h-2 rounded-full" :class="taskIndicatorClass(task)"></span>
              {{ formatTaskName(task) }}
            </div>
            <div class="text-[11.5px] text-[var(--text-mute)] mt-0.5">{{ taskDescriptions[task] || task }}</div>
          </div>
          <select
            class="select"
            :value="taskRoutes[task]"
            @change="(e) => { taskRoutes[task] = (e.target as HTMLSelectElement).value; markDirty(); }"
          >
            <option v-for="profile in enabledProfileNames" :key="profile" :value="profile">{{ profile }}</option>
            <option v-if="!enabledProfileNames.includes(taskRoutes[task])" :value="taskRoutes[task]">{{ taskRoutes[task] }}</option>
          </select>
          <div>
            <span class="badge" :class="routeStatusLabel(task) === 'Ready' ? 'badge-green' : routeStatusLabel(task) === 'Disabled profile' ? 'badge-yellow' : 'badge-red'">
              {{ routeStatusLabel(task) }}
            </span>
          </div>
          <div>
            <button class="btn btn-ghost" style="padding: 4px 10px; font-size: 12px;" @click="resetTask(task)">Reset</button>
          </div>
        </div>
      </template>
    </div>

    <!-- Default Route + Routing Behavior -->
    <div class="grid grid-cols-2 gap-4">
      <!-- Default Route -->
      <div class="card p-5">
        <div class="text-[14px] font-semibold mb-1">Default Route</div>
        <div class="text-[11.5px] text-[var(--text-mute)] mb-4">Digunakan saat task tidak dikenali</div>
        <select
          class="select"
          v-model="routingConfig.default_route"
          @change="markDirty"
        >
          <option v-for="profile in enabledProfileNames" :key="profile" :value="profile">{{ profile }}</option>
        </select>
        <div class="divider"></div>
        <div class="text-[12px] text-[var(--text-dim)]">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="inline mr-1"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>
          Safe passthrough akan aktif jika classifier gagal
        </div>
      </div>

      <!-- Routing Behavior -->
      <div class="card p-5">
        <div class="text-[14px] font-semibold mb-1">Routing Behavior</div>
        <div class="text-[11.5px] text-[var(--text-mute)] mb-4">Kontrol perilaku routing otomatis</div>
        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <div>
              <div class="text-[13px] font-medium">Auto-routing</div>
              <div class="text-[11.5px] text-[var(--text-mute)]">Klasifikasi otomatis setiap request</div>
            </div>
            <div
              class="toggle"
              :class="{ on: routingConfig.auto_routing }"
              @click="routingConfig.auto_routing = !routingConfig.auto_routing; markDirty()"
            ></div>
          </div>
          <div class="flex items-center justify-between">
            <div>
              <div class="text-[13px] font-medium">Confidence Threshold</div>
              <div class="text-[11.5px] text-[var(--text-mute)]">Minimum confidence untuk auto-route</div>
            </div>
            <div class="flex items-center gap-2">
              <input
                type="range" min="0" max="100"
                :value="Math.round(routingConfig.confidence_threshold * 100)"
                class="w-24"
                @input="(e) => { routingConfig.confidence_threshold = +(+(e.target as HTMLInputElement).value / 100).toFixed(2); markDirty(); }"
              >
              <span class="mono text-[12px] w-8">{{ routingConfig.confidence_threshold.toFixed(2) }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Config error -->
    <div v-if="configStore.error" class="mt-4 p-4 rounded-lg border border-[var(--red)] bg-[rgba(248,113,113,.1)] text-[var(--red)] text-[13px]">
      {{ configStore.error }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useConfigStore } from "../stores/config";

const configStore = useConfigStore();
const dirty = ref(false);
const validationError = ref<string | null>(null);

const taskRoutes = ref<Record<string, string>>({});
const routingConfig = ref({
  auto_routing: true,
  default_route: "route.default",
  low_confidence_route: "route.default",
  confidence_threshold: 0.55,
  smart_fast_route: "route.low_cost",
  metadata_transport: "header",
});

const defaultRoutes: Record<string, string> = {
  general_chat: "route.default",
  design_task: "route.design",
  backend_engineering: "route.backend",
  frontend_engineering: "route.frontend",
  fullstack_engineering: "route.fullstack",
  debugging: "route.debugging",
  refactoring: "route.refactoring",
  test_generation: "route.testing",
  documentation: "route.documentation",
  architecture_design: "route.architect",
  security_review: "route.security",
  long_context_analysis: "route.long_context",
  lightweight_task: "route.low_cost",
  unknown: "route.default",
};

const defaultRoutingConfig = {
  auto_routing: true,
  default_route: "route.default",
  low_confidence_route: "route.default",
  confidence_threshold: 0.55,
  smart_fast_route: "route.low_cost",
  metadata_transport: "header",
};

const taskDescriptions: Record<string, string> = {
  general_chat: "General assistance and broad conversation routing.",
  design_task: "UI/UX, visual, product design",
  backend_engineering: "API, database, service, auth",
  frontend_engineering: "Component, CSS, UI behavior",
  fullstack_engineering: "Frontend + backend sekaligus",
  debugging: "Error analysis, root cause",
  refactoring: "Code cleanup, structure",
  test_generation: "Unit test, integration test",
  documentation: "README, docstring, comments",
  architecture_design: "System design, blueprint",
  security_review: "Risk analysis, vulnerability",
  long_context_analysis: "Multi-file, large context",
  lightweight_task: "Simple, short, low-complexity",
  unknown: "Fallback classification for uncategorized input.",
};

const taskOrder = [
  "design_task", "backend_engineering", "frontend_engineering",
  "fullstack_engineering", "debugging", "refactoring", "documentation",
  "architecture_design", "lightweight_task", "security_review",
  "long_context_analysis", "general_chat", "test_generation", "unknown",
];

const profileNames = computed(() => {
  if (!configStore.profiles) return [];
  return Object.keys(configStore.profiles.route_profiles);
});

const enabledProfileNames = computed(() => {
  if (!configStore.profiles) return [];
  return Object.entries(configStore.profiles.route_profiles)
    .filter(([_, profile]) => profile.enabled)
    .map(([name]) => name);
});

const orderedTaskRoutes = computed(() => {
  const keys = new Set<string>([...taskOrder, ...Object.keys(taskRoutes.value || {})]);
  return Array.from(keys).filter((task) => taskRoutes.value[task] !== undefined);
});

const taskRoutesEmpty = computed(() => orderedTaskRoutes.value.length === 0);
const isLoading = computed(() => configStore.loading);

function formatTaskName(task: string) {
  return task.replace(/_/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
}

function isProfileEnabled(profileName: string): boolean {
  if (!configStore.profiles) return true;
  const profile = configStore.profiles.route_profiles[profileName];
  return profile ? profile.enabled : true;
}

function validate(): string | null {
  for (const [task, route] of Object.entries(taskRoutes.value)) {
    if (!profileNames.value.includes(route)) {
      return `Task "${task}" references non-existent profile "${route}"`;
    }
  }
  if (routingConfig.value.default_route && !enabledProfileNames.value.includes(routingConfig.value.default_route)) {
    return `Default route "${routingConfig.value.default_route}" is disabled or does not exist`;
  }
  return null;
}

function load() {
  if (configStore.routes) taskRoutes.value = { ...configStore.routes.task_routes };
  if (configStore.config) routingConfig.value = { ...configStore.config.routing };
  validationError.value = null;
  dirty.value = false;
}

function markDirty() { dirty.value = true; validationError.value = null; }

function resetTask(task: string) {
  taskRoutes.value[task] = defaultRoutes[task] || "route.default";
  dirty.value = true;
}

function resetAll() {
  taskRoutes.value = { ...defaultRoutes };
  routingConfig.value = { ...defaultRoutingConfig };
  dirty.value = true;
}

async function save() {
  const error = validate();
  if (error) { validationError.value = error; return; }
  try {
    await configStore.saveRoutes({ task_routes: taskRoutes.value });
    await configStore.saveConfig({ routing: routingConfig.value });
    validationError.value = null;
    dirty.value = false;
  } catch (e: any) {
    validationError.value = e.message;
  }
}

function taskIndicatorClass(task: string) {
  const route = taskRoutes.value[task];
  if (!route) return "bg-slate-500";
  if (route.includes("security")) return "bg-emerald-400";
  if (route.includes("debug")) return "bg-amber-400";
  if (route.includes("long_context")) return "bg-violet-400";
  return "bg-cyan-400";
}

function routeStatusLabel(task: string) {
  const route = taskRoutes.value[task];
  if (!route) return "Unassigned";
  if (!profileNames.value.includes(route)) return "Profile missing";
  if (!isProfileEnabled(route)) return "Disabled profile";
  return "Ready";
}

onMounted(load);
</script>
