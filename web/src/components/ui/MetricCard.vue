<template>
  <GlassCard>
    <div class="flex items-center justify-between gap-4">
      <div class="min-w-0">
        <p class="text-xs uppercase tracking-[0.22em] text-slate-500">
          {{ label }}
        </p>
        <div class="mt-2 text-2xl font-semibold tracking-tight text-white">
          {{ value }}
        </div>
      </div>

      <div v-if="delta" class="rounded-full border px-3 py-1 text-xs" :class="deltaClass">
        {{ delta }}
      </div>
    </div>

    <p v-if="description" class="mt-3 text-sm text-slate-400">
      {{ description }}
    </p>
  </GlassCard>
</template>

<script setup lang="ts">
import { computed } from "vue";
import GlassCard from "./GlassCard.vue";

const props = defineProps<{
  label: string;
  value: string | number;
  description?: string;
  delta?: string;
  tone?: "default" | "success" | "warning" | "error";
}>();

const deltaClass = computed(() => {
  switch (props.tone) {
    case "success":
      return "border-emerald-400/20 bg-emerald-400/10 text-emerald-200";
    case "warning":
      return "border-amber-400/20 bg-amber-400/10 text-amber-200";
    case "error":
      return "border-rose-400/20 bg-rose-400/10 text-rose-200";
    default:
      return "border-violet-400/20 bg-violet-400/10 text-violet-200";
  }
});
</script>
