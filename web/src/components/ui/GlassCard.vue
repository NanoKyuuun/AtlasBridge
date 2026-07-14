<template>
  <div class="card" :class="$attrs.class" :style="$attrs.style">
    <slot />
  </div>
</template>

<script setup lang="ts">
defineOptions({ inheritAttrs: false });
</script>
