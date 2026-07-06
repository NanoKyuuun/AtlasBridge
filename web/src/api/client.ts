const BASE = "/admin/api";

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    ...options,
    headers: { "Content-Type": "application/json", ...options?.headers },
  });
  if (!res.ok) {
    const err = await res
      .json()
      .catch(() => ({ error: { message: res.statusText } }));
    throw new Error(err.error?.message || `HTTP ${res.status}`);
  }
  return res.json();
}

export interface AppConfig {
  app: { name: string; mode: string; first_run_completed: boolean };
  server: {
    host: string;
    port: number;
    admin_path: string;
    api_base_path: string;
  };
  downstream: { type: string; base_url: string; timeout_seconds: number };
  security: {
    admin_auth_enabled: boolean;
    admin_token_hash: string;
    bind_localhost_only: boolean;
    allow_lan_access: boolean;
  };
  startup: {
    run_at_login: boolean;
    start_proxy_on_app_launch: boolean;
    restart_after_crash: boolean;
  };
  routing: {
    auto_routing: boolean;
    default_route: string;
    low_confidence_route: string;
    explicit_override_enabled: boolean;
    confidence_threshold: number;
  };
  logging: {
    level: string;
    privacy_mode: string;
    prompt_logging_enabled: boolean;
    metadata_logging_enabled: boolean;
    retention_days: number;
  };
}

export interface RoutesConfig {
  task_routes: Record<string, string>;
}

export interface RouteProfile {
  label: string;
  description: string;
  downstream_alias: string;
  priority: string;
  enabled: boolean;
}

export interface ProfilesConfig {
  route_profiles: Record<string, RouteProfile>;
}

export interface StatusResponse {
  status: string;
  version: string;
  uptime: string;
  port: number;
  host: string;
  downstream: string;
  mode: string;
  privacy: string;
  go_version: string;
  pid: number;
}

export interface DownstreamHealth {
  status: string;
  status_code?: number;
  url: string;
  message?: string;
}

export interface DryRunResult {
  analysis: any;
  classification: any;
  decision: any;
}

export interface ComboTestResult {
  model: string;
  resolved_model?: string;
  success: boolean;
  status_code?: number;
  error?: string;
  latency: number;
}

export const api = {
  getStatus: () => request<StatusResponse>("/status"),
  getConfig: () => request<AppConfig>("/config"),
  updateConfig: (cfg: Partial<AppConfig>) =>
    request<{ status: string }>("/config", {
      method: "PUT",
      body: JSON.stringify(cfg),
    }),
  getRoutes: () => request<RoutesConfig>("/routes"),
  updateRoutes: (routes: RoutesConfig) =>
    request<{ status: string }>("/routes", {
      method: "PUT",
      body: JSON.stringify(routes),
    }),
  getProfiles: () => request<ProfilesConfig>("/profiles"),
  updateProfiles: (profiles: ProfilesConfig) =>
    request<{ status: string }>("/profiles", {
      method: "PUT",
      body: JSON.stringify(profiles),
    }),
  runtimeStart: () =>
    request<{ status: string; mode: string }>("/runtime/start", {
      method: "POST",
    }),
  runtimeStop: () =>
    request<{ status: string; mode: string }>("/runtime/stop", {
      method: "POST",
    }),
  runtimeRestart: () =>
    request<{ status: string; mode: string }>("/runtime/restart", {
      method: "POST",
    }),
  getStartup: () => request<AppConfig["startup"]>("/startup"),
  updateStartup: (s: AppConfig["startup"]) =>
    request<{ status: string }>("/startup", {
      method: "PUT",
      body: JSON.stringify(s),
    }),
  getDownstreamHealth: () => request<DownstreamHealth>("/downstream/health"),
  getLogs: () => request<any>("/logs"),
  exportDiagnostics: () =>
    request<any>("/diagnostics/export", { method: "POST" }),
  dryRun: (prompt: string, model: string) =>
    request<DryRunResult>("/routing/dry-run", {
      method: "POST",
      body: JSON.stringify({ prompt, model }),
    }),
  exportConfig: () =>
    request<{
      config: AppConfig;
      routes: RoutesConfig;
      profiles: ProfilesConfig;
    }>("/config/export"),
  importConfig: (data: any) =>
    request<{ status: string }>("/config/import", {
      method: "POST",
      body: JSON.stringify(data),
    }),
  resetConfig: () =>
    request<{ status: string }>("/config/reset", { method: "POST" }),
  clearLogs: () =>
    request<{ status: string }>("/logs/clear", { method: "POST" }),
  testCombo: (model: string) =>
    request<ComboTestResult>("/combo/test", {
      method: "POST",
      body: JSON.stringify({ model }),
    }),
};
