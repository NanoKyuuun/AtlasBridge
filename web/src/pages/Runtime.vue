<template>
  <div>
    <div class="grid grid-cols-2 gap-4 mb-6">
      <!-- Proxy Engine Control -->
      <div class="card p-5">
        <div class="text-[14px] font-semibold mb-1">Proxy Engine</div>
        <div class="text-[11.5px] text-[var(--text-mute)] mb-4">Start, stop, atau restart proxy engine</div>
        <div class="flex gap-2 mb-4">
          <button class="btn btn-success" @click="start" :disabled="status?.status === 'running'">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="5 3 19 12 5 21 5 3"/></svg>
            Start
          </button>
          <button class="btn btn-secondary" @click="stop" :disabled="status?.status !== 'running'">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="6" y="6" width="12" height="12"/></svg>
            Stop
          </button>
          <button class="btn btn-ghost" @click="restart">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
            Restart
          </button>
        </div>
        <div v-if="message" class="p-3 rounded-lg text-[13px]"
          :class="message.startsWith('Failed') ? 'border border-[var(--red)] bg-[rgba(248,113,113,.1)] text-[var(--red)]' : 'border border-[var(--green)] bg-[rgba(52,211,153,.1)] text-[var(--green)]'"
        >{{ message }}</div>
      </div>

      <!-- Current State -->
      <div class="card p-5">
        <div class="text-[14px] font-semibold mb-4">Current State</div>
        <div class="space-y-3 text-[13px]">
          <div class="flex items-center justify-between py-2 border-b border-[var(--border)]">
            <span class="text-[var(--text-mute)]">Status</span>
            <span class="badge" :class="status?.status === 'running' ? 'badge-green' : 'badge-red'">
              <span class="status-dot" :class="status?.status === 'running' ? 'running' : 'stopped'"></span>
              {{ status?.status || 'unknown' }}
            </span>
          </div>
          <div class="flex items-center justify-between py-2 border-b border-[var(--border)]">
            <span class="text-[var(--text-mute)]">Mode</span>
            <span class="badge badge-blue">{{ mode }}</span>
          </div>
          <div class="flex items-center justify-between py-2">
            <span class="text-[var(--text-mute)]">Uptime</span>
            <span class="mono">{{ status?.uptime || '-' }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Quick Actions -->
    <div class="card p-5">
      <div class="text-[14px] font-semibold mb-1">Quick Actions</div>
      <div class="text-[11.5px] text-[var(--text-mute)] mb-4">Salin endpoint atau buka 9Router Dashboard</div>
      <div class="flex gap-3">
        <button class="btn btn-secondary" @click="copyEndpoint">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
          Copy API Endpoint
        </button>
        <a :href="ninerDashboardUrl" target="_blank" class="btn btn-ghost">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/><polyline points="15 3 21 3 21 9"/><line x1="10" y1="14" x2="21" y2="3"/></svg>
          Open 9Router Dashboard
        </a>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useAppStore } from "../stores/app";
import { api } from "../api/client";

const appStore = useAppStore();
const message = ref("");

const status = computed(() => appStore.status);
const mode = computed(() => status.value?.mode || "unknown");

const ninerDashboardUrl = computed(() => {
  const downstream = status.value?.downstream || "http://127.0.0.1:20128/v1";
  try {
    const url = new URL(downstream);
    return `${url.protocol}//${url.hostname}:${url.port}/dashboard`;
  } catch {
    return "http://127.0.0.1:20128/dashboard";
  }
});

async function start() {
  try { await api.runtimeStart(); message.value = "Proxy started"; await appStore.fetchStatus(); }
  catch (e: any) { message.value = `Failed to start: ${e.message}`; }
}
async function stop() {
  try { await api.runtimeStop(); message.value = "Proxy stopped"; await appStore.fetchStatus(); }
  catch (e: any) { message.value = `Failed to stop: ${e.message}`; }
}
async function restart() {
  try { await api.runtimeRestart(); message.value = "Proxy restarted"; await appStore.fetchStatus(); }
  catch (e: any) { message.value = `Failed to restart: ${e.message}`; }
}

async function copyEndpoint() {
  const addr = `http://${status.value?.host || "127.0.0.1"}:${status.value?.port || 20127}/v1`;
  await navigator.clipboard.writeText(addr);
  message.value = "Endpoint copied!";
}

onMounted(() => appStore.fetchStatus());
</script>
