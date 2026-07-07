<template>
  <div class="space-y-6 lg:space-y-8">
    <PageHeader
      eyebrow="Runtime"
      title="Proxy Runtime"
      description="Control the proxy engine state and access runtime shortcuts."
    />

    <div class="grid gap-4 lg:grid-cols-[1.1fr_0.9fr]">
      <GlassCard>
        <div class="space-y-5">
          <div>
            <h2 class="text-lg font-semibold text-white">Proxy Engine</h2>
            <p class="mt-1 text-sm text-slate-400">Start, stop, or restart the local proxy engine.</p>
          </div>

          <div class="flex flex-col gap-3 sm:flex-row sm:flex-wrap">
            <GradientButton @click="start" :disabled="mode === 'always_on'">Start</GradientButton>
            <GhostButton @click="stop" :disabled="mode === 'disabled'">Stop</GhostButton>
            <GhostButton @click="restart">Restart</GhostButton>
          </div>

          <div v-if="message" class="rounded-2xl border px-4 py-3 text-sm" :class="message.startsWith('Failed') ? 'border-rose-400/20 bg-rose-400/10 text-rose-100' : 'border-emerald-400/20 bg-emerald-400/10 text-emerald-100'">
            {{ message }}
          </div>
        </div>
      </GlassCard>

      <GlassCard>
        <div class="space-y-4">
          <h2 class="text-lg font-semibold text-white">Current State</h2>
          <div class="space-y-3">
            <div class="flex items-center justify-between gap-4 rounded-2xl border border-white/10 bg-white/5 px-4 py-3">
              <span class="text-sm text-slate-400">Mode</span>
              <StatusBadge :status="modeStatus" :label="mode" />
            </div>
            <div class="flex items-center justify-between gap-4 rounded-2xl border border-white/10 bg-white/5 px-4 py-3">
              <span class="text-sm text-slate-400">Status</span>
              <StatusBadge :status="status?.status || 'inactive'" :label="status?.status || 'unknown'" />
            </div>
            <div class="flex items-center justify-between gap-4 rounded-2xl border border-white/10 bg-white/5 px-4 py-3">
              <span class="text-sm text-slate-400">Uptime</span>
              <span class="font-mono text-sm text-white">{{ status?.uptime || '-' }}</span>
            </div>
          </div>
        </div>
      </GlassCard>
    </div>

    <GlassCard>
      <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
        <div>
          <h2 class="text-lg font-semibold text-white">Quick Actions</h2>
          <p class="mt-1 text-sm text-slate-400">Copy the API endpoint or open the downstream dashboard.</p>
        </div>
        <div class="flex flex-col gap-3 sm:flex-row sm:flex-wrap">
          <GhostButton @click="copyEndpoint">Copy API Endpoint</GhostButton>
          <a :href="ninerDashboardUrl" target="_blank" class="inline-flex items-center justify-center gap-2 rounded-2xl border border-white/10 bg-white/5 px-4 py-2.5 text-sm font-medium text-slate-200 transition-all duration-200 hover:border-cyan-400/25 hover:bg-white/10 hover:text-white">
            Open 9Router Dashboard
          </a>
        </div>
      </div>
    </GlassCard>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useAppStore } from "../stores/app";
import { api } from "../api/client";
import GlassCard from "../components/ui/GlassCard.vue";
import PageHeader from "../components/ui/PageHeader.vue";
import GradientButton from "../components/ui/GradientButton.vue";
import GhostButton from "../components/ui/GhostButton.vue";
import StatusBadge from "../components/ui/StatusBadge.vue";

const appStore = useAppStore();
const message = ref("");

const status = computed(() => appStore.status);
const mode = computed(() => status.value?.mode || "unknown");
const modeStatus = computed(() => {
  if (mode.value === "always_on") return "active";
  if (mode.value === "disabled") return "disabled";
  return "warning";
});

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
  try {
    await api.runtimeStart();
    message.value = "Proxy started";
    await appStore.fetchStatus();
  } catch (e: any) {
    message.value = `Failed to start: ${e.message}`;
  }
}

async function stop() {
  try {
    await api.runtimeStop();
    message.value = "Proxy stopped";
    await appStore.fetchStatus();
  } catch (e: any) {
    message.value = `Failed to stop: ${e.message}`;
  }
}

async function restart() {
  try {
    await api.runtimeRestart();
    message.value = "Proxy restarted";
    await appStore.fetchStatus();
  } catch (e: any) {
    message.value = `Failed to restart: ${e.message}`;
  }
}

async function copyEndpoint() {
  const addr = `http://${status.value?.host || "127.0.0.1"}:${status.value?.port || 20127}/v1`;
  await navigator.clipboard.writeText(addr);
}

onMounted(() => appStore.fetchStatus());
</script>
