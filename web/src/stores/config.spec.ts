import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useConfigStore } from '../stores/config'

vi.mock('../api/client', () => ({
  api: {
    getConfig: vi.fn().mockResolvedValue({
      app: { name: 'AtlasBridge', mode: 'manual', first_run_completed: true },
      server: { host: '127.0.0.1', port: 20127 },
      downstream: { base_url: 'http://127.0.0.1:20128/v1', timeout_seconds: 120 },
      security: { admin_auth_enabled: true, token_configured: true, bind_localhost_only: true, allow_lan_access: false },
      startup: { run_at_login: false, start_proxy_on_app_launch: true, restart_after_crash: true },
      routing: { auto_routing: true, default_route: 'route.default', low_confidence_route: 'route.default', confidence_threshold: 0.55, smart_fast_route: 'route.low_cost', metadata_transport: 'model_alias' },
      logging: { level: 'info', privacy_mode: 'standard', prompt_logging_enabled: false, metadata_logging_enabled: true, retention_days: 7 },
    }),
    getRoutes: vi.fn().mockResolvedValue({ task_routes: { debugging: 'route.debugging' } }),
    getProfiles: vi.fn().mockResolvedValue({ route_profiles: {} }),
    updateConfig: vi.fn().mockResolvedValue({ status: 'ok' }),
    updateRoutes: vi.fn().mockResolvedValue({ status: 'ok' }),
    updateProfiles: vi.fn().mockResolvedValue({ status: 'ok' }),
    resetConfig: vi.fn().mockResolvedValue({ status: 'ok' }),
  },
  AuthError: class AuthError extends Error {},
}))

describe('config store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('fetchAll loads config, routes, and profiles', async () => {
    const store = useConfigStore()
    await store.fetchAll()
    expect(store.config).not.toBeNull()
    expect(store.config?.server.port).toBe(20127)
    expect(store.routes).not.toBeNull()
    expect(store.routes?.task_routes.debugging).toBe('route.debugging')
    expect(store.profiles).not.toBeNull()
  })

  it('saveConfig calls API and refreshes', async () => {
    const store = useConfigStore()
    await store.fetchAll()
    await store.saveConfig({ logging: { level: 'debug', privacy_mode: 'standard', prompt_logging_enabled: false, metadata_logging_enabled: true, retention_days: 7 } })
    expect(store.saveMessage).toBe('Config saved')
  })

  it('resetConfig calls API and refreshes', async () => {
    const store = useConfigStore()
    await store.fetchAll()
    await store.resetConfig()
    expect(store.saveMessage).toBe('Config reset to defaults')
  })
})
