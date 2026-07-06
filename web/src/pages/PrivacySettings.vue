<template>
  <div>
    <div class="card bg-base-100 shadow-md mb-6">
      <div class="card-body">
        <h2 class="card-title">Privacy & Logging</h2>
        <p class="text-sm text-base-content/60 mb-4">
          Configure privacy mode and logging behavior.
        </p>
        <div class="space-y-4">
          <div class="form-control">
            <label class="label"
              ><span class="label-text">Privacy Mode</span></label
            >
            <select
              class="select select-bordered"
              v-model="logging.privacy_mode"
              @change="dirty = true"
            >
              <option value="standard">
                Standard - Log metadata routing without full prompt
              </option>
              <option value="strict">
                Strict - Log only request ID, status, latency, and route
              </option>
              <option value="debug">
                Debug - Additional info (explicit opt-in)
              </option>
            </select>
          </div>
          <div v-if="logging.privacy_mode === 'debug'" class="alert alert-warning">
            <span>Debug mode increases logging verbosity. Do not use in production or shared environments.</span>
          </div>
          <div class="form-control">
            <label class="label cursor-pointer justify-start gap-4">
              <input
                type="checkbox"
                class="toggle toggle-primary"
                v-model="logging.metadata_logging_enabled"
                @change="dirty = true"
              />
              <div>
                <span class="label-text font-medium">Metadata Logging</span>
                <p class="text-xs text-base-content/50">
                  Log request metadata (task type, route, latency)
                </p>
              </div>
            </label>
          </div>
          <div class="form-control">
            <label class="label cursor-pointer justify-start gap-4">
              <input
                type="checkbox"
                class="toggle toggle-primary"
                v-model="logging.prompt_logging_enabled"
                @change="dirty = true"
              />
              <div>
                <span class="label-text font-medium">Prompt Logging</span>
                <p class="text-xs text-base-content/50">
                  Log full prompts (WARNING: contains sensitive data)
                </p>
              </div>
            </label>
          </div>
          <div class="form-control">
            <label class="label cursor-pointer justify-start gap-4">
              <input
                type="checkbox"
                class="toggle toggle-primary"
                v-model="redactSecrets"
                @change="dirty = true"
              />
              <div>
                <span class="label-text font-medium">Redact Secrets</span>
                <p class="text-xs text-base-content/50">
                  Automatically redact API keys, tokens, and passwords from logs
                </p>
              </div>
            </label>
          </div>
          <div class="form-control">
            <label class="label"
              ><span class="label-text">Retention Days</span></label
            >
            <input
              type="number"
              class="input input-bordered w-32"
              v-model.number="logging.retention_days"
              @input="dirty = true"
              min="1"
              max="90"
            />
          </div>
        </div>
      </div>
    </div>

    <div class="card bg-base-100 shadow-md mb-6">
      <div class="card-body">
        <h2 class="card-title">Logs</h2>
        <div class="flex gap-2">
          <button class="btn btn-outline btn-sm" @click="exportLogs">
            Export Diagnostics
          </button>
          <button class="btn btn-error btn-sm" @click="clearLogs">
            Clear Logs
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
const redactSecrets = ref(true);

const logging = ref({
  level: "info",
  privacy_mode: "standard",
  metadata_logging_enabled: true,
  prompt_logging_enabled: false,
  retention_days: 7,
});

function load() {
  if (configStore.config) {
    logging.value = { ...configStore.config.logging };
  }
  dirty.value = false;
}

async function save() {
  try {
    await configStore.saveConfig({ logging: logging.value });
    dirty.value = false;
  } catch (e: any) {
    console.error("Failed to save privacy settings:", e.message);
  }
}

async function exportLogs() {
  try {
    const data = await api.exportDiagnostics();
    const blob = new Blob([JSON.stringify(data, null, 2)], {
      type: "application/json",
    });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "diagnostics.json";
    a.click();
    URL.revokeObjectURL(url);
  } catch (e: any) {
    console.error("Failed to export diagnostics:", e.message);
  }
}

async function clearLogs() {
  if (confirm("Clear all metadata logs? This cannot be undone.")) {
    try {
      await api.clearLogs();
    } catch (e: any) {
      console.error("Failed to clear logs:", e.message);
    }
  }
}

onMounted(load);
</script>
