<template>
  <div class="space-y-6 lg:space-y-8">
    <PageHeader
      eyebrow="Startup"
      title="Startup Settings"
      description="Control how AtlasBridge runs when your device starts."
    />

    <GlassCard>
      <div class="grid gap-6 lg:grid-cols-[1.2fr_0.8fr]">
        <div class="space-y-4">
          <div>
            <h2 class="text-lg font-semibold text-white">Startup Mode</h2>
            <p class="mt-1 text-sm text-slate-400">
              Choose how AtlasBridge should initialize when the system or app launches.
            </p>
          </div>

          <div class="grid gap-3 md:grid-cols-3">
            <label
              class="group relative cursor-pointer overflow-hidden rounded-3xl border p-4 transition-all duration-200"
              :class="modeCardClass('always_on')"
            >
              <input
                type="radio"
                class="sr-only"
                v-model="appMode"
                value="always_on"
                @change="dirty = true"
              />
              <div class="flex items-start justify-between gap-3">
                <div>
                  <div class="flex items-center gap-2 text-sm font-semibold text-white">
                    <span class="text-base text-cyan-300">◉</span>
                    Always On
                  </div>
                  <p class="mt-2 text-xs leading-5 text-slate-400">
                    Proxy starts automatically and remains ready in the background.
                  </p>
                </div>
                <StatusBadge :status="appMode === 'always_on' ? 'active' : 'inactive'" :label="appMode === 'always_on' ? 'selected' : 'ready'" />
              </div>
            </label>

            <label
              class="group relative cursor-pointer overflow-hidden rounded-3xl border p-4 transition-all duration-200"
              :class="modeCardClass('manual')"
            >
              <input
                type="radio"
                class="sr-only"
                v-model="appMode"
                value="manual"
                @change="dirty = true"
              />
              <div class="flex items-start justify-between gap-3">
                <div>
                  <div class="flex items-center gap-2 text-sm font-semibold text-white">
                    <span class="text-base text-violet-300">▣</span>
                    Manual
                  </div>
                  <p class="mt-2 text-xs leading-5 text-slate-400">
                    AtlasBridge stays idle until you start the proxy from the app.
                  </p>
                </div>
                <StatusBadge :status="appMode === 'manual' ? 'active' : 'inactive'" :label="appMode === 'manual' ? 'selected' : 'ready'" />
              </div>
            </label>

            <label
              class="group relative cursor-pointer overflow-hidden rounded-3xl border p-4 transition-all duration-200"
              :class="modeCardClass('disabled')"
            >
              <input
                type="radio"
                class="sr-only"
                v-model="appMode"
                value="disabled"
                @change="dirty = true"
              />
              <div class="flex items-start justify-between gap-3">
                <div>
                  <div class="flex items-center gap-2 text-sm font-semibold text-white">
                    <span class="text-base text-rose-300">⟡</span>
                    Disabled
                  </div>
                  <p class="mt-2 text-xs leading-5 text-slate-400">
                    Proxy does not accept requests until startup mode changes.
                  </p>
                </div>
                <StatusBadge :status="appMode === 'disabled' ? 'warning' : 'inactive'" :label="appMode === 'disabled' ? 'selected' : 'ready'" />
              </div>
            </label>
          </div>
        </div>

        <div class="rounded-3xl border border-white/10 bg-white/5 p-5">
          <h3 class="text-base font-semibold text-white">Selected mode</h3>
          <p class="mt-2 text-sm text-slate-400">
            {{ appModeDescription }}
          </p>
          <div class="mt-4 rounded-2xl border border-white/10 bg-slate-950/60 p-4">
            <div class="text-xs uppercase tracking-[0.24em] text-slate-500">Current selection</div>
            <div class="mt-2 text-lg font-semibold text-white">{{ appModeLabel }}</div>
            <div class="mt-2 text-sm text-slate-400">
              {{ startup.run_at_login ? 'Startup registration is active.' : 'Startup registration is inactive.' }}
            </div>
          </div>
        </div>
      </div>
    </GlassCard>

    <GlassCard>
      <div class="space-y-5">
        <div>
          <h2 class="text-lg font-semibold text-white">Startup Options</h2>
          <p class="mt-1 text-sm text-slate-400">
            Fine-tune how AtlasBridge behaves during login, app launch, and crash recovery.
          </p>
        </div>

        <div class="grid gap-4 lg:grid-cols-3">
          <label class="flex cursor-pointer flex-col gap-4 rounded-3xl border border-white/10 bg-white/5 p-4 transition-all duration-200 hover:border-cyan-400/25 hover:bg-white/8">
            <div class="flex items-start justify-between gap-3">
              <div>
                <div class="text-sm font-medium text-white">Auto-start on boot</div>
                <div class="mt-1 text-xs leading-5 text-slate-400">
                  Automatically start AtlasBridge when you log in.
                </div>
              </div>
              <span class="text-cyan-300">⌁</span>
            </div>
            <ToggleSwitch v-model="startup.run_at_login" label="Enabled" @update:model-value="dirty = true" />
          </label>

          <label class="flex cursor-pointer flex-col gap-4 rounded-3xl border border-white/10 bg-white/5 p-4 transition-all duration-200 hover:border-cyan-400/25 hover:bg-white/8">
            <div class="flex items-start justify-between gap-3">
              <div>
                <div class="text-sm font-medium text-white">Run in background</div>
                <div class="mt-1 text-xs leading-5 text-slate-400">
                  Start the proxy engine automatically when the app opens.
                </div>
              </div>
              <span class="text-violet-300">◔</span>
            </div>
            <ToggleSwitch v-model="startup.start_proxy_on_app_launch" label="Enabled" @update:model-value="dirty = true" />
          </label>

          <label class="flex cursor-pointer flex-col gap-4 rounded-3xl border border-white/10 bg-white/5 p-4 transition-all duration-200 hover:border-cyan-400/25 hover:bg-white/8">
            <div class="flex items-start justify-between gap-3">
              <div>
                <div class="text-sm font-medium text-white">Restart after crash</div>
                <div class="mt-1 text-xs leading-5 text-slate-400">
                  Recover the proxy automatically if it crashes unexpectedly.
                </div>
              </div>
              <span class="text-emerald-300">⟲</span>
            </div>
            <ToggleSwitch v-model="startup.restart_after_crash" label="Enabled" @update:model-value="dirty = true" />
          </label>
        </div>
      </div>
    </GlassCard>

    <GlassCard>
      <div class="grid grid-cols-1 gap-4 md:grid-cols-3">
        <div class="rounded-2xl border border-white/10 bg-white/5 p-4">
          <div class="text-xs uppercase tracking-[0.22em] text-slate-500">Run at Login</div>
          <div class="mt-2 text-lg font-semibold text-white">
            <span :class="startup.run_at_login ? 'text-emerald-300' : 'text-slate-400'">
              {{ startup.run_at_login ? 'Active' : 'Inactive' }}
            </span>
          </div>
        </div>
        <div class="rounded-2xl border border-white/10 bg-white/5 p-4">
          <div class="text-xs uppercase tracking-[0.22em] text-slate-500">Mode</div>
          <div class="mt-2 text-lg font-semibold text-white">{{ appModeLabel }}</div>
        </div>
        <div class="rounded-2xl border border-white/10 bg-white/5 p-4">
          <div class="text-xs uppercase tracking-[0.22em] text-slate-500">Crash Recovery</div>
          <div class="mt-2 text-lg font-semibold text-white">
            <span :class="startup.restart_after_crash ? 'text-emerald-300' : 'text-slate-400'">
              {{ startup.restart_after_crash ? 'Enabled' : 'Disabled' }}
            </span>
          </div>
        </div>
      </div>
    </GlassCard>

    <div class="flex justify-start">
      <button class="gradient-button inline-flex items-center justify-center rounded-2xl px-4 py-2.5 text-sm font-medium disabled:cursor-not-allowed disabled:opacity-50" @click="save" :disabled="!dirty">
        Save Changes
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useConfigStore } from "../stores/config";
import GlassCard from "../components/ui/GlassCard.vue";
import PageHeader from "../components/ui/PageHeader.vue";
import StatusBadge from "../components/ui/StatusBadge.vue";
import ToggleSwitch from "../components/ui/ToggleSwitch.vue";

const configStore = useConfigStore();
const dirty = ref(false);

const startup = ref({
  run_at_login: false,
  start_proxy_on_app_launch: true,
  restart_after_crash: true,
});
const appMode = ref("manual");

const appModeLabel = computed(() => {
  const labels: Record<string, string> = {
    always_on: "Always On",
    manual: "Manual",
    disabled: "Disabled",
  };
  return labels[appMode.value] || appMode.value;
});

const appModeDescription = computed(() => {
  const descriptions: Record<string, string> = {
    always_on: "Proxy starts automatically on device startup",
    manual: "Proxy starts only when you open the app",
    disabled: "Proxy does not accept requests",
  };
  return descriptions[appMode.value] || "";
});

function modeCardClass(mode: string) {
  const selected = appMode.value === mode;
  if (mode === "disabled") {
    return selected
      ? "border-rose-400/30 bg-rose-400/10 shadow-[0_0_26px_rgba(239,68,68,0.14)]"
      : "border-white/10 bg-white/5 hover:border-rose-400/20 hover:bg-white/8";
  }
  return selected
    ? "border-blue-400/30 bg-gradient-to-br from-blue-500/16 via-violet-500/14 to-cyan-400/10 shadow-[0_0_28px_rgba(53,215,242,0.16)]"
    : "border-white/10 bg-white/5 hover:border-cyan-400/20 hover:bg-white/8";
}

function load() {
  if (configStore.config) {
    startup.value = { ...configStore.config.startup };
    appMode.value = configStore.config.app.mode;
  }
  dirty.value = false;
}

async function save() {
  try {
    await configStore.saveConfig({
      startup: startup.value,
      app: { ...configStore.config!.app, mode: appMode.value },
    });
    dirty.value = false;
  } catch (e: any) {
    console.error("Failed to save startup settings:", e.message);
  }
}

onMounted(load);
</script>
