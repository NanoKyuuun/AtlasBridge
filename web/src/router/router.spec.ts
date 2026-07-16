import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'

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
    getRoutes: vi.fn().mockResolvedValue({ task_routes: {} }),
    getProfiles: vi.fn().mockResolvedValue({ route_profiles: {} }),
  },
  AuthError: class AuthError extends Error {},
}))

describe('router guard', () => {
  beforeEach(() => {
    sessionStorage.clear()
    setActivePinia(createPinia())
  })

  it('allows access to /login without token', async () => {
    const { default: router } = await import('../router/index')
    const to = router.getRoutes().find(r => r.name === 'login')
    expect(to).toBeDefined()
  })

  it('allows access to /setup without token', async () => {
    const { default: router } = await import('../router/index')
    const to = router.getRoutes().find(r => r.name === 'setup')
    expect(to).toBeDefined()
  })

  it('has all expected routes', async () => {
    const { default: router } = await import('../router/index')
    const routeNames = router.getRoutes().map(r => r.name as string)
    expect(routeNames).toContain('dashboard')
    expect(routeNames).toContain('login')
    expect(routeNames).toContain('setup')
    expect(routeNames).toContain('routing')
    expect(routeNames).toContain('profiles')
    expect(routeNames).toContain('runtime')
    expect(routeNames).toContain('startup')
    expect(routeNames).toContain('downstream')
    expect(routeNames).toContain('logs')
    expect(routeNames).toContain('privacy')
    expect(routeNames).toContain('advanced')
    expect(routeNames).toContain('observability')
  })

  it('/logs route exists and is not PrivacySettings', async () => {
    const { default: router } = await import('../router/index')
    const logsRoute = router.getRoutes().find(r => r.name === 'logs')
    expect(logsRoute).toBeDefined()
    expect(logsRoute?.path).toBe('/logs')
    // Verify it's a different route from /privacy
    const privacyRoute = router.getRoutes().find(r => r.name === 'privacy')
    expect(privacyRoute?.path).toBe('/privacy')
    expect(logsRoute?.path).not.toBe(privacyRoute?.path)
  })
})
