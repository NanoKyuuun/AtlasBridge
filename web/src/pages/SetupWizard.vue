<template>
  <div class="mx-auto max-w-2xl space-y-6 lg:space-y-8">
    <PageHeader
      eyebrow="Setup"
      title="Setup Wizard"
      description="Configure AtlasBridge AI Proxy for the first time."
    />

    <GlassCard>
      <div class="space-y-6">
        <ul class="steps steps-horizontal w-full">
          <li class="step" :class="{ 'step-primary': step >= 1 }">Welcome</li>
          <li class="step" :class="{ 'step-primary': step >= 2 }">Endpoint</li>
          <li class="step" :class="{ 'step-primary': step >= 3 }">Downstream</li>
          <li class="step" :class="{ 'step-primary': step >= 4 }">Routing</li>
          <li class="step" :class="{ 'step-primary': step >= 5 }">Done</li>
        </ul>

        <div v-if="step === 1" class="space-y-4">
          <div>
            <h3 class="text-lg font-semibold text-white">Welcome to AtlasBridge AI Proxy</h3>
            <p class="mt-1 text-sm text-slate-400">
              This wizard will help you configure AtlasBridge AI Proxy for the
              first time. You can change these settings later from the Web UI.
            </p>
          </div>
          <div class="rounded-2xl border border-cyan-400/20 bg-cyan-400/10 px-4 py-3 text-sm text-cyan-100">
            AtlasBridge AI Proxy acts as an intelligent routing layer between
            your AI coding assistant and 9Router.
          </div>
          <GradientButton @click="step = 2">Get Started</GradientButton>
        </div>

        <div v-if="step === 2" class="space-y-4">
          <div>
            <h3 class="text-lg font-semibold text-white">API Endpoint</h3>
            <p class="mt-1 text-sm text-slate-400">
              Configure where AtlasBridge AI Proxy listens for requests.
            </p>
          </div>
          <FormField label="Host">
            <input
              class="w-full bg-transparent font-mono text-sm text-white outline-none"
              v-model="config.server.host"
            />
          </FormField>
          <FormField label="Port">
            <input
              type="number"
              class="w-full bg-transparent text-sm text-white outline-none"
              v-model.number="config.server.port"
            />
          </FormField>
          <div class="rounded-2xl border border-emerald-400/20 bg-emerald-400/10 px-4 py-3 text-sm text-emerald-100">
            Endpoint: http://{{ config.server.host }}:{{ config.server.port }}/v1
          </div>
          <div class="flex gap-2">
            <GhostButton @click="step = 1">Back</GhostButton>
            <GradientButton @click="step = 3">Next</GradientButton>
          </div>
        </div>

        <div v-if="step === 3" class="space-y-4">
          <div>
            <h3 class="text-lg font-semibold text-white">9Router Downstream</h3>
            <p class="mt-1 text-sm text-slate-400">
              Configure the 9Router endpoint to forward requests to.
            </p>
          </div>
          <FormField label="Base URL">
            <input
              class="w-full bg-transparent font-mono text-sm text-white outline-none"
              v-model="config.downstream.base_url"
            />
          </FormField>
          <FormField label="Timeout (seconds)">
            <input
              type="number"
              class="w-full bg-transparent text-sm text-white outline-none"
              v-model.number="config.downstream.timeout_seconds"
            />
          </FormField>
          <div class="flex gap-2">
            <GhostButton @click="step = 2">Back</GhostButton>
            <GradientButton @click="step = 4">Next</GradientButton>
          </div>
        </div>

        <div v-if="step === 4" class="space-y-4">
          <div>
            <h3 class="text-lg font-semibold text-white">Routing Mode</h3>
            <p class="mt-1 text-sm text-slate-400">
              Choose how AtlasBridge AI Proxy routes requests.
            </p>
          </div>
          <div class="space-y-2">
            <label
              class="flex cursor-pointer items-center gap-3 rounded-2xl border border-white/10 bg-white/5 p-4 transition-all duration-200 hover:border-white/20"
              :class="{ 'border-cyan-400/40 bg-cyan-400/10': config.routing.auto_routing }"
            >
              <input
                type="radio"
                class="radio radio-primary"
                :checked="config.routing.auto_routing"
                @change="config.routing.auto_routing = true"
              />
              <div>
                <span class="font-medium text-white">Auto Routing (Recommended)</span>
                <p class="text-xs text-slate-400">
                  Automatically classify tasks and select the best route
                </p>
              </div>
            </label>
            <label
              class="flex cursor-pointer items-center gap-3 rounded-2xl border border-white/10 bg-white/5 p-4 transition-all duration-200 hover:border-white/20"
              :class="{ 'border-cyan-400/40 bg-cyan-400/10': !config.routing.auto_routing }"
            >
              <input
                type="radio"
                class="radio radio-primary"
                :checked="!config.routing.auto_routing"
                @change="config.routing.auto_routing = false"
              />
              <div>
                <span class="font-medium text-white">Manual Routing</span>
                <p class="text-xs text-slate-400">
                  Use the model field to determine the route
                </p>
              </div>
            </label>
          </div>
          <div class="flex gap-2">
            <GhostButton @click="step = 3">Back</GhostButton>
            <GradientButton @click="step = 5">Next</GradientButton>
          </div>
        </div>

        <div v-if="step === 5" class="space-y-4">
          <div>
            <h3 class="text-lg font-semibold text-white">Setup Complete</h3>
            <p class="mt-1 text-sm text-slate-400">
              AtlasBridge AI Proxy is configured and ready to use.
            </p>
          </div>
          <div class="rounded-2xl border border-emerald-400/20 bg-emerald-400/10 px-4 py-3">
            <span class="font-medium text-emerald-100">Your endpoint:</span>
            <p class="mt-1 font-mono text-sm text-white">
              http://{{ config.server.host }}:{{ config.server.port }}/v1
            </p>
            <p class="mt-1 text-xs text-emerald-200/70">
              Set this as the base URL in your AI coding assistant.
            </p>
          </div>
          <div class="rounded-2xl border border-cyan-400/20 bg-cyan-400/10 px-4 py-3 text-sm text-cyan-100">
            Use model <code class="font-mono">atlas-auto</code> or
            <code class="font-mono">smart-auto</code> for automatic routing,
            or pick a specific alias like
            <code class="font-mono">atlas-debug</code> /
            <code class="font-mono">smart-debug</code>.
          </div>
          <div class="flex gap-2">
            <GhostButton @click="step = 4">Back</GhostButton>
            <GradientButton @click="finish">Finish & Save</GradientButton>
          </div>
        </div>
      </div>
    </GlassCard>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from "vue";
import { useRouter } from "vue-router";
import { useConfigStore } from "../stores/config";
import GlassCard from "../components/ui/GlassCard.vue";
import PageHeader from "../components/ui/PageHeader.vue";
import FormField from "../components/ui/FormField.vue";
import GradientButton from "../components/ui/GradientButton.vue";
import GhostButton from "../components/ui/GhostButton.vue";

const router = useRouter();
const configStore = useConfigStore();
const step = ref(1);

const config = reactive({
  server: {
    host: "127.0.0.1",
    port: 20127,
    admin_path: "/admin",
  },
  downstream: {
    base_url: "http://127.0.0.1:20128/v1",
    timeout_seconds: 120,
  },
  routing: {
    auto_routing: true,
    default_route: "route.default",
    low_confidence_route: "route.default",
    confidence_threshold: 0.55,
    smart_fast_route: "route.low_cost",
    metadata_transport: "header",
  },
  app: {
    name: "AtlasBridge AI Proxy",
    mode: "always_on",
    first_run_completed: true,
  },
});

async function finish() {
  await configStore.saveConfig(config);
  router.push("/");
}

onMounted(() => {
  if (configStore.config) {
    Object.assign(config.server, configStore.config.server);
    Object.assign(config.downstream, configStore.config.downstream);
    Object.assign(config.routing, configStore.config.routing);
    Object.assign(config.app, configStore.config.app);
  }
});
</script>
