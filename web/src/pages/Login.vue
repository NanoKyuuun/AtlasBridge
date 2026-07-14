<template>
  <div class="flex items-center justify-center min-h-screen bg-[var(--bg-0)]">
    <div class="card w-full max-w-[400px] p-8 mx-4">
      <!-- Logo + Title -->
      <div class="text-center mb-8">
        <div class="inline-flex items-center justify-center w-14 h-14 rounded-2xl bg-gradient-to-br from-[var(--purple)] to-[var(--blue)] shadow-lg mb-4">
          <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
          </svg>
        </div>
        <h1 class="text-[20px] font-bold mb-1">AtlasBridge</h1>
        <p class="text-[12.5px] text-[var(--text-mute)]">Masukkan password untuk mengakses Admin UI</p>
      </div>

      <!-- Form -->
      <form @submit.prevent="login" class="space-y-4">
        <div>
          <label class="text-[12px] text-[var(--text-mute)] mb-1.5 block" for="password-input">
            Admin Password
          </label>
          <div class="relative">
            <input
              id="password-input"
              ref="passwordInput"
              v-model="password"
              :type="showPassword ? 'text' : 'password'"
              placeholder="Masukkan password admin"
              class="input w-full pr-10"
              autocomplete="current-password"
              :disabled="loading"
            />
            <button
              type="button"
              class="absolute right-3 top-1/2 -translate-y-1/2 text-[var(--text-mute)] hover:text-[var(--text)] transition-colors"
              @click="showPassword = !showPassword"
            >
              <svg v-if="!showPassword" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>
              <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"/><line x1="1" y1="1" x2="23" y2="23"/></svg>
            </button>
          </div>
        </div>

        <!-- Error Message -->
        <div v-if="errorMsg" class="p-3 rounded-lg border border-[var(--red)] bg-[rgba(248,113,113,.1)] text-[var(--red)] text-[13px]">
          {{ errorMsg }}
        </div>

        <!-- Submit Button -->
        <button
          type="submit"
          class="btn btn-primary w-full"
          :disabled="!password || loading"
        >
          <svg v-if="loading" class="animate-spin" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12a9 9 0 1 1-6.219-8.56"/></svg>
          <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4"/><polyline points="10 17 15 12 10 7"/><line x1="15" y1="12" x2="3" y2="12"/></svg>
          {{ loading ? 'Memverifikasi...' : 'Masuk' }}
        </button>
      </form>

      <!-- Info footer -->
      <p class="text-center text-[11.5px] text-[var(--text-dim)] mt-6">
        Password diatur pertama kali di Advanced Settings setelah login perdana.
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useRouter } from "vue-router";
import { useAuthStore } from "../stores/auth";
import { loginWithPassword } from "../api/client";

const router = useRouter();
const auth = useAuthStore();

const password = ref("");
const showPassword = ref(false);
const loading = ref(false);
const errorMsg = ref<string | null>(null);
const passwordInput = ref<HTMLInputElement | null>(null);

onMounted(() => {
  if (auth.token) auth.clearToken();
  passwordInput.value?.focus();
});

async function login() {
  if (!password.value || loading.value) return;
  loading.value = true;
  errorMsg.value = null;

  try {
    const { token } = await loginWithPassword(password.value.trim());
    auth.setToken(token);
    router.push("/");
  } catch (e: any) {
    auth.clearToken();
    errorMsg.value = e.message === "invalid password"
      ? "Password salah. Silakan coba lagi."
      : (e.message || "Tidak dapat terhubung ke server.");
  } finally {
    loading.value = false;
  }
}
</script>
