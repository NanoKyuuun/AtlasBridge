<template>
  <button
    type="button"
    :disabled="disabled"
    class="btn btn-secondary inline-flex items-center justify-center gap-2 px-4 py-2.5 text-sm font-medium disabled:cursor-not-allowed disabled:opacity-50"
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
