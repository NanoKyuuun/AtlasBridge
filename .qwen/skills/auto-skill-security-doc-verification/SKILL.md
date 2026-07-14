---
name: security-doc-verification
description: Verify security documentation against codebase and perform deep code-level security audit
source: auto-skill
extracted_at: '2026-07-14T06:27:46.577Z'
---

## Security Verification & Audit Procedure

Two modes: **(A)** verify a security doc against the codebase, **(B)** deep code-level audit. Both share the same file map.

### Shared File Map

**Authentication & Authorization:**
- `internal/security/security.go` — Token generation, hashing, constant-time verify
- `internal/server/middleware.go` — AdminAuth, dataPlaneAuth, SecurityHeaders, HostGuard, SameOriginAdmin
- `internal/server/config_service.go` — Token rotation, config persistence, reset safety

**SSRF & Network Policy:**
- `internal/netutil/ssrf.go` — IP validation (`IsAllowedIP`), URL validation, redirect policy
- `internal/server/downstream_validate.go` — Downstream URL validation wrapper

**Concurrency / Rate Limiting:**
- `internal/server/bulkhead.go` — WeightedBulkhead (concurrency limiter, NOT rate limiter)
- `internal/server/limits.go` — Body size limits (`MaxBytesReader`)

**Web Security:**
- `internal/server/admin_ui.go` — Admin SPA serving
- `web/src/stores/auth.ts` — Frontend token storage
- `web/src/api/client.ts` — API client auth header injection

**Logging & Redaction:**
- `internal/redactor/redactor.go` — Sensitive key/value redaction
- `internal/logging/structured.go` — Structured logger with prompt logging toggle

---

## Mode A: Document Verification

### 1. Parse the Document

- Identify each test case ID (e.g., SEC-01, SEC-02)
- Group by category (Authentication, SSRF, Web Security, etc.)
- Note the expected result for each case

### 2. Cross-Reference Each Test Case

For each SEC-XX item:
1. Find the code that implements the expected behavior
2. Verify the implementation matches the expected result exactly
3. Check if tests exist for this scenario
4. Note any discrepancies

### 3. Output Format

| ID | Scenario | Status | Notes |
|---|---|---|---|
| SEC-XX | Description | ✅/⚠️/❌ | Implementation details or discrepancy |

**Status:** ✅ Correct | ⚠️ Partial/Discrepancy | ❌ Not implemented

---

## Mode B: Deep Code Audit

When asked to "check the project" or do a security audit beyond the markdown:

### 1. Cryptographic Randomness
```
grep pattern: math/rand in *.go files
```
Must use `crypto/rand` for tokens/secrets. `math/rand` is insecure.

### 2. Hardcoded Secrets
```
grep pattern: password|secret|api_key|apikey|credential in *.{go,ts,yaml,json}
```
Exclude test files. Flag real credentials.

### 3. Token Storage (Frontend)
Check `web/src/stores/*.ts` for:
- `localStorage` — BAD (persists forever, XSS-vulnerable)
- `sessionStorage` — MEDIUM (XSS-vulnerable but dies on tab close)
- In-memory only — BEST (but dies on refresh)

### 4. Authentication Bypass
Read the chi router setup in `internal/server/server.go`:
- Every route under `/admin/api` MUST have `security.AdminAuth` middleware
- `/v1/*` routes use `dataPlaneAuth` (conditional on LAN mode)
- `/health` is intentionally unauthenticated

### 5. Log Injection / Secret Leakage
```
grep pattern: log\.(Printf|Println|Print) in internal/
```
Check if user-controlled data (request IDs, headers) flows into log statements unsanitized. Check if `Authorization` header is ever logged.

### 6. CORS Configuration
```
grep pattern: CORS|Access-Control|cors in *.go
```
No CORS headers = good for localhost-only admin. Overly permissive CORS = vulnerability.

### 7. Body Size Limits
Verify `http.MaxBytesReader` is used for ALL request bodies:
- Chat completions: 16MB (`MaxChatBody`)
- Admin API: 1MB (`MaxAdminBody`)
- Config import: 8MB (`MaxImportBody`)

### 8. Error Handling Safety
- `web/assets.go`: Check for `panic()` on missing embed — potential DoS
- Forwarder: Check if request body is properly closed on error paths
- `writeJSONAfterCommit`: Already handles committed response correctly

### 9. Hop-by-Hop Header Stripping
Verify `internal/server/headers.go` strips: Connection, Keep-Alive, Proxy-Auth, Set-Cookie, Transfer-Encoding, Upgrade.

### 10. Atomic State Swaps
Verify `internal/server/state.go` uses `atomic.Pointer[Snapshot]` — no torn reads under concurrent config updates.

### 11. Test Coverage Gaps
Check `internal/server/server_test.go` for:
- `TestAdminAuthRejectsNoToken` / `TestAdminAuthRejectsWrongToken`
- `TestDataPlaneAuthRejectsNoToken` / `TestDataPlaneAuthAcceptsValidToken`
- `TestSecurityHeadersPresent`
- `TestHopByHopHeadersStripped`
- `TestNoRawPromptInObservabilityLog`

---

## Common Discrepancies to Watch For

**Rate Limiting vs Concurrency Limiting:**
- Rate limiting = requests per time window (e.g., 100 req/min)
- Concurrency limiting = simultaneous connections (e.g., 64 in-flight)
- `WeightedBulkhead` is concurrency limiting only

**IPv6 Handling:**
- IPv6 loopback (`::1`) is ALLOWED (same as IPv4 `127.0.0.1`)
- IPv6 link-local (`fe80::/10`) is BLOCKED
- Documentation often conflates these

**Dead Code:**
- `ResolveAndValidateIP()` in `downstream_validate.go` exists but is NOT called during forwarding
- DNS resolution validation only happens at URL construction time, not per-request

**Token Rotation:**
- Old token hash IS replaced in config (good)
- Old `.token` file on disk is overwritten (acceptable — file is just for CLI convenience)

**Security Headers Missing:**
- `X-Robots-Tag: noindex` — NOT implemented (medium risk, admin UI only)
- `Strict-Transport-Security` — NOT implemented (acceptable for localhost HTTP)
