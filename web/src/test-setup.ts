// Stub localStorage/sessionStorage before any Vue/Pinia imports.
// @vue/devtools-kit tries to access localStorage at import time.
const localStorageStore = new Map<string, string>()
const sessionStorageStore = new Map<string, string>()
const makeStorage = (store: Map<string, string>): Storage => ({
  getItem: (k: string) => store.get(k) ?? null,
  setItem: (k: string, v: string) => { store.set(k, String(v)) },
  removeItem: (k: string) => { store.delete(k) },
  clear: () => { store.clear() },
  get length() { return store.size },
  key: (i: number) => [...store.keys()][i] ?? null,
})
vi.stubGlobal('localStorage', makeStorage(localStorageStore))
vi.stubGlobal('sessionStorage', makeStorage(sessionStorageStore))
