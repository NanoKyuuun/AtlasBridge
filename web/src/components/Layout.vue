<template>
  <div class="min-h-screen bg-base-200 flex">
    <aside class="w-64 bg-base-100 shadow-md flex flex-col">
      <div class="p-4 border-b border-base-200">
        <div class="flex items-center gap-3">
          <img
            src="/atlasbridge-logo-mark-256.png"
            alt="AtlasBridge logo"
            class="h-10 w-10 rounded-xl bg-base-200 p-1"
          />
          <div>
            <h1 class="text-lg font-bold">AtlasBridge</h1>
            <p class="text-xs text-base-content/50">AI Proxy | v{{ version }}</p>
          </div>
        </div>
      </div>
      <nav class="flex-1 p-2">
        <router-link
          v-for="item in navItems"
          :key="item.path"
          :to="item.path"
          class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm hover:bg-base-200 transition-colors mb-1"
          :class="{
            'bg-primary/10 text-primary font-medium': isActive(item.path),
          }"
        >
          <span class="text-lg">{{ item.icon }}</span>
          <span>{{ item.label }}</span>
        </router-link>
      </nav>
      <div class="p-4 border-t border-base-200">
        <div class="flex items-center gap-2">
          <div class="badge badge-sm" :class="statusBadgeClass">
            {{ statusText }}
          </div>
          <span class="text-xs text-base-content/50">{{
            status?.mode || "unknown"
          }}</span>
        </div>
      </div>
    </aside>
    <main class="flex-1 overflow-auto">
      <div class="navbar bg-base-100 shadow-sm px-6">
        <div class="flex-1">
          <h2 class="text-lg font-semibold">{{ currentPageTitle }}</h2>
        </div>
        <div class="navbar-end gap-2">
          <span
            v-if="saveMessage"
            class="alert alert-success alert-sm py-1 text-xs"
            >{{ saveMessage }}</span
          >
          <span v-if="error" class="alert alert-error alert-sm py-1 text-xs">{{
            error
          }}</span>
        </div>
      </div>
      <div class="p-6">
        <slot />
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from "vue";
import { useRoute } from "vue-router";
import { useAppStore } from "../stores/app";
import { useConfigStore } from "../stores/config";

const route = useRoute();
const appStore = useAppStore();
const configStore = useConfigStore();

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
</script>
