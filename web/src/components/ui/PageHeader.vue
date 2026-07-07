<template>
  <header class="mb-6 flex flex-col gap-4 sm:mb-8 lg:flex-row lg:items-end lg:justify-between">
    <div class="max-w-3xl">
      <div v-if="eyebrow" class="mb-2 flex items-center gap-2 text-[11px] uppercase tracking-[0.24em] text-slate-500">
        <span class="neon-dot scale-75"></span>
        <span>{{ eyebrow }}</span>
      </div>
      <h1 class="text-2xl font-semibold tracking-tight text-white sm:text-3xl lg:text-4xl">
        {{ title }}
      </h1>
      <p v-if="description" class="mt-2 max-w-2xl text-sm leading-6 text-slate-400 sm:text-base">
        {{ description }}
      </p>
    </div>

    <div v-if="$slots.actions" class="flex flex-wrap items-center gap-3">
      <slot name="actions" />
    </div>
  </header>
</template>

<script setup lang="ts">
defineProps<{
  title: string;
  description?: string;
  eyebrow?: string;
}>();
</script>
