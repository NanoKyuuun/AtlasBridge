<template>
  <header class="mb-5 flex items-center justify-between">
    <div>
      <div class="text-[14px] font-semibold">{{ title }}</div>
      <div v-if="description" class="text-[11.5px] text-[var(--text-mute)]">{{ description }}</div>
    </div>

    <div v-if="$slots.actions" class="flex flex-wrap items-center gap-2">
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
