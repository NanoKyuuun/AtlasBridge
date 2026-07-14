<template>
  <div class="space-y-1.5">
    <label v-if="label" :for="forId" class="text-[12px] text-[var(--text-mute)] block">
      {{ label }}
    </label>

    <div class="input">
      <slot />
    </div>

    <p v-if="hint && !error" class="text-[11.5px] text-[var(--text-dim)]">
      {{ hint }}
    </p>
    <p v-if="error" class="text-[11.5px] text-[var(--red)]">
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
