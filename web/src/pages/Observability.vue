<template>
  <div>
    <!-- Stats Mini Grid — derived from real log data -->
    <div class="grid grid-cols-4 gap-4 mb-6">
      <div class="card p-4">
        <div class="text-[11px] text-[var(--text-mute)] uppercase tracking-wider font-semibold">Total Requests</div>
        <div class="text-[22px] font-bold mt-1">{{ totalRequests }}</div>
        <div class="text-[11px] mt-1 text-[var(--text-dim)]">Log entries tersimpan</div>
      </div>
      <div class="card p-4">
        <div class="text-[11px] text-[var(--text-mute)] uppercase tracking-wider font-semibold">Success Rate</div>
        <div class="text-[22px] font-bold mt-1">{{ successRate }}</div>
        <div class="text-[11px] mt-1 text-[var(--text-dim)]">Status success</div>
      </div>
      <div class="card p-4">
        <div class="text-[11px] text-[var(--text-mute)] uppercase tracking-wider font-semibold">Avg Latency</div>
        <div class="text-[22px] font-bold mt-1">{{ avgLatency }}</div>
        <div class="text-[11px] mt-1 text-[var(--text-dim)]">Rata-rata latency</div>
      </div>
      <div class="card p-4">
        <div class="text-[11px] text-[var(--text-mute)] uppercase tracking-wider font-semibold">Streaming</div>
        <div class="text-[22px] font-bold mt-1">{{ streamingCount }}</div>
        <div class="text-[11px] text-[var(--text-mute)] mt-1">{{ streamingPct }}% of total</div>
      </div>
    </div>

    <!-- Distribution Charts -->
    <div class="grid grid-cols-2 gap-4 mb-6">
      <!-- Route Distribution -->
      <div class="card p-5">
        <div class="text-[14px] font-semibold mb-4">Route Distribution</div>
        <div v-if="routeDist.length === 0" class="text-[12.5px] text-[var(--text-mute)] py-4 text-center">Belum ada data routing</div>
        <div v-else class="space-y-3">
          <div v-for="r in routeDist" :key="r.label">
            <div class="flex justify-between mb-1.5 text-[12.5px]">
              <span class="font-medium">{{ r.label }}</span>
              <span class="text-[var(--text-mute)] mono">{{ r.value }}</span>
            </div>
            <div class="h-2 bg-[var(--bg-3)] rounded-full overflow-hidden">
              <div class="h-full rounded-full" :style="{ width: r.pct + '%', background: r.color }"></div>
            </div>
          </div>
        </div>
      </div>

      <!-- Task Type Distribution -->
      <div class="card p-5">
        <div class="text-[14px] font-semibold mb-4">Task Type Distribution</div>
        <div v-if="taskTypeDist.length === 0" class="text-[12.5px] text-[var(--text-mute)] py-4 text-center">Belum ada data task type</div>
        <div v-else class="space-y-3">
          <div v-for="t in taskTypeDist" :key="t.label">
            <div class="flex justify-between mb-1.5 text-[12.5px]">
              <span class="font-medium">{{ t.label }}</span>
              <span class="text-[var(--text-mute)] mono">{{ t.value }}</span>
            </div>
            <div class="h-2 bg-[var(--bg-3)] rounded-full overflow-hidden">
              <div class="h-full rounded-full" :style="{ width: t.pct + '%', background: t.color }"></div>
            </div>
          </div>
          <div class="divider"></div>
          <div class="flex justify-between text-[12px]">
            <span class="text-[var(--text-mute)]">Total entries</span>
            <span class="mono">{{ allLogs.length }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Full Request Log -->
    <div class="card">
      <div class="p-5 border-b border-[var(--border)] flex items-center justify-between">
        <div>
          <div class="text-[14px] font-semibold">Full Request Log</div>
          <div class="text-[11.5px] text-[var(--text-mute)]">Semua aktivitas routing</div>
        </div>
        <div class="flex gap-2">
          <input
            type="text" class="input" placeholder="Search request ID..."
            style="width: 200px;"
            v-model="searchQuery"
          >
          <select class="select" style="width: 160px;" v-model="filterRoute">
            <option value="">All routes</option>
            <option v-for="r in availableRoutes" :key="r" :value="r">{{ r }}</option>
          </select>
          <button class="btn btn-secondary" @click="refreshLogs">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
            Refresh
          </button>
          <button class="btn btn-secondary" @click="exportLogs">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
            Export
          </button>
        </div>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="p-8 text-center text-[var(--text-mute)] text-[13px]">Loading logs...</div>

      <!-- Header -->
      <div v-else class="log-row" style="padding: 10px 20px; background: var(--bg-2); color: var(--text-mute); font-weight: 600; font-size: 11px; text-transform: uppercase; letter-spacing: .05em;">
        <div>Timestamp</div>
        <div>Request ID</div>
        <div>Task Type</div>
        <div>Route</div>
        <div>Latency</div>
        <div>Status</div>
      </div>

      <!-- Rows -->
      <div v-if="!loading && filteredLogs.length === 0" class="p-8 text-center text-[var(--text-mute)] text-[13px]">
        {{ allLogs.length === 0 ? 'Belum ada log routing.' : 'No matching logs found.' }}
      </div>
      <div v-for="log in filteredLogs" :key="log.id" class="log-row">
        <div class="text-[var(--text-dim)]">{{ log.time }}</div>
        <div class="mono text-[var(--text)]">{{ log.id }}</div>
        <div>
          <span v-if="log.taskType" class="badge badge-blue">{{ log.taskType }}</span>
          <span v-else class="text-[var(--text-mute)]">-</span>
        </div>
        <div><span class="code-tag">{{ log.route }}</span></div>
        <div class="mono">{{ log.latency }}</div>
        <div><span class="badge" :class="log.status === 'success' ? 'badge-green' : log.status === 'failed' ? 'badge-red' : 'badge-gray'">{{ log.status }}</span></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { api } from "../api/client";

const searchQuery = ref("");
const filterRoute = ref("");
const loading = ref(false);

interface LogEntry {
  timestamp?: string;
  request_id?: string;
  task_type?: string;
  route_key?: string;
  selected_route?: string;
  status?: string;
  latency_ms?: number;
  is_stream?: boolean;
}

const allLogs = ref<LogEntry[]>([]);

async function refreshLogs() {
  loading.value = true;
  try {
    const data = await api.getLogs();
    allLogs.value = data.logs || [];
  } catch (e: any) {
    console.error("Failed to fetch logs:", e.message);
  } finally {
    loading.value = false;
  }
}

function exportLogs() {
  const blob = new Blob([JSON.stringify(allLogs.value, null, 2)], { type: "application/json" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = "observability-logs.json";
  a.click();
  URL.revokeObjectURL(url);
}

// --- Computed stats from real data ---

const totalRequests = computed(() => allLogs.value.length || "—");

const successRate = computed(() => {
  if (allLogs.value.length === 0) return "—";
  const success = allLogs.value.filter(l => l.status === "success").length;
  return `${((success / allLogs.value.length) * 100).toFixed(1)}%`;
});

const avgLatency = computed(() => {
  const withLatency = allLogs.value.filter(l => l.latency_ms != null);
  if (withLatency.length === 0) return "—";
  const avg = withLatency.reduce((s, l) => s + (l.latency_ms || 0), 0) / withLatency.length;
  return `${Math.round(avg)}ms`;
});

const streamingCount = computed(() => allLogs.value.filter(l => l.is_stream).length);
const streamingPct = computed(() => {
  if (allLogs.value.length === 0) return 0;
  return ((streamingCount.value / allLogs.value.length) * 100).toFixed(1);
});

const distColors = [
  "linear-gradient(90deg, #7c5cff, #4f8cff)",
  "linear-gradient(90deg, #34d399, #22d3ee)",
  "linear-gradient(90deg, #f59e0b, #ef4444)",
  "linear-gradient(90deg, #a78bfa, #7c5cff)",
  "linear-gradient(90deg, #22d3ee, #4f8cff)",
  "linear-gradient(90deg, #fbbf24, #f59e0b)",
];

const routeDist = computed(() => {
  const counts: Record<string, number> = {};
  for (const l of allLogs.value) {
    const key = l.route_key || l.selected_route;
    if (key) counts[key] = (counts[key] || 0) + 1;
  }
  const total = allLogs.value.length || 1;
  return Object.entries(counts)
    .sort((a, b) => b[1] - a[1])
    .slice(0, 6)
    .map(([label, count], i) => ({
      label,
      value: `${count} (${((count / total) * 100).toFixed(1)}%)`,
      pct: (count / total) * 100,
      color: distColors[i % distColors.length],
    }));
});

const taskTypeDist = computed(() => {
  const counts: Record<string, number> = {};
  for (const l of allLogs.value) {
    if (l.task_type) counts[l.task_type] = (counts[l.task_type] || 0) + 1;
  }
  const total = allLogs.value.length || 1;
  return Object.entries(counts)
    .sort((a, b) => b[1] - a[1])
    .slice(0, 5)
    .map(([label, count], i) => ({
      label,
      value: `${count} (${((count / total) * 100).toFixed(1)}%)`,
      pct: (count / total) * 100,
      color: distColors[i % distColors.length],
    }));
});

const availableRoutes = computed(() => {
  const routes = new Set<string>();
  for (const l of allLogs.value) {
    const r = l.route_key || l.selected_route;
    if (r) routes.add(r);
  }
  return [...routes].sort();
});

// Mapped logs for table display
const mappedLogs = computed(() =>
  allLogs.value.map((l) => ({
    id: l.request_id ? l.request_id.substring(0, 12) : "—",
    time: l.timestamp ? new Date(l.timestamp).toLocaleTimeString() : "—",
    taskType: l.task_type || "",
    route: l.route_key || l.selected_route || "—",
    latency: l.latency_ms != null ? `${l.latency_ms}ms` : "—",
    status: l.status || "unknown",
  }))
);

const filteredLogs = computed(() =>
  mappedLogs.value.filter((log) => {
    const matchSearch = !searchQuery.value || log.id.includes(searchQuery.value);
    const matchRoute = !filterRoute.value || log.route === filterRoute.value;
    return matchSearch && matchRoute;
  })
);

onMounted(refreshLogs);
</script>
