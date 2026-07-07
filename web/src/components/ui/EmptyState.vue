<template>
  <GlassCard class="text-center">
    <div class="mx-auto flex h-14 w-14 items-center justify-center rounded-2xl border border-cyan-400/15 bg-white/5 text-cyan-300 shadow-[0_0_24px_rgba(53,215,242,0.1)]">
      <slot name="icon">
        <span class="text-xl">◌</span>
      </slot>
    </div>

    <h3 class="mt-4 text-lg font-semibold text-white">{{ title }}</h3>
    <p v-if="description" class="mx-auto mt-2 max-w-md text-sm leading-6 text-slate-400">
      {{ description }}
    </p>

    <div v-if="$slots.actions" class="mt-5 flex flex-wrap items-center justify-center gap-3">
      <slot name="actions" />
    </div>
  </GlassCard>
</template>

<script setup lang="ts">
import GlassCard from "./GlassCard.vue";

defineProps<{
  title: string;
  description?: string;
}>();
</script>
