<template>
  <div>
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mb-6">
      <div class="card bg-base-100 shadow-md">
        <div class="card-body">
          <h2 class="card-title text-sm">Proxy Status</h2>
          <div class="flex items-center gap-2">
            <div class="badge badge-lg" :class="statusBadge">
              {{ status?.status || "Loading..." }}
            </div>
          </div>
          <p class="text-xs text-base-content/50 mt-1">
            Uptime: {{ status?.uptime || "-" }}
          </p>
        </div>
      </div>
      <div class="card bg-base-100 shadow-md">
        <div class="card-body">
          <h2 class="card-title text-sm">API Endpoint</h2>
          <p class="font-mono text-sm">
            http://{{ status?.host }}:{{ status?.port }}/v1
          </p>
          <button class="btn btn-xs btn-outline mt-2" @click="copyEndpoint">
            Copy
          </button>
        </div>
      </div>
      <div class="card bg-base-100 shadow-md">
        <div class="card-body">
          <h2 class="card-title text-sm">Admin URL</h2>
          <p class="font-mono text-sm">
            http://{{ status?.host }}:{{ status?.port }}/admin
          </p>
          <button class="btn btn-xs btn-outline mt-2" @click="copyAdminUrl">
            Copy
          </button>
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mb-6">
      <div class="card bg-base-100 shadow-md">
        <div class="card-body">
          <h2 class="card-title text-sm">9Router Downstream</h2>
          <div class="flex items-center gap-2 mb-2">
            <div class="badge badge-sm" :class="downstreamBadge">
              {{ downstreamStatus }}
            </div>
            <span class="text-xs text-base-content/50">{{
              status?.downstream
            }}</span>
          </div>
          <a
            :href="ninerDashboardUrl"
            target="_blank"
            class="btn btn-xs btn-outline"
          >
            Open 9Router Dashboard
          </a>
        </div>
      </div>
      <div class="card bg-base-100 shadow-md">
        <div class="card-body">
          <h2 class="card-title text-sm">Configuration</h2>
          <div class="space-y-2 text-sm">
            <div class="flex justify-between">
              <span class="text-base-content/60">Mode</span>
              <span class="badge badge-outline badge-sm">{{
                status?.mode || "-"
              }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-base-content/60">Privacy</span>
              <span class="badge badge-outline badge-sm">{{
                status?.privacy || "-"
              }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-base-content/60">Auto Routing</span>
              <span class="badge badge-outline badge-sm">{{
                config?.routing?.auto_routing ? "ON" : "OFF"
              }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-base-content/60">Run at Startup</span>
              <span class="badge badge-outline badge-sm">{{
                config?.startup?.run_at_login ? "ON" : "OFF"
              }}</span>
            </div>
          </div>
        </div>
      </div>
      <div class="card bg-base-100 shadow-md">
        <div class="card-body">
          <h2 class="card-title text-sm">Quick Actions</h2>
          <div class="flex flex-wrap gap-2">
            <button
              class="btn btn-primary btn-sm"
              @click="startProxy"
              :disabled="status?.mode === 'always_on'"
            >
              Start
            </button>
            <button
              class="btn btn-warning btn-sm"
              @click="stopProxy"
              :disabled="status?.mode === 'disabled'"
            >
              Stop
            </button>
            <button class="btn btn-info btn-sm" @click="restartProxy">
              Restart
            </button>
          </div>
        </div>
      </div>
    </div>

    <div class="card bg-base-100 shadow-md">
      <div class="card-body">
        <h2 class="card-title text-sm">Smart Aliases</h2>
        <div class="overflow-x-auto">
          <table class="table table-sm">
            <thead>
              <tr>
                <th>Alias</th>
                <th>Purpose</th>
                <th>Route</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="alias in smartAliases" :key="alias.id">
                <td class="font-mono text-sm">{{ alias.id }}</td>
                <td>{{ alias.description }}</td>
                <td>
                  <span class="badge badge-outline badge-sm">{{
                    alias.route
                  }}</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <div class="card bg-base-100 shadow-md mt-4">
      <div class="card-body">
        <h2 class="card-title text-sm">Combo Tester</h2>
        <p class="text-xs text-base-content/50 mb-3">
          Test model combos through 9Router. Enter a model name or select from
          the list.
        </p>
        <div class="flex gap-2 mb-3">
          <input
            v-model="comboModel"
            type="text"
            placeholder="Model name (e.g. combo.default, COding)"
            class="input input-bordered input-sm flex-1 font-mono"
          />
          <button
            class="btn btn-primary btn-sm"
            :class="{ loading: comboLoading }"
            :disabled="!comboModel || comboLoading"
            @click="testCombo"
          >
            Test
          </button>
        </div>
        <div class="flex flex-wrap gap-1 mb-3">
          <button
            v-for="m in comboPresets"
            :key="m"
            class="btn btn-xs btn-outline"
            :class="{ 'btn-active': comboModel === m }"
            @click="comboModel = m"
          >
            {{ m }}
          </button>
        </div>
        <div v-if="comboResult" class="alert alert-sm" :class="comboResult.success ? 'alert-success' : 'alert-error'">
          <div class="text-xs">
            <div v-if="comboResult.success">
              <span class="font-mono font-bold">{{ comboResult.model }}</span>
              <span v-if="comboResult.resolved_model && comboResult.resolved_model !== comboResult.model">
                → <span class="font-mono">{{ comboResult.resolved_model }}</span>
              </span>
              <span class="ml-2 opacity-70">{{ comboResult.latency }}ms</span>
            </div>
            <div v-else>
              <span class="font-mono">{{ comboResult.model }}</span> failed:
              <span class="font-mono">{{ comboResult.error }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { useAppStore } from "../stores/app";
import { useConfigStore } from "../stores/config";
import { api, type ComboTestResult } from "../api/client";

const appStore = useAppStore();
const configStore = useConfigStore();

const status = computed(() => appStore.status);
const config = computed(() => configStore.config);
const downstreamStatus = computed(
  () => appStore.downstreamHealth?.status || "unknown",
);

const statusBadge = computed(() => {
  const s = status.value?.status;
  if (s === "running") return "badge-success";
  if (s === "stopped") return "badge-warning";
  return "badge-ghost";
});

const downstreamBadge = computed(() => {
  const s = downstreamStatus.value;
  if (s === "connected") return "badge-success";
  return "badge-error";
});

const smartAliases = [
  {
    id: "smart-auto",
    description: "Auto-route based on request analysis",
    route: "auto",
  },
  {
    id: "smart-debug",
    description: "Force debugging route",
    route: "route.debugging",
  },
  {
    id: "smart-cheap",
    description: "Force low-cost route",
    route: "route.low_cost",
  },
  {
    id: "smart-docs",
    description: "Force documentation route",
    route: "route.documentation",
  },
  {
    id: "smart-architect",
    description: "Force architecture route",
    route: "route.architect",
  },
  {
    id: "smart-code",
    description: "Force code/engineering route",
    route: "route.backend",
  },
  {
    id: "smart-fast",
    description: "Force low-latency optimized route",
    route: "route.low_cost",
  },
  {
    id: "smart-long-context",
    description: "Force long context analysis route",
    route: "route.long_context",
  },
];

const ninerDashboardUrl = computed(() => {
  const downstream = status.value?.downstream || "http://127.0.0.1:20128/v1";
  try {
    const url = new URL(downstream);
    return `${url.protocol}//${url.hostname}:${url.port}/dashboard`;
  } catch {
    return "http://127.0.0.1:20128/dashboard";
  }
});

async function copyEndpoint() {
  const addr = `http://${status.value?.host || "127.0.0.1"}:${status.value?.port || 20127}/v1`;
  await navigator.clipboard.writeText(addr);
}

async function copyAdminUrl() {
  const addr = `http://${status.value?.host || "127.0.0.1"}:${status.value?.port || 20127}/admin`;
  await navigator.clipboard.writeText(addr);
}

async function startProxy() {
  try {
    await api.runtimeStart();
    await appStore.fetchStatus();
  } catch (e: any) {
    console.error("Failed to start proxy:", e.message);
  }
}

async function stopProxy() {
  try {
    await api.runtimeStop();
    await appStore.fetchStatus();
  } catch (e: any) {
    console.error("Failed to stop proxy:", e.message);
  }
}

async function restartProxy() {
  try {
    await api.runtimeRestart();
    await appStore.fetchStatus();
  } catch (e: any) {
    console.error("Failed to restart proxy:", e.message);
  }
}

// Data is fetched by Layout.vue on mount - no duplicate fetch needed

const comboModel = ref("");
const comboLoading = ref(false);
const comboResult = ref<ComboTestResult | null>(null);

const comboPresets = [
  "combo.default",
  "combo.backend",
  "combo.frontend",
  "combo.fullstack",
  "combo.debugging",
  "combo.refactor",
  "combo.test_generation",
  "combo.documentation",
  "combo.deep_reasoning",
  "combo.design",
  "combo.security_review",
  "combo.long_context",
  "combo.low_cost",
  "COding",
  "opencode",
  "smart-auto",
  "smart-debug",
  "smart-code",
];

async function testCombo() {
  if (!comboModel.value) return;
  comboLoading.value = true;
  comboResult.value = null;
  try {
    comboResult.value = await api.testCombo(comboModel.value);
  } catch (e: any) {
    comboResult.value = {
      model: comboModel.value,
      success: false,
      error: e.message,
      latency: 0,
    };
  } finally {
    comboLoading.value = false;
  }
}
</script>
