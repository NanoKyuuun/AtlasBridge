<template>
  <section :class="cardClass">
    <slot />
  </section>
</template>

<script setup lang="ts">
import { computed } from "vue";

const props = withDefaults(
  defineProps<{
    padded?: boolean;
    rounded?: string;
    class?: string;
  }>(),
  {
    padded: true,
    rounded: "rounded-3xl",
    class: "",
  },
);

const cardClass = computed(() => [
  "glass-card",
  "glow-border",
  props.rounded,
  props.padded ? "p-5 sm:p-6" : "",
  props.class,
].join(" "));
</script>
