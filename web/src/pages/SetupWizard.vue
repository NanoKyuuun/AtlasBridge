<template>
  <div class="max-w-2xl mx-auto">
    <div class="card bg-base-100 shadow-md">
      <div class="card-body">
        <h2 class="card-title text-xl mb-4">Setup Wizard</h2>

        <ul class="steps steps-horizontal w-full mb-6">
          <li class="step" :class="{ 'step-primary': step >= 1 }">Welcome</li>
          <li class="step" :class="{ 'step-primary': step >= 2 }">Endpoint</li>
          <li class="step" :class="{ 'step-primary': step >= 3 }">
            Downstream
          </li>
          <li class="step" :class="{ 'step-primary': step >= 4 }">Routing</li>
          <li class="step" :class="{ 'step-primary': step >= 5 }">Done</li>
        </ul>

        <div v-if="step === 1" class="space-y-4">
          <h3 class="text-lg font-semibold">Welcome to AtlasBridge AI Proxy</h3>
          <p class="text-sm text-base-content/70">
            This wizard will help you configure AtlasBridge AI Proxy for the
            first time. You can change these settings later from the Web UI.
          </p>
          <div class="alert alert-info">
            <span
              >AtlasBridge AI Proxy acts as an intelligent routing layer
              between your AI coding assistant and 9Router.</span
            >
          </div>
          <button class="btn btn-primary" @click="step = 2">Get Started</button>
        </div>

        <div v-if="step === 2" class="space-y-4">
          <h3 class="text-lg font-semibold">API Endpoint</h3>
          <p class="text-sm text-base-content/70">
            Configure where AtlasBridge AI Proxy listens for requests.
          </p>
          <div class="form-control">
            <label class="label"><span class="label-text">Host</span></label>
            <input
              class="input input-bordered font-mono"
              v-model="config.server.host"
            />
          </div>
          <div class="form-control">
            <label class="label"><span class="label-text">Port</span></label>
            <input
              type="number"
              class="input input-bordered font-mono"
              v-model.number="config.server.port"
            />
          </div>
          <div class="alert alert-success">
            <span
              >Endpoint: http://{{ config.server.host }}:{{
                config.server.port
              }}/v1</span
            >
          </div>
          <div class="flex gap-2">
            <button class="btn btn-ghost" @click="step = 1">Back</button>
            <button class="btn btn-primary" @click="step = 3">Next</button>
          </div>
        </div>

        <div v-if="step === 3" class="space-y-4">
          <h3 class="text-lg font-semibold">9Router Downstream</h3>
          <p class="text-sm text-base-content/70">
            Configure the 9Router endpoint to forward requests to.
          </p>
          <div class="form-control">
            <label class="label"
              ><span class="label-text">Base URL</span></label
            >
            <input
              class="input input-bordered font-mono"
              v-model="config.downstream.base_url"
            />
          </div>
          <div class="form-control">
            <label class="label"
              ><span class="label-text">Timeout (seconds)</span></label
            >
            <input
              type="number"
              class="input input-bordered"
              v-model.number="config.downstream.timeout_seconds"
            />
          </div>
          <div class="flex gap-2">
            <button class="btn btn-ghost" @click="step = 2">Back</button>
            <button class="btn btn-primary" @click="step = 4">Next</button>
          </div>
        </div>

        <div v-if="step === 4" class="space-y-4">
          <h3 class="text-lg font-semibold">Routing Mode</h3>
          <p class="text-sm text-base-content/70">
            Choose how AtlasBridge AI Proxy routes requests.
          </p>
          <div class="space-y-2">
            <label
              class="flex items-center gap-3 p-4 rounded-lg border cursor-pointer"
              :class="{
                'border-primary bg-primary/5': config.routing.auto_routing,
              }"
            >
              <input
                type="radio"
                class="radio radio-primary"
                :checked="config.routing.auto_routing"
                @change="config.routing.auto_routing = true"
              />
              <div>
                <span class="font-medium">Auto Routing (Recommended)</span>
                <p class="text-xs text-base-content/50">
                  Automatically classify tasks and select the best route
                </p>
              </div>
            </label>
            <label
              class="flex items-center gap-3 p-4 rounded-lg border cursor-pointer"
              :class="{
                'border-primary bg-primary/5': !config.routing.auto_routing,
              }"
            >
              <input
                type="radio"
                class="radio radio-primary"
                :checked="!config.routing.auto_routing"
                @change="config.routing.auto_routing = false"
              />
              <div>
                <span class="font-medium">Manual Routing</span>
                <p class="text-xs text-base-content/50">
                  Use the model field to determine the route
                </p>
              </div>
            </label>
          </div>
          <div class="flex gap-2">
            <button class="btn btn-ghost" @click="step = 3">Back</button>
            <button class="btn btn-primary" @click="step = 5">Next</button>
          </div>
        </div>

        <div v-if="step === 5" class="space-y-4">
          <h3 class="text-lg font-semibold">Setup Complete</h3>
          <p class="text-sm text-base-content/70">
            AtlasBridge AI Proxy is configured and ready to use.
          </p>
          <div class="alert alert-success">
            <div>
              <span class="font-medium">Your endpoint:</span>
              <p class="font-mono text-sm">
                http://{{ config.server.host }}:{{ config.server.port }}/v1
              </p>
              <p class="text-xs mt-1">
                Set this as the base URL in your AI coding assistant.
              </p>
            </div>
          </div>
          <div class="alert alert-info">
            <span
              >Use model <code class="font-mono">atlas-auto</code> or
              <code class="font-mono">smart-auto</code> for automatic routing,
              or pick a specific alias like
              <code class="font-mono">atlas-debug</code> /
              <code class="font-mono">smart-debug</code>.</span
            >
          </div>
          <div class="flex gap-2">
            <button class="btn btn-ghost" @click="step = 4">Back</button>
            <button class="btn btn-primary" @click="finish">
              Finish & Save
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from "vue";
import { useRouter } from "vue-router";
import { useConfigStore } from "../stores/config";

const router = useRouter();
const configStore = useConfigStore();
const step = ref(1);

const config = reactive({
  server: {
    host: "127.0.0.1",
    port: 20127,
    admin_path: "/admin",
    api_base_path: "/v1",
  },
  downstream: {
    type: "9router",
    base_url: "http://127.0.0.1:20128/v1",
    timeout_seconds: 120,
  },
  routing: {
    auto_routing: true,
    default_route: "route.default",
    low_confidence_route: "route.default",
    explicit_override_enabled: true,
    confidence_threshold: 0.55,
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
