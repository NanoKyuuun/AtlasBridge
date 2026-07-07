<template>
  <div class="space-y-6 lg:space-y-8">
    <PageHeader
      eyebrow="Privacy"
      title="Privacy & Logging"
      description="Configure privacy mode and logging behavior."
    />

    <GlassCard>
      <div class="space-y-5">
        <div>
          <h2 class="text-lg font-semibold text-white">Privacy Mode</h2>
          <p class="mt-1 text-sm text-slate-400">
            Choose how much request detail AtlasBridge keeps in logs.
          </p>
        </div>

        <div class="grid gap-3 lg:grid-cols-3">
          <label
            class="relative cursor-pointer overflow-hidden rounded-3xl border p-4 transition-all duration-200"
            :class="modeCardClass('standard')"
          >
            <input type="radio" class="sr-only" v-model="logging.privacy_mode" value="standard" @change="dirty = true" />
            <div class="flex items-start justify-between gap-3">
              <div>
                <div class="text-sm font-semibold text-white">Standard</div>
                <div class="mt-2 text-xs leading-5 text-slate-400">
                  Log metadata routing without full prompt content.
                </div>
              </div>
              <StatusBadge :status="logging.privacy_mode === 'standard' ? 'active' : 'inactive'" :label="logging.privacy_mode === 'standard' ? 'selected' : 'ready'" />
            </div>
          </label>

          <label
            class="relative cursor-pointer overflow-hidden rounded-3xl border p-4 transition-all duration-200"
            :class="modeCardClass('strict')"
          >
            <input type="radio" class="sr-only" v-model="logging.privacy_mode" value="strict" @change="dirty = true" />
            <div class="flex items-start justify-between gap-3">
              <div>
                <div class="text-sm font-semibold text-white">Strict</div>
                <div class="mt-2 text-xs leading-5 text-slate-400">
                  Log only request ID, status, latency, and route.
                </div>
              </div>
              <StatusBadge :status="logging.privacy_mode === 'strict' ? 'active' : 'inactive'" :label="logging.privacy_mode === 'strict' ? 'selected' : 'ready'" />
            </div>
          </label>

          <label
            class="relative cursor-pointer overflow-hidden rounded-3xl border p-4 transition-all duration-200"
            :class="modeCardClass('debug')"
          >
            <input type="radio" class="sr-only" v-model="logging.privacy_mode" value="debug" @change="dirty = true" />
            <div class="flex items-start justify-between gap-3">
              <div>
                <div class="text-sm font-semibold text-white">Debug</div>
                <div class="mt-2 text-xs leading-5 text-slate-400">
                  Additional info for troubleshooting, explicitly opt-in only.
                </div>
              </div>
              <StatusBadge :status="logging.privacy_mode === 'debug' ? 'warning' : 'inactive'" :label="logging.privacy_mode === 'debug' ? 'selected' : 'ready'" />
            </div>
          </label>
        </div>

        <div v-if="logging.privacy_mode === 'debug'" class="rounded-2xl border border-amber-400/20 bg-amber-400/10 px-4 py-3 text-sm text-amber-100">
          Debug mode increases logging verbosity. Do not use in production or shared environments.
        </div>
      </div>
    </GlassCard>

    <div class="grid gap-4 lg:grid-cols-2">
      <GlassCard>
        <div class="space-y-4">
          <div>
            <h2 class="text-lg font-semibold text-white">Logging Controls</h2>
            <p class="mt-1 text-sm text-slate-400">
              Toggle what AtlasBridge records for diagnostics and privacy protection.
            </p>
          </div>

          <label class="flex cursor-pointer items-start gap-4 rounded-3xl border border-white/10 bg-white/5 p-4 transition-all duration-200 hover:border-cyan-400/25 hover:bg-white/8">
            <ToggleSwitch v-model="logging.metadata_logging_enabled" label="Metadata logging" @update:model-value="dirty = true" />
            <div class="min-w-0">
              <div class="text-sm font-medium text-white">Metadata logging</div>
              <div class="mt-1 text-xs leading-5 text-slate-400">
                Log request metadata such as task type, route, and latency.
              </div>
            </div>
          </label>

          <div class="rounded-3xl border border-white/10 bg-white/5 p-4">
            <div class="flex items-start justify-between gap-4">
              <div>
                <div class="text-sm font-medium text-white">Full prompt logging</div>
                <div class="mt-1 text-xs leading-5 text-slate-400">
                  Log full prompts only when explicitly enabled.
                </div>
              </div>
              <StatusBadge :status="logging.prompt_logging_enabled ? 'warning' : 'inactive'" :label="logging.prompt_logging_enabled ? 'enabled' : 'disabled'" />
            </div>
            <div class="mt-3 rounded-2xl border border-white/10 bg-slate-950/60 px-4 py-3 text-xs text-slate-300">
              Full prompt logging disabled by default.
            </div>
          </div>

          <label class="flex cursor-pointer items-start gap-4 rounded-3xl border border-white/10 bg-white/5 p-4 transition-all duration-200 hover:border-cyan-400/25 hover:bg-white/8">
            <ToggleSwitch v-model="redactSecrets" label="Secret redaction" @update:model-value="dirty = true" />
            <div class="min-w-0">
              <div class="text-sm font-medium text-white">Secret redaction</div>
              <div class="mt-1 text-xs leading-5 text-slate-400">
                Automatically redact API keys, tokens, and passwords from logs.
              </div>
            </div>
          </label>

          <div class="grid gap-3 sm:grid-cols-2">
            <div class="rounded-2xl border border-emerald-400/20 bg-emerald-400/10 p-4">
              <div class="text-xs uppercase tracking-[0.22em] text-emerald-200">Redaction status</div>
              <div class="mt-2 text-sm font-semibold text-white">
                {{ redactSecrets ? 'Protected' : 'Unprotected' }}
              </div>
            </div>
            <div class="rounded-2xl border border-white/10 bg-white/5 p-4">
              <div class="text-xs uppercase tracking-[0.22em] text-slate-500">Retention days</div>
              <input
                type="number"
                class="mt-2 w-full rounded-2xl border border-white/10 bg-slate-950/70 px-3 py-2 text-sm text-white outline-none focus:border-cyan-400/40 focus:ring-2 focus:ring-cyan-400/20"
                v-model.number="logging.retention_days"
                @input="dirty = true"
                min="1"
                max="90"
              />
            </div>
          </div>
        </div>
      </GlassCard>

      <GlassCard>
        <div class="space-y-4">
          <div>
            <h2 class="text-lg font-semibold text-white">Security Snapshot</h2>
            <p class="mt-1 text-sm text-slate-400">
              Quick readout of the current privacy posture.
            </p>
          </div>

          <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-1 xl:grid-cols-2">
            <div class="rounded-2xl border border-white/10 bg-white/5 p-4">
              <div class="text-xs uppercase tracking-[0.22em] text-slate-500">Metadata logging</div>
              <div class="mt-2 flex items-center gap-2 text-sm text-white">
                <span class="h-2 w-2 rounded-full" :class="logging.metadata_logging_enabled ? 'bg-emerald-400 shadow-[0_0_10px_rgba(34,197,94,0.45)]' : 'bg-slate-500'"></span>
                {{ logging.metadata_logging_enabled ? 'Enabled' : 'Disabled' }}
              </div>
            </div>

            <div class="rounded-2xl border border-white/10 bg-white/5 p-4">
              <div class="text-xs uppercase tracking-[0.22em] text-slate-500">Prompt logging</div>
              <div class="mt-2 flex items-center gap-2 text-sm text-white">
                <span class="h-2 w-2 rounded-full" :class="logging.prompt_logging_enabled ? 'bg-amber-400 shadow-[0_0_10px_rgba(245,158,11,0.45)]' : 'bg-slate-500'"></span>
                {{ logging.prompt_logging_enabled ? 'Enabled' : 'Disabled' }}
              </div>
            </div>

            <div class="rounded-2xl border border-white/10 bg-white/5 p-4">
              <div class="text-xs uppercase tracking-[0.22em] text-slate-500">Secret redaction</div>
              <div class="mt-2 flex items-center gap-2 text-sm text-white">
                <span class="h-2 w-2 rounded-full" :class="redactSecrets ? 'bg-cyan-400 shadow-[0_0_10px_rgba(53,215,242,0.45)]' : 'bg-slate-500'"></span>
                {{ redactSecrets ? 'Protected' : 'Off' }}
              </div>
            </div>

            <div class="rounded-2xl border border-white/10 bg-white/5 p-4">
              <div class="text-xs uppercase tracking-[0.22em] text-slate-500">Privacy mode</div>
              <div class="mt-2 flex items-center gap-2 text-sm text-white">
                <StatusBadge :status="logging.privacy_mode === 'debug' ? 'warning' : logging.privacy_mode === 'strict' ? 'active' : 'inactive'" :label="logging.privacy_mode" />
              </div>
            </div>
          </div>
        </div>
      </GlassCard>
    </div>

    <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div class="flex gap-2">
        <GhostButton @click="exportLogs">Export Diagnostic Report</GhostButton>
        <button class="inline-flex items-center justify-center rounded-2xl border border-rose-400/20 bg-rose-400/10 px-4 py-2.5 text-sm font-medium text-rose-100 transition-all hover:border-rose-400/30 hover:bg-rose-400/15" @click="clearLogs">Clear Logs</button>
      </div>
      <GradientButton @click="save" :disabled="!dirty">Save Changes</GradientButton>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useConfigStore } from "../stores/config";
import { api } from "../api/client";
import GlassCard from "../components/ui/GlassCard.vue";
import PageHeader from "../components/ui/PageHeader.vue";
import GradientButton from "../components/ui/GradientButton.vue";
import GhostButton from "../components/ui/GhostButton.vue";
import StatusBadge from "../components/ui/StatusBadge.vue";
import ToggleSwitch from "../components/ui/ToggleSwitch.vue";

const configStore = useConfigStore();
const dirty = ref(false);
const redactSecrets = ref(true);

const logging = ref({
  level: "info",
  privacy_mode: "standard",
  metadata_logging_enabled: true,
  prompt_logging_enabled: false,
  retention_days: 7,
});

function load() {
  if (configStore.config) {
    logging.value = { ...configStore.config.logging };
  }
  dirty.value = false;
}

function modeCardClass(mode: string) {
  const selected = logging.value.privacy_mode === mode;
  if (mode === "debug") {
    return selected
      ? "border-amber-400/30 bg-amber-400/10 shadow-[0_0_24px_rgba(245,158,11,0.12)]"
      : "border-white/10 bg-white/5 hover:border-amber-400/20 hover:bg-white/8";
  }
  return selected
    ? "border-cyan-400/30 bg-gradient-to-br from-blue-500/16 via-violet-500/14 to-cyan-400/10 shadow-[0_0_24px_rgba(53,215,242,0.12)]"
    : "border-white/10 bg-white/5 hover:border-cyan-400/20 hover:bg-white/8";
}

async function save() {
  try {
    await configStore.saveConfig({ logging: logging.value });
    dirty.value = false;
  } catch (e: any) {
    console.error("Failed to save privacy settings:", e.message);
  }
}

async function exportLogs() {
  try {
    const data = await api.exportDiagnostics();
    const blob = new Blob([JSON.stringify(data, null, 2)], {
      type: "application/json",
    });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "diagnostics.json";
    a.click();
    URL.revokeObjectURL(url);
  } catch (e: any) {
    console.error("Failed to export diagnostics:", e.message);
  }
}

async function clearLogs() {
  if (confirm("Clear all metadata logs? This cannot be undone.")) {
    try {
      await api.clearLogs();
    } catch (e: any) {
      console.error("Failed to clear logs:", e.message);
    }
  }
}

onMounted(load);
</script>
