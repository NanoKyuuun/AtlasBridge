<template>
  <div class="space-y-6 lg:space-y-8">
    <PageHeader
      eyebrow="Advanced"
      title="Advanced Settings"
      description="Fine-tune server, downstream, security, and configuration workflows."
    />

    <SettingsSection
      title="Server Settings"
      description="Control the local host and port AtlasBridge binds to, plus the downstream endpoint used for routing."
    >
      <div class="grid gap-4 lg:grid-cols-2">
        <FormField label="Proxy host" hint="Hostname or IP address AtlasBridge listens on.">
          <input
            class="w-full bg-transparent text-sm text-white outline-none placeholder:text-slate-500"
            v-model="server.host"
            @input="dirty = true"
          />
        </FormField>

        <FormField label="Proxy port" hint="Valid range: 1024 to 65535.">
          <input
            type="number"
            class="w-full bg-transparent text-sm text-white outline-none placeholder:text-slate-500"
            v-model.number="server.port"
            @input="dirty = true"
            min="1024"
            max="65535"
          />
        </FormField>

        <FormField class="lg:col-span-2" label="9Router base URL" hint="Base URL for the downstream 9Router service used by AtlasBridge.">
          <input
            class="w-full bg-transparent text-sm text-white outline-none placeholder:text-slate-500 font-mono"
            v-model="downstream.base_url"
            @input="dirty = true"
          />
        </FormField>

        <FormField label="Timeout setting" hint="Request timeout in seconds for downstream calls.">
          <input
            type="number"
            class="w-full bg-transparent text-sm text-white outline-none placeholder:text-slate-500"
            v-model.number="downstream.timeout_seconds"
            @input="dirty = true"
          />
        </FormField>

        <div class="rounded-2xl border border-white/10 bg-white/5 p-4">
          <div class="text-xs uppercase tracking-[0.22em] text-slate-500">Authentication forwarding mode</div>
          <div class="mt-2 text-sm text-white">Uses the current config shape available in this build.</div>
          <div class="mt-2 text-xs leading-5 text-slate-400">
            This UI keeps the existing configuration fields intact. If auth forwarding is added to the config schema later, this section can be expanded without changing the layout.
          </div>
        </div>
      </div>
    </SettingsSection>

    <AdminTokenNotice
      v-if="configStore.generatedToken"
      :token="configStore.generatedToken"
      @dismiss="configStore.dismissToken()"
    />

    <SettingsSection
      title="Security"
      description="Adjust bind behavior, admin authentication, and access scope.">
      <div class="grid gap-4 lg:grid-cols-2">
        <label class="flex cursor-pointer items-start gap-4 rounded-3xl border border-white/10 bg-white/5 p-4 transition-all duration-200 hover:border-cyan-400/25 hover:bg-white/8">
          <ToggleSwitch v-model="security.admin_auth_enabled" label="Admin auth" @update:model-value="dirty = true" />
          <div>
            <div class="text-sm font-medium text-white">Admin authentication</div>
            <p class="mt-1 text-xs text-slate-400">Require a bearer token for all /admin/api/* requests.</p>
          </div>
        </label>

        <label class="flex cursor-pointer items-start gap-4 rounded-3xl border border-white/10 bg-white/5 p-4 transition-all duration-200 hover:border-cyan-400/25 hover:bg-white/8">
          <ToggleSwitch v-model="security.bind_localhost_only" label="Localhost only" @update:model-value="dirty = true" />
          <div>
            <div class="text-sm font-medium text-white">Localhost only</div>
            <p class="mt-1 text-xs text-slate-400">Bind to 127.0.0.1 only, which is recommended for most setups.</p>
          </div>
        </label>

        <label class="flex cursor-pointer items-start gap-4 rounded-3xl border border-white/10 bg-white/5 p-4 transition-all duration-200 hover:border-cyan-400/25 hover:bg-white/8">
          <ToggleSwitch v-model="security.allow_lan_access" label="LAN access" @update:model-value="dirty = true" />
          <div>
            <div class="text-sm font-medium text-white">Allow LAN access</div>
            <p class="mt-1 text-xs text-slate-400">Allow connections from other devices on your network.</p>
          </div>
        </label>
      </div>
    </SettingsSection>

    <SettingsSection
      title="Debug"
      description="Enable verbose output for troubleshooting when needed."
    >
      <div class="space-y-4">
        <label class="flex cursor-pointer items-start gap-4 rounded-3xl border border-white/10 bg-white/5 p-4 transition-all duration-200 hover:border-amber-400/25 hover:bg-white/8">
          <ToggleSwitch v-model="debugMode" label="Debug mode" @update:model-value="dirty = true" />
          <div>
            <div class="text-sm font-medium text-white">Debug mode</div>
            <p class="mt-1 text-xs text-slate-400">Enable verbose logging for troubleshooting.</p>
          </div>
        </label>
        <div v-if="debugMode" class="rounded-2xl border border-amber-400/20 bg-amber-400/10 px-4 py-3 text-sm text-amber-100">
          Debug mode increases logging verbosity. Disable when troubleshooting is complete.
        </div>
      </div>
    </SettingsSection>

    <SettingsSection
      title="Config Import / Export"
      description="Move configuration in and out of AtlasBridge without changing the saved structure."
    >
      <div class="flex flex-col gap-3 sm:flex-row sm:flex-wrap">
        <GhostButton @click="exportConfig">Export Config</GhostButton>
        <GhostButton @click="triggerImport">Import Config</GhostButton>
        <button class="inline-flex items-center justify-center rounded-2xl border border-rose-400/20 bg-rose-400/10 px-4 py-2.5 text-sm font-medium text-rose-100 transition-all hover:border-rose-400/30 hover:bg-rose-400/15" @click="resetConfig">
          Reset Configuration
        </button>
        <input
          ref="fileInput"
          type="file"
          accept=".json"
          class="hidden"
          @change="importConfig"
        />
      </div>
      <div class="mt-4 rounded-2xl border border-white/10 bg-white/5 p-4 text-sm text-slate-400">
        Export generates a full configuration bundle. Import replaces the current configuration with the uploaded JSON file.
      </div>
    </SettingsSection>

    <div class="flex justify-start">
      <GradientButton @click="save" :disabled="!dirty">Save Changes</GradientButton>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useConfigStore } from "../stores/config";
import { api } from "../api/client";
import PageHeader from "../components/ui/PageHeader.vue";
import SettingsSection from "../components/ui/SettingsSection.vue";
import FormField from "../components/ui/FormField.vue";
import ToggleSwitch from "../components/ui/ToggleSwitch.vue";
import GradientButton from "../components/ui/GradientButton.vue";
import GhostButton from "../components/ui/GhostButton.vue";
import AdminTokenNotice from "../components/ui/AdminTokenNotice.vue";

const configStore = useConfigStore();
const dirty = ref(false);
const fileInput = ref<HTMLInputElement>();

const server = ref({
  host: "127.0.0.1",
  port: 20127,
  admin_path: "/admin",
});
const downstream = ref({
  base_url: "http://127.0.0.1:20128/v1",
  timeout_seconds: 120,
});
const security = ref({
  admin_auth_enabled: false,
  admin_token_hash: "",
  bind_localhost_only: true,
  allow_lan_access: false,
});
const debugMode = ref(false);

function load() {
  if (configStore.config) {
    server.value = { ...configStore.config.server };
    downstream.value = { ...configStore.config.downstream };
    security.value = { ...configStore.config.security };
    debugMode.value = configStore.config.logging.level === "debug";
  }
  dirty.value = false;
}

async function save() {
  try {
    await configStore.saveConfig({
      server: server.value,
      downstream: downstream.value,
      security: security.value,
      logging: {
        ...configStore.config!.logging,
        level: debugMode.value ? "debug" : "info",
      },
    });
    load();
  } catch (e: any) {
    console.error("Failed to save settings:", e.message);
  }
}

async function exportConfig() {
  try {
    const data = await api.exportConfig();
    const blob = new Blob([JSON.stringify(data, null, 2)], {
      type: "application/json",
    });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "atlasbridge-config.json";
    a.click();
    URL.revokeObjectURL(url);
  } catch (e: any) {
    console.error("Failed to export config:", e.message);
  }
}

function triggerImport() {
  fileInput.value?.click();
}

async function importConfig(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0];
  if (!file) return;
  try {
    const text = await file.text();
    const data = JSON.parse(text);
    await api.importConfig(data);
    await configStore.fetchAll();
    load();
  } catch (e: any) {
    console.error("Failed to import config:", e.message);
  }
}

async function resetConfig() {
  if (confirm("Reset all settings to defaults? This cannot be undone.")) {
    await configStore.resetConfig();
    load();
  }
}

onMounted(load);
</script>
