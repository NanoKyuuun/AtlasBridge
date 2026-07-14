<template>
  <span class="badge" :class="badgeClass">
    <span v-if="status === 'running'" class="status-dot running"></span>
    <span v-else-if="status === 'stopped'" class="status-dot stopped"></span>
    <span v-else-if="status === 'error'" class="status-dot error"></span>
    {{ label || status }}
  </span>
</template>

<script setup lang="ts">
import { computed } from "vue";

const props = defineProps<{
  status?: string;
  label?: string;
}>();

const badgeClass = computed(() => {
  switch (props.status) {
    case "running": return "badge-green";
    case "stopped": return "badge-red";
    case "error": return "badge-red";
    case "disabled": return "badge-yellow";
    case "Active":
    case "active": return "badge-green";
    case "Inactive":
    case "inactive": return "badge-gray";
    default: return "badge-gray";
  }
});
</script>
