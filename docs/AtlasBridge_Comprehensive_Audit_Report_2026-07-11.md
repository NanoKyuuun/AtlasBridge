# Laporan Audit Komprehensif AtlasBridge

**Tanggal audit:** 11 Juli 2026  
**Peran audit:** Software Architecture, Application Security, DevOps, Reliability Engineering  
**Objek:** Repositori `AtlasBridge` dari arsip yang diunggah  
**Status:** Audit statis mendalam, penelusuran alur eksekusi, build frontend, pemeriksaan dependency frontend, dan pengujian parsial backend

---

## 1. Ringkasan Eksekutif

### Kesimpulan utama

AtlasBridge memiliki fondasi arsitektur yang cukup baik untuk proyek tahap awal: backend Go dipisahkan ke paket analyzer, classifier, routing, forwarder, server, runtime, security, config, observability, dan startup; frontend menggunakan Vue 3/TypeScript; serta terdapat test suite yang luas untuk routing, forwarding, streaming, konfigurasi, dan perilaku admin.

Namun, dalam kondisi kode saat ini, AtlasBridge **belum layak diaktifkan pada LAN atau digunakan sebagai gateway produksi**. Risiko terbesar bukan pada SQL injection atau penyimpanan API key hardcoded, melainkan pada **control plane admin yang tidak diautentikasi secara default, tidak adanya proteksi origin/cross-site, state konfigurasi yang dimutasi secara langsung saat request berjalan, buffering body tanpa batas, split-brain antara konfigurasi tersimpan dan forwarder aktif, serta lifecycle streaming/shutdown yang belum aman**.

### Putusan kesiapan

| Skenario | Putusan |
|---|---|
| Pengembangan lokal, localhost, satu pengguna | Dapat dipakai secara terbatas setelah blocker P0 diperbaiki |
| Penggunaan lokal harian dengan data sensitif | Belum direkomendasikan sebelum auth, body limit, state snapshot, dan stream handling diperbaiki |
| LAN kantor/rumah | Tidak direkomendasikan |
| Internet-facing/public | Tidak layak |
| Produksi multi-user/high concurrency | Tidak layak |

### Skor audit internal

Skor berikut merupakan indikator prioritas internal, bukan sertifikasi formal.

| Dimensi | Skor | Penilaian |
|---|---:|---|
| Keamanan | 42/100 | Lemah pada control plane dan boundary jaringan |
| Arsitektur | 65/100 | Modular, tetapi runtime state belum dikelola secara konsisten |
| Performa | 48/100 | Banyak full-buffer copy dan tanpa limit/backpressure |
| Keandalan/resiliensi | 45/100 | Streaming, reload, lock, dan shutdown perlu redesign |
| Maintainability | 68/100 | Struktur paket dan test cukup baik; beberapa setting hanya dekoratif |
| Testing/DevOps | 61/100 | Banyak test, tetapi belum ada CI quality gate dan race/security test |
| **Keseluruhan** | **55/100** | **Fondasi baik, tetapi belum production-ready** |

### Ringkasan temuan

- **1 temuan kritis bersyarat**
- **6 temuan tingkat tinggi**
- **10 temuan tingkat menengah**
- **6 temuan tingkat rendah**
- **Tidak ditemukan indikasi jelas API key/password hardcoded** dari pencarian pola statis.
- **SQL injection tidak relevan pada implementasi saat ini** karena tidak ditemukan lapisan database/query.
- **Tidak ditemukan sink XSS eksplisit** seperti `v-html`, `innerHTML`, atau `eval` pada frontend yang ditinjau.
- Raw prompt secara default tidak disimpan ke observability entry; ini merupakan praktik positif.
- `Authorization` tidak diteruskan ke downstream; ini juga praktik positif.

---

## 2. Ruang Lingkup dan Metode

### Struktur dan stack yang teridentifikasi

- Backend: Go, Chi router
- Frontend: Vue 3, TypeScript, Vite, Pinia, Tailwind/DaisyUI
- Runtime: desktop/local proxy dengan system tray
- Data plane:
  - `POST /v1/chat/completions`
  - `GET /v1/models`
- Control plane:
  - `/admin/api/*`
  - konfigurasi, route, profile, runtime, log, diagnostics, import/export/reset
- Downstream default:
  - 9Router pada `http://127.0.0.1:20128/v1`

### Ukuran codebase

- 35 file Go
- sekitar 8.605 baris Go
- 31 file Vue/TypeScript
- sekitar 4.009 baris Vue/TypeScript
- 10 file test Go
- 210 fungsi test Go

### Validasi yang berhasil

- `npm ci --ignore-scripts`: berhasil
- `npm run build`: berhasil
- `npm audit`: 0 temuan known vulnerability pada dependency frontend saat audit
- Test paket standard-library-only berikut berhasil dijalankan pada salinan kompatibilitas Go 1.23.2:
  - `internal/security`
  - `internal/observability`
  - `internal/runtime`
  - `internal/analyzer`
  - `internal/classifier`

### Batasan audit dinamis

Repositori mendeklarasikan Go 1.25.5. Lingkungan audit hanya menyediakan Go 1.23.2 dan tidak dapat mengunduh toolchain/modul dari `proxy.golang.org`, sehingga seluruh `go test ./...`, `go test -race ./...`, `go vet ./...`, dan `govulncheck ./...` belum dapat dieksekusi. Karena itu, status seluruh test backend dan vulnerability dependency Go **belum dapat dinyatakan lulus**. Temuan race di bawah didasarkan pada alur mutasi/read source code yang nyata dan harus dikonfirmasi lagi dengan test konkurensi + `-race` pada CI.

---

# 3. Masalah Kritis dan Tingkat Tinggi

## C-01 — Control plane admin terbuka secara default dan tidak memiliki boundary cross-origin yang memadai

**Severity:** Kritis bersyarat; Tinggi pada localhost ketat  
**Kategori:** Broken Access Control, CSRF/local service abuse, unsafe network exposure  
**Lokasi utama:**

- `internal/config/config.go:98-102`
- `internal/server/server.go:103-147`
- `internal/security/security.go:51-74`
- `internal/server/admin.go:506-561`
- `internal/app/app.go:33-37`

### Bukti

Konfigurasi default menetapkan:

```go
AdminAuthEnabled:  false,
BindLocalhostOnly: true,
AllowLANAccess:    false,
```

Seluruh endpoint admin memang dipasangi middleware `AdminAuth`, tetapi middleware melewatkan request tanpa pemeriksaan saat `enabled == false`. Dengan demikian, pada instalasi default semua operasi admin dapat dipanggil tanpa token:

- mengganti konfigurasi;
- mengganti route/profile;
- start/stop/restart;
- import/export/reset konfigurasi;
- menghapus log;
- menjalankan combo test ke downstream.

Tidak ditemukan validasi `Origin`, `Referer`, atau `Host` untuk state-changing request. Tidak ada kewajiban `Content-Type: application/json`. Endpoint `POST /admin/api/config/import` membaca body apa pun dan langsung mencoba `json.Unmarshal`. Ini membuka peluang halaman web jahat untuk mencoba mengirim simple cross-origin request dengan `text/plain`, tergantung kebijakan browser/Private Network Access dan topologi jaringan. Serangan tidak selalu dapat membaca respons karena CORS, tetapi perubahan state dapat tetap menjadi sasaran. Selain itu, proses lokal lain dan host LAN dapat memanggil endpoint secara langsung.

Risiko meningkat karena `downstream.base_url` hanya divalidasi sebagai URL `http/https` dengan host nonkosong. Bila control plane dapat dikendalikan, AtlasBridge dapat diarahkan untuk membuat request server-side ke tujuan internal yang tidak dimaksudkan.

### Dampak

- Pengambilalihan konfigurasi AtlasBridge
- Penghentian proxy
- Perubahan rute model
- Penyalahgunaan kuota/akses downstream
- Potensi SSRF ke host internal/private
- Eksposur data operasional melalui diagnostics/export
- Perubahan keamanan secara diam-diam

### Solusi wajib

1. Aktifkan admin authentication secara default.
2. Jangan izinkan LAN aktif kecuali admin auth aktif.
3. Tambahkan auth terpisah untuk data plane `/v1`, terutama pada mode LAN.
4. Tolak state-changing browser request dengan origin tidak dipercaya.
5. Wajibkan `application/json` untuk endpoint JSON.
6. Validasi `Host` terhadap listener yang diharapkan.
7. Tambahkan header keamanan UI: CSP, `frame-ancestors`, `X-Content-Type-Options`, dan `Referrer-Policy`.
8. Batasi downstream default ke loopback/allowlist; remote downstream harus menjadi mode eksplisit.
9. Validasi redirect dan resolved IP agar tidak berpindah ke link-local/private address di luar kebijakan.

### Patch desain yang disarankan

```go
func RequireJSON(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodPost ||
           r.Method == http.MethodPut ||
           r.Method == http.MethodPatch {
            mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
            if err != nil || mediaType != "application/json" {
                http.Error(w, "application/json required", http.StatusUnsupportedMediaType)
                return
            }
        }
        next.ServeHTTP(w, r)
    })
}
```

Tambahkan guard origin khusus control plane. CLI non-browser tetap dapat diizinkan ketika membawa Bearer token valid.

```go
func SameOriginAdmin(allowedOrigin string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            origin := r.Header.Get("Origin")
            if origin != "" && origin != allowedOrigin {
                http.Error(w, "forbidden origin", http.StatusForbidden)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

### Kriteria selesai

- Instalasi baru selalu menghasilkan token admin.
- Semua endpoint admin gagal dengan `401` tanpa token.
- Cross-origin state-changing request gagal dengan `403`.
- Mode LAN gagal diaktifkan bila auth/TLS policy belum memenuhi syarat.
- Test mencakup `Origin: https://evil.example`, `text/plain`, malformed Bearer, dan direct LAN access policy.

---

## H-01 — Hash token yang dimasking dapat ditulis kembali dan mengunci admin secara permanen

**Severity:** Tinggi  
**Kategori:** Authentication state corruption, availability, credential lifecycle  
**Lokasi:**

- `internal/server/admin.go:37-40`
- `internal/server/admin.go:78-80`
- `internal/server/admin.go:108-126`
- `internal/server/admin.go:492-503`
- `internal/server/admin.go:604-609`
- `web/src/pages/AdvancedSettings.vue:178-193`

### Bukti alur

1. `GET /admin/api/config` memanggil `maskConfig`.
2. `maskConfig` mengganti `admin_token_hash` dengan bintang dan empat karakter terakhir.
3. Frontend memasukkan nilai masking itu ke `security.value`.
4. Saat pengguna menyimpan Advanced Settings, frontend mengirim seluruh objek `security`, termasuk hash masking.
5. Backend melakukan unmarshal ke `merged.Security`.
6. `EnsureToken` melihat string tidak kosong, sehingga tidak membuat token baru.
7. `AuthConfig.Set` dan `config.Save` menyimpan hash masking sebagai verifier baru.
8. Token lama tidak mungkin cocok lagi.

Ekspor konfigurasi juga memasukkan hash masking. Ekspor lalu impor dapat menimbulkan korupsi verifier yang sama.

### Dampak

- Seluruh token admin lama invalid.
- Pengguna dapat terkunci dari admin UI setelah sekadar mengubah timeout/host.
- Lockout bertahan setelah restart karena nilai masking tersimpan ke disk.
- Recovery memerlukan edit manual file konfigurasi.

### Solusi

**Jangan pernah masukkan verifier ke DTO publik.** Pisahkan model persistence dari model API.

```go
type SecurityView struct {
    AdminAuthEnabled  bool `json:"admin_auth_enabled"`
    TokenConfigured   bool `json:"token_configured"`
    BindLocalhostOnly bool `json:"bind_localhost_only"`
    AllowLANAccess    bool `json:"allow_lan_access"`
}

type SecurityUpdate struct {
    AdminAuthEnabled  *bool `json:"admin_auth_enabled,omitempty"`
    BindLocalhostOnly *bool `json:"bind_localhost_only,omitempty"`
    AllowLANAccess    *bool `json:"allow_lan_access,omitempty"`
}
```

Buat endpoint khusus:

```text
POST /admin/api/security/token/rotate
```

Endpoint tersebut menghasilkan token baru satu kali, menyimpan hash sebenarnya secara atomik, lalu mengembalikan raw token sekali saja.

### Kriteria selesai

- Field `admin_token_hash` tidak pernah muncul pada GET/export.
- PUT konfigurasi tidak menerima `admin_token_hash`.
- Export→import tidak mengubah credential.
- Tersedia test regresi “load advanced settings → save → token lama tetap valid”.
- Tersedia test rotate token.

---

## H-02 — Shared mutable configuration menimbulkan data race dan potensi crash concurrent map access

**Severity:** Tinggi  
**Kategori:** Concurrency safety, integrity, availability  
**Lokasi:**

- `internal/server/server.go:56-63`
- `internal/server/server.go:290-293`
- `internal/server/admin.go:126`
- `internal/server/admin.go:168`
- `internal/server/admin.go:203`
- `internal/server/admin.go:529-555`
- `internal/routing/routing.go:91-129`

### Bukti

Server menyimpan pointer global ke:

- `*config.Config`
- `*config.RoutesConfig`
- `*config.ProfilesConfig`

Request proxy membaca config, route, profile, dan map routing saat request berjalan. Pada saat yang sama, endpoint admin mengganti object melalui assignment langsung:

```go
*deps.Config = merged
*deps.Routes = updated
*deps.Profiles = updated
```

Tidak ada `sync.RWMutex`, immutable snapshot, ataupun atomic swap untuk keseluruhan state. `AuthConfig` sudah memakai lock, tetapi config/routes/profiles tidak.

### Dampak

- Data race
- Keputusan routing menggunakan campuran versi lama dan baru
- Potensi fatal error ketika map dibaca saat diganti/dimutasi
- Respons tidak deterministik
- Sulit direproduksi pada test biasa

### Solusi arsitektural

Gunakan immutable runtime snapshot dan atomic swap.

```go
type Snapshot struct {
    Config    config.Config
    Routes    config.RoutesConfig
    Profiles  config.ProfilesConfig
    Forwarder *forwarder.Forwarder
    Version   uint64
}

type StateStore struct {
    current   atomic.Pointer[Snapshot]
    persistMu sync.Mutex
}

func (s *StateStore) Load() *Snapshot {
    return s.current.Load()
}
```

Alur update:

1. Decode ke object baru.
2. Deep-copy seluruh map.
3. Validasi `ValidateFull`.
4. Bangun forwarder baru bila downstream berubah.
5. Simpan seluruh file secara transaksional/atomik.
6. Setelah persistence berhasil, `current.Store(next)`.
7. Request yang sudah berjalan menyelesaikan pekerjaan dengan snapshot lama; request baru memakai snapshot baru.

### Kriteria selesai

- Tidak ada lagi mutasi langsung config/routes/profiles yang sedang dibaca.
- `go test -race ./...` lulus.
- Ada test yang menjalankan ratusan request proxy sambil mengubah route/profile/config berulang.
- Setiap request menggunakan satu `snapshot.Version` yang konsisten.

---

## H-03 — Body request dan response dibuffer tanpa batas; mudah menyebabkan memory exhaustion

**Severity:** Tinggi  
**Kategori:** DoS, resource exhaustion, performance  
**Lokasi:**

- `internal/server/server.go:275`
- `internal/server/server.go:395-404`
- `internal/forwarder/forwarder.go:67-92`
- `internal/forwarder/forwarder.go:109-140`
- `internal/server/admin.go:46`
- `internal/server/admin.go:151`
- `internal/server/admin.go:186`
- `internal/server/admin.go:384`
- `internal/server/admin.go:427`
- `internal/server/admin.go:508`

### Bukti

Alur chat melakukan beberapa kali full buffering/copy:

1. `io.ReadAll(r.Body)` di server.
2. JSON di-unmarshal beberapa kali.
3. Analyzer membaca body.
4. `strings.ToLower(string(body))` membuat salinan.
5. Forwarder kembali menjalankan `io.ReadAll(req.Body)`.
6. `strings.NewReader(string(body))` membuat salinan string.
7. Response nonstream kembali `io.ReadAll(resp.Body)` tanpa limit.

Admin import/config/routes/profile/dry-run/combo juga memakai `io.ReadAll` tanpa batas.

Tidak ada global request size cap, response size cap, semaphore concurrent request, atau memory budget.

### Dampak

- Satu request besar dapat membuat beberapa salinan payload di heap.
- Banyak request paralel dapat menyebabkan GC pressure, latency spike, OOM, atau crash.
- `/v1` tidak diautentikasi, sehingga risiko meningkat saat listener dapat dijangkau pengguna lain.
- Response downstream yang sangat besar juga dapat menghabiskan memori.

### Solusi

Gunakan `http.MaxBytesReader` di boundary HTTP dan batas berbeda per endpoint.

```go
const (
    maxChatBody   = 16 << 20 // sesuaikan kebutuhan long-context
    maxAdminBody  = 1 << 20
    maxImportBody = 8 << 20
)

func readLimitedJSON(w http.ResponseWriter, r *http.Request, max int64) ([]byte, error) {
    r.Body = http.MaxBytesReader(w, r.Body, max)
    return io.ReadAll(r.Body)
}
```

Tambahkan:

- `413 Payload Too Large`;
- bounded response reader untuk nonstream;
- `bytes.NewReader(body)`, bukan `strings.NewReader(string(body))`;
- satu kali decode ke typed request envelope;
- concurrency semaphore;
- per-client/IP rate limit pada mode LAN;
- stream response langsung tanpa full buffering.

### Kriteria selesai

- Body di atas limit mendapat `413`.
- Test memastikan memory tidak bertumbuh linear akibat multiple copy.
- Load test menunjukkan RSS stabil pada concurrency target.
- Ada batas maksimum request aktif dan queue timeout.

---

## H-04 — Konfigurasi tersimpan dapat berbeda dari perilaku runtime aktif

**Severity:** Tinggi  
**Kategori:** Configuration integrity, split-brain, operational correctness  
**Lokasi:**

- `internal/app/app.go:55-77`
- `internal/server/admin.go:44-140`
- `internal/server/admin.go:506-585`

### Bukti

Forwarder dibuat satu kali pada startup:

```go
fwd, err := forwarder.New(cfg.Downstream.BaseURL, cfg.Downstream.TimeoutSeconds)
```

Ketika admin mengubah `downstream.base_url` atau `timeout_seconds`, handler hanya memutasi config dan menyimpan file. `deps.Forwarder` tidak dibangun ulang. Akibatnya:

- UI dan diagnostics dapat menampilkan URL baru;
- file config berisi URL baru;
- request proxy masih dikirim ke URL lama dengan timeout lama.

Hal serupa terjadi untuk host/port. Perubahan config tidak melakukan rebind listener. Endpoint “restart” hanya mengubah state, bukan merekonstruksi server/listener/forwarder.

Import/reset juga nontransaksional:

- config diterapkan sebelum seluruh bundle divalidasi;
- error `config.Save(deps.Config)` pada import diabaikan;
- validasi referensi route/profile dilewati bila hanya salah satu bagian diimpor;
- AuthConfig tidak selalu disinkronkan;
- kegagalan di tengah menyisakan memory/disk setengah berubah.

### Dampak

- Operator percaya perubahan sudah aktif padahal belum.
- Health/status dapat menampilkan state yang salah.
- Restart semu tidak memperbaiki keadaan.
- Crash/restart berikutnya mengubah perilaku secara mendadak.
- Import gagal dapat merusak konsistensi.

### Solusi

- Gunakan `StateStore` immutable pada H-02.
- Bangun seluruh “candidate runtime” sebelum commit.
- Terapkan config transaction:
  1. decode;
  2. merge;
  3. validate full;
  4. construct dependencies;
  5. write temp files;
  6. fsync;
  7. atomic rename;
  8. atomic state swap.
- Bedakan setting:
  - **hot-reloadable**: routing, profile, timeout, downstream client;
  - **restart-required**: host, port, TLS listener.
- API harus mengembalikan `restart_required: true` untuk setting listener.
- “Restart” harus benar-benar membuat listener/runtime baru atau nama aksi diubah menjadi “start/stop routing”.

### Kriteria selesai

- Mengubah downstream langsung mengubah tujuan request berikutnya.
- Gagal menyimpan tidak mengubah state runtime.
- Import bersifat all-or-nothing.
- Status menampilkan `effective_config_version`.
- Test memverifikasi tujuan downstream sebelum/sesudah hot reload.

---

## H-05 — Implementasi streaming dapat memutus stream valid dan menulis error ke response yang sudah dimulai

**Severity:** Tinggi  
**Kategori:** Streaming reliability, protocol correctness  
**Lokasi:**

- `internal/forwarder/forwarder.go:28-37`
- `internal/forwarder/forwarder.go:121-159`
- `internal/server/server.go:313-329`
- `internal/server/server.go:491-493`

### Bukti

1. Satu `http.Client` memakai `Timeout` global 120 detik. Pada Go, timeout client mencakup keseluruhan pertukaran termasuk pembacaan response body. Stream AI yang valid lebih dari 120 detik dapat dihentikan.
2. SSE dibaca dengan `bufio.Scanner` dengan token maksimum 1 MiB. Satu event/line lebih besar dari itu menghasilkan error.
3. `headersWritten(w)` selalu mengembalikan `false`.
4. Bila error terjadi setelah header/event sudah dikirim, handler tetap mencoba menulis JSON 502 ke response SSE yang telah dimulai.
5. Forwarder memaksa `Content-Type: text/event-stream` bahkan jika downstream mengembalikan error JSON.
6. Header response downstream untuk stream tidak dipropagasikan secara selektif.

### Dampak

- Stream panjang berhenti secara tidak terduga.
- Client menerima campuran SSE dan JSON.
- Status code tidak dapat diubah setelah header terkirim.
- Error downstream sulit ditangani dengan benar.
- Chunk besar gagal walau valid.

### Solusi

- Pisahkan client nonstream dan stream.
- Untuk stream, jangan gunakan total `Client.Timeout`; gunakan:
  - `DialContext` timeout;
  - TLS handshake timeout;
  - `ResponseHeaderTimeout`;
  - context cancellation;
  - idle-progress watchdog bila diperlukan.
- Ganti Scanner dengan reader yang dapat menangani event besar/segmentasi.
- Track commit status pada wrapped response writer.
- Sebelum menulis header ke client, periksa status/content type downstream.
- Setelah stream dimulai, log dan tutup koneksi; jangan mencoba mengirim JSON error kedua.

Contoh status writer:

```go
type commitWriter struct {
    http.ResponseWriter
    committed bool
}

func (w *commitWriter) WriteHeader(code int) {
    if !w.committed {
        w.committed = true
        w.ResponseWriter.WriteHeader(code)
    }
}
```

### Kriteria selesai

- Stream 10+ menit tetap aktif selama downstream mengirim data.
- Event di atas 1 MiB tidak gagal karena batas Scanner.
- Tidak ada JSON yang ditulis setelah SSE dimulai.
- Client cancellation segera membatalkan request downstream.
- Test mencakup late stream failure.

---

## H-06 — Mode LAN tidak memiliki autentikasi data plane maupun TLS, dan invariant bind dapat dibypass oleh kombinasi config

**Severity:** Tinggi  
**Kategori:** Network boundary, confidentiality, unauthorized use  
**Lokasi:**

- `internal/app/app.go:33-37`
- `internal/server/server.go:98-106`
- `internal/config/config.go:98-102`

### Bukti

`effectiveHost` hanya memaksa loopback bila dua kondisi sekaligus terpenuhi:

```go
if cfg.Security.BindLocalhostOnly && !cfg.Security.AllowLANAccess {
    return "127.0.0.1"
}
return cfg.Server.Host
```

Konfigurasi berikut dapat bind ke semua interface walau `allow_lan_access == false`:

```yaml
server:
  host: 0.0.0.0
security:
  bind_localhost_only: false
  allow_lan_access: false
```

Selain itu:

- `/v1/chat/completions` tidak memiliki API auth;
- admin Bearer token berjalan melalui HTTP biasa bila LAN diaktifkan;
- tidak ada TLS/mTLS;
- tidak ada rate limit atau quota.

### Dampak

- Pengguna LAN dapat memakai kapasitas downstream tanpa izin.
- Token admin dapat disadap pada jaringan tidak tepercaya.
- Gateway dapat menjadi pivot ke downstream.
- Biaya provider dan data prompt berisiko.

### Solusi

Enforce invariant satu arah:

```go
if !cfg.Security.AllowLANAccess {
    cfg.Server.Host = "127.0.0.1"
}
```

Untuk mode LAN:

- wajib admin auth;
- wajib data-plane API key atau mTLS;
- gunakan TLS langsung atau reverse proxy lokal yang terdokumentasi;
- allowlist CIDR;
- rate limit dan concurrency quota;
- tampilkan warning eksplisit di UI;
- jangan kirim prompt melalui HTTP plaintext.

---

## H-07 — Single-instance lock dan shutdown tidak tahan crash/hang

**Severity:** Tinggi  
**Kategori:** Lifecycle, availability, graceful shutdown  
**Lokasi:**

- `internal/startup/startup.go:25-46`
- `internal/app/app.go:151-160`
- `cmd/atlasbridge/main.go:29-33`

### Bukti

Lock instance memakai pola:

```go
os.Stat(lockFile)
os.Create(lockFile)
```

Ini memiliki race TOCTOU: dua proses dapat sama-sama melihat file belum ada lalu sama-sama membuatnya. Lock `sync.Mutex` hanya berlaku dalam satu proses dan tidak menyinkronkan dua proses berbeda.

Jika proses crash, file lock tertinggal dan startup berikutnya selalu dianggap instance kedua. Tidak ada PID, process liveness check, atau OS-level file lock.

Shutdown memanggil:

```go
a.server.Shutdown(context.Background())
```

tanpa deadline. Stream panjang dapat membuat shutdown menunggu tidak terbatas. Di sisi lain, error `application.Run()` memanggil `os.Exit(1)` dari goroutine sehingga deferred cleanup tidak berjalan.

### Dampak

- Dua instance dapat lolos pada race tertentu.
- Crash meninggalkan stale lock dan mencegah restart.
- Shutdown dapat hang.
- Lock dapat tidak dilepas pada fatal error.

### Solusi

- Gunakan OS-level lock:
  - Windows named mutex/LockFileEx;
  - Unix `flock`;
  - atau library lintas platform yang matang.
- Simpan PID + timestamp hanya sebagai metadata, bukan sebagai lock utama.
- Gunakan context timeout untuk shutdown:

```go
ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
defer cancel()

if err := srv.Shutdown(ctx); err != nil {
    _ = srv.Close()
}
```

- Jangan `os.Exit` dari goroutine. Kirim error ke channel dan koordinasikan shutdown dari `main`.

---

# 4. Masalah Tingkat Menengah dan Rendah

## M-01 — Downstream URL terlalu permisif dan redirect policy tidak dibatasi

**Severity:** Menengah setelah admin auth diperbaiki; bagian dari chain kritis bila control plane terbuka  
**Lokasi:** `internal/config/config.go:220-231`

Validasi hanya memastikan scheme `http/https` dan host nonkosong. Tidak ada:

- allowlist host/port;
- penolakan credentials di URL;
- penolakan fragment/query yang tidak dibutuhkan;
- validasi resolved IP;
- blok link-local/metadata/private address;
- validasi tujuan redirect.

**Rekomendasi:** default hanya loopback 9Router. Remote downstream harus diaktifkan secara eksplisit dengan allowlist. Revalidasi setiap redirect dan resolved address.

---

## M-02 — Error internal dan header downstream dapat bocor ke client

**Severity:** Menengah  
**Lokasi:**

- `internal/server/server.go:322-340`
- `internal/server/server.go:347-353`
- `internal/server/admin.go:291-310`

Raw error downstream dimasukkan ke response:

```go
"message": fmt.Sprintf("downstream error: %v", err)
```

Ini dapat memaparkan URL internal, alamat, detail DNS, atau detail koneksi. Nonstream response menyalin hampir seluruh header downstream kecuali dua header. Header hop-by-hop dan `Set-Cookie` seharusnya tidak diteruskan tanpa kebijakan.

**Rekomendasi:**

- client menerima error code generik + correlation ID;
- detail hanya masuk structured log;
- gunakan response header allowlist;
- strip `Connection`, `Keep-Alive`, `Proxy-*`, `TE`, `Trailer`, `Transfer-Encoding`, `Upgrade`, dan `Set-Cookie` kecuali memang dibutuhkan.

---

## M-03 — Privacy settings memberi kesan perlindungan yang belum diimplementasikan

**Severity:** Menengah  
**Lokasi:**

- `web/src/pages/PrivacySettings.vue:109-124`
- `web/src/pages/PrivacySettings.vue:207-240`
- `internal/config/config.go:117-122`
- `internal/observability/observability.go`

UI menampilkan “Secret redaction — Protected”, tetapi `redactSecrets` hanya `ref(true)` lokal dan tidak dikirim ke backend. `PromptLoggingEnabled`, `RetentionDays`, dan sebagian `PrivacyMode` tidak mengubah pipeline logging secara nyata.

Ini merupakan risiko product security: pengguna dapat mengambil keputusan berdasarkan kontrol yang sebenarnya tidak aktif.

**Rekomendasi:**

- hapus badge “Protected” sampai backend benar-benar menegakkan redaction;
- buat redactor terpusat sebelum semua sink log;
- definisikan perilaku strict/standard/debug;
- retention harus benar-benar diterapkan;
- prompt logging harus explicit opt-in, dengan warning dan expiry otomatis.

---

## M-04 — File konfigurasi tidak private dan penulisannya tidak atomik

**Severity:** Menengah  
**Lokasi:** `internal/config/config.go:152-186`

Direktori dibuat `0755` dan file ditulis `0644`. File saat ini menyimpan hash token, bukan raw token; namun verifier dan topologi konfigurasi tetap sebaiknya private. `os.WriteFile` langsung juga dapat meninggalkan file korup/terpotong jika proses mati saat write.

**Rekomendasi:**

- direktori `0700`;
- file `0600`;
- write ke temp file pada direktori yang sama;
- `Sync`;
- atomic rename;
- serialize writer dengan mutex;
- backup last-known-good.

---

## M-05 — Validasi input dan konstruksi JSON belum cukup kuat

**Severity:** Menengah  
**Lokasi:**

- `internal/server/server.go:372-392`
- `internal/server/admin.go:382-405`
- `internal/server/admin.go:425-452`
- `internal/server/server.go:427-433`

Masalah:

- chat hanya mengecek JSON dan `messages` nonempty;
- tidak ada batas jumlah message, panjang string, nama model, header, atau route key;
- dry-run/combo membuat JSON dengan `fmt.Sprintf`, sehingga quote/newline pada input dapat merusak struktur;
- rewrite model memakai `map[string]interface{}` lalu marshal ulang, yang dapat mengubah representasi unknown field dan angka besar;
- request ID dari client tidak dibatasi panjang/karakter.

**Rekomendasi:**

- typed DTO + `json.Decoder`;
- `DisallowUnknownFields` hanya untuk admin API yang stabil, bukan pass-through OpenAI yang perlu kompatibel;
- gunakan `json.Marshal` untuk body sintetis;
- pertahankan unknown fields dengan `json.RawMessage` atau patch token-level;
- batasi model/request-id/path length.

---

## M-06 — Tidak ada rate limit, concurrency limit, queue limit, atau backpressure

**Severity:** Menengah pada localhost; Tinggi pada LAN  
**Lokasi:** server/data plane secara umum

Walau 9Router bertanggung jawab atas provider failover dan rate limit downstream, AtlasBridge tetap membutuhkan proteksi resource lokal:

- maximum in-flight request;
- queue timeout;
- per-client limit pada LAN;
- separate budget untuk stream/nonstream;
- overload response `429`/`503`.

**Rekomendasi:** gunakan weighted semaphore berdasarkan tipe request/payload dan expose metric saturation.

---

## M-07 — Observability belum cukup untuk operasi dan incident response

**Severity:** Menengah  
**Lokasi:**

- `internal/observability/observability.go`
- `internal/server/server.go:465-488`
- `internal/logging/logging.go`
- `internal/storage/storage.go`

Field seperti latency, status code, dan privacy mode tersedia tetapi tidak diisi lengkap pada jalur utama. Logger masih in-memory. Paket logging/storage/proxy sebagian masih placeholder.

**Rekomendasi:**

- structured JSON log;
- correlation ID internal;
- status, latency, bytes in/out, route, outcome;
- metric Prometheus/OpenTelemetry opsional;
- jangan log raw prompt/header auth;
- redaction sebelum sink;
- health/readiness terpisah;
- downstream circuit state metric.

---

## M-08 — CI/CD belum memiliki quality gate dan supply-chain hardening

**Severity:** Menengah  
**Lokasi:** `.github/workflows/release.yml:1-58`

Workflow hanya berjalan pada tag/manual dan langsung build/release. Belum ada PR CI untuk:

- `go test ./...`;
- `go test -race ./...`;
- `go vet`;
- staticcheck/golangci-lint;
- `govulncheck`;
- frontend typecheck/test/build;
- `npm audit`;
- secret scan;
- CodeQL/SAST;
- SBOM;
- license check.

GitHub Actions direferensikan dengan tag seperti `@v4`, bukan full commit SHA.

**Rekomendasi:**

- buat `ci.yml` pada pull request;
- pin action ke full SHA;
- minimal permission;
- dependency review/Dependabot;
- build reproducible;
- checksum + signing/provenance;
- SBOM CycloneDX/SPDX.

---

## M-09 — Opsi konfigurasi yang tidak efektif menimbulkan configuration drift

**Severity:** Menengah  
**Lokasi:** berbagai

Ditemukan setting yang didefinisikan tetapi tidak menegakkan perilaku:

- `Server.APIBasePath` hanya dipakai untuk log;
- `Server.AdminPath` dipakai untuk log/tray, router tetap hardcoded;
- `Routing.ExplicitOverrideEnabled` tidak digunakan oleh resolver;
- `Startup.RestartAfterCrash` tidak digunakan;
- `Logging.PromptLoggingEnabled` tidak digunakan;
- `Logging.RetentionDays` tidak digunakan;
- `Downstream.Type` tidak digunakan.

**Rekomendasi:** implementasikan atau hapus dari UI/config sampai siap. Setiap setting harus memiliki test “changing setting changes behavior”.

---

## M-10 — Caching perlu selektif; jangan cache chat completion secara default

**Severity:** Menengah sebagai rekomendasi performa/privacy

Tidak ada cache. Itu bukan bug untuk chat completion karena cache prompt dapat menambah risiko privasi dan stale response.

**Rekomendasi cache aman:**

- cache `/v1/models` dengan TTL singkat;
- cache downstream health 1–5 detik;
- cache compiled routing tables/snapshot;
- jangan cache prompt/completion secara default;
- semantic response cache hanya opt-in, tenant-aware, encrypted, bounded, dan memiliki retention jelas.

---

## L-01 — Verifikasi token tidak constant-time

**Severity:** Rendah  
**Lokasi:** `internal/security/security.go:27-28`

Saat ini:

```go
return HashToken(token) == hash
```

Gunakan `crypto/subtle.ConstantTimeCompare`. Token memiliki entropi tinggi sehingga exploit timing lokal mungkin tidak praktis, tetapi perbaikan murah dan sesuai hardening.

---

## L-02 — Request ID client-controlled dapat mengotori log

**Severity:** Rendah  
**Lokasi:** `internal/server/middleware.go:16-29`

`X-Request-ID` diterima apa adanya tanpa batas panjang/karakter.

**Rekomendasi:**

- maksimum 64–128 karakter;
- hanya `[A-Za-z0-9._-]`;
- atau selalu generate internal ID dan simpan external ID terpisah;
- log dalam format terstruktur.

---

## L-03 — CLI wrapper tidak selaras dengan auth dan Windows browser launch

**Severity:** Rendah  
**Lokasi:** `npm-wrapper/bin/cli.js:58-76`

- `stopViaAPI()` tidak mengirim Bearer token; akan gagal setelah auth default diaktifkan.
- Windows `start` adalah builtin `cmd.exe`, bukan executable biasa; `spawn("start", ...)` berpotensi gagal.

**Rekomendasi:**

- baca token dari secure config/credential store;
- gunakan `cmd.exe /c start "" <url>`;
- tangani status `401` sebagai error, bukan “request sent”.

---

## L-04 — Penggunaan shell untuk clipboard tidak perlu

**Severity:** Rendah  
**Lokasi:** tray platform implementation

Gunakan stdin proses `pbcopy`, `xclip`, atau API clipboard native; hindari `bash -c`/shell interpolation.

---

## L-05 — Ring logger dapat dioptimalkan

**Severity:** Rendah  
**Lokasi:** observability logger

Saat penuh, shifting slice bersifat O(n). Dengan cap 1.000 dampaknya kecil, tetapi ring buffer lebih sederhana untuk beban panjang.

---

## L-06 — Dokumentasi dan implementasi mulai drift

**Severity:** Rendah  
**Lokasi:** `docs/*`, README, package metadata

Terdapat audit lama dan checklist fase yang tidak sepenuhnya mencerminkan implementasi terbaru. README menyebut license TBD sementara package metadata menyebut MIT.

**Rekomendasi:**

- jadikan docs bagian release gate;
- tambahkan threat model;
- dokumentasikan batas trust antara AtlasBridge dan 9Router;
- dokumentasikan mode localhost/LAN;
- tandai fitur “planned” versus “enforced”.

---

# 5. Evaluasi per Dimensi

## 5.1 Keamanan dan Kerentanan

### Autentikasi dan otorisasi

**Positif:**

- Admin middleware tersedia.
- Token dibuat dengan 32 byte random dan raw token tidak dipersist.
- Hash token yang disimpan adalah SHA-256 dari token berentropi tinggi.
- `Authorization` tidak diteruskan ke downstream.

**Masalah:**

- auth admin default off;
- `/v1` tanpa auth;
- tidak ada role/scope;
- tidak ada brute-force rate limit;
- token dapat rusak karena masked roundtrip;
- LAN tanpa TLS;
- no origin/host policy.

### Validasi dan sanitasi input

| Risiko | Status |
|---|---|
| SQL injection | Tidak ditemukan surface database/query; saat ini tidak applicable |
| XSS | Tidak ditemukan sink eksplisit; tetap perlu CSP/security headers |
| Command injection | Tidak ada request-driven command execution; shell clipboard perlu dirapikan |
| SSRF | Ada surface melalui downstream URL yang terlalu bebas |
| JSON injection | Ada pada dry-run/combo karena `fmt.Sprintf` |
| DoS payload | Tinggi karena `io.ReadAll` tanpa limit |

### Secrets

- Tidak ditemukan file `.env`, private key, atau pola kredensial hardcoded yang jelas.
- Raw admin token dicetak sekali ke stdout. Ini sesuai pola bootstrap, tetapi dapat tertangkap log launcher/terminal. Pertimbangkan first-run secure UI atau file berpermission ketat yang otomatis dihapus setelah dibaca.
- Config menyimpan hash/verifier, bukan raw token.
- File permission tetap perlu diperketat.

### Paparan data

- Raw prompt tidak dimasukkan ke observability entry pada jalur normal; baik.
- Error downstream terlalu detail.
- Diagnostics memuat URL/topologi/runtime metadata.
- Privacy UI belum menjamin redaction nyata.
- Header downstream terlalu luas.

---

## 5.2 Arsitektur dan Desain Sistem

### Kekuatan

- Pemisahan package cukup jelas.
- Analyzer, classifier, routing, dan forwarder dapat diuji terpisah.
- Routing decision menjadi object eksplisit.
- Runtime state dan auth config sudah memiliki abstraksi tersendiri.
- Frontend state dikelola dengan Pinia.

### Kelemahan

- `ServerDeps` menjadi container pointer mutable global.
- Config persistence dan runtime activation tercampur dalam handler.
- Tidak ada application service/use-case layer untuk transaction.
- Listener lifecycle tidak dimodelkan.
- Forwarder tidak hot-swappable.
- Setting UI tidak selalu merepresentasikan capability backend.

### Target desain

```text
HTTP/API
  │
  ├── Control Plane ── Auth/Origin/Validation
  │                     │
  │                     └── ConfigService
  │                            ├── Validate
  │                            ├── Build Candidate
  │                            ├── Atomic Persist
  │                            └── Atomic Snapshot Swap
  │
  └── Data Plane ─── Request Limits/API Auth
                        │
                        ├── Snapshot.Load()
                        ├── Analyzer
                        ├── Classifier
                        ├── Router
                        └── ForwarderManager
                              ├── Stream Client
                              ├── Nonstream Client
                              ├── Bulkhead
                              └── Circuit State
```

---

## 5.3 Performa dan Manajemen Resource

### Connection management

**Positif:** `http.Client` disimpan di `Forwarder`, sehingga default transport dapat menggunakan connection pooling/keep-alive.

**Perlu diperbaiki:**

- transport belum dituning;
- stream dan nonstream memakai timeout policy yang sama;
- tidak ada `MaxIdleConns`, `MaxIdleConnsPerHost`, `IdleConnTimeout`, `ResponseHeaderTimeout`;
- tidak ada `CloseIdleConnections` saat forwarder diganti.

### Memori

Risiko utama adalah full buffering berulang. Untuk chat panjang, alokasi dapat menjadi beberapa kali ukuran payload.

Prioritas:

1. limit body;
2. one-pass envelope parsing;
3. `bytes.Reader`;
4. stream response;
5. bounded nonstream response;
6. concurrency budget.

### Caching

Cache hanya metadata/health/routing. Hindari cache prompt secara default.

---

## 5.4 Keandalan dan Resiliensi

### Fault tolerance

Belum ditemukan:

- circuit breaker;
- bulkhead;
- retry policy;
- jittered backoff;
- health-aware routing;
- idempotency key.

Namun README menyatakan 9Router bertanggung jawab atas provider failover/load balancing/rate limit. Karena itu AtlasBridge **tidak boleh menambahkan retry buta pada POST chat**, sebab dapat menggandakan request, biaya, atau output.

Rekomendasi:

- circuit breaker hanya pada konektivitas AtlasBridge→9Router;
- retry terbatas pada kegagalan sebelum response header dan hanya bila idempotency dapat dijamin;
- exponential backoff + jitter;
- jangan retry 4xx;
- hormati context cancellation;
- fallback lokal hanya berupa error terstruktur, bukan memilih provider sendiri.

### Graceful degradation

Saat downstream gagal, AtlasBridge kini mengembalikan 502. Perlu ditingkatkan menjadi:

- stable public error code;
- `Retry-After` bila circuit open;
- readiness menunjukkan downstream degraded;
- admin tetap dapat diakses;
- queue dibatasi;
- model list dapat memakai cache pendek;
- stream yang sudah dimulai ditutup bersih tanpa JSON tambahan.

---

## 5.5 Kualitas Kode dan Pemeliharaan

### DRY/SOLID/Clean Code

**Baik:**

- package responsibilities cukup jelas;
- typed config;
- default config terpusat;
- decision object eksplisit;
- test naming cukup sistematis.

**Perlu ditingkatkan:**

- handler terlalu bertanggung jawab atas decode, merge, validate, mutate, persist, dan runtime activation;
- banyak `map[string]interface{}`;
- duplicate JSON error generation;
- field config tidak efektif;
- placeholder package;
- state mutation tersebar.

### Testing

Kekuatan:

- 210 fungsi test;
- test routing dan classifier cukup luas;
- ada test streaming dan pass-through;
- ada test agar raw secret prompt tidak masuk analysis result.

Gap utama:

- auth success/failure/malformed scheme;
- masked hash roundtrip;
- CSRF/origin/content-type;
- body limit;
- slowloris/server timeout;
- concurrent config update + proxy;
- race detector;
- config transaction rollback;
- dynamic forwarder reload;
- stale lock/crash recovery;
- graceful shutdown dengan active stream;
- fuzz test JSON/analyzer/routing;
- property test bahwa rewrite tidak merusak unknown fields;
- load/soak test.

### Dokumentasi

README cukup jelas mengenai tujuan produk dan pembagian tanggung jawab dengan 9Router. Yang belum ada:

- threat model;
- trust boundary;
- security deployment guide;
- config schema/versioning;
- migration strategy;
- incident/runbook;
- backup/recovery;
- release verification;
- LAN/TLS guidance.

---

# 6. Rekomendasi Arsitektur Jangka Panjang

## A. Pisahkan control plane dan data plane secara tegas

Control plane harus selalu membutuhkan auth dan policy origin. Data plane harus memiliki auth opsional untuk localhost, wajib untuk LAN.

Gunakan middleware stack berbeda:

```text
/admin/api:
RequestID → SecurityHeaders → HostGuard → OriginGuard → AdminAuth → BodyLimit → Handler

/v1:
RequestID → APIAuth(if LAN) → BodyLimit → RateLimit → Bulkhead → Handler
```

## B. Gunakan immutable versioned snapshot

Semua request mengambil satu snapshot pada awal request. Jangan membaca pointer config global berulang di tengah request.

Snapshot mencakup:

- config effective;
- compiled routing table;
- profile map;
- forwarder;
- version;
- created_at.

## C. Buat ConfigService transaksional

ConfigService menjadi satu-satunya komponen yang boleh:

- merge;
- validate;
- migrate;
- persist;
- swap runtime state.

Handler hanya memanggil use case.

## D. Pisahkan client stream dan nonstream

Nonstream:

- total deadline;
- bounded body;
- response size cap.

Stream:

- dial/header timeout;
- no total timeout;
- idle watchdog;
- direct streaming;
- cancellation propagation.

## E. Tambahkan bulkhead dan circuit breaker

- separate semaphore untuk stream/nonstream;
- max in-flight;
- max queued;
- circuit breaker ke 9Router;
- half-open probe;
- metric state transition.

## F. Jadikan privacy sebagai enforced policy, bukan UI label

Buat interface:

```go
type Redactor interface {
    RedactLog(fields map[string]any) map[string]any
    RedactError(err error) string
}
```

Semua log dan diagnostics melewati redactor.

## G. Bangun secure-by-default deployment model

Mode:

1. `local-secure`:
   - loopback only;
   - admin token required;
   - data plane optional auth.
2. `lan-secure`:
   - TLS required;
   - admin token required;
   - data-plane key required;
   - CIDR allowlist.
3. Jangan sediakan mode “LAN open”.

## H. Tambahkan CI quality gate

Pull request wajib lulus:

```bash
go test ./...
go test -race ./...
go vet ./...
staticcheck ./...
govulncheck ./...
npm ci
npm run typecheck
npm test
npm run build
npm audit --audit-level=high
```

Tambahkan secret scan, CodeQL, dependency review, SBOM, dan artifact signing.

---

# 7. Rencana Tindakan Prioritas

## P0 — Blocker sebelum rilis berikutnya

- [ ] Ubah admin auth menjadi default aktif.
- [ ] Hapus `admin_token_hash` dari GET/export/PUT DTO.
- [ ] Tambahkan endpoint rotate token dan test regresi lockout.
- [ ] Wajibkan JSON + origin/host guard pada control plane.
- [ ] Paksa loopback bila `allow_lan_access=false`.
- [ ] Tambahkan auth untuk `/v1` pada mode LAN.
- [ ] Tambahkan body limit untuk chat dan seluruh admin endpoint.
- [ ] Tambahkan concurrency cap/bulkhead.
- [ ] Ganti shared mutable config dengan immutable snapshot/atomic swap.
- [ ] Rebuild/swap forwarder ketika downstream/timeout berubah.
- [ ] Jadikan config import/reset/update transaksional.
- [ ] Perbaiki stream timeout, Scanner limit, dan `headersWritten`.
- [ ] Tambahkan timeout graceful shutdown dan hapus `os.Exit` dari goroutine.

## P1 — Satu sprint setelah P0

- [ ] Implement TLS/reverse-proxy guidance untuk LAN.
- [ ] Terapkan downstream allowlist dan redirect/IP validation.
- [ ] Gunakan error publik generik + correlation ID.
- [ ] Terapkan response-header allowlist.
- [ ] Simpan config dengan directory `0700`, file `0600`, atomic write.
- [ ] Implement OS-level single-instance lock.
- [ ] Implement structured logging dan redaction nyata.
- [ ] Implement metric latency/status/bytes/in-flight/circuit state.
- [ ] Tambahkan test concurrency + `go test -race`.
- [ ] Tambahkan fuzz test dan large-payload test.
- [ ] Pisahkan stream/nonstream transport.

## P2 — Hardening dan kesiapan produksi

- [ ] Tambahkan circuit breaker/bulkhead lengkap.
- [ ] Terapkan retry terbatas yang idempotency-safe.
- [ ] Tambahkan health, readiness, dan degraded-state.
- [ ] Implement cache metadata/health yang aman.
- [ ] Buat config schema version + migration.
- [ ] Implement CI PR gate, govulncheck, npm audit, SAST, secret scan.
- [ ] Pin GitHub Actions ke full commit SHA.
- [ ] Hasilkan SBOM, signature, provenance, dan checksum release.
- [ ] Lengkapi threat model, runbook, security guide, dan recovery guide.
- [ ] Hapus atau implement semua setting dekoratif/tidak efektif.

---

# 8. Definition of Done untuk Menyatakan “Layak Local Production”

AtlasBridge baru dapat dinilai layak untuk penggunaan lokal sensitif setelah seluruh kondisi berikut terpenuhi:

- admin auth selalu aktif pada instalasi baru;
- token tidak dapat rusak melalui config roundtrip;
- tidak ada data race pada `go test -race ./...`;
- semua request memiliki size limit;
- config update bersifat transaction + atomic snapshot;
- downstream hot reload benar-benar efektif;
- streaming 10+ menit stabil;
- shutdown memiliki deadline;
- stale lock pulih otomatis;
- privacy badge sesuai enforcement nyata;
- full Go test/vet/govulncheck lulus;
- frontend build/typecheck/test/audit lulus;
- CI menjadi gate sebelum merge/release.

Untuk mode LAN, tambahkan syarat:

- TLS;
- admin auth;
- data-plane auth;
- CIDR allowlist;
- rate limit;
- threat model dan penetration test.

---

# 9. Prioritas Perbaikan Paling Bernilai

Urutan implementasi yang paling efektif adalah:

1. **Perbaiki boundary keamanan:** auth default on, token DTO, origin/content-type/host guard.
2. **Perbaiki resource safety:** body limit, response limit, semaphore.
3. **Perbaiki state model:** immutable snapshot, transaction, forwarder hot swap.
4. **Perbaiki streaming dan lifecycle:** timeout policy, commit tracking, shutdown, process lock.
5. **Perbaiki privacy dan observability:** redaction nyata, structured logs, metrics.
6. **Jadikan kualitas dapat dibuktikan:** race test, fuzz/load test, CI, vulnerability scan, supply-chain hardening.

---

## Kesimpulan Akhir

AtlasBridge bukan codebase yang buruk. Struktur modular, pemisahan analyzer/classifier/routing/forwarder, penggunaan Go, bounded in-memory observability, token random berentropi tinggi, tidak diteruskannya Authorization ke downstream, serta cakupan test yang luas menunjukkan fondasi engineering yang serius.

Masalah utamanya adalah beberapa fitur keamanan dan konfigurasi tampak “tersedia” pada level UI/config, tetapi belum benar-benar menjadi invariant runtime. Control plane masih terlalu percaya pada localhost, state update belum concurrency-safe, payload belum dibatasi, dan stream lifecycle belum robust. Dengan menyelesaikan P0, AtlasBridge dapat naik signifikan dari prototipe lokal menjadi middleware yang aman dan dapat diprediksi. Tanpa P0, mengaktifkan LAN atau memakai AtlasBridge untuk prompt sensitif tetap memiliki risiko yang tidak dapat diterima.
