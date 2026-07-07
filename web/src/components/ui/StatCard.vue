<template>
  <GlassCard>
    <div class="flex items-start justify-between gap-4">
      <div class="min-w-0">
        <p class="text-xs uppercase tracking-[0.22em] text-slate-500">
          {{ label }}
        </p>
        <div class="mt-2 flex items-end gap-3">
          <div class="text-2xl font-semibold tracking-tight text-white">
            {{ value }}
          </div>
          <span v-if="trend" class="rounded-full border px-2.5 py-1 text-xs" :class="trendClass">
            {{ trend }}
          </span>
        </div>
        <p v-if="description" class="mt-2 text-sm text-slate-400">
          {{ description }}
        </p>
      </div>

      <div v-if="$slots.icon" class="flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl border border-white/10 bg-white/5 text-cyan-300 shadow-[0_0_18px_rgba(53,215,242,0.12)]">
        <slot name="icon" />
      </div>
    </div>
  </GlassCard>
</template>

<script setup lang="ts">
import { computed } from "vue";
import GlassCard from "./GlassCard.vue";

const props = defineProps<{
  label: string;
  value: string | number;
  description?: string;
  trend?: string;
  tone?: "default" | "success" | "warning" | "error";
}>();

const trendClass = computed(() => {
  switch (props.tone) {
    case "success":
      return "border-emerald-400/20 bg-emerald-400/10 text-emerald-200";
    case "warning":
      return "border-amber-400/20 bg-amber-400/10 text-amber-200";
    case "error":
      return "border-rose-400/20 bg-rose-400/10 text-rose-200";
    default:
      return "border-blue-400/20 bg-blue-400/10 text-blue-200";
  }
});
</script>
