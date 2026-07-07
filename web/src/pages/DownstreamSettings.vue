<template>
  <div class="space-y-6 lg:space-y-8">
    <PageHeader
      eyebrow="9Router"
      title="9Router Downstream"
      description="Configure and verify the downstream endpoint AtlasBridge forwards requests to."
    />

    <div class="grid gap-4 lg:grid-cols-[1.1fr_0.9fr]">
      <SettingsSection title="Connection Settings" description="Set the 9Router base URL and request timeout.">
        <div class="space-y-4">
          <FormField label="Base URL" hint="The downstream OpenAI-compatible 9Router endpoint.">
            <input
              class="w-full bg-transparent text-sm text-white outline-none placeholder:text-slate-500 font-mono"
              v-model="downstream.base_url"
              @input="dirty = true"
              placeholder="http://127.0.0.1:20128/v1"
            />
          </FormField>

          <FormField label="Timeout" hint="Timeout in seconds. Recommended range: 10 to 300.">
            <input
              type="number"
              class="w-full bg-transparent text-sm text-white outline-none placeholder:text-slate-500"
              v-model.number="downstream.timeout_seconds"
              @input="dirty = true"
              min="10"
              max="300"
            />
          </FormField>
        </div>
      </SettingsSection>

      <GlassCard>
        <div class="space-y-5">
          <div>
            <h2 class="text-lg font-semibold text-white">Connection Status</h2>
            <p class="mt-1 text-sm text-slate-400">Check whether AtlasBridge can reach the configured downstream.</p>
          </div>

          <div class="rounded-2xl border border-white/10 bg-white/5 p-4">
            <div class="flex flex-wrap items-center gap-3">
              <StatusBadge :status="health?.status === 'connected' ? 'active' : 'error'" :label="health?.status || 'unknown'" />
              <span class="break-all text-sm text-slate-400">{{ health?.message || health?.url || downstream.base_url }}</span>
            </div>
          </div>

          <div class="flex flex-col gap-3 sm:flex-row sm:flex-wrap">
            <GhostButton @click="checkHealth">Check Connection</GhostButton>
            <a class="inline-flex items-center justify-center gap-2 rounded-2xl border border-white/10 bg-white/5 px-4 py-2.5 text-sm font-medium text-slate-200 transition-all duration-200 hover:border-cyan-400/25 hover:bg-white/10 hover:text-white" :href="dashboardUrl" target="_blank">
              Open 9Router Dashboard
            </a>
          </div>
        </div>
      </GlassCard>
    </div>

    <div class="flex justify-start">
      <GradientButton @click="save" :disabled="!dirty">Save Changes</GradientButton>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useConfigStore } from "../stores/config";
import { api, type DownstreamHealth } from "../api/client";
import GlassCard from "../components/ui/GlassCard.vue";
import PageHeader from "../components/ui/PageHeader.vue";
import SettingsSection from "../components/ui/SettingsSection.vue";
import FormField from "../components/ui/FormField.vue";
import GradientButton from "../components/ui/GradientButton.vue";
import GhostButton from "../components/ui/GhostButton.vue";
import StatusBadge from "../components/ui/StatusBadge.vue";

const configStore = useConfigStore();
const dirty = ref(false);
const health = ref<DownstreamHealth | null>(null);

const downstream = ref({
  type: "9router",
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
