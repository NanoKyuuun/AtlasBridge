<template>
  <section class="space-y-4">
    <header v-if="title || description" class="space-y-1">
      <h2 class="text-lg font-semibold text-white">{{ title }}</h2>
      <p v-if="description" class="text-sm leading-6 text-slate-400">
        {{ description }}
      </p>
    </header>

    <GlassCard :padded="false" class="overflow-hidden">
      <div class="p-5 sm:p-6">
        <slot />
      </div>
    </GlassCard>
  </section>
</template>

<script setup lang="ts">
import GlassCard from "./GlassCard.vue";

defineProps<{
  title?: string;
  description?: string;
}>();
</script>
