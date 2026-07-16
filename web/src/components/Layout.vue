<template>
  <div class="flex min-h-screen">
    <!-- MOBILE OVERLAY -->
    <div
      v-if="sidebarOpen"
      class="sidebar-mobile-overlay fixed inset-0 z-40 bg-black/50 lg:hidden"
      @click="sidebarOpen = false"
    ></div>

    <!-- SIDEBAR -->
    <aside
      class="sidebar-desktop w-[260px] flex-shrink-0 border-r border-[var(--border)] bg-[var(--bg-1)] flex flex-col fixed inset-y-0 left-0 z-50 lg:relative lg:translate-x-0 transition-transform"
      :class="sidebarOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'"
    >
      <!-- Logo -->
      <div class="px-5 py-5 border-b border-[var(--border)] flex items-center gap-3">
        <img src="/atlasbridge-logo-mark-256.png" alt="AtlasBridge" class="w-9 h-9 rounded-xl" />
        <div>
          <div class="font-bold text-[15px]">AtlasBridge</div>
          <div class="text-[11px] text-[var(--text-mute)]">v1.1 · Intelligent Proxy</div>
        </div>
      </div>

      <!-- Navigation -->
      <nav class="flex-1 p-3 overflow-y-auto">
        <div class="section-title">Overview</div>
        <router-link
          to="/"
          class="nav-item"
          :class="{ active: route.path === '/' }"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="7" height="9"/><rect x="14" y="3" width="7" height="5"/><rect x="14" y="12" width="7" height="9"/><rect x="3" y="16" width="7" height="5"/></svg>
          Dashboard
        </router-link>
        <router-link
          to="/routing"
          class="nav-item"
          :class="{ active: route.path.startsWith('/routing') }"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="16 3 21 3 21 8"/><line x1="4" y1="20" x2="21" y2="3"/><polyline points="21 16 21 21 16 21"/><line x1="15" y1="15" x2="20" y2="20"/><line x1="4" y1="4" x2="9" y2="9"/></svg>
          Routing Settings
        </router-link>
        <router-link
          to="/profiles"
          class="nav-item"
          :class="{ active: route.path.startsWith('/profiles') }"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 7h-9"/><path d="M14 17H5"/><circle cx="17" cy="17" r="3"/><circle cx="7" cy="7" r="3"/></svg>
          Route Profiles
        </router-link>

        <div class="section-title mt-5">System</div>
        <router-link
          to="/startup"
          class="nav-item"
          :class="{ active: route.path.startsWith('/startup') }"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18.36 6.64a9 9 0 1 1-12.73 0"/><line x1="12" y1="2" x2="12" y2="12"/></svg>
          Startup
        </router-link>
        <router-link
          to="/runtime"
          class="nav-item"
          :class="{ active: route.path.startsWith('/runtime') }"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="5 3 19 12 5 21 5 3"/></svg>
          Runtime Control
        </router-link>
        <router-link
          to="/logs"
          class="nav-item"
          :class="{ active: route.path.startsWith('/logs') || route.path.startsWith('/privacy') }"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/></svg>
          Privacy &amp; Logs
        </router-link>
        <router-link
          to="/advanced"
          class="nav-item"
          :class="{ active: route.path.startsWith('/advanced') || route.path.startsWith('/downstream') }"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>
          Advanced Settings
        </router-link>

        <div class="section-title mt-5">Insights</div>
        <router-link
          to="/observability"
          class="nav-item"
          :class="{ active: route.path.startsWith('/observability') }"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="20" x2="18" y2="10"/><line x1="12" y1="20" x2="12" y2="4"/><line x1="6" y1="20" x2="6" y2="14"/></svg>
          Observability
        </router-link>
      </nav>

      <!-- Footer / Endpoint Info -->
      <div class="p-3 border-t border-[var(--border)]">
        <div class="card-soft p-3 flex items-center gap-3">
          <div class="w-8 h-8 rounded-full flex items-center justify-center text-xs font-bold" style="background: linear-gradient(135deg, #f59e0b, #ef4444);">AB</div>
          <div class="flex-1 min-w-0">
            <div class="text-[12.5px] font-medium truncate">AtlasBridge</div>
            <div class="text-[11px] text-[var(--text-mute)] truncate">{{ proxyEndpoint }}</div>
          </div>
        </div>
      </div>
    </aside>

    <!-- MAIN -->
    <main class="flex-1 flex flex-col min-w-0">
      <!-- Top Bar / Header -->
      <header class="h-[64px] border-b border-[var(--border)] bg-[var(--bg-1)] flex items-center justify-between px-6 flex-shrink-0">
        <div class="flex items-center gap-4">
          <!-- Mobile hamburger -->
          <button class="mobile-menu-btn btn btn-ghost p-2" @click="sidebarOpen = !sidebarOpen">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="3" y1="12" x2="21" y2="12"/><line x1="3" y1="6" x2="21" y2="6"/><line x1="3" y1="18" x2="21" y2="18"/></svg>
          </button>
          <div>
            <h1 class="text-[16px] font-semibold">{{ currentPageTitle }}</h1>
            <p class="text-[11.5px] text-[var(--text-mute)]">{{ currentPageSubtitle }}</p>
          </div>
        </div>
        <div class="flex items-center gap-3">
          <!-- Global Status Indicator -->
          <div class="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-[var(--bg-2)] border border-[var(--border)]">
            <span class="status-dot" :class="statusClass"></span>
            <span class="text-[12px] font-medium">{{ statusText }}</span>
          </div>
          <!-- Refresh -->
          <button class="btn btn-ghost" @click="refreshAll">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
            Refresh
          </button>
          <!-- Stop / Start Proxy -->
          <button
            class="btn"
            :class="proxyRunning ? 'btn-secondary' : 'btn-success'"
            :disabled="proxyToggling"
            @click="toggleProxy"
          >
            <svg v-if="proxyToggling" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="animate-spin"><path d="M21 12a9 9 0 1 1-6.219-8.56"/></svg>
            <svg v-else-if="proxyRunning" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="6" y="4" width="4" height="16"/><rect x="14" y="4" width="4" height="16"/></svg>
            <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="5 3 19 12 5 21 5 3"/></svg>
            {{ proxyRunning ? 'Stop' : 'Start' }}
          </button>
        </div>
      </header>

      <!-- Page Content -->
      <div class="flex-1 overflow-y-auto p-6">
        <slot />
      </div>
    </main>

    <!-- Toast Notifications -->
    <Toast />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useAppStore } from "../stores/app";
import { useConfigStore } from "../stores/config";
import { useAuthStore } from "../stores/auth";
import { useToast } from "../composables/useToast";
import { api } from "../api/client";
import Toast from "./ui/Toast.vue";

const route = useRoute();
const router = useRouter();
const appStore = useAppStore();
const configStore = useConfigStore();
const authStore = useAuthStore();
const { showToast } = useToast();

const sidebarOpen = ref(false);
const proxyToggling = ref(false);

const status = computed(() => appStore.status);

const proxyRunning = computed(() => {
  const s = status.value?.status;
  return s === "running";
});

const proxyEndpoint = computed(() => {
  if (configStore.config?.server) {
    return `${configStore.config.server.host}:${configStore.config.server.port}`;
  }
  if (status.value?.host && status.value?.port) {
    return `${status.value.host}:${status.value.port}`;
  }
  return "—";
});

const statusText = computed(() => {
  if (!status.value) return "Offline";
  const s = status.value.status || "unknown";
  return s.charAt(0).toUpperCase() + s.slice(1);
});

const statusClass = computed(() => {
  if (!status.value) return "stopped";
  const s = status.value?.status;
  if (s === "running") return "running";
  if (s === "stopped") return "stopped";
  if (s === "error") return "error";
  return "disabled";
});

const pageInfo: Record<string, { title: string; subtitle: string }> = {
  "/": { title: "Dashboard", subtitle: "Monitoring status proxy dan aktivitas routing" },
  "/routing": { title: "Routing Settings", subtitle: "Atur task-to-route mapping dan perilaku routing" },
  "/profiles": { title: "Route Profiles", subtitle: "Kelola abstraksi routing untuk 9Router" },
  "/startup": { title: "Startup", subtitle: "Konfigurasi auto-start dan perilaku startup" },
  "/runtime": { title: "Runtime Control", subtitle: "Kontrol start/stop/restart proxy" },
  "/logs": { title: "Privacy & Logs", subtitle: "Pengaturan privasi dan viewing log routing" },
  "/privacy": { title: "Privacy & Logs", subtitle: "Pengaturan privasi dan viewing log routing" },
  "/advanced": { title: "Advanced Settings", subtitle: "Konfigurasi teknis lanjutan" },
  "/downstream": { title: "Advanced Settings", subtitle: "Konfigurasi teknis lanjutan" },
  "/observability": { title: "Observability", subtitle: "Metrics, analytics, dan full request log" },
};

const currentPageTitle = computed(() => pageInfo[route.path]?.title || "AtlasBridge");
const currentPageSubtitle = computed(() => pageInfo[route.path]?.subtitle || "");

async function toggleProxy() {
  proxyToggling.value = true;
  try {
    if (proxyRunning.value) {
      await api.runtimeStop();
      showToast("Proxy stopped", "warning");
    } else {
      await api.runtimeStart();
      showToast("Proxy started", "success");
    }
    await appStore.fetchStatus();
  } catch (e: any) {
    showToast(`Failed: ${e.message}`, "error");
  } finally {
    proxyToggling.value = false;
  }
}

function refreshAll() {
  appStore.fetchStatus();
  appStore.fetchDownstreamHealth();
  configStore.fetchAll();
  showToast("Status refreshed", "info");
}

onMounted(() => {
  appStore.fetchStatus();
  appStore.fetchDownstreamHealth();
  configStore.fetchAll();
});

watch(
  () => route.path,
  () => { sidebarOpen.value = false; }
);

watch(
  () => authStore.authRequired,
  (required) => {
    if (required && !authStore.token && route.name !== "login") {
      router.replace("/login");
    }
  },
);
</script>
