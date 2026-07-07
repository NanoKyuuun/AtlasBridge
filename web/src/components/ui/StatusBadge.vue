<template>
  <span class="inline-flex items-center gap-2 rounded-full border px-3 py-1 text-xs font-medium capitalize" :class="badgeClass">
    <span class="h-1.5 w-1.5 rounded-full" :class="dotClass"></span>
    {{ label }}
  </span>
</template>

<script setup lang="ts">
import { computed } from "vue";

const props = defineProps<{
  status: "running" | "stopped" | "error" | "disabled" | "active" | "inactive" | "warning" | string;
  label?: string;
}>();

const badgeClass = computed(() => {
  switch (props.status) {
    case "running":
    case "active":
      return "border-emerald-400/20 bg-emerald-400/10 text-emerald-200";
    case "stopped":
    case "disabled":
    case "inactive":
      return "border-slate-500/20 bg-slate-500/10 text-slate-300";
    case "error":
      return "border-rose-400/20 bg-rose-400/10 text-rose-200";
    case "warning":
      return "border-amber-400/20 bg-amber-400/10 text-amber-200";
    default:
      return "border-blue-400/20 bg-blue-400/10 text-blue-200";
  }
});

const dotClass = computed(() => {
  switch (props.status) {
    case "running":
    case "active":
      return "bg-emerald-400 shadow-[0_0_10px_rgba(34,197,94,0.55)]";
    case "stopped":
    case "disabled":
    case "inactive":
      return "bg-slate-400 shadow-[0_0_10px_rgba(148,163,184,0.3)]";
    case "error":
      return "bg-rose-400 shadow-[0_0_10px_rgba(239,68,68,0.55)]";
    case "warning":
      return "bg-amber-400 shadow-[0_0_10px_rgba(245,158,11,0.55)]";
    default:
      return "bg-cyan-400 shadow-[0_0_10px_rgba(53,215,242,0.55)]";
  }
});

const label = computed(() => props.label || props.status);
</script>
