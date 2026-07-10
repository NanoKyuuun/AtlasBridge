<template>
  <div class="min-h-screen relative overflow-hidden text-slate-100">
    <div class="app-bg-noise"></div>
    <div class="absolute inset-0 pointer-events-none bg-[radial-gradient(circle_at_top_left,rgba(47,128,255,0.12),transparent_30%),radial-gradient(circle_at_top_right,rgba(124,58,237,0.10),transparent_28%),radial-gradient(circle_at_bottom_center,rgba(53,215,242,0.08),transparent_32%)]"></div>

    <div class="relative z-10 flex min-h-screen">
      <aside
        class="hidden lg:flex w-[260px] shrink-0 flex-col border-r border-white/10 glass-card bg-[rgba(6,10,20,0.82)]"
      >
        <div class="px-5 pt-6 pb-5 border-b border-white/10">
          <div class="flex items-center gap-3">
            <div
              class="flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-br from-blue-500 via-violet-500 to-cyan-400 shadow-[0_0_24px_rgba(59,130,246,0.35)]"
              aria-hidden="true"
            >
              <span class="text-xl font-black text-white">A</span>
            </div>
            <div class="min-w-0">
              <h1 class="text-lg font-semibold tracking-wide text-white">
                AtlasBridge
              </h1>
              <p class="text-[11px] uppercase tracking-[0.24em] text-slate-400">
                Intelligent Routing. Maximum Potential.
              </p>
            </div>
          </div>

          <div class="mt-5 rounded-2xl border border-white/10 bg-white/5 px-4 py-3">
            <div class="flex items-center gap-2">
              <span class="neon-dot"></span>
              <span class="text-sm font-medium text-slate-100">{{ statusText }}</span>
            </div>
            <p class="mt-1 text-xs text-slate-400">
              {{ status?.mode || "unknown" }} · v{{ version }}
            </p>
          </div>
        </div>

        <nav class="flex-1 px-3 py-4 overflow-y-auto">
          <div class="mb-3 px-3 text-[11px] font-medium uppercase tracking-[0.22em] text-slate-500">
            Navigation
          </div>
          <router-link
            v-for="item in navItems"
            :key="item.path"
            :to="item.path"
            class="group mb-1 flex items-center gap-3 rounded-2xl px-3 py-3 text-sm transition-all duration-200"
            :class="isActive(item.path) ? 'bg-gradient-to-r from-blue-500/20 via-violet-500/18 to-cyan-400/12 text-white shadow-[0_0_24px_rgba(47,128,255,0.14)] border border-blue-400/20' : 'text-slate-300 hover:bg-white/5 hover:text-white border border-transparent'"
          >
            <span
              class="flex h-9 w-9 items-center justify-center rounded-xl bg-white/5 text-base transition-all duration-200 group-hover:bg-white/10"
              :class="isActive(item.path) ? 'bg-white/10 shadow-[0_0_18px_rgba(53,215,242,0.16)]' : ''"
            >
              {{ item.icon }}
            </span>
            <span class="font-medium">{{ item.label }}</span>
          </router-link>
        </nav>

        <div class="px-5 pb-5 pt-3 border-t border-white/10">
          <div class="rounded-2xl border border-cyan-400/15 bg-white/5 px-4 py-3">
            <div class="flex items-center justify-between gap-3">
              <div>
                <p class="text-[11px] uppercase tracking-[0.22em] text-slate-500">
                  Connection
                </p>
                <p class="mt-1 text-sm text-slate-100">{{ currentPageTitle }}</p>
              </div>
              <div class="badge badge-sm border-0 bg-white/10 text-slate-100" :class="statusBadgeClass">
                {{ statusText }}
              </div>
            </div>
          </div>
        </div>
      </aside>

      <div class="flex min-w-0 flex-1 flex-col">
        <header class="sticky top-0 z-20 border-b border-white/10 bg-[rgba(5,7,13,0.72)] backdrop-blur-xl">
          <div class="flex items-center justify-between gap-4 px-4 py-4 sm:px-6 lg:px-8">
            <div class="flex min-w-0 items-center gap-3">
              <button
                class="btn btn-ghost btn-sm lg:hidden border border-white/10 bg-white/5 text-slate-100 hover:bg-white/10"
                type="button"
                @click="mobileNavOpen = !mobileNavOpen"
                aria-label="Open navigation menu"
              >
                ☰
              </button>
              <div class="min-w-0">
                <h2 class="truncate text-lg font-semibold text-white sm:text-xl">
                  {{ currentPageTitle }}
                </h2>
                <p class="truncate text-xs text-slate-400 sm:text-sm">
                  AtlasBridge control surface
                </p>
              </div>
            </div>

            <div class="flex items-center gap-2 sm:gap-3">
              <span
                v-if="saveMessage"
                class="hidden rounded-full border border-emerald-400/20 bg-emerald-400/10 px-3 py-1 text-xs text-emerald-200 sm:inline-flex"
              >
                {{ saveMessage }}
              </span>
              <span
                v-if="error"
                class="hidden rounded-full border border-rose-400/20 bg-rose-400/10 px-3 py-1 text-xs text-rose-200 sm:inline-flex"
              >
                {{ error }}
              </span>
              <div class="hidden md:flex items-center gap-2 rounded-full border border-white/10 bg-white/5 px-3 py-2">
                <span class="neon-dot scale-75"></span>
                <span class="text-xs text-slate-300">{{ status?.mode || 'unknown' }}</span>
              </div>
            </div>
          </div>
        </header>

        <div class="flex-1 p-4 sm:p-6 lg:p-8 xl:p-10">
          <div class="mx-auto max-w-[1600px]">
            <div class="relative overflow-hidden rounded-[1.75rem] border border-white/10 bg-[rgba(8,12,22,0.38)] p-4 sm:p-6 lg:p-8 shadow-[0_24px_80px_rgba(0,0,0,0.38)] backdrop-blur-xl">
              <div class="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_top_right,rgba(53,215,242,0.08),transparent_28%),radial-gradient(circle_at_bottom_left,rgba(124,58,237,0.08),transparent_30%)]"></div>
              <div class="relative z-10">
                <slot />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="mobileNavOpen" class="fixed inset-0 z-30 lg:hidden">
      <button
        class="absolute inset-0 bg-black/60 backdrop-blur-sm"
        type="button"
        aria-label="Close navigation menu"
        @click="mobileNavOpen = false"
      ></button>
      <aside
        class="absolute left-0 top-0 flex h-full w-[min(86vw,320px)] flex-col border-r border-white/10 glass-card bg-[rgba(6,10,20,0.96)]"
      >
        <div class="px-5 pt-6 pb-5 border-b border-white/10">
          <div class="flex items-center gap-3">
            <div
              class="flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-br from-blue-500 via-violet-500 to-cyan-400 shadow-[0_0_24px_rgba(59,130,246,0.35)]"
              aria-hidden="true"
            >
              <span class="text-xl font-black text-white">A</span>
            </div>
            <div>
              <h1 class="text-lg font-semibold text-white">AtlasBridge</h1>
              <p class="text-[11px] uppercase tracking-[0.24em] text-slate-400">
                Intelligent Routing. Maximum Potential.
              </p>
            </div>
          </div>
        </div>

        <nav class="flex-1 px-3 py-4 overflow-y-auto">
          <router-link
            v-for="item in navItems"
            :key="item.path"
            :to="item.path"
            class="group mb-1 flex items-center gap-3 rounded-2xl px-3 py-3 text-sm transition-all duration-200"
            :class="isActive(item.path) ? 'bg-gradient-to-r from-blue-500/20 via-violet-500/18 to-cyan-400/12 text-white border border-blue-400/20' : 'text-slate-300 hover:bg-white/5 hover:text-white border border-transparent'"
            @click="mobileNavOpen = false"
          >
            <span class="flex h-9 w-9 items-center justify-center rounded-xl bg-white/5 text-base">
              {{ item.icon }}
            </span>
            <span class="font-medium">{{ item.label }}</span>
          </router-link>
        </nav>
      </aside>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useAppStore } from "../stores/app";
import { useConfigStore } from "../stores/config";
import { useAuthStore } from "../stores/auth";

const route = useRoute();
const router = useRouter();
const appStore = useAppStore();
const configStore = useConfigStore();
const authStore = useAuthStore();
const mobileNavOpen = ref(false);

const navItems = [
  { path: "/", label: "Dashboard", icon: "📊" },
  { path: "/routing", label: "Routing Settings", icon: "🔀" },
  { path: "/profiles", label: "Route Profiles", icon: "📋" },
  { path: "/runtime", label: "Runtime", icon: "⚙️" },
  { path: "/startup", label: "Startup", icon: "🚀" },
  { path: "/downstream", label: "9Router", icon: "🔗" },
  { path: "/logs", label: "Logs", icon: "📝" },
  { path: "/privacy", label: "Privacy", icon: "🔒" },
  { path: "/advanced", label: "Advanced", icon: "🛠️" },
];

const version = computed(() => appStore.status?.version || "0.1.0");
const status = computed(() => appStore.status);
const saveMessage = computed(() => configStore.saveMessage);
const error = computed(() => configStore.error);

const statusText = computed(() => {
  if (!status.value) return "Loading...";
  return status.value.status || "unknown";
});

const statusBadgeClass = computed(() => {
  const s = statusText.value;
  if (s === "running") return "badge-success";
  if (s === "stopped") return "badge-warning";
  if (s === "error") return "badge-error";
  return "badge-ghost";
});

const currentPageTitle = computed(() => {
  if (route.path === "/") return "AtlasBridge Dashboard";
  const item = navItems.find((i) => i.path === route.path);
  return item?.label || "AtlasBridge";
});

function isActive(path: string) {
  if (path === "/") return route.path === "/";
  return route.path.startsWith(path);
}

onMounted(() => {
  appStore.fetchStatus();
  appStore.fetchDownstreamHealth();
  configStore.fetchAll();
});

watch(
  () => route.path,
  () => {
    mobileNavOpen.value = false;
  },
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
