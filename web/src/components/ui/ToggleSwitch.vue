<template>
  <button
    type="button"
    role="switch"
    :aria-checked="modelValue"
    :disabled="disabled"
    class="inline-flex cursor-pointer items-center gap-3 rounded-full outline-none disabled:cursor-not-allowed disabled:opacity-50"
    @click="emit('update:modelValue', !modelValue)"
  >
    <span v-if="label" class="text-sm text-slate-200">{{ label }}</span>
    <span
      class="relative h-7 w-12 rounded-full border border-white/10 bg-slate-800/80 transition-all duration-200 focus-visible:ring-2 focus-visible:ring-cyan-400/50 focus-visible:ring-offset-2 focus-visible:ring-offset-slate-950"
      :class="modelValue ? 'border-cyan-400/30 bg-gradient-to-r from-blue-500 to-violet-500' : ''"
    >
      <span
        class="absolute left-1 top-1 h-5 w-5 rounded-full bg-white shadow-[0_0_14px_rgba(255,255,255,0.22)] transition-all duration-200"
        :class="modelValue ? 'translate-x-5 shadow-[0_0_18px_rgba(53,215,242,0.28)]' : ''"
      ></span>
    </span>
  </button>
</template>

<script setup lang="ts">
defineProps<{
  modelValue: boolean;
  label?: string;
  disabled?: boolean;
}>();

const emit = defineEmits<{
  (event: "update:modelValue", value: boolean): void;
}>();
</script>
