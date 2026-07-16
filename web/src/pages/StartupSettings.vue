<template>
  <div>
    <div class="grid grid-cols-2 gap-4 mb-6">
      <!-- Startup Mode -->
      <div class="card p-6">
        <div class="text-[14px] font-semibold mb-1">Startup Mode</div>
        <div class="text-[11.5px] text-[var(--text-mute)] mb-5">Pilih bagaimana AtlasBridge berjalan saat device restart</div>
        <div class="space-y-3">
          <label class="block cursor-pointer">
            <div
              class="card-soft p-4 flex items-start gap-3 border-2 transition-all"
              :class="appMode === 'always_on' ? 'border-[var(--accent)]' : 'border-transparent hover:border-[var(--accent)]'"
              @click="appMode = 'always_on'; dirty = true"
            >
              <input type="radio" name="startup" class="mt-1" :checked="appMode === 'always_on'" style="accent-color: var(--accent);">
              <div class="flex-1">
                <div class="flex items-center gap-2">
                  <div class="font-medium">Always On</div>
                  <span class="badge badge-green">Recommended</span>
                </div>
                <div class="text-[12px] text-[var(--text-dim)] mt-1">Proxy otomatis aktif saat device menyala dan berjalan di background</div>
              </div>
            </div>
          </label>

          <label class="block cursor-pointer">
            <div
              class="card-soft p-4 flex items-start gap-3 border-2 transition-all"
              :class="appMode === 'manual' ? 'border-[var(--accent)]' : 'border-transparent hover:border-[var(--accent)]'"
              @click="appMode = 'manual'; dirty = true"
            >
              <input type="radio" name="startup" class="mt-1" :checked="appMode === 'manual'" style="accent-color: var(--accent);">
              <div class="flex-1">
                <div class="font-medium">Manual</div>
                <div class="text-[12px] text-[var(--text-dim)] mt-1">Proxy hanya aktif ketika Anda jalankan manual dari UI</div>
              </div>
            </div>
          </label>

          <label class="block cursor-pointer">
            <div
              class="card-soft p-4 flex items-start gap-3 border-2 transition-all"
              :class="appMode === 'disabled' ? 'border-[var(--red)]' : 'border-transparent hover:border-[var(--accent)]'"
              @click="appMode = 'disabled'; dirty = true"
            >
              <input type="radio" name="startup" class="mt-1" :checked="appMode === 'disabled'" style="accent-color: var(--accent);">
              <div class="flex-1">
                <div class="flex items-center gap-2">
                  <div class="font-medium">Disabled</div>
                  <span class="badge badge-yellow">Not recommended</span>
                </div>
                <div class="text-[12px] text-[var(--text-dim)] mt-1">Proxy tidak aktif sampai Anda mengaktifkannya kembali</div>
              </div>
            </div>
          </label>
        </div>
      </div>

      <!-- Runtime Options -->
      <div class="card p-6">
        <div class="text-[14px] font-semibold mb-1">Runtime Options</div>
        <div class="text-[11.5px] text-[var(--text-mute)] mb-5">Opsi tambahan untuk perilaku runtime</div>
        <div class="space-y-4">
          <div class="flex items-center justify-between">
            <div>
              <div class="text-[13px] font-medium">Auto-start on boot</div>
              <div class="text-[11.5px] text-[var(--text-mute)]">Daftarkan ke startup manager OS</div>
            </div>
            <div
              class="toggle"
              :class="{ on: startup.run_at_login }"
              @click="startup.run_at_login = !startup.run_at_login; dirty = true"
            ></div>
          </div>

          <div class="flex items-center justify-between">
            <div>
              <div class="text-[13px] font-medium">Restart after crash</div>
              <div class="text-[11.5px] text-[var(--text-mute)]">Otomatis restart jika proxy crash</div>
            </div>
            <div
              class="toggle"
              :class="{ on: startup.restart_after_crash }"
              @click="startup.restart_after_crash = !startup.restart_after_crash; dirty = true"
            ></div>
          </div>

          <div class="flex items-center justify-between">
            <div>
              <div class="text-[13px] font-medium">Start proxy on launch</div>
              <div class="text-[11.5px] text-[var(--text-mute)]">Mulai proxy otomatis saat aplikasi dibuka</div>
            </div>
            <div
              class="toggle"
              :class="{ on: startup.start_proxy_on_app_launch }"
              @click="startup.start_proxy_on_app_launch = !startup.start_proxy_on_app_launch; dirty = true"
            ></div>
          </div>
        </div>

        <div class="divider"></div>
        <div class="flex gap-2">
          <button class="btn btn-primary flex-1 justify-center" :disabled="!dirty" @click="save">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"/><polyline points="17 21 17 13 7 13 7 21"/><polyline points="7 3 7 8 15 8"/></svg>
            Save Changes
          </button>
          <button class="btn btn-ghost" :disabled="!dirty" @click="load">Discard</button>
        </div>
      </div>
    </div>

    <!-- Runtime History -->
    <div class="card">
      <div class="p-5 border-b border-[var(--border)]">
        <div class="text-[14px] font-semibold">Runtime History</div>
        <div class="text-[11.5px] text-[var(--text-mute)]">Log aktivitas start/stop/restart</div>
      </div>
      <div class="log-row" style="padding: 10px 20px; background: var(--bg-2); color: var(--text-mute); font-weight: 600; font-size: 11px; text-transform: uppercase; letter-spacing: .05em;">
        <div>Timestamp</div>
        <div>Event</div>
        <div>Mode</div>
        <div>Duration</div>
        <div>Status</div>
        <div></div>
      </div>
      <div v-if="runtimeHistory.length === 0" class="p-8 text-center text-[var(--text-mute)] text-[13px]">
        Belum ada event runtime
      </div>
      <div v-for="entry in runtimeHistory" :key="entry.ts" class="log-row">
        <div class="text-[var(--text-dim)]">{{ entry.ts }}</div>
        <div><span class="badge" :class="entry.eventBadge">{{ entry.event }}</span></div>
        <div><span class="code-tag">{{ entry.mode }}</span></div>
        <div class="mono">{{ entry.duration }}</div>
        <div><span :style="{ color: 'var(--green)' }">{{ entry.status }}</span></div>
        <div></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useConfigStore } from "../stores/config";

const configStore = useConfigStore();
const dirty = ref(false);

const startup = ref({
  run_at_login: false,
  start_proxy_on_app_launch: true,
  restart_after_crash: true,
});
const appMode = ref("manual");

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

const runtimeHistory = ref<Array<{ ts: string; event: string; eventBadge: string; mode: string; duration: string; status: string }>>([]);

onMounted(load);
</script>
