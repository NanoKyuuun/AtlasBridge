<template>
  <div class="space-y-8 lg:space-y-10">
    <PageHeader
      eyebrow="Dashboard"
      title="AtlasBridge Control Center"
      description="Route every AI coding task to the right model path."
    >
      <template #actions>
        <GradientButton :disabled="status?.status === 'running'" @click="startProxy">
          Start Proxy
        </GradientButton>
        <GhostButton @click="openRoutingSettings">
          Open Routing Settings
        </GhostButton>
      </template>
    </PageHeader>

    <section class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
      <StatCard
        label="Proxy Status"
        :value="statusText"
        :description="status?.uptime ? `Uptime ${status.uptime}` : 'Waiting for status signal'"
        :tone="statusTone"
      >
        <template #icon>
          <StatusBadge :status="statusStatus" :label="statusStatus" />
        </template>
      </StatCard>

      <MetricCard
        label="OpenAI-compatible endpoint"
        :value="openAiEndpoint"
        description="Base URL used by coding assistants."
      />

      <MetricCard
        label="Downstream 9Router endpoint"
        :value="downstreamEndpoint"
        :description="downstreamStatusLabel"
      />

      <MetricCard
        label="Startup mode"
        :value="startupMode"
        :description="startupDescription"
        :delta="startupDelta"
        :tone="startupTone"
      />

      <StatCard
        label="Default route"
        :value="defaultRoute"
        description="Fallback route used when no override applies."
      >
        <template #icon>
          <span class="text-cyan-300">↪</span>
        </template>
      </StatCard>

      <MetricCard
        label="Request count today"
        :value="requestCountToday"
        description="Derived from currently available dashboard state."
      />

      <MetricCard
        label="Most used task type"
        :value="mostUsedTaskType"
        description="Best available task classification view."
      />

      <MetricCard
        label="Most used route profile"
        :value="mostUsedRouteProfile"
        description="Top profile based on current configuration data."
      />
    </section>

    <section class="grid gap-4 lg:grid-cols-[1.2fr_0.8fr]">
      <GlassCard class="relative overflow-hidden">
        <div class="relative z-10 space-y-5">
          <div class="flex items-start justify-between gap-4">
            <div>
              <h2 class="text-lg font-semibold text-white">Architecture Flow</h2>
              <p class="mt-1 text-sm text-slate-400">
                OpenCode / AI Coding Assistant → AtlasBridge → 9Router → AI Providers
              </p>
            </div>
            <StatusBadge :status="statusStatus" :label="statusStatus" />
          </div>

          <div class="grid gap-3 md:grid-cols-[1fr_auto_1fr_auto_1fr_auto_1fr] md:items-center">
            <div class="rounded-2xl border border-white/10 bg-white/5 px-4 py-4 text-center shadow-[0_0_22px_rgba(47,128,255,0.08)]">
              <div class="text-sm font-medium text-white">OpenCode / AI Assistant</div>
              <div class="mt-2 text-xs text-slate-400">Request source</div>
            </div>
            <div class="hidden justify-center text-cyan-300 md:flex">→</div>

            <div class="rounded-2xl border border-blue-400/25 bg-gradient-to-br from-blue-500/18 via-violet-500/18 to-cyan-400/12 px-5 py-5 text-center shadow-[0_0_30px_rgba(47,128,255,0.14)]">
              <div class="text-sm font-semibold text-white">AtlasBridge</div>
              <div class="mt-2 text-xs text-slate-300">Routing core</div>
              <div class="mt-3 flex items-center justify-center gap-2">
                <span class="neon-dot"></span>
                <span class="text-[11px] uppercase tracking-[0.22em] text-slate-400">Center</span>
              </div>
            </div>
            <div class="hidden justify-center text-cyan-300 md:flex">→</div>

            <div class="rounded-2xl border border-white/10 bg-white/5 px-4 py-4 text-center shadow-[0_0_22px_rgba(124,58,237,0.08)]">
              <div class="text-sm font-medium text-white">9Router</div>
              <div class="mt-2 text-xs text-slate-400">Policy + alias routing</div>
            </div>
            <div class="hidden justify-center text-cyan-300 md:flex">→</div>

            <div class="rounded-2xl border border-white/10 bg-white/5 px-4 py-4 text-center shadow-[0_0_22px_rgba(53,215,242,0.08)] md:col-span-1 col-span-full">
              <div class="text-sm font-medium text-white">AI Providers</div>
              <div class="mt-2 text-xs text-slate-400">Final model execution</div>
            </div>
          </div>
        </div>
      </GlassCard>

      <GlassCard>
        <div class="space-y-4">
          <div>
            <h2 class="text-lg font-semibold text-white">Quick Actions</h2>
            <p class="mt-1 text-sm text-slate-400">Common runtime actions and routing shortcuts.</p>
          </div>
          <div class="flex flex-wrap gap-3">
            <GradientButton @click="startProxy" :disabled="status?.status === 'running'">
              Start Proxy
            </GradientButton>
            <GhostButton @click="stopProxy" :disabled="status?.status !== 'running'">
              Stop Proxy
            </GhostButton>
            <GhostButton @click="restartProxy">Restart Proxy</GhostButton>
          </div>

          <div class="space-y-3 rounded-2xl border border-white/10 bg-white/5 p-4">
            <div class="flex items-center justify-between gap-3 text-sm">
              <span class="text-slate-400">OpenAI-compatible endpoint</span>
              <button class="text-cyan-300 transition hover:text-cyan-200" @click="copyEndpoint">
                Copy
              </button>
            </div>
            <div class="font-mono text-sm text-white break-all">{{ openAiEndpoint }}</div>
            <div class="flex items-center justify-between gap-3 text-sm">
              <span class="text-slate-400">Downstream endpoint</span>
              <button class="text-cyan-300 transition hover:text-cyan-200" @click="copyDownstreamUrl">
                Copy
              </button>
            </div>
            <div class="font-mono text-sm text-white break-all">{{ downstreamEndpoint }}</div>
          </div>
        </div>
      </GlassCard>
    </section>

    <section class="grid gap-4 lg:grid-cols-[1.1fr_0.9fr]">
      <GlassCard>
        <div class="space-y-4">
          <div class="flex items-start justify-between gap-4">
            <div>
              <h2 class="text-lg font-semibold text-white">Model Aliases</h2>
              <p class="mt-1 text-sm text-slate-400">Reference aliases currently exposed through the dashboard.</p>
            </div>
            <StatusBadge status="active" label="ready" />
          </div>

          <div class="overflow-hidden rounded-2xl border border-white/10">
            <div class="grid grid-cols-[1.1fr_1.6fr_1fr] gap-3 border-b border-white/10 bg-white/5 px-4 py-3 text-[11px] uppercase tracking-[0.22em] text-slate-500">
              <span>Alias</span>
              <span>Purpose</span>
              <span>Route</span>
            </div>
            <div v-for="alias in smartAliases" :key="alias.id" class="grid grid-cols-[1.1fr_1.6fr_1fr] gap-3 border-b border-white/5 px-4 py-3 text-sm last:border-b-0">
              <span class="font-mono text-slate-100">{{ alias.id }}</span>
              <span class="text-slate-400">{{ alias.description }}</span>
              <span class="text-cyan-200">{{ alias.route }}</span>
            </div>
          </div>
        </div>
      </GlassCard>

      <GlassCard>
        <div class="space-y-4">
          <div>
            <h2 class="text-lg font-semibold text-white">Combo Tester</h2>
            <p class="mt-1 text-sm text-slate-400">
              Test model combos through 9Router. Enter a model name or select from the list.
            </p>
          </div>

          <FormField label="Model name" forId="combo-model" hint="Examples: combo.default, COding, atlas-auto">
            <input
              id="combo-model"
              v-model="comboModel"
              type="text"
              placeholder="Model name (e.g. combo.default, COding)"
              class="w-full bg-transparent text-sm text-white outline-none placeholder:text-slate-500"
            />
          </FormField>

          <div class="flex flex-wrap gap-2">
            <GhostButton
              v-for="m in comboPresets"
              :key="m"
              :class="comboModel === m ? 'border-cyan-400/35 bg-cyan-400/10 text-cyan-100' : ''"
              @click="comboModel = m"
            >
              {{ m }}
            </GhostButton>
          </div>

          <div class="flex flex-wrap gap-3">
            <GradientButton :disabled="!comboModel || comboLoading" @click="testCombo">
              Test Combo
            </GradientButton>
          </div>

          <div v-if="comboResult" class="rounded-2xl border px-4 py-3 text-sm" :class="comboResult.success ? 'border-emerald-400/20 bg-emerald-400/10 text-emerald-100' : 'border-rose-400/20 bg-rose-400/10 text-rose-100'">
            <div v-if="comboResult.success">
              <span class="font-mono font-semibold">{{ comboResult.model }}</span>
              <span v-if="comboResult.resolved_model && comboResult.resolved_model !== comboResult.model">
                → <span class="font-mono">{{ comboResult.resolved_model }}</span>
              </span>
              <span class="ml-2 text-white/70">{{ comboResult.latency }}ms</span>
            </div>
            <div v-else>
              <span class="font-mono">{{ comboResult.model }}</span> failed:
              <span class="font-mono">{{ comboResult.error }}</span>
            </div>
          </div>
        </div>
      </GlassCard>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { useRouter } from "vue-router";
import { useAppStore } from "../stores/app";
import { useConfigStore } from "../stores/config";
import { api, type ComboTestResult } from "../api/client";
import GlassCard from "../components/ui/GlassCard.vue";
import PageHeader from "../components/ui/PageHeader.vue";
import StatCard from "../components/ui/StatCard.vue";
import MetricCard from "../components/ui/MetricCard.vue";
import StatusBadge from "../components/ui/StatusBadge.vue";
import GradientButton from "../components/ui/GradientButton.vue";
import GhostButton from "../components/ui/GhostButton.vue";
import FormField from "../components/ui/FormField.vue";

const appStore = useAppStore();
const configStore = useConfigStore();
const router = useRouter();

const status = computed(() => appStore.status);
const config = computed(() => configStore.config);
const routes = computed(() => configStore.routes);
const profiles = computed(() => configStore.profiles);
const downstreamStatus = computed(
  () => appStore.downstreamHealth?.status || "unknown",
);

const statusStatus = computed(() => status.value?.status || "inactive");

const statusTone = computed(() => {
  const s = statusStatus.value;
  if (s === "running") return "success";
  if (s === "stopped") return "warning";
  if (s === "error") return "error";
  return "default";
});

const statusText = computed(() => status.value?.status || "Loading...");

const openAiEndpoint = computed(() => {
  const host = status.value?.host || "127.0.0.1";
  const port = status.value?.port || 20127;
  return `http://${host}:${port}/v1`;
});

const downstreamEndpoint = computed(() => status.value?.downstream || "http://127.0.0.1:20128/v1");

const downstreamStatusLabel = computed(() => {
  const state = downstreamStatus.value;
  if (state === "connected") return "Connected and ready";
  if (state === "unavailable") return "Downstream health unavailable";
  return `Status: ${state}`;
});

const startupMode = computed(() => config.value?.startup?.run_at_login ? "Auto start" : "Manual start");
const startupDescription = computed(() => config.value?.startup?.run_at_login ? "Proxy starts with the system." : "Proxy starts manually from the dashboard.");
const startupDelta = computed(() => config.value?.startup?.start_proxy_on_app_launch ? "App launch enabled" : "App launch disabled");
const startupTone = computed(() => config.value?.startup?.run_at_login ? "success" : "warning");

const defaultRoute = computed(() => config.value?.routing?.default_route || "route.default");

const requestCountToday = computed(() => "Unavailable");

const mostUsedTaskType = computed(() => {
  const taskRouteCount = Object.keys(routes.value?.task_routes || {}).length;
  return taskRouteCount > 0 ? "Derived routing map" : "Unavailable";
});

const mostUsedRouteProfile = computed(() => {
  const profileEntries = Object.entries(profiles.value?.route_profiles || {});
  if (!profileEntries.length) return "Unavailable";
  const [name, profile] = profileEntries[0];
  return profile.label || name;
});

const smartAliases = [
  {
    id: "atlas-auto",
    description: "Auto-route based on request analysis",
    route: "auto",
  },
  {
    id: "atlas-debug",
    description: "Force debugging route",
    route: "route.debugging",
  },
  {
    id: "atlas-cheap",
    description: "Force low-cost route",
    route: "route.low_cost",
  },
  {
    id: "atlas-docs",
    description: "Force documentation route",
    route: "route.documentation",
  },
  {
    id: "atlas-architect",
    description: "Force architecture route",
    route: "route.architect",
  },
  {
    id: "atlas-fast",
    description: "Force low-latency optimized route",
    route: "route.low_cost",
  },
  {
    id: "atlas-long-context",
    description: "Force long context analysis route",
    route: "route.long_context",
  },
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

async function copyEndpoint() {
  await navigator.clipboard.writeText(openAiEndpoint.value);
}

async function copyDownstreamUrl() {
  await navigator.clipboard.writeText(downstreamEndpoint.value);
}

function openRoutingSettings() {
  router.push("/routing");
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
  "atlas-auto",
  "atlas-debug",
  "atlas-cheap",
  "atlas-docs",
  "atlas-architect",
  "atlas-fast",
  "atlas-long-context",
  "smart-auto",
  "smart-debug",
  "smart-cheap",
  "smart-docs",
  "smart-architect",
  "smart-fast",
  "smart-long-context",
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
