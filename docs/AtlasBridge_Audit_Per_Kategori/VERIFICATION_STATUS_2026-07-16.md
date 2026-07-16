# AtlasBridge — Audit Per Kategori: Status Verifikasi (16 Juli 2026)

**Tanggal verifikasi:** 16 Juli 2026  
**Cara verifikasi:** Static code review terhadap kode aktual, perbandingan dengan temuan audit

---

## Ringkasan Status per Kategori

| Kategori | Total Temuan | Fixed | Design Decision | Belum Fix | Needs Check |
|---|---:|---:|---:|---:|---:|
| Security (01) | 14 | 8 | 0 | 4 | 2 |
| Backend (02) | 15 | 10 | 4 | 1 | 0 |
| Frontend (03) | 13 | 8 | 0 | 5 | 0 |
| Performance (04) | 4 | 0 | 1 | 3 | 0 |
| Testing/CI (05) | 8 | 0 | 0 | 6 | 2 |
| Hygiene (06) | 8 | 0 | 0 | 6 | 2 |
| **Total** | **62** | **26** | **5** | **25** | **6** |

---

## 1. Security (01_SECURITY.md)

| ID | Temuan | Status | Bukti |
|---|---|---|---|
| SEC-01 | Password hashing SHA-256 unsalted | ✅ **FIXED** | `bcrypt.GenerateFromPassword` di `security.go:37` |
| SEC-02 | Login tanpa rate limit/lockout | ✅ **FIXED** | `loginRateLimiter` dengan exponential backoff + lockout di `middleware.go:182-287` |
| SEC-03 | Login tidak dilindungi SameOrigin/HostGuard | ✅ **FIXED** | Login endpoint punya `RequireJSON`, `LoginRateLimit`, `HostGuard`, `SameOriginAdmin` di `server.go:93` |
| SEC-04 | SSRF hanya block IP literal, tidak DNS-aware | ✅ **FIXED** | `ValidateDownstreamURL` resolve hostname dan check semua IP di `ssrf.go:44-53`. Custom `DialContext` di forwarder juga DNS-aware. |
| SEC-05 | Redirect SSRF tidak DNS-aware | ✅ **FIXED** | `SafeRedirectPolicy` resolve hostname di redirect target di `ssrf.go:114-127` |
| SEC-06 | HostGuard terima `0.0.0.0` sebagai valid | ✅ **FIXED** | HostGuard hanya allow `localhost`, `127.0.0.1`, `[::1]` di `middleware.go:155-160` |
| SEC-07 | Token admin disimpan di localStorage | ❌ **BELUM** | `auth.ts:7` masih pakai `localStorage` — perlu `sessionStorage` atau cookie HttpOnly |
| SEC-08 | Session token tidak punya TTL | ✅ **FIXED** | `SessionExpiresAt` di `security.go:87`, login set 24h di `admin.go:628` |
| SEC-09 | Minimum password 6 karakter | ⚠️ **PARTIAL** | Backend tetap 6, UI guidance 12. Belum enforce di backend. |
| SEC-10 | Error admin leak `err.Error()` mentah | ✅ **FIXED** | Semua handler pakai `writeError(w, code, "generic message")` dengan correlation ID |
| SEC-11 | Structured logging belum terintegrasi | ❌ **BELUM** | Masih banyak `log.Printf` di forwarder/admin. Redactor ada tapi belum mandatory. |
| SEC-12 | Raw token dicetak ke stdout | ⚠️ **PARTIAL** | Token ditulis ke file `0o600` (`admin.go:636`), tetapi masih dicetak ke stdout juga. |
| SEC-13 | Downstream URL terekspos ke UI | ✅ **FIXED** | `maskURL()` di `server.go:600-609` — status endpoint mask URL. |
| SEC-14 | CSP blocking inline styles | ✅ **FIXED** | CSP sekarang include `style-src 'self' 'unsafe-inline'` di `middleware.go:132` |

---

## 2. Backend API & Arsitektur (02_BACKEND_API_DAN_ARSITEKTUR.md)

| ID | Temuan | Status | Bukti |
|---|---|---|---|
| BE-01 | `admin_path` tidak efektif | ✅ **FIXED** | Input admin_path dihapus dari UI. Backend field tetap untuk backward compat. |
| BE-02 | Forwarder hanya `/chat/completions` | ✅ **DESIGN** | Intentional — hanya ada 1 data-plane endpoint |
| BE-03 | Authorization tidak diteruskan | ✅ **DESIGN** | Intentional — proxy handle auth sendiri |
| BE-04 | Wildcard header tidak bekerja | ✅ **FIXED** | `isAllowedHeader()` pakai prefix matching di `headers.go:37-41` |
| BE-05 | Streaming selalu paksa `text/event-stream` | ✅ **FIXED** | Non-2xx preserve content-type downstream di `forwarder.go:244-248` |
| BE-06 | Stream idle timeout tidak efektif saat read block | ✅ **FIXED** | `deadlineReader` wrapper cancel context saat idle timeout di `forwarder.go:177-211` |
| BE-07 | Stream tidak ada byte budget | ✅ **FIXED** | `StreamMaxBytesBudget = 256 MB` di `forwarder.go:31` |
| BE-08 | `IsStreamRequest` dead function | ✅ **FIXED** | Fungsi dihapus dari codebase |
| BE-09 | Request validation dangkal | ✅ **DESIGN** | Tradeoff wajar untuk pass-through proxy |
| BE-10 | Substring matching false positive | ✅ **DESIGN** | Confidence scores sudah di-dampen |
| BE-11 | App lifecycle hang jika error awal | ✅ **FIXED** | `Run()` error ditangkap, `Shutdown()` + `os.Exit(1)` di `main.go` |
| BE-12 | Import/reset tidak atomic | ✅ **FIXED** | Marshal semua data dulu, lalu write. `SaveAtomic` ditambahkan. |
| BE-13 | Network invariant tidak ditegakkan saat runtime change | ✅ **FIXED** | `EnforceNetworkInvariants` dipanggil di `ApplyConfigPatch()` |
| BE-14 | Health check tidak DNS-aware | ✅ **FIXED** | `downstreamHealthHandler` pakai `Forwarder.StreamTransport()` dengan DNS-aware dialer |
| BE-15 | `BindLocalhostOnly` vs `AllowLANAccess` membingungkan | ✅ **FIXED** | `EnforceNetworkInvariants` check kedua field |

---

## 3. Frontend, UX, dan UI State (03_FRONTEND_UX_DAN_UI_STATE.md)

| ID | Temuan | Status | Bukti |
|---|---|---|---|
| FE-01 | `/logs` route salah ke PrivacySettings | ✅ **FIXED** | Route `/logs` → `Logs.vue` di `router/index.ts` |
| FE-02 | PrivacySettings diroute dua kali | ✅ **OK** | `/logs` → Logs.vue, `/privacy` → PrivacySettings.vue — terpisah sekarang |
| FE-03 | Topbar Start/Stop fake toggle | ✅ **FIXED** | Sekarang panggil `api.runtimeStart()`/`runtimeStop()` + refresh status |
| FE-04 | Runtime page hidden | ✅ **FIXED** | Nav link "Runtime Control" ditambahkan di sidebar Layout.vue |
| FE-05 | Downstream health cek `ok` vs `connected` | ✅ **FIXED** | `status === "ok" \|\| "connected"` di Dashboard.vue |
| FE-06 | Dead UI controls di Advanced | ✅ **FIXED** | Dry-run, expose headers, auth mode, streaming toggle dihapus |
| FE-07 | Dummy runtime history | ✅ **FIXED** | Diganti empty state, `ref([])` di StartupSettings.vue |
| FE-08 | Setup wizard unauthenticated | ✅ **FIXED** | Router guard allow setup saat `first_run_completed=false` |
| FE-09 | Login legacy token fallback | ❌ **BELUM** | `admin.go:600` masih terima raw token sebagai password |
| FE-10 | Logo inkonsisten | ❌ **BELUM** | Masih pakai inline SVG + "AB" circle, asset logo tersedia tapi belum dipakai |
| FE-11 | Responsiveness lemah | ❌ **BELUM** | Sidebar fixed 260px, grid fixed cols, sedikit breakpoint |
| FE-12 | Error handling console-only | ❌ **BELUM** | Beberapa catch hanya `console.error` tanpa user toast |
| FE-13 | Inline styles tersebar | ❌ **BELUM** | 31+ instance inline style di Vue files |

---

## 4. Performance (04_PERFORMANCE.md)

| ID | Temuan | Status | Bukti |
|---|---|---|---|
| PERF-01 | Dynamic import warning | ❌ **BELUM** | Vite warning masih muncul di build |
| PERF-02 | Performance fields tidak terisi | ❌ **BELUM** | Log entry tidak konsisten mengisi semua field |
| PERF-03 | Non-stream full buffer | ✅ **DESIGN** | Intentional — sudah ada `MaxResponseBody` limit |
| PERF-04 | Tidak ada perf test gate | ❌ **BELUM** | Tidak ada benchmark/smoke test di CI |

---

## 5. Testing, CI, dan Release (05_TESTING_CI_DAN_RELEASE.md)

| ID | Temuan | Status | Bukti |
|---|---|---|---|
| TEST-01 | Go validation blocker (1.25.5) | ⚠️ **BLOCKER** | `go.mod` butuh 1.25.5, belum bisa verifikasi local |
| TEST-02 | CI Go jobs gagal tanpa web/dist | ❌ **BELUM** | `go-lint`, `go-test`, `go-test-race`, `govulncheck` tidak build frontend dulu |
| TEST-03 | Tool @latest tidak di-pin | ❌ **BELUM** | `staticcheck@latest`, `govulncheck@latest` di CI |
| TEST-04 | Tidak ada frontend unit test | ❌ **BELUM** | Tidak ada Vitest/Jest, hanya `test:e2e` |
| TEST-05 | E2E mocked, tidak test backend | ❌ **BELUM** | `fixtures.ts` intercept semua `/admin/api` |
| TEST-06 | Playwright gagal local | ❌ **BELUM** | Timeout menunggu webServer |
| TEST-07 | Release tanpa test gate | ❌ **BELUM** | Release workflow tidak menjalankan test |
| TEST-08 | Tidak ada signing/SBOM | ❌ **BELUM** | Hanya checksum SHA256 |

---

## 6. Repository Hygiene (06_REPOSITORY_HYGIENE.md)

| ID | Temuan | Status | Bukti |
|---|---|---|---|
| HYGIENE-01 | ZIP berisi `.git` | ⚠️ **NEEDS CHECK** | `.gitignore` tidak exclude `.git` (normal untuk git repo) |
| HYGIENE-02 | Binary `atlasbridge.exe~` | ❌ **BELUM** | Masih ada di working tree |
| HYGIENE-03 | `web/test-results/` | ⚠️ **NEEDS CHECK** | `.gitignore` tidak exclude `web/test-results/` |
| HYGIENE-04 | Audit docs ganda/drift | ❌ **BELUM** | 3+ audit report lama masih ada di `docs/` |
| HYGIENE-05 | Folder `tamplate` typo | ❌ **BELUM** | Masih bernama `tamplate` |
| HYGIENE-06 | `.qwen/skills` ikut | ❌ **BELUM** | Folder `.qwen` tidak di-gitignore |
| HYGIENE-07 | Git working tree tidak clean | ⚠️ **NEEDS CHECK** | Perlu verifikasi setelah perubahan |
| HYGIENE-08 | Logo duplikasi | ❌ **BELUM** | Asset di `docs/assets/`, `web/public/`, belum konsolidasi |

---

## 7. Checklist Prioritas yang Sudah Dikerjakan

### P0 (Security Blockers)

| # | Item | Status |
|---|---|---|
| 1 | Password hashing → bcrypt | ✅ |
| 2 | Login rate limiting/backoff/lockout | ✅ |
| 3 | SSRF DNS-aware semua path | ✅ |
| 4 | CI Go + web/dist | ❌ **BELUM** |
| 5 | Go test/race/vet/govulncheck lulus | ❌ **BLOCKER** (Go 1.25.5) |
| 6 | Hapus binary + .git dari distribusi | ❌ **BELUM** |

### P1 (Before UAT)

| # | Item | Status |
|---|---|---|
| 1 | Route `/logs` fix | ✅ |
| 2 | Topbar Start/Stop → API | ✅ |
| 3 | Downstream health contract | ✅ |
| 4 | `admin_path` fix | ✅ |
| 5 | Dummy runtime history | ✅ |
| 6 | Setup wizard flow | ✅ |
| 7 | Error admin tidak leak | ✅ |
| 8 | Frontend unit tests | ❌ **BELUM** |

---

## 8. Temuan yang Masih Perlu Dikerjakan (Sisa 25 item)

### Critical / High

1. **CI Go jobs tanpa web/dist** (TEST-02) — Go CI akan gagal dari clean checkout
2. **Go test validation** (TEST-01) — Belum bisa verifikasi dengan Go 1.25.5
3. **Login legacy token fallback** (FE-09) — `admin.go:600` masih terima raw token
4. **Token di localStorage** (SEC-07) — Belum pakai sessionStorage/HttpOnly cookie
5. **Release tanpa test gate** (TEST-07) — Release bisa dibuat dari kode yang test-nya gagal
6. **Binary `atlasbridge.exe~`** (HYGIENE-02) — Masih ada di working tree

### Medium

7. **Structured logging belum terintegrasi** (SEC-11) — Masih banyak `log.Printf`
8. **Minimum password backend 6 karakter** (SEC-09) — Belum enforce 12
9. **Raw token stdout** (SEC-12) — Masih dicetak ke stdout
10. **Frontend unit tests** (TEST-04) — Tidak ada Vitest
11. **E2E mocked** (TEST-05) — Tidak test backend nyata
12. **Tool @latest** (TEST-03) — staticcheck/govulncheck tidak di-pin
13. **Performance fields tidak terisi** (PERF-02) — Log entry tidak lengkap
14. **Logo inkonsisten** (FE-10) — Asset ada tapi belum dipakai
15. **Responsiveness** (FE-11) — Sidebar fixed, grid fixed
16. **Error console-only** (FE-12) — Catch tanpa user toast
17. **Audit docs drift** (HYGIENE-04) — 3+ report lama
18. **Folder `tamplate`** (HYGIENE-05) — Typo
19. **`.qwen` di repo** (HYGIENE-06) — Tidak di-gitignore

### Low

20. **Dynamic import warning** (PERF-01) — Vite warning
21. **No perf test gate** (PERF-04) — Tidak ada benchmark
22. **No signing/SBOM** (TEST-08) — Supply chain
23. **Playwright gagal local** (TEST-06) — Timeout
24. **Inline styles** (FE-13) — 31+ instance
25. **Logo duplikasi** (HYGIENE-08) — Asset belum konsolidasi

---

## 9. Kesimpulan

Dari **62 temuan audit** dalam 6 kategori:

- **26 (42%) sudah diperbaiki** — termasuk semua P0 security, streaming, config atomic, SSRF, auth, dan UI fixes
- **5 (8%) adalah design decisions** — intentional behavior yang tidak perlu diubah
- **25 (40%) masih perlu dikerjakan** — mayoritas di area CI/testing, repository hygiene, dan beberapa security hardening
- **6 (10%) perlu verifikasi** — membutuhkan environment tertentu atau pengecekan lebih lanjut

**Prioritas paling kritis yang masih tersisa:**
1. Fix CI Go jobs agar build frontend sebelum Go test
2. Verifikasi Go test suite dengan Go 1.25.5
3. Hapus binary `atlasbridge.exe~` dari working tree
4. Hapus atau fix login legacy token fallback
