import { defineStore } from "pinia";
import { ref } from "vue";
import {
  api,
  AuthError,
  type AppConfig,
  type RoutesConfig,
  type ProfilesConfig,
} from "../api/client";

export const useConfigStore = defineStore("config", () => {
  const config = ref<AppConfig | null>(null);
  const routes = ref<RoutesConfig | null>(null);
  const profiles = ref<ProfilesConfig | null>(null);
  const loading = ref(false);
  const error = ref<string | null>(null);
  const saveMessage = ref<string | null>(null);

  async function fetchAll() {
    loading.value = true;
    error.value = null;
    try {
      const [c, r, p] = await Promise.all([
        api.getConfig(),
        api.getRoutes(),
        api.getProfiles(),
      ]);
      config.value = c;
      routes.value = r;
      profiles.value = p;
    } catch (e: any) {
      if (!(e instanceof AuthError)) {
        error.value = e.message;
      }
    } finally {
      loading.value = false;
    }
  }

  const generatedToken = ref<string | null>(null);

  async function saveConfig(cfg: Partial<AppConfig>) {
    try {
      const res = await api.updateConfig(cfg);
      generatedToken.value = res.admin_token ?? null;
      if (generatedToken.value) {
        const { useAuthStore } = await import("../stores/auth");
        useAuthStore().setToken(generatedToken.value);
      }
      saveMessage.value = "Config saved";
      setTimeout(() => (saveMessage.value = null), 3000);
      await fetchAll();
    } catch (e: any) {
      if (!(e instanceof AuthError)) {
        error.value = e.message;
      }
    }
  }

  function dismissToken() {
    generatedToken.value = null;
  }

  async function saveRoutes(r: RoutesConfig) {
    try {
      await api.updateRoutes(r);
      saveMessage.value = "Routes saved";
      setTimeout(() => (saveMessage.value = null), 3000);
      routes.value = r;
    } catch (e: any) {
      if (!(e instanceof AuthError)) {
        error.value = e.message;
      }
    }
  }

  async function saveProfiles(p: ProfilesConfig) {
    try {
      await api.updateProfiles(p);
      saveMessage.value = "Profiles saved";
      setTimeout(() => (saveMessage.value = null), 3000);
      profiles.value = p;
    } catch (e: any) {
      if (!(e instanceof AuthError)) {
        error.value = e.message;
      }
    }
  }

  async function resetConfig() {
    try {
      await api.resetConfig();
      saveMessage.value = "Config reset to defaults";
      setTimeout(() => (saveMessage.value = null), 3000);
      await fetchAll();
    } catch (e: any) {
      if (!(e instanceof AuthError)) {
        error.value = e.message;
      }
    }
  }

  return {
    config,
    routes,
    profiles,
    loading,
    error,
    saveMessage,
    generatedToken,
    fetchAll,
    saveConfig,
    saveRoutes,
    saveProfiles,
    resetConfig,
    dismissToken,
  };
});
