# Audit Project AtlasBridge

**Tanggal audit:** 7 Juli 2026  
**Nama project:** AtlasBridge  
**Sumber audit:** `AtlasBridge.zip`  
**Fokus audit:** perencanaan di folder `docs`, implementasi backend, frontend, konfigurasi, keamanan, runtime, build/release, dan gap antara dokumen dengan kode.

---

## 1. Ringkasan Eksekutif

AtlasBridge adalah project local AI routing proxy yang dirancang sebagai jembatan antara AI coding assistant dan downstream service seperti 9Router. Dari sisi perencanaan, project ini terlihat serius: folder `docs` berisi PRD, technical plan, implementation plan, architecture decisions, dan phase completion checklist. Struktur repository juga sudah cukup profesional dengan pemisahan backend Go, frontend Vue, konfigurasi, script, wrapper npm, dan workflow release.

Namun, hasil audit menunjukkan bahwa **implementasi belum sejalan dengan klaim kelengkapan pada dokumen**. Beberapa fitur yang tampak sudah dianggap selesai di checklist ternyata masih belum berjalan secara fungsional di kode. Temuan paling kritis adalah fitur inti smart routing: default config menggunakan `metadata_transport: model_alias`, tetapi server belum benar-benar mengubah field `model` pada request sebelum diteruskan ke downstream. Akibatnya, keputusan routing bisa hanya menjadi log/metadata, bukan mempengaruhi model atau route yang dipakai downstream.

Secara status, project ini lebih tepat disebut:

> **Prototype alpha dengan dokumentasi matang, struktur bagus, dan fondasi teknis awal, tetapi belum siap disebut MVP stabil atau production-ready.**

Project ini **bukan project gagal**. Fondasinya cukup bagus. Tetapi perlu hardening sprint untuk memperbaiki fungsi inti, runtime control, config update, streaming, admin security, dan release pipeline.

---

## 2. Skor Audit Keseluruhan

| Area Audit | Nilai | Status | Catatan Singkat |
|---|---:|---|---|
| Perencanaan dan dokumentasi | 85/100 | Baik | Dokumen lengkap dan arah produk jelas. |
| Struktur repository | 78/100 | Cukup baik | Modular, rapi, tetapi masih ada package placeholder. |
| Backend core proxy | 58/100 | Perlu perbaikan | Proxy dasar ada, tetapi routing default dan streaming bermasalah. |
| Routing intelligence | 52/100 | Belum matang | Analyzer/classifier ada, tetapi hasil routing belum efektif di default mode. |
| Frontend Web UI | 55/100 | Perlu perbaikan | UI cukup lengkap, tetapi beberapa save config berpotensi gagal. |
| Runtime control | 35/100 | Lemah | Admin start/stop hanya ubah config, tidak update runtime state aktif. |
| Security | 35/100 | Lemah | Admin auth masih placeholder dan LAN exposure belum dikunci. |
| Observability/logging | 60/100 | Cukup awal | Logger ada, tetapi privacy setting belum sepenuhnya dihormati. |
| Build dan release readiness | 40/100 | Lemah | Workflow release ada, tetapi test/build lokal dan script perlu diperbaiki. |
| Kesiapan MVP | 45/100 | Belum siap | Cocok untuk alpha internal, belum untuk user umum. |

---

## 3. Ruang Lingkup Audit

Audit dilakukan terhadap area berikut:

1. **Dokumentasi perencanaan**
   - `docs/prd.md`
   - `docs/technical-plan.md`
   - `docs/implementation-plan.md`
   - `docs/architecture-decisions.md`
   - `docs/phase-completion-checklist.md`

2. **Backend Go**
   - `cmd/atlasbridge`
   - `internal/server`
   - `internal/forwarder`
   - `internal/config`
   - `internal/routing`
   - `internal/analyzer`
   - `internal/classifier`
   - `internal/runtime`
   - `internal/observability`
   - `internal/startup`
   - `internal/security`

3. **Frontend Vue**
   - `web/src`
   - `web/package.json`
   - `web/assets.go`

4. **Konfigurasi dan contoh config**
   - `configs/config.example.yaml`
   - `configs/routes.example.yaml`
   - `configs/profiles.example.yaml`

5. **Script dan release**
   - `scripts/*.sh`
   - `scripts/*.ps1`
   - `npm-wrapper/bin/cli.js`
   - `.github/workflows/release.yml`

6. **Testing**
   - File `*_test.go`
   - Percobaan menjalankan `go test ./...`

---

## 4. Metode Audit

Metode audit yang digunakan:

- Membaca struktur repository.
- Membandingkan dokumen perencanaan dengan implementasi kode.
- Static code review terhadap backend dan frontend.
- Memeriksa config default, routes, dan profiles.
- Memeriksa endpoint admin dan runtime.
- Memeriksa kemungkinan bug pada request forwarding dan streaming.
- Memeriksa script build/lint/run.
- Mencoba menjalankan test Go.

Keterbatasan audit:

- Test Go tidak dapat dijalankan penuh karena `go.mod` membutuhkan Go `1.25.5`, sedangkan environment audit memiliki Go `1.23.2` dan tidak dapat mengunduh toolchain dari internet.
- Audit ini tidak melakukan dynamic integration test dengan 9Router asli.
- Audit ini tidak melakukan penetration test aktif.
- Audit ini berfokus pada static correctness dan kesesuaian dokumen dengan kode.

Hasil percobaan test:

```text
$ GOTOOLCHAIN=local go test ./...
go: go.mod requires go >= 1.25.5 (running go 1.23.2; GOTOOLCHAIN=local)
```

---

## 5. Gambaran Struktur Project

Struktur utama project:

```text
AtlasBridge/
├── cmd/atlasbridge/
│   └── main.go
├── configs/
│   ├── config.example.yaml
│   ├── profiles.example.yaml
│   └── routes.example.yaml
├── docs/
│   ├── architecture-decisions.md
│   ├── implementation-plan.md
│   ├── phase-completion-checklist.md
│   ├── prd.md
│   └── technical-plan.md
├── internal/
│   ├── analyzer/
│   ├── app/
│   ├── classifier/
│   ├── config/
│   ├── forwarder/
│   ├── logging/
│   ├── observability/
│   ├── proxy/
│   ├── routing/
│   ├── runtime/
│   ├── security/
│   ├── server/
│   ├── startup/
│   ├── storage/
│   └── tray/
├── npm-wrapper/
├── scripts/
├── testdata/
└── web/
```

Catatan:

- Struktur project sudah cukup baik untuk project Go + Vue.
- Pemisahan internal package cukup jelas.
- Ada 9 file test Go.
- Ada beberapa package placeholder yang belum berisi implementasi nyata.

---

## 6. Audit Dokumentasi Perencanaan di Folder `docs`

### 6.1 Dokumen yang tersedia

| Dokumen | Fungsi | Penilaian |
|---|---|---|
| `prd.md` | Menjelaskan kebutuhan produk, tujuan, user, dan fitur utama | Baik |
| `technical-plan.md` | Menjelaskan rencana teknis, arsitektur, dan teknologi | Baik |
| `implementation-plan.md` | Menjelaskan fase implementasi | Baik |
| `architecture-decisions.md` | Mencatat keputusan arsitektur | Baik |
| `phase-completion-checklist.md` | Checklist status fase | Perlu dikoreksi |

### 6.2 Kekuatan dokumentasi

Dokumentasi memiliki beberapa kelebihan:

1. **Visi produk jelas**  
   AtlasBridge diposisikan sebagai local OpenAI-compatible intelligent routing proxy.

2. **Target use case cukup spesifik**  
   Fokus pada AI coding assistant, local proxy, smart routing, privacy, dan integration dengan downstream service.

3. **Rencana fase lengkap**  
   Ada fase mulai dari technical foundation, proxy MVP, routing intelligence, Web UI, tray, packaging, QA, hingga release.

4. **Keputusan arsitektur cukup terdokumentasi**  
   Pemilihan Go, Vue, same-port UI/API, YAML/JSON config, Windows-first, dan local-first cukup masuk akal.

5. **Dokumen sudah bisa menjadi dasar roadmap engineering**  
   Dokumen bisa digunakan untuk memandu sprint berikutnya.

### 6.3 Masalah utama dokumentasi

Masalah terbesarnya adalah **status checklist terlalu optimistis**.

Di `docs/phase-completion-checklist.md`, beberapa fase ditandai selesai, misalnya:

- Phase 0: Done
- Phase 1: Done
- Phase 2: Done
- Phase 3: Done

Namun implementasi menunjukkan beberapa fitur inti dari phase tersebut belum benar-benar berjalan. Contoh:

| Klaim di checklist | Kondisi kode aktual | Status audit |
|---|---|---|
| Core proxy MVP done | Proxy dasar ada, tetapi streaming flush bermasalah | Sebagian benar |
| Routing intelligence done | Classifier ada, tetapi default `model_alias` tidak rewrite request | Belum selesai |
| Local Web UI settings usable | UI ada, tetapi partial config save berpotensi gagal | Belum stabil |
| Runtime control tersedia | Endpoint hanya ubah config, tidak update runtime state | Belum benar |
| Admin/security tersedia | `internal/security` masih placeholder | Belum selesai |
| Embedded Web UI | `//go:embed dist` ada, tetapi `web/dist` tidak tersedia di zip | Perlu build step jelas |

### 6.4 Rekomendasi dokumentasi

Dokumen sebaiknya direvisi agar jujur dengan status aktual:

```text
Phase 0: Done
Phase 1: Partial / Needs hardening
Phase 2: Partial / Core routing integration incomplete
Phase 3: Partial / UI exists but save behavior needs fix
Phase 4: In progress
Phase 5: In progress
Phase 6: Not started / limited tests only
Phase 7: Not ready
```

---

## 7. Temuan Kritis P0

P0 adalah masalah yang berpotensi membuat fitur utama tidak berjalan, membuat project gagal dipakai, atau membuat klaim utama produk tidak benar.

---

### P0-01 — Smart routing default belum benar-benar mempengaruhi request downstream

**Severity:** P0 Critical  
**Area:** Backend routing, request forwarding  
**File terkait:**

- `internal/config/config.go`
- `internal/server/server.go`
- `internal/forwarder/forwarder.go`

#### Bukti kode

Default config menggunakan metadata transport `model_alias`:

```go
// internal/config/config.go
MetadataTransport: "model_alias",
```

Validasi juga menganggap `model_alias` sebagai mode yang sah:

```go
validMetadataTransport := map[string]bool{"model_alias": true, "header": true}
```

Namun di request handler, kode hanya melakukan tindakan jika transport mode adalah `header`:

```go
if deps.Config.Routing.MetadataTransport == "header" && decision.DownstreamAlias != "" {
    r.Header.Set("X-Route-Intent", decision.DownstreamAlias)
}
```

Tidak ada proses untuk mengubah body JSON request dari:

```json
{
  "model": "smart-auto"
}
```

menjadi:

```json
{
  "model": "combo.backend"
}
```

atau route hasil keputusan classifier.

#### Dampak

Ini adalah masalah paling serius karena fitur utama AtlasBridge adalah smart routing. Jika model tidak diubah dan downstream tidak membaca header khusus, maka routing decision hanya menjadi log internal.

Risikonya:

- User memilih `smart-auto`, tetapi downstream tetap menerima `smart-auto`.
- Downstream mungkin tidak mengenali model tersebut.
- Smart routing tidak mempengaruhi hasil inference.
- Produk tidak memenuhi PRD utama.
- Test bisa memberi rasa aman palsu jika hanya mengecek request tetap diteruskan, bukan route dipakai.

#### Rekomendasi teknis

Pilih salah satu strategi utama.

##### Opsi A — Implement model rewrite

Jika `metadata_transport = model_alias`, maka server harus membaca body JSON, mengganti field `model`, lalu membuat ulang `r.Body` sebelum diteruskan ke forwarder.

Contoh pendekatan:

```go
func rewriteModelInBody(body []byte, model string) ([]byte, error) {
    var payload map[string]interface{}
    if err := json.Unmarshal(body, &payload); err != nil {
        return nil, err
    }

    payload["model"] = model
    return json.Marshal(payload)
}
```

Di handler:

```go
if deps.Config.Routing.MetadataTransport == "model_alias" && decision.DownstreamAlias != "" {
    rewritten, err := rewriteModelInBody(body, decision.DownstreamAlias)
    if err != nil {
        // return 400 atau fallback sesuai policy
    }
    body = rewritten
    r.Body = io.NopCloser(bytes.NewReader(body))
    r.ContentLength = int64(len(body))
}
```

##### Opsi B — Gunakan header routing sebagai default

Jika downstream 9Router memang membaca header:

```http
X-Route-Intent: combo.backend
```

maka ubah default config menjadi:

```yaml
metadata_transport: header
```

Tetapi pendekatan ini hanya benar jika 9Router downstream sudah pasti mendukung header tersebut.

#### Prioritas perbaikan

Wajib diperbaiki sebelum MVP.

---

### P0-02 — Runtime Start/Stop/Restart dari Web UI tidak mengubah runtime state aktif

**Severity:** P0 Critical  
**Area:** Runtime control, Admin API, Web UI  
**File terkait:**

- `internal/server/admin.go`
- `internal/server/server.go`
- `internal/runtime/runtime.go`
- `internal/app/app.go`

#### Bukti kode

`ServerDeps` memiliki `RuntimeState`:

```go
RuntimeState *runtimemod.State
```

`chatCompletionsHandler` membaca runtime state:

```go
mode := deps.RuntimeState.GetMode()
status := deps.RuntimeState.GetStatus()
```

Namun `AdminDeps` tidak membawa `RuntimeState`:

```go
type AdminDeps struct {
    Config        *config.Config
    Routes        *config.RoutesConfig
    Profiles      *config.ProfilesConfig
    Forwarder     *forwarder.Forwarder
    Observability *observability.Logger
}
```

Endpoint runtime hanya mengubah config:

```go
func runtimeStartHandler(deps *AdminDeps) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        deps.Config.App.Mode = "always_on"
        config.Save(deps.Config)
        writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "message": "proxy started"})
    }
}
```

Stop juga hanya mengubah config:

```go
deps.Config.App.Mode = "disabled"
```

Restart bahkan hanya mengirim response sukses tanpa melakukan operasi state nyata:

```go
log.Printf("ADMIN: proxy engine restarted")
writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "message": "proxy restarted"})
```

#### Dampak

Tombol Start/Stop/Restart di Web UI bisa terlihat berhasil, tetapi runtime gate yang dipakai oleh `/v1/chat/completions` tidak berubah secara nyata.

Risiko:

- User menekan Start, tetapi proxy tetap menolak request.
- User menekan Stop, tetapi status server bisa tidak sesuai ekspektasi.
- Restart tidak melakukan restart.
- Dashboard memberikan status palsu.
- UX menjadi membingungkan.

#### Rekomendasi teknis

Tambahkan `RuntimeState` ke `AdminDeps`:

```go
type AdminDeps struct {
    Config        *config.Config
    Routes        *config.RoutesConfig
    Profiles      *config.ProfilesConfig
    Forwarder     *forwarder.Forwarder
    Observability *observability.Logger
    RuntimeState  *runtimemod.State
}
```

Saat membuat `AdminDeps`, teruskan state:

```go
adminDeps := &AdminDeps{
    Config: deps.Config,
    Routes: deps.Routes,
    Profiles: deps.Profiles,
    Forwarder: deps.Fwd,
    Observability: deps.ObsLogger,
    RuntimeState: deps.RuntimeState,
}
```

Handler start:

```go
func runtimeStartHandler(deps *AdminDeps) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        deps.Config.App.Mode = "always_on"
        if deps.RuntimeState != nil {
            deps.RuntimeState.SetMode(runtimemod.ModeAlwaysOn)
            _ = deps.RuntimeState.Start()
        }
        // save config and response
    }
}
```

Handler stop:

```go
if deps.RuntimeState != nil {
    deps.RuntimeState.SetMode(runtimemod.ModeDisabled)
    _ = deps.RuntimeState.Stop()
}
```

Handler restart:

```go
if deps.RuntimeState != nil {
    _ = deps.RuntimeState.Stop()
    _ = deps.RuntimeState.Start()
}
```

#### Prioritas perbaikan

Wajib diperbaiki sebelum UI runtime dianggap usable.

---

### P0-03 — Frontend mengirim partial config, backend mengharapkan full config

**Severity:** P0 Critical  
**Area:** Frontend, Admin API, configuration persistence  
**File terkait:**

- `web/src/stores/config.ts`
- `web/src/pages/DownstreamSettings.vue`
- `web/src/pages/PrivacySettings.vue`
- `web/src/pages/RoutingSettings.vue`
- `web/src/pages/StartupSettings.vue`
- `internal/server/admin.go`

#### Bukti kode frontend

Store menerima partial config:

```ts
async function saveConfig(cfg: Partial<AppConfig>) {
  await api.updateConfig(cfg);
  await fetchAll();
}
```

Beberapa halaman memang mengirim sebagian config saja:

```ts
await configStore.saveConfig({ downstream: downstream.value });
```

```ts
await configStore.saveConfig({ logging: logging.value });
```

```ts
await configStore.saveConfig({ routing: routingConfig.value });
```

#### Bukti kode backend

Backend melakukan unmarshal ke struct config penuh:

```go
var updated config.Config
if err := json.Unmarshal(body, &updated); err != nil {
    ...
}

if err := config.Validate(&updated); err != nil {
    ...
}

*deps.Config = updated
```

Jika frontend hanya mengirim:

```json
{
  "logging": {
    "level": "info"
  }
}
```

maka field lain seperti `server.port`, `server.host`, `downstream.base_url`, dan `app.mode` menjadi empty/zero value. Validasi bisa gagal atau config bisa rusak jika validasi tidak cukup ketat.

#### Dampak

Banyak tombol save di Web UI bisa gagal.

Risiko:

- User mengubah privacy setting, tetapi gagal disimpan.
- User mengubah downstream setting, tetapi backend menolak request.
- Config menjadi tidak lengkap jika validasi luput.
- Web UI terlihat lengkap tetapi tidak usable.

#### Rekomendasi teknis

##### Opsi terbaik — Backend menerima partial update dan melakukan merge

Backend harus membedakan antara full replace dan partial patch.

Contoh konsep:

```go
func patchConfigHandler(deps *AdminDeps) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var patch map[string]json.RawMessage
        if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
            writeError(w, http.StatusBadRequest, "invalid JSON")
            return
        }

        updated := *deps.Config

        if raw, ok := patch["logging"]; ok {
            if err := json.Unmarshal(raw, &updated.Logging); err != nil {
                writeError(w, http.StatusBadRequest, "invalid logging config")
                return
            }
        }

        if raw, ok := patch["downstream"]; ok {
            if err := json.Unmarshal(raw, &updated.Downstream); err != nil {
                writeError(w, http.StatusBadRequest, "invalid downstream config")
                return
            }
        }

        if err := config.Validate(&updated); err != nil {
            writeError(w, http.StatusBadRequest, "validation failed: "+err.Error())
            return
        }

        *deps.Config = updated
        _ = config.Save(deps.Config)
        writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
    }
}
```

##### Opsi alternatif — Frontend merge sebelum kirim

```ts
async function saveConfig(patch: Partial<AppConfig>) {
  if (!config.value) return;

  const merged = {
    ...config.value,
    ...patch,
  };

  await api.updateConfig(merged);
  await fetchAll();
}
```

Namun, untuk nested object, perlu deep merge agar tidak menimpa field lain.

#### Prioritas perbaikan

Wajib diperbaiki sebelum Web UI diklaim usable.

---

### P0-04 — Streaming SSE berpotensi tidak real-time karena response writer wrapper tidak meneruskan Flush

**Severity:** P0 Critical  
**Area:** Streaming, OpenAI-compatible API, middleware  
**File terkait:**

- `internal/server/middleware.go`
- `internal/forwarder/forwarder.go`

#### Bukti kode

Middleware logger membungkus `http.ResponseWriter`:

```go
wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
next.ServeHTTP(wrapped, r)
```

Wrapper hanya mengimplementasikan:

```go
func (rw *responseWriter) WriteHeader(code int)
func (rw *responseWriter) Write(b []byte) (int, error)
```

Namun streaming forwarder mengecek apakah writer bisa flush:

```go
flusher, canFlush := w.(Flusher)
...
if canFlush {
    flusher.Flush()
}
```

Karena wrapper tidak punya method `Flush()`, maka `canFlush` kemungkinan `false`.

#### Dampak

Streaming dari downstream bisa tidak langsung dikirim ke client. Untuk client AI coding assistant, ini masalah besar.

Risiko:

- Token tidak muncul real-time.
- Response terasa macet sampai buffer selesai.
- Client mengira server lambat.
- Compatibility dengan Cursor/Cline/Continue/OpenCode bisa buruk.

#### Rekomendasi teknis

Tambahkan implementasi `Flush()` di wrapper:

```go
func (rw *responseWriter) Flush() {
    if f, ok := rw.ResponseWriter.(http.Flusher); ok {
        f.Flush()
    }
}
```

Lebih baik lagi, buat wrapper mendukung interface tambahan bila diperlukan:

```go
var _ http.ResponseWriter = (*responseWriter)(nil)
var _ http.Flusher = (*responseWriter)(nil)
```

#### Prioritas perbaikan

Wajib sebelum compatibility testing streaming.

---

## 8. Temuan High P1

---

### P1-01 — Admin auth belum diimplementasikan

**Severity:** P1 High  
**Area:** Security, Admin API  
**File terkait:**

- `internal/security/security.go`
- `internal/server/server.go`
- `internal/server/admin.go`
- `internal/config/config.go`

#### Bukti kode

Security config sudah ada:

```go
type SecurityConfig struct {
    AdminAuthEnabled  bool
    AdminTokenHash    string
    BindLocalhostOnly bool
    AllowLANAccess    bool
}
```

Namun package security masih placeholder:

```go
// Package security provides secret redaction and admin auth.
// This is a placeholder for Phase 3 implementation.
package security
```

Tidak terlihat middleware autentikasi pada route:

```go
r.Route("/admin", func(r chi.Router) {
    r.Route("/api", func(r chi.Router) {
        ...
    })
})
```

#### Dampak

Endpoint admin lokal bisa dipanggil tanpa token/password.

Risiko:

- Config bisa diubah oleh proses lokal lain.
- Jika host dibuka ke LAN, admin API bisa terekspos.
- Import/export config bisa disalahgunakan.
- Tidak sesuai klaim privacy/security-first.

#### Rekomendasi

Implement minimal local admin auth:

1. Generate admin token saat first run.
2. Simpan hash token, bukan plaintext token.
3. Web UI memakai secure cookie atau bearer token lokal.
4. Semua `/admin/api/*` wajib melewati middleware auth jika `admin_auth_enabled=true`.
5. Berikan opsi reset token.

Contoh middleware:

```go
func AdminAuth(cfg *config.Config) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !cfg.Security.AdminAuthEnabled {
                next.ServeHTTP(w, r)
                return
            }

            token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
            if !verifyToken(token, cfg.Security.AdminTokenHash) {
                writeError(w, http.StatusUnauthorized, "unauthorized")
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

---

### P1-02 — `BindLocalhostOnly` dan `AllowLANAccess` belum ditegakkan

**Severity:** P1 High  
**Area:** Security, networking  
**File terkait:**

- `internal/config/config.go`
- `internal/app/app.go`

#### Bukti kode

Config memiliki field:

```go
BindLocalhostOnly bool
AllowLANAccess    bool
```

Default-nya:

```go
BindLocalhostOnly: true,
AllowLANAccess: false,
```

Namun server bind langsung memakai config host:

```go
addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
ln, err := net.Listen("tcp", addr)
```

Tidak ada enforcement bahwa jika `allow_lan_access=false`, maka host harus `127.0.0.1` atau `localhost`.

#### Dampak

Jika config berubah menjadi:

```yaml
server:
  host: 0.0.0.0
security:
  allow_lan_access: false
```

server tetap bisa bind ke semua interface.

Risiko:

- Admin UI/API terekspos ke LAN.
- Downstream API key dan config bisa bocor.
- Bertentangan dengan local-first security model.

#### Rekomendasi

Tambahkan validasi/enforcement:

```go
func effectiveHost(cfg *config.Config) string {
    if cfg.Security.BindLocalhostOnly && !cfg.Security.AllowLANAccess {
        return "127.0.0.1"
    }
    return cfg.Server.Host
}
```

Atau tolak config:

```go
if !cfg.Security.AllowLANAccess && cfg.Server.Host != "127.0.0.1" && cfg.Server.Host != "localhost" {
    return fmt.Errorf("LAN access disabled but server.host is not localhost")
}
```

---

### P1-03 — Privacy logging setting belum dihormati sepenuhnya

**Severity:** P1 High  
**Area:** Logging, privacy  
**File terkait:**

- `internal/config/config.go`
- `internal/server/server.go`
- `internal/observability/observability.go`

#### Bukti kode

Config punya:

```go
PromptLoggingEnabled   bool
MetadataLoggingEnabled bool
PrivacyMode            string
```

Namun request observation tetap dicatat dengan metadata routing di `recordObservation(...)`, tanpa terlihat enforcement penuh terhadap `MetadataLoggingEnabled` dan `PrivacyMode`.

#### Dampak

UI mungkin menampilkan privacy setting, tetapi backend belum sepenuhnya mengikuti setting tersebut.

Risiko:

- Metadata tetap tercatat saat user mengira sudah dimatikan.
- Mode strict/debug tidak konsisten.
- Trust terhadap privacy-first claim turun.

#### Rekomendasi

Buat policy logging eksplisit:

```go
type LoggingPolicy struct {
    PromptLoggingEnabled bool
    MetadataLoggingEnabled bool
    PrivacyMode string
}
```

Lalu di `recordObservation`:

```go
if !cfg.Logging.MetadataLoggingEnabled {
    metadata = nil
}

if !cfg.Logging.PromptLoggingEnabled {
    promptPreview = ""
}

if cfg.Logging.PrivacyMode == "strict" {
    stripSensitiveFields(...)
}
```

---

### P1-04 — Config import bisa menyimpan state tidak valid atau gagal diam-diam

**Severity:** P1 High  
**Area:** Config import/export  
**File terkait:**

- `internal/server/admin.go`

#### Bukti kode

Saat import config:

```go
*deps.Config = *imported.Config
config.Save(deps.Config)
```

Return value `config.Save(...)` tidak dicek.

Validasi full hanya dilakukan jika `routes` dan `profiles` sama-sama ada:

```go
if imported.Routes != nil {
    if imported.Profiles != nil {
        ValidateFull(...)
    }
    *deps.Routes = *imported.Routes
}
```

Jika hanya import `routes`, validasi terhadap profiles existing bisa terlewat.

#### Dampak

Risiko config/routes/profiles tidak sinkron.

Contoh:

- Routes mengacu ke profile yang tidak ada.
- Profiles mengacu ke route yang tidak valid.
- Config tersimpan gagal tetapi API tetap terlihat berhasil.

#### Rekomendasi

Gunakan staging copy:

```go
nextCfg := *deps.Config
nextRoutes := *deps.Routes
nextProfiles := *deps.Profiles
```

Apply import ke copy, lalu validasi full:

```go
if err := config.ValidateFull(&nextCfg, &nextRoutes, &nextProfiles); err != nil {
    writeError(...)
    return
}
```

Baru commit ke memory dan disk jika semua valid.

---

### P1-05 — Windows startup registry function ada tetapi belum dipanggil dari admin update

**Severity:** P1 High  
**Area:** Startup behavior, desktop integration  
**File terkait:**

- `internal/startup/startup.go`
- `internal/server/admin.go`
- `web/src/pages/StartupSettings.vue`

#### Bukti kode

Ada fungsi:

```go
func SetRunAtLogin(enable bool) error
```

Namun pada `putStartupHandler`, backend hanya menyimpan config:

```go
deps.Config.Startup = updated
config.Save(deps.Config)
```

Tidak ada pemanggilan:

```go
startup.SetRunAtLogin(updated.RunAtLogin)
```

#### Dampak

Toggle “run at login” di Web UI bisa tersimpan di config, tetapi tidak benar-benar mengubah Windows startup registry.

#### Rekomendasi

Pada `putStartupHandler`, panggil `startup.SetRunAtLogin(...)` jika nilai berubah:

```go
if updated.RunAtLogin != deps.Config.Startup.RunAtLogin {
    if err := startup.SetRunAtLogin(updated.RunAtLogin); err != nil {
        writeError(w, http.StatusInternalServerError, "failed to update startup setting: "+err.Error())
        return
    }
}
```

---

### P1-06 — Package Windows registry tidak dipisahkan dengan build tag

**Severity:** P1 High  
**Area:** Build portability  
**File terkait:**

- `internal/startup/startup.go`

#### Bukti kode

File `internal/startup/startup.go` mengimpor langsung:

```go
"golang.org/x/sys/windows/registry"
```

Tanpa build tag:

```go
//go:build windows
```

#### Dampak

Project bisa gagal build di Linux/macOS. Memang project tampaknya Windows-first, tetapi repo dan CI perlu jelas.

#### Rekomendasi

Pisahkan file:

```text
internal/startup/startup_windows.go
internal/startup/startup_other.go
```

Isi `startup_windows.go`:

```go
//go:build windows
```

Isi `startup_other.go`:

```go
//go:build !windows

func SetRunAtLogin(enable bool) error {
    return nil // atau error explicit: unsupported platform
}
```

---

## 9. Temuan Medium P2

---

### P2-01 — Status endpoint hardcode `running`

**Severity:** P2 Medium  
**File terkait:** `internal/server/server.go`

#### Bukti kode

```go
json.NewEncoder(w).Encode(map[string]interface{}{
    "status": "running",
    ...
})
```

#### Dampak

Dashboard bisa menampilkan status running walaupun runtime state stopped/disabled.

#### Rekomendasi

Status endpoint harus membaca `RuntimeState`:

```go
"runtime_status": deps.RuntimeState.GetStatus(),
"runtime_mode": deps.RuntimeState.GetMode(),
```

---

### P2-02 — Script lint salah path frontend

**Severity:** P2 Medium  
**File terkait:**

- `scripts/lint.sh`
- `scripts/lint.ps1`

#### Bukti kode

```sh
npx --yes prettier --write "src/**/*.{ts,vue,js}"
```

Padahal frontend source ada di:

```text
web/src
```

#### Dampak

Lint/format frontend tidak mengenai file yang benar jika dijalankan dari root repo.

#### Rekomendasi

Gunakan:

```sh
cd web
npx prettier --write "src/**/*.{ts,vue,js}"
```

atau:

```sh
npx --prefix web prettier --write "web/src/**/*.{ts,vue,js}"
```

---

### P2-03 — File `.ps1` berisi syntax batch, bukan PowerShell

**Severity:** P2 Medium  
**File terkait:**

- `scripts/lint.ps1`
- `scripts/run.ps1`

#### Bukti kode

```bat
@echo off
REM AtlasBridge - Lint and Format Script for Windows
setlocal
```

Itu adalah syntax `.bat`/`.cmd`, bukan PowerShell.

#### Dampak

Script akan gagal atau tidak berjalan benar jika dipanggil sebagai PowerShell.

#### Rekomendasi

Opsi A: rename menjadi `.bat`:

```text
lint.bat
run.bat
```

Opsi B: tulis ulang sebagai PowerShell valid:

```powershell
Write-Host "AtlasBridge - Lint and Format Script"
$failed = $false
```

---

### P2-04 — npm wrapper help mencantumkan `stop`, tetapi command tidak tersedia

**Severity:** P2 Medium  
**File terkait:** `npm-wrapper/bin/cli.js`

#### Bukti kode

Help menampilkan:

```text
stop      Stop the proxy
```

Namun switch command tidak memiliki case:

```js
case 'stop':
```

#### Dampak

User yang mengikuti help akan bingung karena `atlasbridge stop` tidak bekerja.

#### Rekomendasi

Tambahkan command `stop`, minimal memanggil endpoint admin runtime stop:

```js
case 'stop':
    cmdStop();
    break;
```

---

### P2-05 — npm wrapper belum punya mekanisme install/download binary

**Severity:** P2 Medium  
**File terkait:** `npm-wrapper/bin/cli.js`

#### Masalah

Wrapper mencari binary lokal/config dir, tetapi belum ada mekanisme postinstall untuk download binary release.

#### Dampak

Jika package dipublish ke npm, user bisa install package tetapi binary tidak tersedia.

#### Rekomendasi

Tambahkan salah satu:

1. `postinstall` script untuk download binary sesuai platform.
2. Bundled binary dalam package per platform.
3. Instruksi manual yang jelas bahwa npm wrapper bukan installer penuh.

---

### P2-06 — `web/dist` tidak tersedia di zip tetapi di-embed oleh Go

**Severity:** P2 Medium  
**File terkait:** `web/assets.go`

#### Bukti kode

```go
//go:embed dist
var distFS embed.FS
```

Namun folder `web/dist` tidak ada di zip dan memang biasanya dihasilkan oleh `npm run build`.

#### Dampak

Build Go bisa gagal jika frontend belum dibuild dulu.

#### Rekomendasi

Dokumentasikan urutan build:

```sh
cd web
npm ci
npm run build
cd ..
go build ./cmd/atlasbridge
```

Tambahkan script root:

```sh
scripts/build.sh
```

yang memastikan frontend build dulu.

---

### P2-07 — Release workflow belum menjalankan test

**Severity:** P2 Medium  
**File terkait:** `.github/workflows/release.yml`

#### Bukti workflow

Workflow melakukan:

```yaml
- npm run build
- go build
```

Tetapi tidak ada:

```yaml
- go test ./...
- npm test
- npm run lint
```

#### Dampak

Release bisa dibuat dari kode yang test-nya gagal.

#### Rekomendasi

Tambahkan CI sebelum release:

```yaml
- name: Run Go tests
  run: go test ./...

- name: Typecheck frontend
  run: |
    cd web
    npm ci
    npm run build
```

Idealnya release workflow bergantung pada workflow CI yang sudah hijau.

---

## 10. Temuan Low P3

---

### P3-01 — Beberapa package masih placeholder

**Severity:** P3 Low  
**File terkait:**

- `internal/logging/logging.go`
- `internal/proxy/proxy.go`
- `internal/security/security.go`
- `internal/storage/storage.go`

#### Dampak

Tidak selalu salah untuk prototype, tetapi berbahaya jika dokumen menyatakan fitur sudah selesai.

#### Rekomendasi

- Hapus package placeholder yang tidak dipakai.
- Atau ubah status phase menjadi belum selesai.
- Atau isi implementasi nyata.

---

### P3-02 — Version hardcode masih `0.1.0`

**Severity:** P3 Low  
**File terkait:** `internal/server/server.go`

#### Bukti kode

```go
const Version = "0.1.0"
```

Workflow mencoba override via ldflags:

```powershell
-X github.com/atlasbridge/atlasbridge/internal/server.Version=$VERSION
```

Namun Go tidak bisa override `const` dengan ldflags. Yang bisa dioverride adalah `var`.

#### Dampak

Release version mungkin tetap `0.1.0` walaupun tag berbeda.

#### Rekomendasi

Ubah:

```go
const Version = "0.1.0"
```

menjadi:

```go
var Version = "0.1.0"
```

---

### P3-03 — Dokumentasi UI/API perlu menyatakan batasan alpha

**Severity:** P3 Low  
**Area:** Docs, user expectation

#### Masalah

Dokumentasi cukup percaya diri, tetapi implementasi masih alpha.

#### Rekomendasi

Tambahkan status:

```text
Current status: Alpha prototype. Not recommended for production use.
Known limitations: admin auth incomplete, smart routing model_alias pending, runtime control hardening pending.
```

---

## 11. Audit Backend Detail

### 11.1 Server routing

Endpoint utama:

- `/health`
- `/v1/models`
- `/v1/chat/completions`
- `/admin/api/*`
- `/admin/*` untuk UI

Struktur cukup baik. Namun ada masalah desain:

1. Admin API belum protected.
2. Status endpoint tidak membaca runtime state.
3. Smart route decision tidak dipakai untuk rewrite model default.
4. Middleware logger mengganggu streaming flush.
5. Handler terlalu banyak melakukan hal sekaligus: read body, analyze, route, log, modify header, forward.

Rekomendasi refactor:

```text
chatCompletionsHandler
├── validate runtime state
├── read request body
├── parse OpenAI request
├── analyze request
├── classify route
├── apply routing decision to request body/header
├── record observation according to privacy policy
└── forward request
```

### 11.2 Forwarder

Forwarder sudah memisahkan non-stream dan stream. Ini bagus. Tetapi perlu diperiksa:

- Streaming flush.
- Header forwarding policy.
- Sensitive header handling.
- Error response compatibility dengan OpenAI API.
- Request body rewrite support.

### 11.3 Config

Config default cukup lengkap, tetapi ada masalah:

- `metadata_transport` default `model_alias` belum diimplementasikan nyata.
- `BindLocalhostOnly` belum ditegakkan.
- `AllowLANAccess` belum mempengaruhi bind address.
- Partial update tidak didukung oleh backend.

### 11.4 Runtime

Runtime state sudah ada sebagai konsep, tetapi Admin API belum terhubung ke state. Ini perlu dijadikan prioritas.

### 11.5 Observability

Logger in-memory bisa berguna untuk dashboard. Namun privacy policy harus lebih ketat.

---

## 12. Audit Frontend Detail

### 12.1 Kekuatan frontend

Frontend sudah cukup ambisius. Ada halaman seperti:

- Dashboard
- Routing Settings
- Route Profiles
- Runtime
- Startup
- Privacy
- Logs
- Downstream Settings
- Advanced Settings
- Setup Wizard

Ini bagus untuk user lokal yang tidak ingin edit YAML manual.

### 12.2 Masalah frontend utama

Masalah paling besar adalah asumsi bahwa backend menerima partial config. Beberapa halaman mengirim sebagian config saja. Backend tidak menerima patch/partial dengan benar.

### 12.3 Rekomendasi frontend

- Tambahkan error toast yang terlihat user, bukan hanya `console.error`.
- Disable save jika config belum loaded.
- Gunakan deep merge sebelum kirim full config.
- Sinkronkan state runtime UI dengan endpoint status sebenarnya.
- Tampilkan label alpha/beta jika fitur belum final.

---

## 13. Audit Security

### 13.1 Security posture saat ini

Security posture saat ini masih lemah karena:

1. Admin auth belum ada.
2. LAN bind policy belum ditegakkan.
3. Privacy logging belum konsisten.
4. Config import belum atomic.
5. Local token/hash belum diimplementasikan.

### 13.2 Risiko keamanan utama

| Risiko | Dampak | Prioritas |
|---|---|---|
| Admin API terbuka tanpa auth | Config bisa diubah pihak lain | Tinggi |
| Server bind ke LAN tanpa proteksi | Dashboard terekspos jaringan | Tinggi |
| Metadata tetap dicatat | Privacy claim tidak valid | Tinggi |
| Config import tidak atomic | Config rusak | Sedang |
| Secret masking hanya di beberapa tempat | Potensi bocor via log/export | Sedang |

### 13.3 Rekomendasi minimum security sebelum release alpha

Sebelum release alpha publik, minimal lakukan:

- Force localhost bind secara default.
- Tambah admin token local.
- Tambah warning besar jika `host=0.0.0.0`.
- Terapkan privacy logging policy.
- Pastikan exported config tidak menyertakan secret/API key asli.
- Validasi import config secara atomic.

---

## 14. Audit Testing

### 14.1 Kondisi test saat ini

Ada 9 file test Go:

```text
internal/analyzer/analyzer_test.go
internal/classifier/classifier_test.go
internal/classifier/eval_test.go
internal/config/config_test.go
internal/forwarder/forwarder_test.go
internal/observability/observability_test.go
internal/routing/routing_test.go
internal/runtime/runtime_test.go
internal/server/server_test.go
```

Ini positif. Artinya project sudah punya niat QA.

### 14.2 Masalah test coverage secara konsep

Test yang perlu ditambahkan:

1. **Model alias rewrite test**

Input:

```json
{"model":"smart-auto","messages":[...]}
```

Expected downstream body:

```json
{"model":"combo.backend","messages":[...]}
```

2. **Header transport test**

Expected downstream header:

```http
X-Route-Intent: combo.backend
```

3. **Runtime start/stop integration test**

- Stop runtime.
- Request `/v1/chat/completions` harus 503.
- Start runtime via admin API.
- Request harus diterima.

4. **Partial config update test**

- Kirim `{ "logging": {...} }`.
- Pastikan config lain tidak hilang.

5. **Streaming flush test**

- Pastikan writer wrapper tetap support `http.Flusher`.

6. **Admin auth test**

- Tanpa token: 401.
- Token salah: 401.
- Token benar: 200.

7. **LAN binding validation test**

- `allow_lan_access=false` + `host=0.0.0.0` harus ditolak atau dipaksa localhost.

### 14.3 Rekomendasi CI

Tambahkan workflow `.github/workflows/ci.yml`:

```yaml
name: CI

on:
  pull_request:
  push:
    branches: [main]

jobs:
  test:
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
      - name: Build frontend
        run: |
          cd web
          npm ci
          npm run build
      - name: Run Go tests
        run: go test ./...
```

---

## 15. Gap Dokumen vs Implementasi

| Area | Klaim/Harapan di docs | Implementasi aktual | Gap |
|---|---|---|---|
| Smart routing | `smart-auto` memilih route otomatis | Decision dibuat, tetapi default model_alias belum rewrite body | Besar |
| Metadata transport | `model_alias` didukung | Hanya `header` yang punya aksi nyata | Besar |
| Runtime control | Start/stop/restart dari UI | Hanya ubah config, bukan runtime state | Besar |
| Local Web UI | Settings usable | Save partial config bisa gagal | Besar |
| Security | Admin auth/local security | Security package placeholder | Besar |
| Auto-start | Run at login | Fungsi registry ada tapi tidak dipanggil dari admin update | Sedang-besar |
| Streaming | Streaming passthrough | Wrapper tidak support flush | Besar |
| Build/release | Portable release | Workflow ada, test belum ada | Sedang |
| Privacy | Privacy-first logging | Setting belum sepenuhnya ditegakkan | Sedang-besar |
| Cross-platform | Go app dengan Windows-first | Startup registry tidak diberi build tag | Sedang |

---

## 16. Rekomendasi Roadmap Perbaikan

### Sprint 1 — Fix fitur inti proxy dan routing

Target: smart routing benar-benar bekerja.

Checklist:

- [ ] Implement `model_alias` rewrite body.
- [ ] Pastikan request body yang diteruskan ke downstream memakai model hasil route.
- [ ] Tambahkan unit test untuk rewrite model.
- [ ] Tambahkan integration test dengan fake downstream server.
- [ ] Pastikan fallback route bekerja jika confidence rendah.
- [ ] Pastikan explicit override tetap bisa melewati smart routing.

Estimasi prioritas: wajib sebelum semua hal lain.

---

### Sprint 2 — Fix runtime control dan dashboard status

Target: tombol runtime benar-benar sesuai state.

Checklist:

- [ ] Tambahkan `RuntimeState` ke `AdminDeps`.
- [ ] Update `runtime/start` agar memanggil state start.
- [ ] Update `runtime/stop` agar memanggil state stop.
- [ ] Update `runtime/restart` agar benar-benar restart.
- [ ] Update `/admin/api/status` agar membaca runtime state.
- [ ] Tambahkan test runtime API.

---

### Sprint 3 — Fix config update dan import/export

Target: Web UI settings usable.

Checklist:

- [ ] Backend menerima partial config patch atau frontend mengirim full deep-merged config.
- [ ] Tambahkan validation full setelah update config/routes/profiles.
- [ ] Buat import config atomic.
- [ ] Cek error semua `config.Save(...)`.
- [ ] Tambahkan test partial update.
- [ ] Tambahkan test invalid import.

---

### Sprint 4 — Fix streaming compatibility

Target: streaming real-time.

Checklist:

- [ ] Tambahkan `Flush()` ke response writer wrapper.
- [ ] Tambahkan test bahwa wrapped writer support `http.Flusher`.
- [ ] Test SSE chunk diteruskan secara bertahap.
- [ ] Validasi compatibility dengan Cursor/Cline/Continue jika memungkinkan.

---

### Sprint 5 — Minimum security hardening

Target: aman untuk alpha publik.

Checklist:

- [ ] Implement admin auth token.
- [ ] Force localhost bind jika LAN disabled.
- [ ] Tambahkan warning UI untuk LAN access.
- [ ] Terapkan privacy logging policy.
- [ ] Pastikan export config masking secret.
- [ ] Tambahkan security tests.

---

### Sprint 6 — Build, packaging, dan release hardening

Target: release alpha yang bisa diuji user.

Checklist:

- [ ] Ubah `Version` dari `const` menjadi `var`.
- [ ] Tambahkan CI workflow.
- [ ] Fix PowerShell scripts atau rename ke `.bat`.
- [ ] Fix lint path frontend.
- [ ] Tambahkan command `stop` di npm wrapper.
- [ ] Tambahkan mekanisme binary install/download untuk npm wrapper.
- [ ] Pastikan `web/dist` dibuild sebelum Go build.
- [ ] Pisahkan Windows startup code dengan build tag.

---

## 17. Checklist Kelayakan MVP

Project baru layak disebut MVP jika checklist berikut selesai:

### Fitur inti

- [ ] `/v1/chat/completions` bekerja untuk non-streaming.
- [ ] `/v1/chat/completions` bekerja untuk streaming real-time.
- [ ] `smart-auto` benar-benar mengubah route/model downstream.
- [ ] Fallback route bekerja.
- [ ] Explicit route override bekerja.
- [ ] Error downstream diteruskan dengan format yang jelas.

### Web UI

- [ ] Dashboard status akurat.
- [ ] Runtime start/stop/restart benar-benar bekerja.
- [ ] Downstream settings bisa disimpan.
- [ ] Routing settings bisa disimpan.
- [ ] Privacy settings bisa disimpan.
- [ ] Startup settings benar-benar mengubah startup OS.
- [ ] Config import/export aman dan tervalidasi.

### Security

- [ ] Admin API protected minimal token.
- [ ] Localhost-only ditegakkan default.
- [ ] LAN access butuh opt-in eksplisit.
- [ ] Secret/API key tidak bocor di logs/export.
- [ ] Privacy mode strict benar-benar mengurangi logging.

### Build/release

- [ ] `go test ./...` lolos di CI.
- [ ] Frontend build lolos.
- [ ] Go binary build setelah `web/dist` tersedia.
- [ ] Release artifact berisi binary dan docs.
- [ ] npm wrapper bisa menjalankan binary atau memberi instruksi jelas.

### Documentation

- [ ] README menjelaskan status alpha.
- [ ] QUICKSTART sesuai behavior aktual.
- [ ] Troubleshooting mencakup 9Router down, port conflict, invalid config.
- [ ] Phase checklist disesuaikan dengan kondisi nyata.

---

## 18. Prioritas Bug Fix Berdasarkan Risiko

| Prioritas | Issue | Alasan |
|---:|---|---|
| 1 | Implement model alias rewrite/header route nyata | Fitur inti produk |
| 2 | Fix runtime control state | Web UI runtime saat ini misleading |
| 3 | Fix partial config update | Banyak settings UI berpotensi gagal |
| 4 | Fix streaming flush | Compatibility AI assistant |
| 5 | Implement admin auth | Security dasar |
| 6 | Enforce localhost/LAN policy | Mencegah exposure admin API |
| 7 | Apply privacy logging policy | Menjaga klaim privacy |
| 8 | Atomic config import | Mencegah config rusak |
| 9 | Fix release CI/test | Mencegah release broken |
| 10 | Fix scripts/npm wrapper | Developer/user experience |

---

## 19. Saran Revisi Status Phase

Status saat ini di dokumen sebaiknya diubah menjadi lebih realistis:

| Phase | Nama | Status audit yang disarankan |
|---|---|---|
| Phase 0 | Technical Foundation | Done |
| Phase 1 | Core Proxy MVP | Partial, needs streaming hardening |
| Phase 2 | Routing Intelligence MVP | Partial, model_alias integration incomplete |
| Phase 3 | Local Web UI | Partial, config save/runtime control incomplete |
| Phase 4 | Tray and Auto-Start | In Progress |
| Phase 5 | Packaging and Distribution | In Progress |
| Phase 6 | QA, Compatibility, and Hardening | Not Ready / Not Started fully |
| Phase 7 | MVP Release | Not Ready |
| Phase 8 | Post-MVP Improvements | Future |

---

## 20. Kesimpulan Final

AtlasBridge memiliki konsep yang menarik dan dokumentasi yang jauh lebih matang dibanding banyak prototype sejenis. Struktur project juga menunjukkan arah engineering yang cukup rapi. Namun, gap antara dokumen dan implementasi masih cukup besar.

Temuan terpenting adalah:

1. **Smart routing belum efektif di default mode `model_alias`.**
2. **Runtime control UI belum mengubah runtime state aktif.**
3. **Frontend mengirim partial config, sementara backend mengharapkan full config.**
4. **Streaming SSE berpotensi tidak real-time karena wrapper tidak support flush.**
5. **Admin security masih placeholder.**
6. **LAN/local binding policy belum ditegakkan.**
7. **Release workflow belum menjalankan test.**

Penilaian akhir:

> **AtlasBridge layak dilanjutkan, tetapi belum layak dirilis sebagai MVP publik. Status yang paling tepat adalah alpha prototype yang perlu hardening sprint.**

Dengan memperbaiki 4 masalah P0 terlebih dahulu, project ini bisa naik signifikan dari prototype menjadi MVP internal yang cukup kuat. Setelah security, CI, dan packaging diperbaiki, barulah project ini masuk akal untuk alpha release publik.

---

## 21. Lampiran: Bukti File Penting

| Area | File | Catatan |
|---|---|---|
| Config default | `internal/config/config.go` | Default `metadata_transport` adalah `model_alias`. |
| Request handler | `internal/server/server.go` | Hanya mode `header` yang menambahkan `X-Route-Intent`. |
| Runtime admin | `internal/server/admin.go` | Start/stop hanya mengubah config. |
| Runtime state | `internal/runtime/runtime.go` | State ada, tetapi tidak dipakai admin handler. |
| Middleware | `internal/server/middleware.go` | Response writer wrapper tidak punya `Flush()`. |
| Streaming | `internal/forwarder/forwarder.go` | Forwarder membutuhkan writer yang bisa flush. |
| Security | `internal/security/security.go` | Masih placeholder. |
| Startup | `internal/startup/startup.go` | Registry function ada, tetapi perlu build tag dan dipanggil dari admin. |
| Frontend config store | `web/src/stores/config.ts` | `saveConfig` menerima `Partial<AppConfig>`. |
| Frontend settings | `web/src/pages/*.vue` | Banyak halaman mengirim partial config. |
| Embedded UI | `web/assets.go` | `//go:embed dist`, perlu `web/dist` sebelum Go build. |
| Release | `.github/workflows/release.yml` | Build ada, test belum ada. |
| Go version | `go.mod` | Membutuhkan Go `1.25.5`. |

---

## 22. Action Plan Sangat Singkat

Jika hanya punya waktu sedikit, lakukan urutan ini:

1. Fix `model_alias` rewrite.
2. Fix runtime state admin handlers.
3. Fix config partial update.
4. Fix streaming flush.
5. Add admin auth minimal.
6. Add CI test.
7. Revisi docs checklist agar tidak overclaim.

Setelah itu baru siapkan label release:

```text
v0.1.0-alpha.1
```

Dengan catatan jelas:

```text
Alpha release. Local testing only. Not production ready.
```
