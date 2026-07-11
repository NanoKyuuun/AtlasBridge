import { test as base, type Page } from "@playwright/test";

export const MOCK_STATUS = {
  status: "running",
  version: "0.1.0",
  uptime: "2h 30m",
  port: 20127,
  host: "127.0.0.1",
  downstream: "http://127.0.0.1:20128/v1",
  mode: "always_on",
  privacy: "standard",
  go_version: "go1.25.5",
  pid: 12345,
};

export const MOCK_CONFIG = {
  app: { name: "AtlasBridge", mode: "always_on", first_run_completed: true },
  server: { host: "127.0.0.1", port: 20127, admin_path: "/admin" },
  downstream: { base_url: "http://127.0.0.1:20128/v1", timeout_seconds: 120 },
  security: {
    admin_auth_enabled: true,
    admin_token_hash: "abc123",
    bind_localhost_only: true,
    allow_lan_access: false,
  },
  startup: { run_at_login: false, start_proxy_on_app_launch: false, restart_after_crash: true },
  routing: {
    auto_routing: true,
    default_route: "route.default",
    low_confidence_route: "route.default",
    confidence_threshold: 0.55,
    smart_fast_route: "route.low_cost",
    metadata_transport: "header",
  },
  logging: {
    level: "info",
    privacy_mode: "standard",
    metadata_logging_enabled: true,
    prompt_logging_enabled: false,
    retention_days: 7,
  },
};

export const MOCK_ROUTES = {
  task_routes: {
    general_chat: "route.default",
    design_task: "route.design",
    backend_engineering: "route.backend",
    frontend_engineering: "route.frontend",
    debugging: "route.debugging",
    refactoring: "route.refactoring",
    test_generation: "route.testing",
    documentation: "route.documentation",
    architecture_design: "route.architect",
    security_review: "route.security",
    long_context_analysis: "route.long_context",
    lightweight_task: "route.low_cost",
    unknown: "route.default",
  },
};

export const MOCK_PROFILES = {
  route_profiles: {
    "route.default": { label: "Default", description: "Standard routing", downstream_alias: "default", priority: "normal", enabled: true },
    "route.backend": { label: "Backend", description: "Backend engineering", downstream_alias: "backend", priority: "normal", enabled: true },
    "route.frontend": { label: "Frontend", description: "Frontend engineering", downstream_alias: "frontend", priority: "normal", enabled: true },
    "route.debugging": { label: "Debugging", description: "Debug tasks", downstream_alias: "debug", priority: "high", enabled: true },
    "route.design": { label: "Design", description: "Design tasks", downstream_alias: "design", priority: "normal", enabled: true },
    "route.refactoring": { label: "Refactoring", description: "Code refactoring", downstream_alias: "refactor", priority: "normal", enabled: true },
    "route.testing": { label: "Testing", description: "Test generation", downstream_alias: "testing", priority: "normal", enabled: true },
    "route.documentation": { label: "Documentation", description: "Documentation", downstream_alias: "docs", priority: "low", enabled: true },
    "route.architect": { label: "Architecture", description: "Architecture design", downstream_alias: "architect", priority: "high", enabled: true },
    "route.security": { label: "Security", description: "Security review", downstream_alias: "security", priority: "high", enabled: true },
    "route.long_context": { label: "Long Context", description: "Long context analysis", downstream_alias: "long_context", priority: "normal", enabled: true },
    "route.low_cost": { label: "Low Cost", description: "Cost-optimized", downstream_alias: "low_cost", priority: "low", enabled: true },
  },
};

export const MOCK_DOWNSTREAM_HEALTH = {
  status: "connected",
  status_code: 200,
  url: "http://127.0.0.1:20128/v1",
  message: "Downstream is healthy",
};

export const MOCK_COMBO_RESULT = {
  model: "combo.default",
  resolved_model: "gpt-4o",
  success: true,
  status_code: 200,
  latency: 245,
};

type MockAPIOptions = {
  status?: typeof MOCK_STATUS;
  config?: typeof MOCK_CONFIG;
  routes?: typeof MOCK_ROUTES;
  profiles?: typeof MOCK_PROFILES;
  downstreamHealth?: typeof MOCK_DOWNSTREAM_HEALTH;
  authRequired?: boolean;
};

export async function mockAllAPIs(page: Page, opts: MockAPIOptions = {}) {
  const {
    status = MOCK_STATUS,
    config = MOCK_CONFIG,
    routes = MOCK_ROUTES,
    profiles = MOCK_PROFILES,
    downstreamHealth = MOCK_DOWNSTREAM_HEALTH,
    authRequired = false,
  } = opts;

  await page.route("**/admin/api/status", (route) => {
    if (authRequired && !route.request().headers()["authorization"]) {
      return route.fulfill({ status: 401, contentType: "application/json", body: JSON.stringify({ error: { type: "auth_error", message: "Unauthorized" } }) });
    }
    return route.fulfill({ contentType: "application/json", body: JSON.stringify(status) });
  });

  await page.route("**/admin/api/config", (route) => {
    if (route.request().method() === "PUT") {
      return route.fulfill({ contentType: "application/json", body: JSON.stringify({ status: "ok" }) });
    }
    if (authRequired && !route.request().headers()["authorization"]) {
      return route.fulfill({ status: 401, contentType: "application/json", body: JSON.stringify({ error: { type: "auth_error", message: "Unauthorized" } }) });
    }
    return route.fulfill({ contentType: "application/json", body: JSON.stringify(config) });
  });

  await page.route("**/admin/api/routes", (route) => {
    if (route.request().method() === "PUT") {
      return route.fulfill({ contentType: "application/json", body: JSON.stringify({ status: "ok" }) });
    }
    return route.fulfill({ contentType: "application/json", body: JSON.stringify(routes) });
  });

  await page.route("**/admin/api/profiles", (route) => {
    if (route.request().method() === "PUT") {
      return route.fulfill({ contentType: "application/json", body: JSON.stringify({ status: "ok" }) });
    }
    return route.fulfill({ contentType: "application/json", body: JSON.stringify(profiles) });
  });

  await page.route("**/admin/api/downstream/health", (route) => {
    return route.fulfill({ contentType: "application/json", body: JSON.stringify(downstreamHealth) });
  });

  await page.route("**/admin/api/combo/test", (route) => {
    return route.fulfill({ contentType: "application/json", body: JSON.stringify(MOCK_COMBO_RESULT) });
  });

  await page.route("**/admin/api/runtime/**", (route) => {
    return route.fulfill({ contentType: "application/json", body: JSON.stringify({ status: "ok", mode: "always_on" }) });
  });

  await page.route("**/admin/api/startup", (route) => {
    if (route.request().method() === "PUT") {
      return route.fulfill({ contentType: "application/json", body: JSON.stringify({ status: "ok" }) });
    }
    return route.fulfill({ contentType: "application/json", body: JSON.stringify(config.startup) });
  });

  await page.route("**/admin/api/logs", (route) => {
    return route.fulfill({ contentType: "application/json", body: JSON.stringify({ entries: [] }) });
  });

  await page.route("**/admin/api/diagnostics/export", (route) => {
    return route.fulfill({ contentType: "application/json", body: JSON.stringify({ diagnostics: "ok" }) });
  });

  await page.route("**/admin/api/logs/clear", (route) => {
    return route.fulfill({ contentType: "application/json", body: JSON.stringify({ status: "ok" }) });
  });

  await page.route("**/admin/api/config/export", (route) => {
    return route.fulfill({ contentType: "application/json", body: JSON.stringify({ config, routes, profiles }) });
  });

  await page.route("**/admin/api/config/import", (route) => {
    return route.fulfill({ contentType: "application/json", body: JSON.stringify({ status: "ok" }) });
  });

  await page.route("**/admin/api/config/reset", (route) => {
    return route.fulfill({ contentType: "application/json", body: JSON.stringify({ status: "ok" }) });
  });

  await page.route("**/admin/api/routing/dry-run", (route) => {
    return route.fulfill({ contentType: "application/json", body: JSON.stringify({ analysis: {}, classification: {}, decision: {} }) });
  });
}
