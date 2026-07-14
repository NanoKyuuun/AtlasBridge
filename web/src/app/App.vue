<template>
  <div>
    <!-- Login page = no layout wrapper -->
    <router-view v-if="route.name === 'login' || route.name === 'setup'" />
    <!-- All other pages = wrapped in Layout -->
    <Layout v-else>
      <router-view />
    </Layout>
  </div>
</template>

<script setup lang="ts">
import { useRoute } from "vue-router";
import Layout from "../components/Layout.vue";

const route = useRoute();
</script>
