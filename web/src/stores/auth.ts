import { defineStore } from "pinia";
import { ref } from "vue";

const TOKEN_KEY = "atlasbridge_token";

export const useAuthStore = defineStore("auth", () => {
  const token = ref<string | null>(sessionStorage.getItem(TOKEN_KEY));
  const authRequired = ref(false);

  function setToken(t: string) {
    token.value = t;
    authRequired.value = false;
    sessionStorage.setItem(TOKEN_KEY, t);
  }

  function clearToken() {
    token.value = null;
    authRequired.value = false;
    sessionStorage.removeItem(TOKEN_KEY);
  }

  return { token, authRequired, setToken, clearToken };
});
