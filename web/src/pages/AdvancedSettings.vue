<template>
  <div>
    <div class="card bg-base-100 shadow-md mb-6">
      <div class="card-body">
        <h2 class="card-title">Server Settings</h2>
        <div class="space-y-4">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="form-control">
              <label class="label"><span class="label-text">Host</span></label>
              <input
                class="input input-bordered font-mono"
                v-model="server.host"
                @input="dirty = true"
              />
            </div>
            <div class="form-control">
              <label class="label"><span class="label-text">Port</span></label>
              <input
                type="number"
                class="input input-bordered font-mono"
                v-model.number="server.port"
                @input="dirty = true"
                min="1024"
                max="65535"
              />
            </div>
          </div>
          <div class="form-control">
            <label class="label"
              ><span class="label-text">Downstream URL</span></label
            >
            <input
              class="input input-bordered font-mono"
              v-model="downstream.base_url"
              @input="dirty = true"
            />
          </div>
          <div class="form-control">
            <label class="label"
              ><span class="label-text">Timeout (seconds)</span></label
            >
            <input
              type="number"
              class="input input-bordered w-32"
              v-model.number="downstream.timeout_seconds"
              @input="dirty = true"
            />
          </div>
        </div>
      </div>
    </div>

    <div class="card bg-base-100 shadow-md mb-6">
      <div class="card-body">
        <h2 class="card-title">Security</h2>
        <div class="space-y-4">
          <div class="form-control">
            <label class="label cursor-pointer justify-start gap-4">
              <input
                type="checkbox"
                class="toggle toggle-primary"
                v-model="security.bind_localhost_only"
                @change="dirty = true"
              />
              <div>
                <span class="label-text font-medium">Localhost Only</span>
                <p class="text-xs text-base-content/50">
                  Bind to 127.0.0.1 only (recommended)
                </p>
              </div>
            </label>
          </div>
          <div class="form-control">
            <label class="label cursor-pointer justify-start gap-4">
              <input
                type="checkbox"
                class="toggle toggle-primary"
                v-model="security.allow_lan_access"
                @change="dirty = true"
              />
              <div>
                <span class="label-text font-medium">Allow LAN Access</span>
                <p class="text-xs text-base-content/50">
                  Allow connections from other devices on your network
                </p>
              </div>
            </label>
          </div>
        </div>
      </div>
    </div>

    <div class="card bg-base-100 shadow-md mb-6">
      <div class="card-body">
        <h2 class="card-title">Debug</h2>
        <div class="space-y-4">
          <div class="form-control">
            <label class="label cursor-pointer justify-start gap-4">
              <input
                type="checkbox"
                class="toggle toggle-warning"
                v-model="debugMode"
                @change="dirty = true"
              />
              <div>
                <span class="label-text font-medium">Debug Mode</span>
                <p class="text-xs text-base-content/50">
                  Enable verbose logging for troubleshooting
                </p>
              </div>
            </label>
          </div>
          <div v-if="debugMode" class="alert alert-warning">
            <span>Debug mode increases logging verbosity. Disable when troubleshooting is complete.</span>
          </div>
        </div>
      </div>
    </div>

    <div class="card bg-base-100 shadow-md mb-6">
      <div class="card-body">
        <h2 class="card-title">Config Import/Export</h2>
        <div class="flex gap-2">
          <button class="btn btn-outline btn-sm" @click="exportConfig">
            Export Config
          </button>
          <button class="btn btn-outline btn-sm" @click="triggerImport">
            Import Config
          </button>
          <input
            ref="fileInput"
            type="file"
            accept=".json"
            class="hidden"
            @change="importConfig"
          />
          <button class="btn btn-error btn-sm" @click="resetConfig">
            Reset to Defaults
          </button>
        </div>
      </div>
    </div>

    <button class="btn btn-primary" @click="save" :disabled="!dirty">
      Save Changes
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useConfigStore } from "../stores/config";
import { api } from "../api/client";

const configStore = useConfigStore();
const dirty = ref(false);
const fileInput = ref<HTMLInputElement>();

const server = ref({
  host: "127.0.0.1",
  port: 20127,
  admin_path: "/admin",
  api_base_path: "/v1",
});
const downstream = ref({
  type: "9router",
  base_url: "http://127.0.0.1:20128/v1",
  timeout_seconds: 120,
});
const security = ref({
  admin_auth_enabled: false,
  admin_token_hash: "",
  bind_localhost_only: true,
  allow_lan_access: false,
});
const debugMode = ref(false);

function load() {
  if (configStore.config) {
    server.value = { ...configStore.config.server };
    downstream.value = { ...configStore.config.downstream };
    security.value = { ...configStore.config.security };
    debugMode.value = configStore.config.logging.level === "debug";
  }
  dirty.value = false;
}

async function save() {
  try {
    await configStore.saveConfig({
      server: server.value,
      downstream: downstream.value,
      security: security.value,
      logging: {
        ...configStore.config!.logging,
        level: debugMode.value ? "debug" : "info",
      },
    });
    dirty.value = false;
  } catch (e: any) {
    console.error("Failed to save settings:", e.message);
  }
}

async function exportConfig() {
  try {
    const data = await api.exportConfig();
    const blob = new Blob([JSON.stringify(data, null, 2)], {
      type: "application/json",
    });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "smart-ai-proxy-config.json";
    a.click();
    URL.revokeObjectURL(url);
  } catch (e: any) {
    console.error("Failed to export config:", e.message);
  }
}

function triggerImport() {
  fileInput.value?.click();
}

async function importConfig(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0];
  if (!file) return;
  try {
    const text = await file.text();
    const data = JSON.parse(text);
    await api.importConfig(data);
    await configStore.fetchAll();
    load();
  } catch (e: any) {
    console.error("Failed to import config:", e.message);
  }
}

async function resetConfig() {
  if (confirm("Reset all settings to defaults? This cannot be undone.")) {
    await configStore.resetConfig();
    load();
  }
}

onMounted(load);
</script>
