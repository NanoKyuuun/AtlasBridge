<template>
  <div>
    <div class="card bg-base-100 shadow-md mb-6">
      <div class="card-body">
        <div class="flex justify-between items-center mb-4">
          <div>
            <h2 class="card-title">Metadata Logs</h2>
            <p class="text-sm text-base-content/60">
              Routing decisions and request metadata. Prompts, API keys, and
              authorization headers are never shown.
            </p>
          </div>
          <div class="flex gap-2">
            <button class="btn btn-outline btn-sm" @click="refreshLogs">
              Refresh
            </button>
            <button class="btn btn-outline btn-sm" @click="exportLogs">
              Export
            </button>
          </div>
        </div>

        <div class="flex items-center gap-4 mb-4">
          <div class="badge badge-outline">
            Privacy: {{ privacyMode }}
          </div>
          <div class="badge badge-outline">
            Total: {{ totalEntries }}
          </div>
          <div v-if="!metadataEnabled" class="alert alert-warning py-1 text-xs">
            Metadata logging is disabled. Enable it in Privacy settings.
          </div>
        </div>

        <div v-if="logs.length === 0" class="text-center py-12 text-base-content/40">
          <p class="text-lg">No logs yet</p>
          <p class="text-sm">Routing decisions will appear here</p>
        </div>

        <div v-else class="overflow-x-auto">
          <table class="table table-sm">
            <thead>
              <tr>
                <th>Time</th>
                <th>Request ID</th>
                <th>Task Type</th>
                <th>Route</th>
                <th>Status</th>
                <th>Latency</th>
                <th>Stream</th>
                <th>Error</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(log, index) in logs" :key="index">
                <td class="text-xs whitespace-nowrap">{{ formatTime(log.timestamp) }}</td>
                <td class="font-mono text-xs">{{ shortId(log.request_id) }}</td>
                <td>
                  <span class="badge badge-sm badge-outline">{{ log.task_type || "-" }}</span>
                </td>
                <td class="font-mono text-xs">{{ log.route_key || log.selected_route || "-" }}</td>
                <td>
                  <span
                    class="badge badge-sm"
                    :class="log.status === 'success' ? 'badge-success' : log.status === 'failed' ? 'badge-error' : 'badge-warning'"
                  >
                    {{ log.status || "fallback" }}
                  </span>
                </td>
                <td class="text-xs">{{ log.latency_ms || "-" }}ms</td>
                <td>
                  <span v-if="log.is_stream" class="badge badge-xs badge-info">SSE</span>
                  <span v-else class="badge badge-xs badge-ghost">JSON</span>
                </td>
                <td class="text-xs text-error max-w-[120px] truncate">{{ log.error_class || "" }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
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

async function refreshLogs() {
  try {
    const data = await api.getLogs();
    logs.value = data.logs || [];
    totalEntries.value = data.total || 0;
    privacyMode.value = data.privacy_mode || "standard";
    metadataEnabled.value = data.metadata_enabled !== false;
  } catch (e: any) {
    console.error("Failed to fetch logs:", e.message);
  }
}

function exportLogs() {
  const blob = new Blob([JSON.stringify(logs.value, null, 2)], {
    type: "application/json",
  });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = "metadata-logs.json";
  a.click();
  URL.revokeObjectURL(url);
}

function formatTime(ts?: string) {
  if (!ts) return "-";
  try {
    return new Date(ts).toLocaleTimeString();
  } catch {
    return ts;
  }
}

function shortId(id?: string) {
  if (!id) return "-";
  return id.length > 12 ? id.substring(0, 12) + "..." : id;
}

onMounted(refreshLogs);
</script>
