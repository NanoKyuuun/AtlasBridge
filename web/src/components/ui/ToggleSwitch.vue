<template>
  <div
    class="toggle"
    :class="{ on: modelValue }"
    @click="$emit('update:modelValue', !modelValue)"
    role="switch"
    :aria-checked="modelValue"
  ></div>
</template>

<script setup lang="ts">
defineProps<{ modelValue: boolean }>();
defineEmits<{ (e: "update:modelValue", v: boolean): void }>();
</script>
