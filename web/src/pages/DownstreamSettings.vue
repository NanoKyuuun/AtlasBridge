<template>
  <div>
    <!-- Connection Settings -->
    <div class="card p-5 mb-6">
      <div class="text-[14px] font-semibold mb-1">Connection Settings</div>
      <div class="text-[11.5px] text-[var(--text-mute)] mb-4">Set the 9Router base URL and request timeout.</div>
      
      <div class="grid grid-cols-2 gap-4">
        <div>
          <label class="text-[12px] text-[var(--text-mute)] mb-1.5 block">Base URL</label>
          <input
            type="text" class="input mono w-full"
            v-model="downstream.base_url"
            @input="dirty = true"
            placeholder="http://127.0.0.1:20128/v1"
          />
          <div class="text-[11.5px] text-[var(--text-dim)] mt-1.5">The downstream OpenAI-compatible 9Router endpoint.</div>
        </div>
        <div>
          <label class="text-[12px] text-[var(--text-mute)] mb-1.5 block">Timeout</label>
          <input
            type="number" class="input w-full"
            v-model.number="downstream.timeout_seconds"
            @input="dirty = true"
            min="10" max="300"
          />
          <div class="text-[11.5px] text-[var(--text-dim)] mt-1.5">Timeout in seconds. Recommended range: 10 to 300.</div>
        </div>
      </div>
      <div class="mt-4 pt-4 border-t border-[var(--border)] flex justify-end">
        <button class="btn btn-primary" @click="save" :disabled="!dirty">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"/><polyline points="17 21 17 13 7 13 7 21"/><polyline points="7 3 7 8 15 8"/></svg>
          Save Changes
        </button>
      </div>
    </div>

    <!-- Connection Status -->
    <div class="card p-5">
      <div class="text-[14px] font-semibold mb-1">Connection Status</div>
      <div class="text-[11.5px] text-[var(--text-mute)] mb-4">Check whether AtlasBridge can reach the configured downstream.</div>
      
      <div class="p-4 mb-4 rounded-xl border border-[var(--border)] bg-[var(--bg-2)] flex items-center gap-3">
        <span class="badge" :class="health?.status === 'connected' || health?.status === 'ok' ? 'badge-green' : 'badge-red'">
          <span class="status-dot" :class="health?.status === 'connected' || health?.status === 'ok' ? 'running' : 'stopped'"></span>
          {{ health?.status || 'unknown' }}
        </span>
        <span class="text-[13px] text-[var(--text-mute)] break-all">{{ health?.message || health?.url || downstream.base_url }}</span>
      </div>

      <div class="flex gap-2">
        <button class="btn btn-secondary" @click="checkHealth">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
          Check Connection
        </button>
        <a class="btn btn-ghost" :href="dashboardUrl" target="_blank">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/><polyline points="15 3 21 3 21 9"/><line x1="10" y1="14" x2="21" y2="3"/></svg>
          Open 9Router Dashboard
        </a>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useConfigStore } from "../stores/config";
import { api, type DownstreamHealth } from "../api/client";

const configStore = useConfigStore();
const dirty = ref(false);
const health = ref<DownstreamHealth | null>(null);

const downstream = ref({
  base_url: "http://127.0.0.1:20128/v1",
  timeout_seconds: 120,
});

const dashboardUrl = computed(() => {
  try {
    const url = new URL(downstream.value.base_url);
    url.pathname = "/dashboard";
    return url.toString();
  } catch {
    return "http://127.0.0.1:20128/dashboard";
  }
});

function load() {
  if (configStore.config) {
    downstream.value = { ...configStore.config.downstream };
  }
  dirty.value = false;
}

async function checkHealth() {
  try {
    health.value = await api.getDownstreamHealth();
  } catch (e: any) {
    health.value = {
      status: "unavailable",
      url: downstream.value.base_url,
      message: e.message,
    };
  }
}

async function save() {
  try {
    await configStore.saveConfig({ downstream: downstream.value });
    dirty.value = false;
    checkHealth();
  } catch (e: any) {
    console.error("Failed to save downstream settings:", e.message);
  }
}

onMounted(() => {
  load();
  checkHealth();
});
</script>
