<template>
  <button
    type="button"
    :disabled="disabled"
    class="inline-flex items-center justify-center gap-2 rounded-2xl border border-white/10 bg-white/5 px-4 py-2.5 text-sm font-medium text-slate-200 transition-all duration-200 hover:border-cyan-400/25 hover:bg-white/8 hover:text-white active:scale-[0.99] disabled:cursor-not-allowed disabled:opacity-50"
    v-bind="$attrs"
  >
    <slot />
  </button>
</template>

<script setup lang="ts">
defineProps<{
  disabled?: boolean;
}>();
</script>
