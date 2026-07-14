import { createRouter, createWebHistory } from "vue-router";

const router = createRouter({
  history: createWebHistory("/admin"),
  routes: [
    {
      path: "/",
      name: "dashboard",
      component: () => import("../pages/Dashboard.vue"),
    },
    {
      path: "/login",
      name: "login",
      component: () => import("../pages/Login.vue"),
    },
    {
      path: "/setup",
      name: "setup",
      component: () => import("../pages/SetupWizard.vue"),
    },
    {
      path: "/routing",
      name: "routing",
      component: () => import("../pages/RoutingSettings.vue"),
    },
    {
      path: "/profiles",
      name: "profiles",
      component: () => import("../pages/RouteProfiles.vue"),
    },
    {
      path: "/runtime",
      name: "runtime",
      component: () => import("../pages/Runtime.vue"),
    },
    {
      path: "/startup",
      name: "startup",
      component: () => import("../pages/StartupSettings.vue"),
    },
    {
      path: "/downstream",
      name: "downstream",
      component: () => import("../pages/DownstreamSettings.vue"),
    },
    {
      path: "/logs",
      name: "logs",
      component: () => import("../pages/PrivacySettings.vue"),
    },
    {
      path: "/privacy",
      name: "privacy",
      component: () => import("../pages/PrivacySettings.vue"),
    },
    {
      path: "/advanced",
      name: "advanced",
      component: () => import("../pages/AdvancedSettings.vue"),
    },
    {
      path: "/observability",
      name: "observability",
      component: () => import("../pages/Observability.vue"),
    },
  ],
});

router.beforeEach(async (to) => {
  if (to.name === "login" || to.name === "setup") return true;

  const { useAuthStore } = await import("../stores/auth");
  const auth = useAuthStore();

  // Redirect to login if no token (either first visit or session expired)
  if (!auth.token) {
    return { name: "login" };
  }

  return true;
});

export default router;
