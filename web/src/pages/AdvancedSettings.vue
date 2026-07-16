<template>
  <div>
    <!-- Proxy Config + Downstream -->
    <div class="grid grid-cols-2 gap-4 mb-6">
      <!-- Proxy Configuration -->
      <div class="card p-5">
        <div class="text-[14px] font-semibold mb-1">Proxy Configuration</div>
        <div class="text-[11.5px] text-[var(--text-mute)] mb-4">Pengaturan endpoint proxy lokal</div>
        <div class="space-y-3">
          <div>
            <label class="text-[12px] text-[var(--text-mute)] mb-1.5 block">Listen Host</label>
            <input type="text" class="input" v-model="server.host" @input="dirty = true" placeholder="127.0.0.1">
          </div>
          <div>
            <label class="text-[12px] text-[var(--text-mute)] mb-1.5 block">Listen Port</label>
            <input type="number" class="input" v-model="server.port" @input="dirty = true">
          </div>
          <div class="flex items-center justify-between pt-2">
            <div>
              <div class="text-[13px] font-medium">Localhost only</div>
              <div class="text-[11px] text-[var(--text-mute)]">Jangan expose ke public network</div>
            </div>
            <div
              class="toggle"
              :class="{ on: security.bind_localhost_only }"
              @click="security.bind_localhost_only = !security.bind_localhost_only; dirty = true"
            ></div>
          </div>
        </div>
      </div>

      <!-- Downstream 9Router -->
      <div class="card p-5">
        <div class="text-[14px] font-semibold mb-1">Downstream 9Router</div>
        <div class="text-[11.5px] text-[var(--text-mute)] mb-4">Konfigurasi router downstream</div>
        <div class="space-y-3">
          <div>
            <label class="text-[12px] text-[var(--text-mute)] mb-1.5 block">Base URL</label>
            <input type="text" class="input mono" v-model="downstream.base_url" @input="dirty = true" placeholder="https://router.internal/v1">
          </div>
          <div>
            <label class="text-[12px] text-[var(--text-mute)] mb-1.5 block">Request Timeout (seconds)</label>
            <input type="number" class="input" v-model="downstream.timeout_seconds" @input="dirty = true">
          </div>
        </div>
      </div>
    </div>

    <!-- Authentication + Debug -->
    <div class="grid grid-cols-2 gap-4 mb-6">
      <!-- Authentication -->
      <div class="card p-5">
        <div class="text-[14px] font-semibold mb-1">Authentication</div>
        <div class="text-[11.5px] text-[var(--text-mute)] mb-4">Proteksi akses proxy dan Web UI</div>
        <div class="space-y-3">
          <div class="flex items-center justify-between pt-2">
            <div>
              <div class="text-[13px] font-medium">Require Web UI auth</div>
              <div class="text-[11px] text-[var(--text-mute)]">Token wajib untuk akses settings</div>
            </div>
            <div
              class="toggle"
              :class="{ on: security.admin_auth_enabled }"
              @click="security.admin_auth_enabled = !security.admin_auth_enabled; dirty = true"
            ></div>
          </div>
          <div class="flex items-center justify-between">
            <div>
              <div class="text-[13px] font-medium">Allow LAN access</div>
              <div class="text-[11px] text-[var(--text-mute)]">Izinkan akses dari jaringan lokal</div>
            </div>
            <div
              class="toggle"
              :class="{ on: security.allow_lan_access }"
              @click="security.allow_lan_access = !security.allow_lan_access; dirty = true"
            ></div>
          </div>
        </div>
      </div>

      <!-- Debug & Diagnostics -->
      <div class="card p-5">
        <div class="text-[14px] font-semibold mb-1">Debug &amp; Diagnostics</div>
        <div class="text-[11.5px] text-[var(--text-mute)] mb-4">Opsi untuk troubleshooting</div>
        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <div>
              <div class="text-[13px] font-medium">Debug mode</div>
              <div class="text-[11px] text-[var(--text-mute)]">Verbose logging</div>
            </div>
            <div
              class="toggle"
              :class="{ on: debugMode }"
              @click="debugMode = !debugMode; dirty = true"
            ></div>
          </div>
        </div>
        <div class="divider"></div>
        <button class="btn btn-primary w-full justify-center" :disabled="!dirty" @click="save">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"/><polyline points="17 21 17 13 7 13 7 21"/><polyline points="7 3 7 8 15 8"/></svg>
          Save Changes
        </button>
      </div>
    </div>

    <!-- Configuration Management -->
    <div class="card p-5">
      <div class="text-[14px] font-semibold mb-1">Configuration Management</div>
      <div class="text-[11.5px] text-[var(--text-mute)] mb-4">Import, export, atau reset konfigurasi</div>
      <div class="flex gap-3">
        <button class="btn btn-secondary" @click="exportConfig">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
          Export Configuration
        </button>
        <button class="btn btn-secondary" @click="triggerImport">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
          Import Configuration
        </button>
        <button class="btn btn-danger" @click="resetConfig">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="1 4 1 10 7 10"/><path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/></svg>
          Reset to Default
        </button>
      </div>
      <input ref="fileInput" type="file" accept=".json" class="hidden" @change="importConfig">
    </div>

    <!-- Security: Change Password -->
    <div class="card p-5 mt-4">
      <div class="text-[14px] font-semibold mb-1">Security</div>
      <div class="text-[11.5px] text-[var(--text-mute)] mb-4">Ganti password login Admin UI</div>
      <div class="grid grid-cols-2 gap-4">
        <div>
          <label class="text-[12px] text-[var(--text-mute)] mb-1.5 block">Password Saat Ini</label>
          <input
            type="password"
            class="input w-full"
            v-model="pwd.current"
            placeholder="Password yang sekarang"
            autocomplete="current-password"
          />
        </div>
        <div>
          <label class="text-[12px] text-[var(--text-mute)] mb-1.5 block">Password Baru</label>
          <input
            type="password"
            class="input w-full"
            v-model="pwd.newPass"
            placeholder="Min. 6 karakter"
            autocomplete="new-password"
          />
        </div>
        <div>
          <label class="text-[12px] text-[var(--text-mute)] mb-1.5 block">Konfirmasi Password Baru</label>
          <input
            type="password"
            class="input w-full"
            v-model="pwd.confirm"
            placeholder="Ulangi password baru"
            autocomplete="new-password"
          />
        </div>
      </div>

      <!-- Password feedback -->
      <div v-if="pwdError" class="mt-3 p-3 rounded-lg border border-[var(--red)] bg-[rgba(248,113,113,.1)] text-[var(--red)] text-[13px]">
        {{ pwdError }}
      </div>
      <div v-if="pwdSuccess" class="mt-3 p-3 rounded-lg border border-[var(--green)] bg-[rgba(52,211,153,.1)] text-[var(--green)] text-[13px]">
        ✓ {{ pwdSuccess }}
      </div>

      <div class="mt-4 pt-4 border-t border-[var(--border)]">
        <button
          class="btn btn-primary"
          :disabled="!pwd.current || !pwd.newPass || !pwd.confirm || pwdLoading"
          @click="changePassword"
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
          {{ pwdLoading ? 'Menyimpan...' : 'Update Password' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useConfigStore } from "../stores/config";
import { useToast } from "../composables/useToast";
import { api } from "../api/client";

const configStore = useConfigStore();
const { showToast } = useToast();
const dirty = ref(false);
const fileInput = ref<HTMLInputElement>();

// Password change state
const pwd = ref({ current: "", newPass: "", confirm: "" });
const pwdError = ref<string | null>(null);
const pwdSuccess = ref<string | null>(null);
const pwdLoading = ref(false);

const server = ref({ host: "127.0.0.1", port: 20127 });
const downstream = ref({ base_url: "http://127.0.0.1:20128/v1", timeout_seconds: 120 });
const security = ref({ admin_auth_enabled: false, token_configured: false, bind_localhost_only: true, allow_lan_access: false });
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
      logging: { ...configStore.config!.logging, level: debugMode.value ? "debug" : "info" },
    });
    load();
  } catch (e: any) {
    console.error("Failed to save settings:", e.message);
  }
}

async function exportConfig() {
  try {
    const data = await api.exportConfig();
    const blob = new Blob([JSON.stringify(data, null, 2)], { type: "application/json" });
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

function triggerImport() { fileInput.value?.click(); }

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

async function changePassword() {
  pwdError.value = null;
  pwdSuccess.value = null;

  if (pwd.value.newPass.length < 6) {
    pwdError.value = "Password baru minimal 6 karakter.";
    return;
  }
  if (pwd.value.newPass !== pwd.value.confirm) {
    pwdError.value = "Konfirmasi password tidak cocok.";
    return;
  }

  pwdLoading.value = true;
  try {
    await api.changePassword(pwd.value.current, pwd.value.newPass);
    pwdSuccess.value = "Password berhasil diperbarui!";
    showToast("Password berhasil diperbarui", "success");
    pwd.value = { current: "", newPass: "", confirm: "" };
  } catch (e: any) {
    pwdError.value = e.message === "current password is incorrect"
      ? "Password saat ini salah."
      : (e.message || "Gagal memperbarui password.");
  } finally {
    pwdLoading.value = false;
  }
}

onMounted(load);
</script>
