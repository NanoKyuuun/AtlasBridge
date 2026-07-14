<template>
  <div>
    <div class="grid grid-cols-3 gap-4 mb-6">
      <!-- Privacy Mode -->
      <div class="card p-5">
        <div class="text-[14px] font-semibold mb-1">Privacy Mode</div>
        <div class="text-[11.5px] text-[var(--text-mute)] mb-4">Tingkat privasi logging</div>
        <div class="space-y-2">
          <label v-for="mode in privacyModes" :key="mode.value" class="block cursor-pointer">
            <div
              class="card-soft p-3 flex items-center gap-3 border-2 transition-all"
              :class="logging.privacy_mode === mode.value ? 'border-[var(--accent)]' : 'border-transparent hover:border-[var(--accent)]'"
              @click="logging.privacy_mode = mode.value; dirty = true"
            >
              <input type="radio" name="privacy" style="accent-color: var(--accent);" :checked="logging.privacy_mode === mode.value">
              <div class="flex-1">
                <div class="flex items-center gap-2">
                  <div class="text-[13px] font-medium">{{ mode.label }}</div>
                  <span v-if="mode.badge" class="badge badge-yellow">{{ mode.badge }}</span>
                </div>
                <div class="text-[11px] text-[var(--text-mute)]">{{ mode.desc }}</div>
              </div>
            </div>
          </label>
        </div>
      </div>

      <!-- Logging Options -->
      <div class="card p-5">
        <div class="text-[14px] font-semibold mb-1">Logging Options</div>
        <div class="text-[11.5px] text-[var(--text-mute)] mb-4">Kontrol apa yang dicatat</div>
        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <div>
              <div class="text-[13px] font-medium">Metadata logs</div>
              <div class="text-[11px] text-[var(--text-mute)]">Task type, route, latency</div>
            </div>
            <div
              class="toggle"
              :class="{ on: logging.metadata_logging_enabled }"
              @click="logging.metadata_logging_enabled = !logging.metadata_logging_enabled; dirty = true"
            ></div>
          </div>
          <div class="flex items-center justify-between">
            <div>
              <div class="text-[13px] font-medium">Prompt logging</div>
              <div class="text-[11px] text-[var(--text-mute)]">Simpan prompt penuh</div>
            </div>
            <div
              class="toggle"
              :class="{ on: logging.prompt_logging_enabled }"
              @click="logging.prompt_logging_enabled = !logging.prompt_logging_enabled; dirty = true"
            ></div>
          </div>
          <div class="flex items-center justify-between">
            <div>
              <div class="text-[13px] font-medium">Redact secrets</div>
              <div class="text-[11px] text-[var(--text-mute)]">Deteksi &amp; samarkan API key</div>
            </div>
            <div
              class="toggle"
              :class="{ on: redactSecrets }"
              @click="redactSecrets = !redactSecrets; dirty = true"
            ></div>
          </div>
          <div class="flex items-center justify-between">
            <div>
              <div class="text-[13px] font-medium">Error details</div>
              <div class="text-[11px] text-[var(--text-mute)]">Stack trace &amp; context</div>
            </div>
            <div
              class="toggle on"
              @click="dirty = true"
            ></div>
          </div>
        </div>
      </div>

      <!-- Retention -->
      <div class="card p-5">
        <div class="text-[14px] font-semibold mb-1">Retention</div>
        <div class="text-[11.5px] text-[var(--text-mute)] mb-4">Berapa lama log disimpan</div>
        <div class="space-y-4">
          <div>
            <div class="text-[12px] text-[var(--text-mute)] mb-2">Log retention period</div>
            <select class="select" v-model="logging.retention_days" @change="dirty = true">
              <option :value="7">7 days</option>
              <option :value="30">30 days</option>
              <option :value="90">90 days</option>
              <option :value="0">Forever</option>
            </select>
          </div>
          <button class="btn btn-danger w-full justify-center" @click="clearLogs">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
            Clear All Logs
          </button>
          <button class="btn btn-secondary w-full justify-center" @click="exportLogs">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
            Export Diagnostic
          </button>
          <button class="btn btn-primary w-full justify-center" :disabled="!dirty" @click="save">
            Save Settings
          </button>
        </div>
      </div>
    </div>

    <!-- Recent Logs Viewer -->
    <div class="card">
      <div class="p-5 border-b border-[var(--border)] flex items-center justify-between">
        <div>
          <div class="text-[14px] font-semibold">Recent Logs</div>
          <div class="text-[11.5px] text-[var(--text-mute)]">Log metadata routing (tanpa prompt penuh)</div>
        </div>
        <div class="flex gap-2">
          <select class="select" style="width: 140px;">
            <option>All levels</option>
            <option>Info</option>
            <option>Warning</option>
            <option>Error</option>
          </select>
          <button class="btn btn-secondary">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3"/></svg>
            Filter
          </button>
        </div>
      </div>
      <div class="font-mono text-[12px] p-4 space-y-1 max-h-[400px] overflow-y-auto bg-[var(--bg-0)]">
        <div v-for="log in recentLogs" :key="log.ts" class="flex gap-3 py-1 hover:bg-[var(--bg-2)] px-2 rounded">
          <span class="text-[var(--text-mute)]">{{ log.ts }}</span>
          <span :style="{ color: levelColor(log.level) }">{{ log.level }}</span>
          <span class="text-[var(--text)]">{{ log.message }}</span>
        </div>
      </div>
    </div>
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

const privacyModes = [
  { value: "standard", label: "Standard", desc: "Metadata routing tanpa prompt", badge: "" },
  { value: "strict", label: "Strict", desc: "Hanya request ID, status, latency", badge: "" },
  { value: "debug", label: "Debug", desc: "Informasi tambahan untuk debugging", badge: "Advanced" },
];

function levelColor(level: string) {
  if (level === "INFO") return "var(--green)";
  if (level === "WARN") return "var(--yellow)";
  if (level === "ERROR") return "var(--red)";
  return "var(--text-dim)";
}

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
    const blob = new Blob([JSON.stringify(data, null, 2)], { type: "application/json" });
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

const recentLogs = [
  { ts: "14:32:18.423", level: "INFO", message: "req_8f3a2c → task=backend_eng route=route.backend confidence=0.92 latency=38ms" },
  { ts: "14:31:42.108", level: "INFO", message: "req_7e2b1d → task=debugging route=route.debugging confidence=0.88 latency=52ms" },
  { ts: "14:30:15.892", level: "INFO", message: "req_6d1a0c → task=design_task route=route.design confidence=0.95 latency=41ms" },
  { ts: "14:29:03.241", level: "WARN", message: "req_5c0z9b → low confidence (0.42), fallback to default route=route.low_cost" },
  { ts: "14:27:51.003", level: "INFO", message: "req_4b9y8a → task=frontend_eng route=route.frontend confidence=0.91 latency=45ms" },
  { ts: "14:26:12.567", level: "INFO", message: "req_3a8x7z → task=documentation route=route.documentation confidence=0.87 latency=29ms" },
  { ts: "14:24:58.891", level: "INFO", message: "req_2z7w6y → task=architecture_design route=route.architect confidence=0.94 latency=68ms" },
  { ts: "14:23:01.234", level: "ERROR", message: "req_1y6v5x → classifier timeout, safe passthrough activated route=default" },
  { ts: "14:21:44.678", level: "INFO", message: "req_0x5u4w → task=lightweight route=route.low_cost confidence=0.96 latency=18ms" },
];

onMounted(load);
</script>
