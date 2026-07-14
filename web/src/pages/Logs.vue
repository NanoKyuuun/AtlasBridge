<template>
  <div>
    <!-- Status bar -->
    <div class="card p-4 mb-4 flex items-center justify-between">
      <div class="flex items-center gap-3">
        <span class="badge" :class="metadataEnabled ? 'badge-green' : 'badge-gray'">
          {{ metadataEnabled ? 'Metadata Enabled' : 'Metadata Disabled' }}
        </span>
        <span class="text-[12px] text-[var(--text-mute)]">Total: <span class="font-medium text-[var(--text)]">{{ totalEntries }}</span></span>
        <span class="text-[12px] text-[var(--text-mute)]">Privacy: <span class="font-medium text-[var(--text)]">{{ privacyMode }}</span></span>
      </div>
      <div class="flex gap-2">
        <button class="btn btn-ghost" @click="refreshLogs">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
          Refresh
        </button>
        <button class="btn btn-secondary" @click="exportLogs">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
          Export
        </button>
      </div>
    </div>

    <!-- Warning if metadata disabled -->
    <div v-if="!metadataEnabled" class="mb-4 p-3 rounded-lg border border-[var(--yellow)] bg-[rgba(251,191,36,.1)] text-[var(--yellow)] text-[13px]">
      Metadata logging is disabled. Enable it in Privacy &amp; Logs settings.
    </div>

    <!-- Log Table -->
    <div class="card">
      <div class="p-5 border-b border-[var(--border)]">
        <div class="text-[14px] font-semibold">Metadata Logs</div>
        <div class="text-[11.5px] text-[var(--text-mute)]">Routing decisions — prompts, API keys, dan authorization headers tidak disimpan</div>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="p-8 text-center text-[var(--text-mute)] text-[13px]">
        Loading logs...
      </div>

      <!-- Empty -->
      <div v-else-if="logs.length === 0" class="p-10 text-center">
        <div class="text-[var(--text-mute)] text-[13px]">No routing logs yet</div>
        <div class="text-[11.5px] text-[var(--text-mute)] mt-1">Routing decisions will appear here</div>
      </div>

      <!-- Rows -->
      <template v-else>
        <div class="log-row" style="padding:10px 20px; background: var(--bg-2); color: var(--text-mute); font-weight:600; font-size:11px; text-transform:uppercase; letter-spacing:.05em;">
          <div>Timestamp</div>
          <div>Request ID</div>
          <div>Task Type</div>
          <div>Route</div>
          <div>Latency</div>
          <div>Status</div>
        </div>
        <div v-for="(log, index) in logs" :key="index" class="log-row">
          <div class="text-[var(--text-dim)]">{{ formatTime(log.timestamp) }}</div>
          <div class="mono">{{ shortId(log.request_id) }}</div>
          <div>
            <span v-if="log.task_type" class="badge badge-blue">{{ log.task_type }}</span>
            <span v-else class="text-[var(--text-mute)]">-</span>
          </div>
          <div><span class="code-tag">{{ log.route_key || log.selected_route || '-' }}</span></div>
          <div class="mono">{{ log.latency_ms ?? '-' }}ms</div>
          <div>
            <span class="badge" :class="statusBadge(log.status)">{{ log.status || 'unknown' }}</span>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { api } from "../api/client";

interface LogEntry {
  timestamp?: string;
  request_id?: string;
  task_type?: string;
  route_key?: string;
  selected_route?: string;
  status?: string;
  latency_ms?: number;
  is_stream?: boolean;
  error_class?: string;
  confidence?: number;
  method?: string;
  path?: string;
}

const logs = ref<LogEntry[]>([]);
const totalEntries = ref(0);
const privacyMode = ref("standard");
const metadataEnabled = ref(true);
const loading = ref(false);

async function refreshLogs() {
  loading.value = true;
  try {
    const data = await api.getLogs();
    logs.value = data.logs || [];
    totalEntries.value = data.total || 0;
    privacyMode.value = data.privacy_mode || "standard";
    metadataEnabled.value = data.metadata_enabled !== false;
  } catch (e: any) {
    console.error("Failed to fetch logs:", e.message);
  } finally {
    loading.value = false;
  }
}

function exportLogs() {
  const blob = new Blob([JSON.stringify(logs.value, null, 2)], { type: "application/json" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = "metadata-logs.json";
  a.click();
  URL.revokeObjectURL(url);
}

function formatTime(ts?: string) {
  if (!ts) return "-";
  try { return new Date(ts).toLocaleString(); } catch { return ts; }
}

function shortId(id?: string) {
  if (!id) return "-";
  return id.length > 12 ? id.substring(0, 12) + "..." : id;
}

function statusBadge(status?: string) {
  if (status === "success") return "badge-green";
  if (status === "failed") return "badge-red";
  if (status === "warning") return "badge-yellow";
  return "badge-gray";
}

onMounted(refreshLogs);
</script>
