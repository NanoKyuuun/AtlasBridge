<template>
  <div class="space-y-6 lg:space-y-8">
    <PageHeader
      eyebrow="Logs"
      title="Metadata Logs"
      description="Routing decisions and request metadata. Prompts, API keys, and authorization headers are never shown."
    >
      <template #actions>
        <GhostButton @click="refreshLogs">Refresh</GhostButton>
        <GhostButton @click="exportLogs">Export</GhostButton>
      </template>
    </PageHeader>

    <GlassCard>
      <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
        <div class="flex flex-wrap items-center gap-2 text-xs text-slate-400">
          <StatusBadge :status="metadataEnabled ? 'active' : 'inactive'" :label="metadataEnabled ? 'metadata enabled' : 'metadata disabled'" />
          <span class="rounded-full border border-white/10 bg-white/5 px-3 py-1">Total: {{ totalEntries }}</span>
          <span class="rounded-full border border-white/10 bg-white/5 px-3 py-1">Privacy: {{ privacyMode }}</span>
        </div>

        <div v-if="!metadataEnabled" class="rounded-2xl border border-amber-400/20 bg-amber-400/10 px-4 py-3 text-sm text-amber-100">
          Metadata logging is disabled. Enable it in Privacy settings.
        </div>
      </div>
    </GlassCard>

    <GlassCard>
      <div class="space-y-4">
        <div class="flex items-start justify-between gap-4">
          <div>
            <h2 class="text-lg font-semibold text-white">Log List</h2>
            <p class="mt-1 text-sm text-slate-400">
              Compact request history with only the non-sensitive fields needed for routing review.
            </p>
          </div>
        </div>

        <div v-if="loading" class="rounded-2xl border border-white/10 bg-white/5 p-6 text-sm text-slate-400">
          Loading logs...
        </div>

        <EmptyState v-else-if="logs.length === 0" title="No logs yet" description="Routing decisions will appear here.">
          <template #icon>
            <span class="text-xl text-cyan-300">📝</span>
          </template>
        </EmptyState>

        <div v-else class="overflow-hidden rounded-3xl border border-white/10">
          <div class="grid grid-cols-12 gap-3 border-b border-white/10 bg-white/5 px-4 py-3 text-[11px] uppercase tracking-[0.22em] text-slate-500">
            <div class="col-span-12 md:col-span-2">Timestamp</div>
            <div class="col-span-12 md:col-span-3">Request ID</div>
            <div class="col-span-12 md:col-span-2">Route</div>
            <div class="col-span-12 md:col-span-2">Status</div>
            <div class="col-span-12 md:col-span-1">Latency</div>
            <div class="col-span-12 md:col-span-2">Type</div>
          </div>

          <div v-for="(log, index) in logs" :key="index" class="grid grid-cols-12 gap-3 border-b border-white/5 px-4 py-4 last:border-b-0">
            <div class="col-span-12 md:col-span-2 text-xs text-slate-300 whitespace-nowrap">
              {{ formatTime(log.timestamp) }}
            </div>
            <div class="col-span-12 md:col-span-3 font-mono text-xs text-white">
              {{ shortId(log.request_id) }}
            </div>
            <div class="col-span-12 md:col-span-2 text-sm text-cyan-200">
              {{ log.route_key || log.selected_route || '-' }}
            </div>
            <div class="col-span-12 md:col-span-2">
              <StatusBadge :status="statusTone(log.status)" :label="log.status || 'fallback'" />
            </div>
            <div class="col-span-12 md:col-span-1 text-sm text-slate-300">
              {{ log.latency_ms ?? '-' }}ms
            </div>
            <div class="col-span-12 md:col-span-2 text-sm text-slate-400">
              <span v-if="log.task_type" class="rounded-full border border-white/10 bg-white/5 px-3 py-1 text-xs">{{ log.task_type }}</span>
              <span v-else class="text-slate-500">-</span>
            </div>
          </div>
        </div>
      </div>
    </GlassCard>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { api } from "../api/client";
import GlassCard from "../components/ui/GlassCard.vue";
import PageHeader from "../components/ui/PageHeader.vue";
import GhostButton from "../components/ui/GhostButton.vue";
import StatusBadge from "../components/ui/StatusBadge.vue";
import EmptyState from "../components/ui/EmptyState.vue";

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
    return new Date(ts).toLocaleString();
  } catch {
    return ts;
  }
}

function shortId(id?: string) {
  if (!id) return "-";
  return id.length > 12 ? id.substring(0, 12) + "..." : id;
}

function statusTone(status?: string) {
  if (status === "success") return "running";
  if (status === "failed") return "error";
  if (status === "warning") return "warning";
  return "inactive";
}

onMounted(refreshLogs);
</script>
