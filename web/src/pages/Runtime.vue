<template>
  <div>
    <div class="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
      <div class="card bg-base-100 shadow-md">
        <div class="card-body">
          <h2 class="card-title">Proxy Engine</h2>
          <p class="text-sm text-base-content/60 mb-4">
            Control the proxy engine state.
          </p>
          <div class="flex gap-2">
            <button
              class="btn btn-success"
              @click="start"
              :disabled="mode === 'always_on'"
            >
              Start
            </button>
            <button
              class="btn btn-warning"
              @click="stop"
              :disabled="mode === 'disabled'"
            >
              Stop
            </button>
            <button class="btn btn-info" @click="restart">Restart</button>
          </div>
          <div v-if="message" class="alert alert-success mt-4">
            <span>{{ message }}</span>
          </div>
        </div>
      </div>

      <div class="card bg-base-100 shadow-md">
        <div class="card-body">
          <h2 class="card-title">Current State</h2>
          <div class="space-y-3">
            <div class="flex justify-between items-center">
              <span class="text-sm text-base-content/60">Mode</span>
              <span class="badge badge-lg" :class="modeBadge">{{ mode }}</span>
            </div>
            <div class="flex justify-between items-center">
              <span class="text-sm text-base-content/60">Status</span>
              <span class="badge badge-outline">{{
                status?.status || "unknown"
              }}</span>
            </div>
            <div class="flex justify-between items-center">
              <span class="text-sm text-base-content/60">Uptime</span>
              <span class="text-sm font-mono">{{ status?.uptime || "-" }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="card bg-base-100 shadow-md">
      <div class="card-body">
        <h2 class="card-title">Quick Actions</h2>
        <p class="text-sm text-base-content/60 mb-4">
          Common runtime actions.
        </p>
        <div class="flex flex-wrap gap-2">
          <button class="btn btn-outline btn-sm" @click="copyEndpoint">
            Copy API Endpoint
          </button>
          <a
            :href="ninerDashboardUrl"
            target="_blank"
            class="btn btn-outline btn-sm"
          >
            Open 9Router Dashboard
          </a>
        </div>
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

const modeBadge = computed(() => {
  if (mode.value === "always_on") return "badge-success";
  if (mode.value === "disabled") return "badge-error";
  return "badge-warning";
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
