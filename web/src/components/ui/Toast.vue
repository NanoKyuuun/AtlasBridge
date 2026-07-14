<template>
  <teleport to="body">
    <div class="toast-container">
      <transition-group name="toast-anim">
        <div
          v-for="toast in toasts"
          :key="toast.id"
          class="toast"
          :style="{ borderLeft: `3px solid ${colors[toast.type]}` }"
        >
          <span v-html="icons[toast.type]"></span>
          <span>{{ toast.message }}</span>
          <button
            class="ml-auto"
            style="background:transparent; border:none; color:var(--text-mute); cursor:pointer; padding:0 0 0 8px; font-size:16px; line-height:1;"
            @click="removeToast(toast.id)"
          >×</button>
        </div>
      </transition-group>
    </div>
  </teleport>
</template>

<script setup lang="ts">
import { useToast } from "../../composables/useToast";

const { toasts, removeToast } = useToast();

const colors: Record<string, string> = {
  success: "#34d399",
  warning: "#fbbf24",
  error: "#f87171",
  info: "#4f8cff",
};

const icons: Record<string, string> = {
  success: `<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#34d399" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>`,
  warning: `<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#fbbf24" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>`,
  error: `<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#f87171" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/></svg>`,
  info: `<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#4f8cff" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>`,
};
</script>

<style scoped>
.toast-anim-enter-active { animation: slideIn .3s ease; }
.toast-anim-leave-active { animation: slideOut .3s ease forwards; }
</style>
