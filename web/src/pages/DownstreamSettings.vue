<template>
  <div>
    <div class="card bg-base-100 shadow-md mb-6">
      <div class="card-body">
        <h2 class="card-title">9Router Downstream</h2>
        <p class="text-sm text-base-content/60 mb-4">
          Configure the 9Router downstream endpoint.
        </p>
        <div class="space-y-4">
          <div class="form-control">
            <label class="label"
              ><span class="label-text">Base URL</span></label
            >
            <input
              class="input input-bordered font-mono"
              v-model="downstream.base_url"
              @input="dirty = true"
              placeholder="http://127.0.0.1:20128/v1"
            />
          </div>
          <div class="form-control">
            <label class="label"
              ><span class="label-text">Timeout (seconds)</span></label
            >
            <input
              type="number"
              class="input input-bordered"
              v-model.number="downstream.timeout_seconds"
              @input="dirty = true"
              min="10"
              max="300"
            />
          </div>
        </div>
      </div>
    </div>

    <div class="card bg-base-100 shadow-md mb-6">
      <div class="card-body">
        <h2 class="card-title">Connection Status</h2>
        <div class="flex items-center gap-4">
          <div class="badge badge-lg" :class="healthBadge">
            {{ health?.status || "unknown" }}
          </div>
          <span class="text-sm text-base-content/60">{{
            health?.message || health?.url
          }}</span>
        </div>
        <div class="flex gap-2 mt-4">
          <button class="btn btn-outline btn-sm" @click="checkHealth">
            Check Connection
          </button>
          <a
            class="btn btn-outline btn-sm"
            :href="dashboardUrl"
            target="_blank"
          >
            Open 9Router Dashboard
          </a>
        </div>
      </div>
    </div>

    <button class="btn btn-primary" @click="save" :disabled="!dirty">
      Save Changes
    </button>
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
  type: "9router",
  base_url: "http://127.0.0.1:20128/v1",
  timeout_seconds: 120,
});

const healthBadge = computed(() => {
  if (health.value?.status === "connected") return "badge-success";
  return "badge-error";
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
