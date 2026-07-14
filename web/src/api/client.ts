const BASE = "/admin/api";

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...((options?.headers as Record<string, string>) || {}),
  };

  const { useAuthStore } = await import("../stores/auth");
  const auth = useAuthStore();
  if (auth.token) {
    headers["Authorization"] = `Bearer ${auth.token}`;
  }

  const res = await fetch(`${BASE}${path}`, {
    ...options,
    headers,
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    const errType = body?.error?.type;
    const errMsg = body?.error?.message || `HTTP ${res.status}`;

    if (res.status === 401 && errType === "auth_error") {
      auth.authRequired = true;
      if (auth.token) {
        auth.clearToken();
      }
      throw new AuthError(errMsg);
    }

    throw new Error(errMsg);
  }

  return res.json();
}

// loginWithPassword calls the PUBLIC /auth/login endpoint (no Bearer token needed).
export async function loginWithPassword(
  password: string
): Promise<{ token: string }> {
  const res = await fetch(`${BASE}/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ password }),
  });
  const body = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error(body?.error?.message || `HTTP ${res.status}`);
  }
  return body;
}

export class AuthError extends Error {
  constructor(msg: string) {
    super(msg);
    this.name = "AuthError";
  }
}

export interface AppConfig {
  app: { name: string; mode: string; first_run_completed: boolean };
  server: {
    host: string;
    port: number;
    admin_path: string;
  };
  downstream: { base_url: string; timeout_seconds: number };
  security: {
    admin_auth_enabled: boolean;
    token_configured: boolean;
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
    confidence_threshold: number;
    smart_fast_route: string;
    metadata_transport: string;
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
    request<{ status: string; admin_token?: string }>("/config", {
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
  changePassword: (currentPassword: string, newPassword: string) =>
    request<{ status: string; message: string }>("/auth/change-password", {
      method: "POST",
      body: JSON.stringify({
        current_password: currentPassword,
        new_password: newPassword,
      }),
    }),
  saveConfig: (partial: Record<string, unknown>) =>
    request<{ status: string }>("/config", {
      method: "PUT",
      body: JSON.stringify(partial),
    }),
};
