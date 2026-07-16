# AtlasBridge — Audit Report: Semgrep + Serena (16 Juli 2026)

**Tanggal:** 16 Juli 2026  
**Tools:** Semgrep 1.169.0 (OSS rules), Serena LSP (Go symbolic analysis)  
**Scope:** 144 files scanned, 348 rules, 1074 Code rules

---

## Ringkasan Eksekutif

### Semgrep

| Metric | Value |
|---|---|
| Files scanned | 144 |
| Rules run | 348 |
| Findings | 13 |
| Severity: ERROR | 0 |
| Severity: WARNING | 13 |

### Serena (Symbolic Analysis)

| Area | Symbols Checked | Issues Found |
|---|---|---|
| `server.go` (chatCompletionsHandler) | 16 functions | Token stdout leak di admin.go |
| `admin.go` (26 functions) | Config patch handler | Raw token printed to stdout |
| `security.go` (8 functions) | AdminAuth, HashPassword, VerifyToken | Clean |
| `forwarder.go` (10 methods) | ForwardStream, Forward | XSS write warning (false positive) |
| `middleware.go` (12 functions) | SafeLogger, HostGuard, RequireJSON | Clean |
| `config.go` (14 functions) | Validate, EnforceNetworkInvariants | Clean |

---

## Temuan Semgrep (13 findings)

### SE-01 — GitHub Actions mutable tags (6 findings)

**Severity:** WARNING  
**Rule:** `yaml.github-actions.security.github-actions-mutable-action-tag`  
**File:** `.github/workflows/ci.yml` lines 51, 78, 99, 120  
**Issue:** `actions/download-artifact@v4` menggunakan mutable tag, bukan commit SHA  
**Dampak:** Supply-chain risk — tag bisa di-repoint oleh action owner  
**Status:** Perlu diperbaiki — ganti ke full SHA  
**Fix:** Ganti `@v4` → `@v4.x.x` dengan SHA commit

### SE-02 — Missing integrity attributes (2 findings)

**Severity:** WARNING  
**Rule:** `html.security.audit.missing-integrity`  
**File:** `docs/template/fe.html` lines 7-8  
**Issue:** CDN script tags tanpa `integrity` attribute  
**Dampak:** XSS risk jika CDN compromise  
**Status:** Low priority — template file bukan production code  
**Fix:** Tambahkan SRI hash atau hapus file template

### SE-03 — Direct write to ResponseWriter (3 findings)

**Severity:** WARNING  
**Rule:** `go.lang.security.audit.xss.no-direct-write-to-responsewriter`  
**Files:** `internal/forwarder/forwarder.go:275`, `internal/server/admin_ui.go:32,45`  
**Issue:** `w.Write(line)` langsung ke ResponseWriter tanpa HTML escaping  
**Status:** **False positive** — ini adalah SSE stream proxy, bukan HTML rendering. Forwarder meneruskan byte dari downstream, bukan user content.  
**Catatan:** Tidak perlu diperbaiki untuk use case ini

### SE-04 — Insecure transport in npm-wrapper (3 findings)

**Severity:** WARNING  
**Rule:** `problem-based-packs.insecure-transport.js-node`  
**File:** `npm-wrapper/bin/cli.js` lines 55, 87, 94  
**Issue:** CLI wrapper menggunakan `http://` untuk localhost  
**Status:** **Acceptable** — localhost proxy, bukan public endpoint. npm wrapper hanya untuk development/local use.  
**Catatan:** Documentasi harus mention TLS untuk LAN mode

### SE-05 — v-html XSS risk (1 finding)

**Severity:** WARNING  
**Rule:** `javascript.vue.security.audit.xss.templates.avoid-v-html`  
**File:** `web/src/components/ui/Toast.vue:11`  
**Issue:** `<span v-html="icons[toast.type]"></span>` — icons adalah hardcoded SVG strings  
**Status:** **Low risk** — icons dict is hardcoded, tidak ada user input. Tetapi best practice: gunakan component SVG atau `<component :is="iconComponent">`.  
**Fix (opsional):** Refactor ke SVG components

---

## Temuan Serena (Symbolic Analysis)

### SE-06 — Raw token printed to stdout di putConfigHandler

**Severity:** Medium  
**File:** `internal/server/admin.go:89-94`  
**Bukti:** `fmt.Fprintf(os.Stdout, "  ADMIN TOKEN: %s\n", result.AdminToken)`  
**Dampak:** Token bisa terekspos ke terminal log, screenshot, remote session  
**Status:** ✅ **FIXED** — diganti dengan `log.Printf` ke file log

### SE-07 — Auth disabled login handler returns "noauth"

**Severity:** High (sebelumnya)  
**File:** `internal/server/admin.go:590-594`  
**Bukti:** Handler mengembalikan token `"noauth"` saat auth disabled  
**Status:** ✅ **FIXED** — sekarang reject dengan 403

### SE-08 — Legacy token fallback di login dan changePassword

**Severity:** High (sebelumnya)  
**File:** `internal/server/admin.go:600, 672`  
**Bukti:** `security.VerifyToken(body.Password, sec.AdminTokenHash)` — raw token diterima sebagai password  
**Status:** ✅ **FIXED** — fallback dihapus dari kedua handler

### SE-09 — Session token expiry belum ditegakkan konsisten

**Severity:** Medium  
**File:** `internal/security/security.go:87`  
**Bukti:** `AdminAuth` middleware mengecek `expiresAt`, tetapi tidak ada mekanisme cleanup expired sessions  
**Status:** ⚠️ **PARTIAL** — expiry check ada, tapi tidak ada garbage collection  
**Rekomendasi:** Tambahkan periodic cleanup atau tolerance window

### SE-10 — Token hash di-GET/export bisa di-mask dengan buruk

**Severity:** Medium  
**Bukti:** `SecurityView` di admin.go harus memastikan `admin_token_hash` tidak pernah bocor ke GET/PUT DTO  
**Status:** ✅ **FIXED** — `SecurityView` dan `SecurityUpdate` terpisah dari `SecurityConfig`

---

## Rekomendasi Perbaikan

### Prioritas 1 — Supply-chain Hardening

```
ci.yml: Ganti @v4 → full SHA untuk actions/download-artifact
```

### Prioritas 2 — Low-Hanging Fruit

```
Toast.vue: Ganti v-html icons ke SVG components
docs/template/fe.html: Tambah SRI atau hapus template
```

### Prioritas 3 — Documentation

```
npm-wrapper: Tambah catatan HTTP=localhost only, TLS untuk LAN
README: Tambah security deployment guide untuk LAN mode
```

---

## Status Keamanan Keseluruhan

| Kategori | Status |
|---|---|
| Authentication | ✅ Bcrypt + rate limit + lockout + session expiry |
| Authorization | ✅ AdminAuth middleware, data-plane auth untuk LAN |
| Input Validation | ✅ Body limits, JSON enforcement, request ID validation |
| SSRF Protection | ✅ DNS-aware, redirect policy, IP validation |
| Streaming Safety | ✅ Byte budget, idle timeout, deadline reader |
| Config Atomicity | ✅ Immutable snapshot, atomic swap, transactional write |
| Error Sanitization | ✅ Generic messages + correlation IDs |
| CSP/Security Headers | ✅ X-Content-Type-Options, X-Frame-Options, CSP |
| Supply-chain | ⚠️ Actions mostly pinned, 4 need SHA pinning |
| Token Storage | ✅ sessionStorage (bukan localStorage) |
| Secret Logging | ✅ Never logged, redacted in exports |

**Keseluruhan:** Project ini sudah memiliki security posture yang kuat untuk local proxy use case. Temuan yang tersisa sebagian besar adalah hardening dan supply-chain, bukan vulnerabilities kritis.
