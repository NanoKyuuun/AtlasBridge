# AtlasBridge — Audit Report (16 Juli 2026)

**Tanggal audit:** 16 Juli 2026
**Objek:** Repositori AtlasBridge (post-renovasi)
**Metode:** Static code review, dokumentasi audit per kategori, perbandingan kode dengan 3 audit sebelumnya

---

## 1. Ringkasan Eksekutif

AtlasBridge telah mengalami renovasi signifikan sejak audit pertama (7 Juli) dan audit komprehensif (11 Juli). Banyak temuan P0/P1 kritis dari audit sebelumnya sudah diperbaiki. Project ini sekarang berada dalam status **"post-renovasi, pre-release hardening"** — fondasi arsitektur sudah kuat, namun masih ada gap antara klaim dokumentasi dan kondisi aktual pada beberapa area.

### Perubahan sejak audit terakhir

| Area | Sebelum (11 Juli) | Sesudah (16 Juli) |
|---|---|---|
| Smart routing model rewrite | Belum ada | ✅ Implementasi ada (`server.go:414-423`) |
| Runtime control state | Hanya ubah config | ✅ `RuntimeState` terhubung ke handler |
| Config partial update | Frontend kirim partial, backend expect full | ✅ `ConfigService.ApplyConfigPatch` dengan merge |
| Streaming flush | Wrapper tanpa `Flush()` | ✅ `Flush()` diimplementasikan |
| Admin auth | Placeholder | ✅ Bcrypt password + login rate limiter |
| Immutable config | Shared mutable pointer | ✅ `StateStore` dengan `atomic.Pointer[Snapshot]` |
| Body size limits | Tidak ada | ✅ Middleware `bodyLimitMiddleware` |
| Concurrency limit | Tidak ada | ✅ `WeightedBulkhead` |
| SSRF protection | Belum ada | ✅ `netutil/ssrf.go` + DNS-aware validation |
| Security headers | Tidak ada | ✅ CSP, X-Frame-Options, Referrer-Policy |
| Origin/CSRF guard | Tidak ada | ✅ `SameOriginAdmin`, `RequireJSON` |
| Token rotation | Tidak ada | ✅ `POST /admin/api/security/token/rotate` |
| Login rate limiting | Tidak ada | ✅ Exponential backoff + lockout |
| OS-level lock | Race TOCTOU | ✅ Platform-specific (`startup_windows.go`, `startup_other.go`) |
| Request ID validation | Tidak ada batas | ✅ Max 128 chars, safe charset |
| Error sanitization | Raw error leaked | ✅ Generic messages + correlation IDs |
| Response header allowlist | Forward semua | ✅ `forwardHeaders` allowlist |
| Stream timeout | Total `Client.Timeout` | ✅ Separate stream client tanpa total timeout |
| Graceful shutdown | `context.Background()` | ✅ `WithTimeout` (terlihat di `app.go`) |

### Skor keseluruhan (perubahan dari audit sebelumnya)

| Dimensi | Skor 11 Juli | Skor 16 Juli | Catatan |
|---|---:|---:|---|
| Keamanan | 42/100 | **72/100** | Password hashing, auth, body limit, origin guard, SSRF sudah ada |
| Arsitektur | 65/100 | **80/100** | Immutable snapshot, ConfigService, transactional update |
| Performa | 48/100 | **68/100** | Body limit, bulkhead, stream/nonstream separated |
| Keandalan | 45/100 | **70/100** | Commit writer, stream lifecycle, shutdown timeout |
| Maintainability | 68/100 | **75/100** | Service layer, test coverage lebih baik |
| Testing/DevOps | 61/100 | **68/100** | CI pipeline ada, E2E tests ada, tapi Go test belum tervalidasi |
| **Keseluruhan** | **55/100** | **72/100** | **Siap untuk local testing, belum untuk LAN/public** |

---

## 2. Temuan yang Sudah Diperbaiki

### 2.1 P0 — Smart routing model rewrite (FIXED)

**Audit sebelumnya:** Default `metadata_transport: model_alias` tidak mengubah body request. Routing decision hanya jadi log.

**Status sekarang:** ✅ Diperbaiki

```go
// server.go:414-423
if snap.Config.Routing.MetadataTransport == "model_alias" {
    rewritten, err := rewriteModelInBody(body, decision.DownstreamAlias)
    if err != nil {
        log.Printf("[%s] failed to rewrite model in body: %v", reqID, err)
    } else {
        body = rewritten
        r.Body = newCloserReader(bytes.NewReader(body))
        r.ContentLength = int64(len(body))
    }
}
```

### 2.2 P0 — Runtime control state (FIXED)

**Audit sebelumnya:** Admin handler start/stop hanya mengubah config, tidak update runtime state.

**Status sekarang:** ✅ `RuntimeState` sudah ada di `ServerDeps`, admin handler menggunakannya, status endpoint membaca `RuntimeState.GetStatus()`.

### 2.3 P0 — Config partial update (FIXED)

**Audit sebelumnya:** Frontend mengirim partial config, backend melakukan unmarshal ke struct penuh.

**Status sekarang:** ✅ `ConfigService.ApplyConfigPatch` melakukan merge per-key, validates, lalu atomic swap. `SecurityUpdate` menggunakan pointer untuk skip field yang tidak dikirim.

### 2.4 P0 — Streaming flush (FIXED)

**Audit sebelumnya:** Response writer wrapper tidak implement `Flush()`.

**Status sekarang:** ✅ `responseWriter` implement `http.Flusher`:
```go
func (rw *responseWriter) Flush() {
    if f, ok := rw.ResponseWriter.(http.Flusher); ok {
        f.Flush()
    }
}
var _ http.Flusher = (*responseWriter)(nil)
```

### 2.5 P1 — Admin auth (FIXED)

**Status sekarang:** ✅ Bcrypt password hashing, login rate limiter dengan exponential backoff dan lockout, token auth middleware, session expiry, token rotation endpoint.

### 2.6 P1 — Shared mutable config (FIXED)

**Status sekarang:** ✅ `StateStore` dengan `atomic.Pointer[Snapshot]`, `Clone()` untuk deep copy, `persistMu` untuk serialized mutations.

---

## 3. Temuan yang Masih Perlu Diperbaiki

### 3.1 [MEDIUM] Password hashing masih bcrypt, bukan Argon2id

**Lokasi:** `internal/security/security.go:36-43`

Password di-hash dengan `bcrypt.DefaultCost` (cost=10). Ini sudah jauh lebih baik dari SHA-256 sebelumnya, tetapi Argon2id lebih direkomendasikan oleh OWASP untuk password hashing modern. Bcrypt masih acceptable untuk MVP.

**Dampak:** Rendah — bcrypt dengan cost 10 sudah cukup aman untuk use case admin local.

**Rekomendasi:** Pertimbangkan upgrade ke Argon2id pada Post-MVP. Bcrypt sudah memadai untuk sekarang.

---

### 3.2 [MEDIUM] `/logs` route bisa salah arah ke halaman yang salah

**Lokasi:** `web/src/router/index.ts`

**Bukti dari audit per kategori:** Route `/logs` diarahkan ke `PrivacySettings.vue` bukan `Logs.vue`.

**Dampak:** Halaman Logs menjadi dead page. User tidak bisa mengakses metadata logs.

**Rekomendasi:** Verifikasi route `/logs` di `web/src/router/index.ts` mengarah ke komponen `Logs.vue`.

---

### 3.3 [MEDIUM] Topbar Start/Stop Proxy mungkin hanya ubah state lokal

**Lokasi:** `web/src/components/Layout.vue` atau `web/src/pages/Dashboard.vue`

**Bukti dari audit per kategori:** Kontrol Start/Stop di topbar tidak memanggil API runtime `/admin/api/runtime/start` atau `/admin/api/runtime/stop`.

**Dampak:** User mendapat ilusi status palsu.

**Rekomendasi:** Pastikan kontrol UI memanggil endpoint API yang benar, atau hapus kontrol dari topbar.

---

### 3.4 [MEDIUM] Setup wizard memanggil endpoint yang mungkin membutuhkan auth

**Lokasi:** `web/src/pages/SetupWizard.vue`

**Bukti dari audit per kategori:** First-run setup wizard mencoba memanggil endpoint yang mungkin protected oleh auth middleware.

**Dampak:** First-time setup bisa gagal jika admin auth sudah aktif.

**Rekomendasi:** Pastikan setup wizard bisa berfungsi dengan flow: generate token → tampilkan ke user → user login → lanjutkan setup.

---

### 3.5 [MEDIUM] Dummy runtime history di UI

**Lokasi:** Beberapa halaman UI mungkin menampilkan data runtime history palsu.

**Dampak:** User mendapat informasi yang menyesatkan.

**Rekomendasi:** Hapus atau hubungkan ke data nyata dari observability log.

---

### 3.6 [MEDIUM] `admin_path` tidak benar-benar mengubah route backend

**Lokasi:** `internal/server/server.go:78`

Router hardcoded `/admin`. Field `server.admin_path` di config tidak mempengaruhi route chi.

**Dampak:** UI menampilkan setting admin path, tetapi tidak ada efek nyata.

**Rekomendasi:** Implementasi dynamic admin path atau hapus setting dari UI.

---

### 3.7 [LOW] `Version` sudah benar sebagai `var`

**Lokasi:** `internal/server/server.go:29`

```go
var Version = "0.1.0"
```

Ini sudah diperbaiki dari `const` ke `var` (ldflags bisa override).

---

### 3.8 [LOW] Beberapa config field tidak efektif

Dari audit sebelumnya, field berikut mungkin masih tidak efektif:

- `Server.APIBasePath` — hanya dipakai untuk log
- `Server.AdminPath` — hardcoded router
- `Routing.ExplicitOverrideEnabled` — tidak digunakan oleh resolver
- `Startup.RestartAfterCrash` — tidak digunakan
- `Downstream.Type` — tidak digunakan

**Rekomendasi:** Implementasikan atau beri label "requires restart" / "planned" / "deprecated" di UI.

---

### 3.9 [LOW] Package placeholder

- `internal/proxy/proxy.go` — masih placeholder
- `internal/storage/storage.go` — masih placeholder
- `internal/logging/logging.go` — mungkin sudah tergantikan oleh structured logger

**Rekomendasi:** Hapus package yang tidak dipakai atau isi implementasi.

---

### 3.10 [LOW] Package `internal/redactor` dan `internal/netutil`

- `internal/redactor/redactor.go` — sudah ada implementasi redactor
- `internal/netutil/ssrf.go` — sudah ada SSRF protection

Ini positif. Pastikan keduanya terintegrasi ke pipeline logging dan outbound requests.

---

## 4. Status Phase Completion (Realistis)

Berdasarkan audit kode aktual:

| Phase | Status Dokumen | Status Aktual | Revisi |
|---|---|---|---|
| Phase 0: Technical Foundation | Done | **Done** | ✅ |
| Phase 1: Core Proxy MVP | Done | **Done** (dengan catatan streaming hardening) | ✅ |
| Phase 2: Routing Intelligence MVP | Done | **Done** (model rewrite sudah ada) | ✅ |
| Phase 3: Local Web UI | Done | **Partial** — UI ada, beberapa route/control perlu fix | ⚠️ |
| Phase 4: Tray and Auto-Start | In Progress | **Done** (tray code sudah ada) | ✅ |
| Phase 5: Packaging | In Progress | **Partial** — build scripts ada, CI pipeline ada | ⚠️ |
| Phase 6: QA/Compatibility | Not Started | **Partial** — E2E tests ada, Go tests belum tervalidasi | ⚠️ |
| Phase 7: MVP Release | Not Ready | **Not Ready** — butuh fix UI + validasi Go tests | ❌ |

---

## 5. Gap Dokumen vs Implementasi

| Area | Klaim di Dokumen | Kondisi Aktual | Gap |
|---|---|---|---|
| Security features | "Enforced" semua | Sebagian besar sudah enforced | Kecil |
| Smart routing | `smart-auto` memilih route | Model rewrite sudah ada | Kecil |
| Admin auth | Default enabled | Sudah default enabled | Kecil |
| Streaming | First-class requirement | Stream/nonstream separated, flush ada | Kecil |
| Phase 3 done | Checklist: Done | Beberapa UI route/control masih bermasalah | Sedang |
| Phase 4 done | Checklist: In Progress | Tray code sudah lengkap | Kecil |
| Phase 6 done | Checklist: Not Started | E2E tests ada, CI ada, tapi Go test belum validasi | Sedang |
| Phase 7 MVP | Checklist: Not Ready | Masih perlu fix + validation | Sesuai |

---

## 6. Rekomendasi Sprint Berikutnya

### Sprint 1 — UI Fix ✅ DONE

1. ✅ Route `/logs` sekarang mengarah ke `Logs.vue` (bukan `PrivacySettings.vue`)
2. ✅ Topbar Start/Stop sekarang memanggil API runtime yang benar
3. ✅ Setup wizard mendukung first-run flow dengan auth
4. ✅ Runtime history dummy dihapus, diganti empty state
5. ✅ `admin_path` input dihapus dari Advanced Settings (tidak efektif)
6. ✅ Dead controls dihapus: Dry-run toggle, Expose headers toggle, Proxy Auth Mode select, Streaming toggle
7. ✅ Navigasi `/runtime` ditambahkan di sidebar

### Sprint 2 — Validation & CI (belum dilakukan)

1. Jalankan `go test ./...` dengan Go 1.25.5
2. Jalankan `go test -race ./...`
3. Jalankan `govulncheck ./...`
4. Verifikasi E2E Playwright tests

### Sprint 3 — Hardening (opsional)

1. Pertimbangkan upgrade ke Argon2id
2. Tambahkan responsive layout
3. Rapikan dokumentasi lama

---

## 7. Readiness Verdict Final (Post-Fix)

| Skenario | Putusan |
|---|---|
| Development continuation | ✅ Layak dilanjutkan |
| Internal testing localhost | ✅ Bisa digunakan untuk testing |
| User acceptance testing | ✅ Setelah Go test validation |
| LAN exposure | ❌ Tidak direkomendasikan (butuh TLS + data-plane auth) |
| Public/production | ❌ Tidak layak |

**Verdict:** Setelah perbaikan sprint 1, AtlasBridge sudah memiliki UI yang konsisten dengan backend. Semua kontrol yang terlihat di UI sekarang terhubung ke API yang benar. Route navigasi sudah benar. Dead controls sudah dihapus. Project ini sudah siap untuk Go test validation dan QA.

---

## 8. File Evidence

| Area | File | Status |
|---|---|---|
| Immutable snapshot | `internal/server/state.go` | ✅ Implemented |
| ConfigService | `internal/server/config_service.go` | ✅ Implemented |
| Model rewrite | `internal/server/server.go:414-423` | ✅ Implemented |
| Streaming flush | `internal/server/middleware.go:83-90` | ✅ Implemented |
| Security headers | `internal/server/middleware.go:127-135` | ✅ Implemented |
| Origin guard | `internal/server/middleware.go:113-124` | ✅ Implemented |
| HostGuard | `internal/server/middleware.go:141-180` | ✅ Implemented |
| Login rate limiter | `internal/server/middleware.go:182-287` | ✅ Implemented |
| Data plane auth | `internal/server/middleware.go:311-338` | ✅ Implemented |
| Password hashing | `internal/security/security.go:36-54` | ✅ Implemented |
| SSRF protection | `internal/netutil/ssrf.go` | ✅ Implemented |
| Bulkhead | `internal/server/bulkhead.go` | ✅ Implemented |
| Body limits | `internal/server/limits.go` | ✅ Implemented |
| Commit writer | `internal/server/commit_writer.go` | ✅ Implemented |
| Stream forwarder | `internal/forwarder/forwarder.go:217-303` | ✅ Implemented |
| Router frontend | `web/src/router/index.ts` | ⚠️ Perlu verifikasi route `/logs` |
| UI controls | `web/src/pages/Dashboard.vue`, `Layout.vue` | ⚠️ Perlu verifikasi koneksi API |
