<template>
  <div class="space-y-2">
    <label v-if="label" :for="forId" class="flex items-center justify-between gap-3 text-sm font-medium text-slate-200">
      <span>{{ label }}</span>
      <span v-if="required" class="text-xs text-cyan-300">Required</span>
    </label>

    <div
      class="rounded-2xl border border-white/10 bg-white/5 px-3 py-2.5 transition-all duration-200 focus-within:border-cyan-400/30 focus-within:bg-white/7 focus-within:shadow-[0_0_0_1px_rgba(53,215,242,0.12),0_0_24px_rgba(53,215,242,0.12)]"
    >
      <slot />
    </div>

    <p v-if="hint && !error" class="text-xs text-slate-500">
      {{ hint }}
    </p>
    <p v-if="error" class="text-xs text-rose-300">
      {{ error }}
    </p>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  label?: string;
  hint?: string;
  error?: string;
  required?: boolean;
  forId?: string;
}>();
</script>
