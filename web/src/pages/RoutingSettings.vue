<template>
  <div>
    <div v-if="validationError" class="alert alert-error mb-6">
      <span>{{ validationError }}</span>
    </div>

    <div class="card bg-base-100 shadow-md mb-6">
      <div class="card-body">
        <h2 class="card-title">Task-to-Route Mapping</h2>
        <p class="text-sm text-base-content/60 mb-4">
          Map detected task types to route profiles. Changes take effect on the
          next request.
        </p>
        <div class="overflow-x-auto">
          <table class="table table-sm">
            <thead>
              <tr>
                <th>Task Category</th>
                <th>Route Profile</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(_, task) in taskRoutes" :key="task">
                <td class="font-mono text-sm">{{ task }}</td>
                <td>
                  <select
                    class="select select-bordered select-sm w-full max-w-xs"
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
                </td>
                <td>
                  <button
                    class="btn btn-ghost btn-xs"
                    @click="resetTask(task as string)"
                  >
                    Reset
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <div class="card bg-base-100 shadow-md mb-6">
      <div class="card-body">
        <h2 class="card-title">Routing Defaults</h2>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div class="form-control">
            <label class="label"
              ><span class="label-text">Default Route</span></label
            >
            <select
              class="select select-bordered"
              v-model="routingConfig.default_route"
              @change="markDirty"
            >
              <option v-for="p in enabledProfileNames" :key="p" :value="p">
                {{ p }}
              </option>
            </select>
            <label class="label" v-if="!isProfileEnabled(routingConfig.default_route)">
              <span class="label-text-alt text-error">Selected profile is disabled</span>
            </label>
          </div>
          <div class="form-control">
            <label class="label"
              ><span class="label-text">Low Confidence Route</span></label
            >
            <select
              class="select select-bordered"
              v-model="routingConfig.low_confidence_route"
              @change="markDirty"
            >
              <option v-for="p in enabledProfileNames" :key="p" :value="p">
                {{ p }}
              </option>
            </select>
            <label class="label" v-if="!isProfileEnabled(routingConfig.low_confidence_route)">
              <span class="label-text-alt text-error">Selected profile is disabled</span>
            </label>
          </div>
          <div class="form-control">
            <label class="label"
              ><span class="label-text">Confidence Threshold</span></label
            >
            <input
              type="number"
              class="input input-bordered"
              v-model.number="routingConfig.confidence_threshold"
              min="0"
              max="1"
              step="0.05"
              @change="markDirty"
            />
          </div>
          <div class="form-control">
            <label class="label"
              ><span class="label-text">Auto Routing</span></label
            >
            <input
              type="checkbox"
              class="toggle toggle-primary"
              v-model="routingConfig.auto_routing"
              @change="markDirty"
            />
          </div>
        </div>
      </div>
    </div>

    <div class="flex gap-2">
      <button class="btn btn-primary" @click="save" :disabled="!dirty">
        Save Changes
      </button>
      <button class="btn btn-ghost" @click="load" :disabled="!dirty">
        Discard
      </button>
      <button class="btn btn-outline btn-warning" @click="resetAll">
        Reset All to Default
      </button>
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
  explicit_override_enabled: true,
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
  explicit_override_enabled: true,
  confidence_threshold: 0.55,
};

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

onMounted(load);
</script>
