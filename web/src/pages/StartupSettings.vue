<template>
  <div>
    <div class="card bg-base-100 shadow-md mb-6">
      <div class="card-body">
        <h2 class="card-title">Startup Settings</h2>
        <p class="text-sm text-base-content/60 mb-4">
          Configure how AtlasBridge behaves when your device starts.
        </p>
        <div class="space-y-4">
          <div class="form-control">
            <label class="label cursor-pointer justify-start gap-4">
              <input
                type="checkbox"
                class="toggle toggle-primary"
                v-model="startup.run_at_login"
                @change="dirty = true"
              />
              <div>
                <span class="label-text font-medium">Run at Login</span>
                <p class="text-xs text-base-content/50">
                  Automatically start AtlasBridge when you log in
                </p>
              </div>
            </label>
          </div>
          <div class="form-control">
            <label class="label cursor-pointer justify-start gap-4">
              <input
                type="checkbox"
                class="toggle toggle-primary"
                v-model="startup.start_proxy_on_app_launch"
                @change="dirty = true"
              />
              <div>
                <span class="label-text font-medium"
                  >Start Proxy on App Launch</span
                >
                <p class="text-xs text-base-content/50">
                  Automatically start the proxy engine when the app opens
                </p>
              </div>
            </label>
          </div>
          <div class="form-control">
            <label class="label cursor-pointer justify-start gap-4">
              <input
                type="checkbox"
                class="toggle toggle-primary"
                v-model="startup.restart_after_crash"
                @change="dirty = true"
              />
              <div>
                <span class="label-text font-medium">Restart After Crash</span>
                <p class="text-xs text-base-content/50">
                  Automatically restart the proxy if it crashes
                </p>
              </div>
            </label>
          </div>
        </div>
      </div>
    </div>

    <div class="card bg-base-100 shadow-md mb-6">
      <div class="card-body">
        <h2 class="card-title">Runtime Mode</h2>
        <p class="text-sm text-base-content/60 mb-4">
          Choose how the proxy behaves on startup.
        </p>
        <div class="flex gap-4">
          <label
            class="flex items-center gap-2 cursor-pointer p-4 rounded-lg border"
            :class="{ 'border-primary bg-primary/5': appMode === 'always_on' }"
          >
            <input
              type="radio"
              class="radio radio-primary"
              v-model="appMode"
              value="always_on"
              @change="dirty = true"
            />
            <div>
              <span class="font-medium">Always On</span>
              <p class="text-xs text-base-content/50">
                Proxy starts automatically
              </p>
            </div>
          </label>
          <label
            class="flex items-center gap-2 cursor-pointer p-4 rounded-lg border"
            :class="{ 'border-primary bg-primary/5': appMode === 'manual' }"
          >
            <input
              type="radio"
              class="radio radio-primary"
              v-model="appMode"
              value="manual"
              @change="dirty = true"
            />
            <div>
              <span class="font-medium">Manual</span>
              <p class="text-xs text-base-content/50">Start proxy manually</p>
            </div>
          </label>
          <label
            class="flex items-center gap-2 cursor-pointer p-4 rounded-lg border"
            :class="{ 'border-primary bg-primary/5': appMode === 'disabled' }"
          >
            <input
              type="radio"
              class="radio radio-primary"
              v-model="appMode"
              value="disabled"
              @change="dirty = true"
            />
            <div>
              <span class="font-medium">Disabled</span>
              <p class="text-xs text-base-content/50">
                Proxy does not accept requests
              </p>
            </div>
          </label>
        </div>
      </div>
    </div>

    <div class="card bg-base-100 shadow-md mb-6">
      <div class="card-body">
        <h2 class="card-title">Registration Status</h2>
        <p class="text-sm text-base-content/60 mb-4">
          Current startup registration state on this device.
        </p>
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div class="stat">
            <div class="stat-title">Run at Login</div>
            <div class="stat-value text-lg">
              <span :class="startup.run_at_login ? 'text-success' : 'text-base-content/40'">
                {{ startup.run_at_login ? 'Active' : 'Inactive' }}
              </span>
            </div>
            <div class="stat-desc">
              {{ startup.run_at_login ? 'Will auto-start on login' : 'No auto-start registered' }}
            </div>
          </div>
          <div class="stat">
            <div class="stat-title">Mode</div>
            <div class="stat-value text-lg">
              {{ appModeLabel }}
            </div>
            <div class="stat-desc">{{ appModeDescription }}</div>
          </div>
          <div class="stat">
            <div class="stat-title">Crash Recovery</div>
            <div class="stat-value text-lg">
              <span :class="startup.restart_after_crash ? 'text-success' : 'text-base-content/40'">
                {{ startup.restart_after_crash ? 'Enabled' : 'Disabled' }}
              </span>
            </div>
            <div class="stat-desc">
              {{ startup.restart_after_crash ? 'Auto-restart on crash' : 'No auto-recovery' }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <button class="btn btn-primary" @click="save" :disabled="!dirty">
      Save Changes
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useConfigStore } from "../stores/config";

const configStore = useConfigStore();
const dirty = ref(false);

const startup = ref({
  run_at_login: false,
  start_proxy_on_app_launch: true,
  restart_after_crash: true,
});
const appMode = ref("manual");

const appModeLabel = computed(() => {
  const labels: Record<string, string> = {
    always_on: "Always On",
    manual: "Manual",
    disabled: "Disabled",
  };
  return labels[appMode.value] || appMode.value;
});

const appModeDescription = computed(() => {
  const descriptions: Record<string, string> = {
    always_on: "Proxy starts automatically on device startup",
    manual: "Proxy starts only when you open the app",
    disabled: "Proxy does not accept requests",
  };
  return descriptions[appMode.value] || "";
});

function load() {
  if (configStore.config) {
    startup.value = { ...configStore.config.startup };
    appMode.value = configStore.config.app.mode;
  }
  dirty.value = false;
}

async function save() {
  try {
    await configStore.saveConfig({
      startup: startup.value,
      app: { ...configStore.config!.app, mode: appMode.value },
    });
    dirty.value = false;
  } catch (e: any) {
    console.error("Failed to save startup settings:", e.message);
  }
}

onMounted(load);
</script>
