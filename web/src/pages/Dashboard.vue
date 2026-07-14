<template>
  <div>
      <!-- Stats Grid -->
    <div class="grid grid-cols-4 gap-4 mb-6">
      <!-- Total Requests -->
      <div class="card stat-card p-5 glow-bg">
        <div class="flex items-start justify-between mb-3">
          <div class="stat-icon" style="background: rgba(52, 211, 153, .12);">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#34d399" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg>
          </div>
          <span class="badge badge-green">Live</span>
        </div>
        <div class="text-[11px] text-[var(--text-mute)] uppercase tracking-wider font-semibold">Total Requests</div>
        <div class="text-[26px] font-bold mt-1">{{ totalRequests }}</div>
        <div class="text-[11px] text-[var(--text-dim)] mt-1">Log entries tersimpan</div>
      </div>

      <!-- Avg Latency -->
      <div class="card stat-card p-5">
        <div class="flex items-start justify-between mb-3">
          <div class="stat-icon" style="background: rgba(124, 92, 255, .12);">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#a78bfa" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
          </div>
          <span class="badge badge-purple">Runtime</span>
        </div>
        <div class="text-[11px] text-[var(--text-mute)] uppercase tracking-wider font-semibold">Uptime</div>
        <div class="text-[22px] font-bold mt-1 mono">{{ uptime }}</div>
        <div class="text-[11px] text-[var(--text-dim)] mt-1">Sejak start terakhir</div>
      </div>

      <!-- Mode -->
      <div class="card stat-card p-5">
        <div class="flex items-start justify-between mb-3">
          <div class="stat-icon" style="background: rgba(79, 140, 255, .12);">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#4f8cff" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
          </div>
          <span class="badge badge-blue">Auto</span>
        </div>
        <div class="text-[11px] text-[var(--text-mute)] uppercase tracking-wider font-semibold">Mode</div>
        <div class="text-[22px] font-bold mt-1">{{ startupMode }}</div>
        <div class="text-[11px] text-[var(--text-dim)] mt-1">Routing mode</div>
      </div>

      <!-- Version -->
      <div class="card stat-card p-5">
        <div class="flex items-start justify-between mb-3">
          <div class="stat-icon" style="background: rgba(251, 191, 36, .12);">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#fbbf24" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
          </div>
          <span class="badge badge-yellow">Info</span>
        </div>
        <div class="text-[11px] text-[var(--text-mute)] uppercase tracking-wider font-semibold">Version</div>
        <div class="text-[22px] font-bold mt-1">{{ version }}</div>
        <div class="text-[11px] text-[var(--text-dim)] mt-1">AtlasBridge</div>
      </div>
    </div>

    <!-- Endpoints + Proxy Control -->
    <div class="grid grid-cols-3 gap-4 mb-6">
      <!-- Endpoint Config -->
      <div class="card p-5 col-span-2">
        <div class="flex items-center justify-between mb-4">
          <div>
            <div class="text-[14px] font-semibold">Endpoint Configuration</div>
            <div class="text-[11.5px] text-[var(--text-mute)]">Konfigurasi proxy dan downstream 9Router</div>
          </div>
          <router-link to="/advanced" class="btn btn-ghost">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4"/></svg>
            Edit
          </router-link>
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div class="card-soft p-4">
            <div class="flex items-center gap-2 mb-2">
              <span class="status-dot" :class="statusClass"></span>
              <span class="text-[11px] text-[var(--text-mute)] uppercase tracking-wider font-semibold">Proxy Endpoint</span>
            </div>
            <div class="code-tag text-[13px]">{{ openAiEndpoint }}</div>
            <div class="text-[11.5px] text-[var(--text-dim)] mt-2">OpenAI-compatible · Model: <span class="code-tag">smart-auto</span></div>
          </div>
          <div class="card-soft p-4">
            <div class="flex items-center gap-2 mb-2">
              <span class="status-dot" :class="downstreamHealthy ? 'running' : 'error'"></span>
              <span class="text-[11px] text-[var(--text-mute)] uppercase tracking-wider font-semibold">Downstream 9Router</span>
            </div>
            <div class="code-tag text-[13px]">{{ downstreamEndpoint }}</div>
            <div class="text-[11.5px] text-[var(--text-dim)] mt-2">Status: <span :style="{ color: downstreamHealthy ? 'var(--green)' : 'var(--red)' }">{{ downstreamHealthy ? 'Connected' : 'Disconnected' }}</span></div>
          </div>
        </div>
        <div class="divider"></div>
        <div class="grid grid-cols-3 gap-4 text-[12.5px]">
          <div>
            <div class="text-[var(--text-mute)] text-[11px] uppercase tracking-wider mb-1">Default Route</div>
            <div class="font-medium"><span class="code-tag">{{ defaultRoute }}</span></div>
          </div>
          <div>
            <div class="text-[var(--text-mute)] text-[11px] uppercase tracking-wider mb-1">Startup Mode</div>
            <div class="font-medium"><span class="badge badge-green">{{ startupMode }}</span></div>
          </div>
          <div>
            <div class="text-[var(--text-mute)] text-[11px] uppercase tracking-wider mb-1">Privacy Mode</div>
            <div class="font-medium"><span class="badge badge-blue">Standard</span></div>
          </div>
        </div>
      </div>

      <!-- Proxy Control -->
      <div class="card p-5">
        <div class="text-[14px] font-semibold mb-1">Proxy Control</div>
        <div class="text-[11.5px] text-[var(--text-mute)] mb-4">Runtime management</div>
        <div class="flex flex-col gap-2">
          <button class="btn btn-success w-full justify-center" @click="handleStart">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="5 3 19 12 5 21 5 3"/></svg>
            Start
          </button>
          <button class="btn btn-danger w-full justify-center" @click="handleStop">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="6" y="6" width="12" height="12"/></svg>
            Stop
          </button>
          <button class="btn btn-secondary w-full justify-center" @click="handleRestart">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
            Restart
          </button>
        </div>
        <div class="divider"></div>
        <div class="text-[12px]">
          <div class="flex justify-between py-1.5"><span class="text-[var(--text-mute)]">Uptime</span><span class="font-medium mono">{{ uptime }}</span></div>
          <div class="flex justify-between py-1.5"><span class="text-[var(--text-mute)]">Version</span><span class="font-medium mono">{{ version }}</span></div>
          <div class="flex justify-between py-1.5"><span class="text-[var(--text-mute)]">PID</span><span class="font-medium mono">{{ pid }}</span></div>
        </div>
      </div>
    </div>

    <!-- Charts + Top Profiles -->
    <div class="grid grid-cols-3 gap-4 mb-6">
      <!-- Request Distribution Chart -->
      <div class="card p-5 col-span-2">
        <div class="flex items-center justify-between mb-4">
          <div>
            <div class="text-[14px] font-semibold">Request Distribution</div>
            <div class="text-[11.5px] text-[var(--text-mute)]">Distribusi task type 24 jam terakhir</div>
          </div>
          <div class="flex gap-1 bg-[var(--bg-2)] p-1 rounded-lg">
            <div class="tab active">24h</div>
            <div class="tab">7d</div>
            <div class="tab">30d</div>
          </div>
        </div>
        <div class="flex items-end gap-2 h-[180px] pt-4">
          <div v-for="bar in chartBars" :key="bar.label" class="flex-1 flex flex-col items-center gap-2">
            <div class="w-full flex items-end justify-center h-[140px]">
              <div class="chart-bar w-full" :style="{ height: bar.height + '%', background: bar.color }" :data-value="bar.value"></div>
            </div>
            <div class="text-[10.5px] text-[var(--text-mute)] text-center">{{ bar.label }}</div>
          </div>
        </div>
      </div>

      <!-- Top Route Profiles -->
      <div class="card p-5">
        <div class="text-[14px] font-semibold mb-1">Top Route Profiles</div>
        <div class="text-[11.5px] text-[var(--text-mute)] mb-4">Paling sering digunakan</div>
        <div v-if="topProfiles.length === 0" class="text-[12.5px] text-[var(--text-mute)] py-4 text-center">Belum ada data routing</div>
        <div v-else class="space-y-3">
          <div v-for="profile in topProfiles" :key="profile.name">
            <div class="flex justify-between mb-1.5 text-[12.5px]">
              <span class="font-medium">{{ profile.name }}</span>
              <span class="text-[var(--text-mute)] mono">{{ profile.count }}</span>
            </div>
            <div class="h-1.5 bg-[var(--bg-3)] rounded-full overflow-hidden">
              <div class="h-full rounded-full" :style="{ width: profile.pct + '%', background: profile.color }"></div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Recent Routing Activity -->
    <div class="card">
      <div class="flex items-center justify-between p-5 border-b border-[var(--border)]">
        <div>
          <div class="text-[14px] font-semibold">Recent Routing Activity</div>
          <div class="text-[11.5px] text-[var(--text-mute)]">Log metadata routing terbaru (tanpa prompt penuh)</div>
        </div>
        <router-link to="/observability" class="btn btn-ghost">View all →</router-link>
      </div>
      <!-- Header row -->
      <div class="log-row" style="padding: 10px 20px; background: var(--bg-2); color: var(--text-mute); font-weight: 600; font-size: 11px; text-transform: uppercase; letter-spacing: .05em;">
        <div>Timestamp</div>
        <div>Request ID</div>
        <div>Task Type</div>
        <div>Route</div>
        <div>Latency</div>
        <div>Status</div>
      </div>
      <!-- Rows -->
      <div v-if="recentLogs.length === 0" class="p-8 text-center text-[var(--text-mute)] text-[13px]">Belum ada log routing</div>
      <div v-for="log in recentLogs" :key="log.id" class="log-row">
        <div class="text-[var(--text-dim)]">{{ log.time }}</div>
        <div class="mono text-[var(--text)]">{{ log.id }}</div>
        <div><span class="badge badge-blue">{{ log.taskType || '-' }}</span></div>
        <div><span class="code-tag">{{ log.route }}</span></div>
        <div class="mono">{{ log.latency }}</div>
        <div><span class="badge" :class="log.status === '200' || log.status === 'success' ? 'badge-green' : 'badge-red'">{{ log.status }}</span></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useAppStore } from "../stores/app";
import { useConfigStore } from "../stores/config";
import { useToast } from "../composables/useToast";
import { api } from "../api/client";

const appStore = useAppStore();
const configStore = useConfigStore();
const { showToast } = useToast();

const status = computed(() => appStore.status);
const statusClass = computed(() => {
  const s = status.value?.status;
  if (s === "running") return "running";
  if (s === "stopped") return "stopped";
  if (s === "error") return "error";
  return "disabled";
});

const openAiEndpoint = computed(() => {
  if (configStore.config?.server) {
    return `http://${configStore.config.server.host}:${configStore.config.server.port}/v1`;
  }
  return "http://localhost:20127/v1";
});

const downstreamEndpoint = computed(() =>
  configStore.config?.downstream?.base_url || "—"
);

const downstreamHealthy = computed(() => appStore.downstreamHealth?.status === "ok");

const defaultRoute = computed(() =>
  configStore.config?.routing?.default_route || "—"
);

const startupMode = computed(() =>
  configStore.config?.app?.mode || "—"
);

const uptime = computed(() => status.value?.uptime || "—");
const version = computed(() => status.value?.version || "—");
const pid = computed(() => status.value?.pid ? String(status.value.pid) : "—");

// Recent logs from real backend
const rawLogs = ref<any[]>([]);

const recentLogs = computed(() =>
  rawLogs.value.slice(0, 8).map((l: any) => ({
    id: l.request_id ? l.request_id.substring(0, 12) : "—",
    time: l.timestamp ? new Date(l.timestamp).toLocaleTimeString() : "—",
    taskType: l.task_type || "-",
    route: l.route_key || l.selected_route || "—",
    latency: l.latency_ms != null ? `${l.latency_ms}ms` : "—",
    status: l.status || "—",
  }))
);

// totalRequests from log count
const totalRequests = computed(() =>
  rawLogs.value.length > 0 ? String(rawLogs.value.length) : "—"
);

// topProfiles derived from actual log data
const profileColors = [
  "linear-gradient(90deg, #7c5cff, #4f8cff)",
  "linear-gradient(90deg, #f59e0b, #ef4444)",
  "linear-gradient(90deg, #34d399, #22d3ee)",
  "linear-gradient(90deg, #a78bfa, #7c5cff)",
  "linear-gradient(90deg, #22d3ee, #4f8cff)",
  "linear-gradient(90deg, #8b91a7, #5a6077)",
];

const topProfiles = computed(() => {
  const counts: Record<string, number> = {};
  for (const l of rawLogs.value) {
    const key = l.route_key || l.selected_route;
    if (key) counts[key] = (counts[key] || 0) + 1;
  }
  const total = rawLogs.value.length || 1;
  return Object.entries(counts)
    .sort((a, b) => b[1] - a[1])
    .slice(0, 6)
    .map(([name, count], i) => ({
      name,
      count,
      pct: Math.round((count / total) * 100),
      color: profileColors[i % profileColors.length],
    }));
});

// Chart bars derived from topProfiles
const chartBars = computed(() => {
  if (topProfiles.value.length === 0) return [];
  const max = topProfiles.value[0]?.count || 1;
  return topProfiles.value.map((p, i) => ({
    label: p.name.replace("route.", ""),
    height: Math.round((p.count / max) * 90),
    value: p.count,
    color: profileColors[i % profileColors.length],
  }));
});

async function loadLogs() {
  try {
    const data = await api.getLogs();
    rawLogs.value = data.logs || [];
  } catch {
    rawLogs.value = [];
  }
}

// Proxy control — calls real API
async function handleStart() {
  try {
    await api.runtimeStart();
    showToast("Proxy started", "success");
    await appStore.fetchStatus();
  } catch (e: any) {
    showToast(`Failed to start: ${e.message}`, "error");
  }
}
async function handleStop() {
  try {
    await api.runtimeStop();
    showToast("Proxy stopped", "warning");
    await appStore.fetchStatus();
  } catch (e: any) {
    showToast(`Failed to stop: ${e.message}`, "error");
  }
}
async function handleRestart() {
  try {
    await api.runtimeRestart();
    showToast("Proxy restarted", "success");
    await appStore.fetchStatus();
  } catch (e: any) {
    showToast(`Failed to restart: ${e.message}`, "error");
  }
}

onMounted(loadLogs);
</script>
