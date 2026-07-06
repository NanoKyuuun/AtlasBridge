import { defineStore } from "pinia";
import { ref } from "vue";
import { api, type StatusResponse, type DownstreamHealth } from "../api/client";

export const useAppStore = defineStore("app", () => {
  const status = ref<StatusResponse | null>(null);
  const downstreamHealth = ref<DownstreamHealth | null>(null);
  const loading = ref(false);
  const error = ref<string | null>(null);

  async function fetchStatus() {
    loading.value = true;
    error.value = null;
    try {
      status.value = await api.getStatus();
    } catch (e: any) {
      error.value = e.message;
    } finally {
      loading.value = false;
    }
  }

  async function fetchDownstreamHealth() {
    try {
      downstreamHealth.value = await api.getDownstreamHealth();
    } catch (e: any) {
      downstreamHealth.value = {
        status: "unavailable",
        url: "",
        message: e.message,
      };
    }
  }

  return {
    status,
    downstreamHealth,
    loading,
    error,
    fetchStatus,
    fetchDownstreamHealth,
  };
});
