import { describe, it, expect, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'

// Mock API client to avoid auth store circular import issues
vi.mock('../api/client', () => ({
  api: {},
  AuthError: class AuthError extends Error {},
}))

import { useAuthStore } from '../stores/auth'

describe('auth store', () => {
  beforeEach(() => {
    sessionStorage.clear()
    setActivePinia(createPinia())
  })

  it('starts with no token', () => {
    const auth = useAuthStore()
    expect(auth.token).toBeNull()
    expect(auth.authRequired).toBe(false)
  })

  it('setToken stores token in sessionStorage', () => {
    const auth = useAuthStore()
    auth.setToken('test-token-abc')
    expect(auth.token).toBe('test-token-abc')
    expect(sessionStorage.getItem('atlasbridge_token')).toBe('test-token-abc')
  })

  it('clearToken removes token from sessionStorage', () => {
    const auth = useAuthStore()
    auth.setToken('test-token')
    auth.clearToken()
    expect(auth.token).toBeNull()
    expect(sessionStorage.getItem('atlasbridge_token')).toBeNull()
  })

  it('loads existing token from sessionStorage on init', () => {
    sessionStorage.setItem('atlasbridge_token', 'existing-token')
    const auth = useAuthStore()
    expect(auth.token).toBe('existing-token')
  })

  it('setToken clears authRequired', () => {
    const auth = useAuthStore()
    auth.authRequired = true
    auth.setToken('new-token')
    expect(auth.authRequired).toBe(false)
  })
})
