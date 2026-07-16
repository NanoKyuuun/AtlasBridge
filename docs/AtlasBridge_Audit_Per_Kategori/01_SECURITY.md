# AtlasBridge — Security

**Catatan:** File ini adalah pecahan laporan audit utama agar temuan dapat dikerjakan per kategori. Tidak ada perubahan source code yang dilakukan.

---

## 7. Temuan Detail — Security

### SEC-01 — Password hashing tidak aman

**Severity:** Critical  
**Evidence:** `internal/security/security.go:32-39`  
**Temuan:** `HashPassword()` hanya memanggil `HashToken()` yang menggunakan SHA-256 tanpa salt dan tanpa work factor. Ini masih dapat diterima untuk token acak 32-byte, tetapi tidak aman untuk password yang dipilih manusia.  
**Dampak:** Jika `AdminPasswordHash` atau config bocor, password dapat di-bruteforce/offline-crack jauh lebih cepat.  
**Rekomendasi audit:** gunakan Argon2id, bcrypt, atau scrypt dengan salt unik dan parameter biaya yang eksplisit. Migrasi hash lama secara bertahap.

### SEC-02 — Login admin tidak memiliki rate limit, lockout, atau backoff

**Severity:** Critical  
**Evidence:** `internal/server/server.go:89-90`, `internal/server/admin.go:563-621`  
**Temuan:** endpoint public `/admin/api/auth/login` hanya menggunakan `RequireJSON`. Tidak terlihat limiter per IP, per username/session, backoff eksponensial, lockout sementara, atau audit counter. Selain itu, login handler memiliki fallback (line 586-588) yang menerima raw admin token sebagai password — menggandakan attack surface. Ketika `AdminAuthEnabled` `false` (default), handler mengembalikan token hardcoded `"noauth"` tanpa verifikasi apa pun.  
**Dampak:** Brute-force password admin dapat dilakukan terus-menerus, terutama jika LAN mode/public exposure aktif. Fallback token membuat attacker bisa mencoba password DAN token sekaligus. Default `noauth` membuat endpoint terbuka sepenuhnya sampai admin mengaktifkan auth.  
**Rekomendasi audit:** tambahkan rate limiting khusus login, lockout sementara, incremental delay, dan event log untuk failed login. Hapus atau batasi fallback token-as-password. Pastikan `AdminAuthEnabled` default ke `true` atau hanya untuk initial setup.

### SEC-03 — Login tidak dilindungi Same-Origin/HostGuard

**Severity:** High  
**Evidence:** `internal/server/server.go:89-100`  
**Temuan:** `SameOriginAdmin`, `HostGuard`, dan `AdminAuth` hanya diterapkan pada protected route group. `/auth/login` berada di luar group tersebut dan hanya dibatasi JSON content type.  
**Dampak:** Permukaan login lebih longgar daripada endpoint admin lain.  
**Rekomendasi audit:** terapkan HostGuard dan SameOriginAdmin juga ke endpoint login, atau definisikan alasan threat-model secara eksplisit.

### SEC-04 — SSRF protection hanya memblokir IP literal, bukan DNS-resolved hostname

**Severity:** Critical  
**Evidence:** `internal/netutil/ssrf.go:14-42`, `internal/server/downstream_validate.go:50-69`, `internal/forwarder/forwarder.go:40-78`  
**Temuan:** `ValidateDownstreamURL()` memanggil `net.ParseIP(host)` dan hanya memeriksa IP jika host berbentuk IP literal. Hostname biasa tidak di-resolve dan divalidasi. Fungsi `ResolveAndValidateIP()` memang ada di `downstream_validate.go:50-69`, tetapi **dead code** — tidak dipanggil di path forwarder utama maupun di konstruksi forwarder (`New()` hanya memanggil `ValidateDownstreamURL`). Selain itu, `IsAllowedIP()` di `ssrf.go:49-51` memiliki bug **inverted logic**: return `true` untuk loopback (`127.0.0.1`, `::1`) — seharusnya *block*, bukan *allow*. `ValidateDownstreamURLRemote()` (`downstream_validate.go:19-40`) juga tidak melakukan IP check sama sekali (hanya validasi scheme/host/credentials). Validasi hanya terjadi saat forwarder dibuat (`New()`), bukan per-request — hostname yang resolve ke IP private setelah pembuatan lolos tanpa deteksi.  
**Dampak:** DNS rebinding atau domain yang resolve ke private/link-local/metadata IP dapat melewati validasi. Loopback address diizinkan melewati SSRF check meskipun seharusnya diblokir.  
**Rekomendasi audit:** wire `ResolveAndValidateIP()` ke path forwarder, perbaiki `IsAllowedIP()` agar block loopback, gunakan custom DialContext yang memvalidasi IP final sebelum connect, dan lakukan validasi DNS per-request (bukan hanya saat konstruksi).

### SEC-05 — Redirect SSRF policy juga belum DNS-aware

**Severity:** High  
**Evidence:** `internal/netutil/ssrf.go:89-100`  
**Temuan:** redirect target divalidasi dengan aturan URL yang sama, tetapi tidak resolve hostname menjadi IP final.  
**Dampak:** Redirect ke hostname yang resolve ke private/internal address masih berpotensi lolos.  
**Rekomendasi audit:** redirect policy perlu DNS/IP validation per redirect target.

### SEC-06 — HostGuard menganggap `0.0.0.0` valid sebagai loopback-like host

**Severity:** Medium  
**Evidence:** `internal/server/middleware.go:154-168`  
**Temuan:** daftar allowed host memasukkan `0.0.0.0`, dan `isLoopbackHost()` juga memperlakukan `0.0.0.0` sebagai host yang diterima. Selain itu, HostGuard mengizinkan empty Host header tanpa validasi (line 144-147) — HTTP/1.0 requests tanpa Host header lolos sepenuhnya.  
**Dampak:** `0.0.0.0` bukan host browser normal untuk same-origin semantics dan dapat melemahkan validasi host. Empty Host header merupakan bypass path tambahan.  
**Rekomendasi audit:** jangan treat `0.0.0.0` sebagai host yang sah untuk Host header; gunakan explicit bind address vs allowed browser host secara terpisah. Tolak request dengan empty Host header atau tentukan handling yang eksplisit.

### SEC-07 — Token admin disimpan di `localStorage`

**Severity:** High  
**Evidence:** `web/src/stores/auth.ts:4-13`  
**Temuan:** bearer token admin disimpan persisten di `localStorage` dengan key `atlasbridge_token`.  
**Dampak:** Jika XSS terjadi, token mudah dicuri dan bertahan setelah browser restart.  
**Rekomendasi audit:** gunakan session in-memory token atau cookie `HttpOnly`, `SameSite`, `Secure` bila arsitektur memungkinkan. Minimal tambahkan TTL dan clear-on-close option.

### SEC-08 — Session token tidak memiliki TTL/expiry yang jelas

**Severity:** Medium  
**Evidence:** `internal/server/admin.go:596-617`, `web/src/stores/auth.ts:7-13`  
**Temuan:** login menghasilkan token baru dan menyimpannya sebagai hash di config, tetapi tidak ada TTL, idle timeout, refresh strategy, atau session list.  
**Dampak:** Token yang dicuri tetap berlaku sampai token diganti/dirotasi.  
**Rekomendasi audit:** tambahkan expiry/idle timeout dan session invalidation.

### SEC-09 — Minimum password terlalu lemah

**Severity:** Medium  
**Evidence:** `internal/server/admin.go:635-640`, UI change password pada `web/src/pages/AdvancedSettings.vue`  
**Temuan:** password baru hanya diwajibkan minimal 6 karakter.  
**Dampak:** Kombinasi hash cepat + password pendek + tanpa rate limit meningkatkan risiko compromise.  
**Rekomendasi audit:** minimal 12 karakter, block common passwords, atau gunakan passphrase guidance.

### SEC-10 — Error response admin banyak mengembalikan `err.Error()` mentah

**Severity:** High  
**Evidence:** `internal/server/admin.go:89`, `95`, `134`, `157`, `168`, `221`, `228`, `234`, `258`, `400`, `411`, `417`, `445`, `461`, `469`, `482`, `538`, `549`, `666`  
**Temuan:** banyak handler admin menggabungkan pesan error internal langsung ke response user. Semua endpoint ini berada di belakang `AdminAuth` middleware.  
**Dampak:** Berpotensi membocorkan path, detail konfigurasi, format internal, downstream URL, atau informasi debug ke session admin yang terkompromi atau malicious.  
**Rekomendasi audit:** gunakan error code user-safe dan simpan detail di log internal yang sudah direduksi.

### SEC-11 — Structured logging dan redaction belum benar-benar terintegrasi

**Severity:** Medium  
**Evidence:** `internal/logging/structured.go`, `internal/redactor/redactor.go`, pemakaian `log.Printf` di `internal/app/app.go`, `internal/forwarder/forwarder.go`, `internal/server/admin.go`  
**Temuan:** package structured logging/redactor ada, tetapi runtime utama masih dominan memakai `log.Printf`.  
**Dampak:** Klaim logging terstruktur/redacted tidak konsisten; risiko secret masuk log lebih tinggi.  
**Rekomendasi audit:** injeksikan logger terstruktur ke server/app/forwarder/admin handler dan jadikan redactor mandatory.

### SEC-12 — Raw admin token dicetak ke stdout

**Severity:** Medium  
**Evidence:** `internal/app/app.go:119-123`, `internal/server/admin.go:100-104`  
**Temuan:** token admin dicetak ke stdout saat generated. Ini mungkin intentional, tetapi tetap risk-prone bila terminal/log direkam.  
**Dampak:** Token dapat terekspos melalui log terminal, screenshot, remote support session, atau logging service.  
**Rekomendasi audit:** tampilkan sekali hanya pada secure UI/CLI mode, beri opsi suppress, dan pastikan tidak masuk app log.

### SEC-13 — Downstream URL terekspos ke UI/diagnostic response

**Severity:** Low/Medium  
**Evidence:** `internal/server/admin.go:276-315`, `admin.go:70`, `admin.go:371`, `internal/server/server.go:188`  
**Temuan:** downstream URL terekspos di beberapa endpoint: health (`admin.go:319`), config (`admin.go:70`), diagnostics export (`admin.go:371`), dan status (`server.go:188`). Semua endpoint berada di belakang admin auth.  
**Dampak:** Bukan secret secara langsung, tetapi dapat mengungkap topology internal (hostname, port, path) ke session admin yang terkompromi.  
**Rekomendasi audit:** mask hostname/internal detail pada UI umum atau tampilkan hanya pada advanced diagnostics.

### SEC-14 — CSP berpotensi tidak cocok dengan inline style pada Vue templates

**Severity:** Medium  
**Evidence:** `internal/server/middleware.go:131`; 31 instance inline `style` di 12 Vue files (`Layout.vue`, `Dashboard.vue`, `Observability.vue`, `StartupSettings.vue`, `PrivacySettings.vue`, `RoutingSettings.vue`, `RouteProfiles.vue`, `Logs.vue`, `Toast.vue`, dan lainnya)  
**Temuan:** CSP `default-src 'self'; frame-ancestors 'none'` tanpa `style-src` eksplisit membuat inline `style="..."` dan `:style` bindings di-block oleh browser modern. Ini bukan hanya theoretical — ditemukan 31 instance inline style yang akan menyebabkan UI rusak.  
**Dampak:** UI admin mengalami broken styles pada browser yang menerapkan CSP ketat. Ini adalah functional bug, bukan cuma theoretical concern.  
**Rekomendasi audit:** audit CSP aktual di browser, refactor inline style ke CSS classes, dan definisikan `style-src 'self' 'unsafe-inline'` (atau better: nonce-based approach) serta `script-src`, `img-src`, `connect-src` secara eksplisit.

---
