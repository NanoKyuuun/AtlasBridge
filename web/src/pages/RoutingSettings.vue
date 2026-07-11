<template>
  <div class="space-y-6 lg:space-y-8">
    <PageHeader
      eyebrow="Routing"
      title="Task-to-Route Mapping"
      description="Map detected task types to route profiles. Changes take effect on the next request."
    />

    <div v-if="validationError" class="rounded-2xl border border-rose-400/20 bg-rose-400/10 px-4 py-3 text-sm text-rose-100">
      <span>{{ validationError }}</span>
    </div>

    <GlassCard>
      <div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <h2 class="text-lg font-semibold text-white">Auto-routing</h2>
          <p class="mt-1 text-sm text-slate-400">
            Control whether task classification routes automatically or falls back to manual overrides.
          </p>
        </div>

        <label class="inline-flex items-center gap-3 rounded-2xl border border-white/10 bg-white/5 px-4 py-3">
          <div class="space-y-1">
            <div class="text-sm font-medium text-white">Automatic task routing</div>
            <div class="text-xs text-slate-400">Enable model-path selection based on detected task type.</div>
          </div>
          <input
            type="checkbox"
            class="toggle toggle-primary toggle-lg border-white/20 bg-slate-800/80 checked:border-cyan-300 checked:bg-gradient-to-r checked:from-blue-500 checked:to-violet-500"
            v-model="routingConfig.auto_routing"
            @change="markDirty"
          />
        </label>
      </div>
    </GlassCard>

    <GlassCard>
      <div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
        <div>
          <h2 class="text-lg font-semibold text-white">Task-to-Route Mapping</h2>
          <p class="mt-1 text-sm text-slate-400">
            Each task category can be mapped to a specific route profile.
          </p>
        </div>
        <div class="flex flex-wrap gap-2 text-xs text-slate-400">
          <span class="rounded-full border border-white/10 bg-white/5 px-3 py-1">{{ orderedTaskRoutes.length }} tasks</span>
          <span class="rounded-full border border-white/10 bg-white/5 px-3 py-1">{{ enabledProfileNames.length }} enabled profiles</span>
        </div>
      </div>

      <div v-if="isLoading" class="mt-5 rounded-2xl border border-white/10 bg-white/5 p-6 text-sm text-slate-400">
        Loading routing configuration...
      </div>

      <div v-else-if="taskRoutesEmpty" class="mt-5">
        <EmptyState
          title="No task mappings available"
          description="Load routing settings to view and edit the task-to-route mapping list."
        >
          <template #icon>
            <span class="text-xl text-cyan-300">↗</span>
          </template>
        </EmptyState>
      </div>

      <div v-else class="mt-5 overflow-hidden rounded-3xl border border-white/10 bg-white/5">
        <div class="grid grid-cols-12 gap-3 border-b border-white/10 bg-white/5 px-4 py-3 text-[11px] uppercase tracking-[0.22em] text-slate-500">
          <div class="col-span-12 md:col-span-4">Task Category</div>
          <div class="col-span-12 md:col-span-4">Route Profile</div>
          <div class="col-span-12 md:col-span-3">Status</div>
          <div class="col-span-12 md:col-span-1 text-right">Reset</div>
        </div>

        <div
          v-for="task in orderedTaskRoutes"
          :key="task"
          class="grid grid-cols-12 gap-3 border-b border-white/5 px-4 py-4 last:border-b-0"
        >
          <div class="col-span-12 md:col-span-4">
            <div class="flex items-start gap-3">
              <span class="mt-1 h-2.5 w-2.5 shrink-0 rounded-full" :class="taskIndicatorClass(task)"></span>
              <div>
                <div class="font-mono text-sm text-white">{{ task }}</div>
                <div class="mt-1 text-xs leading-5 text-slate-400">
                  {{ taskDescriptions[task] || 'Task classification route mapping.' }}
                </div>
              </div>
            </div>
          </div>

          <div class="col-span-12 md:col-span-4">
            <label class="block">
              <span class="sr-only">Route profile for {{ task }}</span>
              <select
                class="w-full rounded-2xl border border-white/10 bg-slate-950/70 px-3 py-2.5 text-sm text-slate-100 outline-none transition-all duration-200 focus:border-cyan-400/40 focus:ring-2 focus:ring-cyan-400/20"
                v-model="taskRoutes[task]"
                @change="markDirty"
              >
                <option
                  v-for="profile in profileNames"
                  :key="profile"
                  :value="profile"
                >
                  {{ profile }}
                </option>
              </select>
            </label>
          </div>

          <div class="col-span-12 md:col-span-3 flex items-center">
            <div class="flex items-center gap-2 text-sm">
              <span class="h-2 w-2 rounded-full" :class="routeStatusDot(task)"></span>
              <span class="text-slate-300">{{ routeStatusLabel(task) }}</span>
            </div>
          </div>

          <div class="col-span-12 md:col-span-1 flex items-center justify-end">
            <button class="rounded-full border border-white/10 bg-white/5 px-3 py-2 text-xs text-slate-200 transition hover:border-cyan-400/25 hover:bg-white/10" @click="resetTask(task)">
              Reset
            </button>
          </div>
        </div>
      </div>
    </GlassCard>

    <GlassCard>
      <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <div class="space-y-4">
          <div>
            <h2 class="text-lg font-semibold text-white">Routing Defaults</h2>
            <p class="mt-1 text-sm text-slate-400">
              Set the fallback route and confidence thresholds used by the classifier.
            </p>
          </div>

          <FormField
            label="Default Route"
            hint="Used when no specific task route is selected."
            :error="!isProfileEnabled(routingConfig.default_route) ? 'Selected profile is disabled' : ''"
          >
            <select
              class="w-full bg-transparent text-sm text-white outline-none"
              v-model="routingConfig.default_route"
              @change="markDirty"
            >
              <option v-for="p in enabledProfileNames" :key="p" :value="p">
                {{ p }}
              </option>
            </select>
          </FormField>

          <FormField
            label="Low Confidence Route"
            hint="Applied when task detection confidence is below the threshold."
            :error="!isProfileEnabled(routingConfig.low_confidence_route) ? 'Selected profile is disabled' : ''"
          >
            <select
              class="w-full bg-transparent text-sm text-white outline-none"
              v-model="routingConfig.low_confidence_route"
              @change="markDirty"
            >
              <option v-for="p in enabledProfileNames" :key="p" :value="p">
                {{ p }}
              </option>
            </select>
          </FormField>
        </div>

        <div class="space-y-4">
          <div>
            <h3 class="text-base font-semibold text-white">Confidence Threshold</h3>
            <p class="mt-1 text-sm text-slate-400">
              Tune how aggressively AtlasBridge auto-routes uncertain tasks.
            </p>
          </div>

          <FormField label="Threshold" hint="Range: 0 to 1">
            <input
              type="number"
              class="w-full bg-transparent text-sm text-white outline-none"
              v-model.number="routingConfig.confidence_threshold"
              min="0"
              max="1"
              step="0.05"
              @change="markDirty"
            />
          </FormField>

          <div class="rounded-2xl border border-white/10 bg-white/5 p-4">
            <div class="flex items-center justify-between gap-4">
              <div>
                <div class="text-sm font-medium text-white">Current routing mode</div>
                <div class="mt-1 text-xs text-slate-400">
                  {{ routingConfig.auto_routing ? 'Auto-routing enabled' : 'Manual routing preferred' }}
                </div>
              </div>
              <span class="rounded-full border px-3 py-1 text-xs" :class="routingConfig.auto_routing ? 'border-cyan-400/20 bg-cyan-400/10 text-cyan-200' : 'border-slate-500/20 bg-slate-500/10 text-slate-300'">
                {{ routingConfig.auto_routing ? 'active' : 'inactive' }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </GlassCard>

    <div class="flex flex-col gap-3 sm:flex-row sm:items-center">
      <GradientButton class="sm:min-w-40" @click="save" :disabled="!dirty">
        Save Changes
      </GradientButton>
      <GhostButton class="sm:min-w-32" @click="load" :disabled="!dirty">
        Discard
      </GhostButton>
      <button
        class="inline-flex items-center justify-center rounded-2xl border border-white/10 bg-white/5 px-4 py-2.5 text-sm font-medium text-slate-200 transition-all duration-200 hover:border-rose-400/25 hover:bg-rose-400/10 hover:text-rose-100 disabled:cursor-not-allowed disabled:opacity-50"
        @click="resetAll"
      >
        Reset to Default
      </button>
    </div>

    <div v-if="configStore.error" class="rounded-2xl border border-rose-400/20 bg-rose-400/10 px-4 py-3 text-sm text-rose-100">
      {{ configStore.error }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useConfigStore } from "../stores/config";
import GlassCard from "../components/ui/GlassCard.vue";
import PageHeader from "../components/ui/PageHeader.vue";
import GradientButton from "../components/ui/GradientButton.vue";
import GhostButton from "../components/ui/GhostButton.vue";
import EmptyState from "../components/ui/EmptyState.vue";
import FormField from "../components/ui/FormField.vue";

const configStore = useConfigStore();
const dirty = ref(false);
const validationError = ref<string | null>(null);

const taskRoutes = ref<Record<string, string>>({});
const routingConfig = ref({
  auto_routing: true,
  default_route: "route.default",
  low_confidence_route: "route.default",
  confidence_threshold: 0.55,
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
};

const taskDescriptions: Record<string, string> = {
  general_chat: "General assistance and broad conversation routing.",
  design_task: "Product, interface, and visual design work.",
  backend_engineering: "Server-side logic, APIs, and data handling.",
  frontend_engineering: "Client-side UI work and component updates.",
  fullstack_engineering: "Cross-layer tasks spanning UI and server code.",
  debugging: "Troubleshooting, tracing, and issue isolation.",
  refactoring: "Structural cleanup and code modernization.",
  test_generation: "Test creation, coverage, and verification tasks.",
  documentation: "Docs, guides, and explanatory writing.",
  architecture_design: "System design, boundaries, and long-term shape.",
  security_review: "Security analysis, review, and hardening.",
  long_context_analysis: "Deep analysis across large code or prompt context.",
  lightweight_task: "Small, fast, low-cost changes.",
  unknown: "Fallback classification for uncategorized input.",
};

const taskOrder = [
  "design_task",
  "backend_engineering",
  "frontend_engineering",
  "fullstack_engineering",
  "debugging",
  "refactoring",
  "documentation",
  "architecture_design",
  "lightweight_task",
  "security_review",
  "long_context_analysis",
  "general_chat",
  "test_generation",
  "unknown",
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
  const keys = new Set<string>([
    ...taskOrder,
    ...Object.keys(taskRoutes.value || {}),
  ]);
  return Array.from(keys).filter((task) => taskRoutes.value[task] !== undefined);
});

const taskRoutesEmpty = computed(() => orderedTaskRoutes.value.length === 0);

const isLoading = computed(() => configStore.loading);

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
  if (routingConfig.value.low_confidence_route && !enabledProfileNames.value.includes(routingConfig.value.low_confidence_route)) {
    return `Low confidence route "${routingConfig.value.low_confidence_route}" is disabled or does not exist`;
  }
  return null;
}

function load() {
  if (configStore.routes) {
    taskRoutes.value = { ...configStore.routes.task_routes };
  }
  if (configStore.config) {
    routingConfig.value = { ...configStore.config.routing };
  }
  validationError.value = null;
  dirty.value = false;
}

function markDirty() {
  dirty.value = true;
  validationError.value = null;
}

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
  if (error) {
    validationError.value = error;
    return;
  }
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
  if (!route) return "bg-slate-500/80 shadow-[0_0_10px_rgba(148,163,184,0.35)]";
  if (route.includes("security")) return "bg-emerald-400 shadow-[0_0_10px_rgba(34,197,94,0.4)]";
  if (route.includes("debug")) return "bg-amber-400 shadow-[0_0_10px_rgba(245,158,11,0.45)]";
  if (route.includes("long_context")) return "bg-violet-400 shadow-[0_0_10px_rgba(167,139,250,0.45)]";
  return "bg-cyan-400 shadow-[0_0_10px_rgba(53,215,242,0.45)]";
}

function routeStatusLabel(task: string) {
  const route = taskRoutes.value[task];
  if (!route) return "Unassigned";
  if (!profileNames.value.includes(route)) return "Profile missing";
  if (!isProfileEnabled(route)) return "Disabled profile";
  return "Ready";
}

function routeStatusDot(task: string) {
  const label = routeStatusLabel(task);
  if (label === "Ready") return "bg-emerald-400 shadow-[0_0_10px_rgba(34,197,94,0.45)]";
  if (label === "Disabled profile") return "bg-amber-400 shadow-[0_0_10px_rgba(245,158,11,0.45)]";
  if (label === "Profile missing") return "bg-rose-400 shadow-[0_0_10px_rgba(239,68,68,0.45)]";
  return "bg-slate-400 shadow-[0_0_10px_rgba(148,163,184,0.35)]";
}

onMounted(load);
</script>
