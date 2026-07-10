<template>
  <div class="mx-auto flex min-h-[60vh] max-w-md items-center justify-center">
    <GlassCard class="w-full">
      <div class="space-y-6">
        <div class="text-center">
          <div
            class="mx-auto flex h-14 w-14 items-center justify-center rounded-2xl bg-gradient-to-br from-blue-500 via-violet-500 to-cyan-400 shadow-[0_0_24px_rgba(59,130,246,0.35)]"
          >
            <span class="text-2xl font-black text-white">A</span>
          </div>
          <h1 class="mt-4 text-xl font-semibold text-white">
            Admin Authentication
          </h1>
          <p class="mt-1 text-sm text-slate-400">
            Enter the admin token shown on first launch to access the dashboard.
          </p>
        </div>

        <form @submit.prevent="login" class="space-y-4">
          <FormField label="Admin Token" forId="token-input">
            <input
              id="token-input"
              ref="tokenInput"
              v-model="tokenValue"
              type="password"
              placeholder="Paste your admin token here"
              class="w-full bg-transparent font-mono text-sm text-white outline-none placeholder:text-slate-500"
              autocomplete="off"
              :disabled="loading"
            />
          </FormField>

          <p v-if="errorMsg" class="rounded-2xl border border-rose-400/20 bg-rose-400/10 px-4 py-3 text-sm text-rose-100">
            {{ errorMsg }}
          </p>

          <GradientButton type="submit" :disabled="!tokenValue || loading" class="w-full">
            {{ loading ? 'Verifying...' : 'Unlock Dashboard' }}
          </GradientButton>
        </form>
      </div>
    </GlassCard>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useRouter } from "vue-router";
import { useAuthStore } from "../stores/auth";
import { api, AuthError } from "../api/client";
import GlassCard from "../components/ui/GlassCard.vue";
import FormField from "../components/ui/FormField.vue";
import GradientButton from "../components/ui/GradientButton.vue";

const router = useRouter();
const auth = useAuthStore();
const tokenValue = ref("");
const loading = ref(false);
const errorMsg = ref<string | null>(null);
const tokenInput = ref<HTMLInputElement | null>(null);

onMounted(() => {
  if (auth.token) {
    auth.clearToken();
  }
  tokenInput.value?.focus();
});

async function login() {
  if (!tokenValue.value) return;
  loading.value = true;
  errorMsg.value = null;

  const testToken = tokenValue.value.trim();
  auth.setToken(testToken);

  try {
    await api.getStatus();
    router.push("/");
  } catch (e: any) {
    auth.clearToken();
    if (e instanceof AuthError) {
      errorMsg.value = "Invalid token. Check and try again.";
    } else {
      errorMsg.value = e.message || "Could not reach the server.";
    }
  } finally {
    loading.value = false;
  }
}
</script>
