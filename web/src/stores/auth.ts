import { defineStore } from "pinia";
import { ref } from "vue";

const STORAGE_KEY = "atlasbridge_admin_token";

export const useAuthStore = defineStore("auth", () => {
  const token = ref<string | null>(sessionStorage.getItem(STORAGE_KEY));
  const authRequired = ref(false);

  function setToken(t: string) {
    token.value = t;
    authRequired.value = true;
    try {
      sessionStorage.setItem(STORAGE_KEY, t);
    } catch {
      // storage unavailable
    }
  }

  function clearToken() {
    token.value = null;
    authRequired.value = false;
    try {
      sessionStorage.removeItem(STORAGE_KEY);
    } catch {
      // storage unavailable
    }
  }

  return { token, authRequired, setToken, clearToken };
});
