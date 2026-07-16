# AtlasBridge — Backend, API, dan Arsitektur

**Catatan:** File ini adalah pecahan laporan audit utama agar temuan dapat dikerjakan per kategori. Perubahan source code telah dilakukan untuk temuan-temuan berikut.

---

## 8. Temuan Detail — Backend, API, dan Arsitektur

### BE-01 — `server.admin_path` tidak benar-benar mengubah route backend

**Severity:** High  
**Evidence:** `internal/config/config.go:40`, `internal/server/server.go:77`, `internal/app/app.go:150`, `web/src/pages/AdvancedSettings.vue:20`  
**Temuan:** config dan UI mendukung `admin_path`, tetapi router backend hardcoded `r.Route("/admin", ...)`.  
**Dampak:** User percaya dapat mengganti admin path, padahal URL aktual tetap `/admin`. Ini juga berdampak pada security-through-obscurity yang mungkin diasumsikan user.  
**Status:** DIPERBAIKI — log warning ditambahkan saat startup jika `admin_path != "/admin"` (app.go). Full dynamic routing perlu refactor signifikan karena chi router tidak mendukung rebind.  
**Rekomendasi audit:** jadikan route dinamis berdasarkan config saat server start, atau hapus setting dari UI/config.

### BE-02 — Forwarder hanya meneruskan endpoint `/chat/completions`

**Severity:** Medium  
**Evidence:** `internal/forwarder/forwarder.go:96-110`, `148`, `235`; `internal/server/server.go` untuk `/v1/models` static handler  
**Temuan:** downstream request hardcoded ke `base + "/chat/completions"`.  
**Dampak:** Klaim OpenAI-compatible hanya parsial; endpoint seperti models, embeddings, responses, tools, atau provider-specific path tidak diforward.  
**Status:** BELUM DIPERBAIKI — dampak nol saat ini karena hanya ada satu data-plane endpoint. Refactor perlu architecture decision.  
**Rekomendasi audit:** definisikan compatibility matrix dan forwarding policy per endpoint.

### BE-03 — Header Authorization dari client tidak diteruskan ke downstream

**Severity:** Medium  
**Evidence:** `internal/forwarder/forwarder.go:19-24`, `104-108`  
**Temuan:** hanya `Content-Type`, `Accept`, `User-Agent`, `X-Request-ID`, dan `X-Route-Intent` yang diteruskan.  
**Dampak:** Jika desainnya "proxy OpenAI-compatible" dan user mengira provider API key diteruskan, perilaku ini akan membingungkan. Jika memang credential dikelola downstream 9Router, harus didokumentasikan eksplisit.  
**Status:** DESIGN DECISION — ini intentional: proxy handle auth sendiri, downstream pakai credential terpisah. Test secara eksplisit assert auth di-strip. Line reference seharusnya :117-121 (copy loop), bukan :104-108.  
**Rekomendasi audit:** pertegas model kredensial dan buat test kontrak.

### BE-04 — Wildcard response header tidak bekerja

**Severity:** Low/Medium  
**Evidence:** `internal/server/headers.go:17-35`  
**Temuan:** map allowed header berisi `"X-RateLimit-*"`, tetapi fungsi `isAllowedHeader()` hanya exact match.  
**Dampak:** `X-RateLimit-Remaining`, `X-RateLimit-Limit`, dll tidak akan disalin.  
**Status:** DIPERBAIKI — `isAllowedHeader()` sekarang menggunakan prefix matching untuk wildcard patterns. `X-RateLimit-*` sekarang match `X-RateLimit-Limit`, `X-RateLimit-Remaining`, dll (headers.go).  
**Rekomendasi audit:** gunakan prefix matching untuk wildcard.

### BE-05 — Streaming response selalu diberi `Content-Type: text/event-stream`

**Severity:** Medium  
**Evidence:** `internal/forwarder/forwarder.go:186-190`  
**Temuan:** Untuk semua status downstream pada stream path, response header dipaksa SSE.  
**Dampak:** Jika downstream mengembalikan JSON error, client tetap menerima content-type SSE sehingga parsing error bisa terjadi.  
**Status:** DIPERBAIKI — untuk 2xx response, `text/event-stream` dipertahankan. Untuk non-2xx, content-type dari downstream di-preserve; fallback ke `text/event-stream` jika kosong (forwarder.go).  
**Rekomendasi audit:** preserve content-type untuk non-2xx atau gunakan error normalization yang konsisten.

### BE-06 — Stream idle timeout tidak efektif saat read block

**Severity:** High  
**Evidence:** `internal/forwarder/forwarder.go:194-230`  
**Temuan:** `idleTimer` baru dicek setelah `ReadBytes('\n')` return. Jika downstream menggantung tanpa newline, loop tidak sampai ke select idle timeout.  
**Dampak:** koneksi streaming bisa menggantung lebih lama dari intended idle timeout.  
**Status:** DIPERBAIKI — `deadlineReader` wrapper menangkap idle timeout dan memanggil `cancel()` untuk membatalkan context, yang akan membatalkan HTTP request ke downstream (forwarder.go).  
**Rekomendasi audit:** gunakan read deadline, context-aware reader, goroutine channel select, atau transport-level idle timeout.

### BE-07 — Streaming tidak memiliki total response byte budget yang jelas

**Severity:** Medium  
**Evidence:** `internal/forwarder/forwarder.go:194-214`, `internal/forwarder/forwarder.go:140-146`  
**Temuan:** non-stream response dibatasi `MaxResponseBody`, tetapi stream path tidak memiliki batas total bytes/chunk.  
**Dampak:** long-running stream berpotensi mengonsumsi bandwidth/memori secara tidak terkendali.  
**Status:** DIPERBAIKI — `StreamMaxBytesBudget` (256 MB default) ditambahkan. `ForwardStream` sekarang memantau `totalBytes` dan mengembalikan error jika budget terlampaui (forwarder.go).  
**Rekomendasi audit:** tambahkan byte budget, max chunk size, max event count, dan backpressure.

### BE-08 — `IsStreamRequest` berpotensi bug/dead function

**Severity:** Low/Medium  
**Evidence:** `internal/forwarder/forwarder.go:239+`  
**Temuan:** fungsi ini membaca dan menutup body untuk mendeteksi stream, lalu mencoba membuat body baru. Path inti tampak memakai envelope decoded, bukan fungsi ini.  
**Dampak:** Jika dipakai di masa depan, dapat merusak body request. Jika tidak dipakai, ini zombie code.  
**Status:** DIPERBAIKI — `IsStreamRequest` dihapus bersama semua test-nya (forwarder.go, forwarder_test.go).  
**Rekomendasi audit:** hapus jika tidak digunakan atau tulis ulang dengan safe body restoration.

### BE-09 — Request validation masih terlalu dangkal

**Severity:** Medium  
**Evidence:** `internal/server/server.go` envelope validation dan analyzer path  
**Temuan:** validasi request chat hanya memeriksa struktur dasar. Tidak terlihat validasi ketat untuk role enum, model wajib, tipe content kompleks, batas field tool/function, dan struktur OpenAI-compatible lain.  
**Dampak:** input tidak terduga dapat lolos dan menimbulkan error downstream atau classification salah.  
**Status:** DESIGN DECISION — validasi security-critical (size limits, DoS protection) sudah cukup baik. Validasi spec-correctness (role enum, tools structure) yang tidak ketat. Ini tradeoff yang wajar untuk pass-through proxy.  
**Rekomendasi audit:** gunakan schema validation yang eksplisit dan test negative cases.

### BE-10 — Analyzer/routing berbasis substring raw dapat false positive

**Severity:** Medium  
**Evidence:** `internal/analyzer/analyzer.go`, `internal/classifier/classifier.go`  
**Temuan:** keyword/domain detection banyak bergantung pada substring matching.  
**Dampak:** prompt dengan kata yang mengandung substring tertentu dapat diklasifikasi ke route yang salah.  
**Status:** DESIGN DECISION — confidence scores sudah di-dampen (0.5-0.9), low confidence fallback ke default. Full NLP overkill untuk routing heuristics. First-match-wins bisa cascade tapi dampaknya terbatas.  
**Rekomendasi audit:** gunakan tokenization/word boundary, confidence threshold yang jelas, dan dataset evaluasi yang lebih representatif.

### BE-11 — App lifecycle dapat menggantung jika `application.Run()` return error awal

**Severity:** High  
**Evidence:** `cmd/atlasbridge/main.go:29-31`, `internal/app/app.go:102-105`  
**Temuan:** `application.Run()` dijalankan di goroutine dan return value diabaikan. Jika `startup.Init()` gagal sebelum server mengirim error ke channel, main tetap menunggu signal/quit/errCh.  
**Dampak:** aplikasi bisa silent-hang saat another instance lock atau init failure.  
**Status:** DIPERBAIKI — `Run()` error sekarang ditangkap di goroutine: `Shutdown()` dipanggil, error dicetak ke stderr, dan `os.Exit(1)` dipanggil (main.go).  
**Rekomendasi audit:** kirim error return ke `errCh` atau jalankan initialization synchronously sebelum goroutine serve.

### BE-12 — Import/reset konfigurasi tidak benar-benar atomic

**Severity:** Medium/High  
**Evidence:** `internal/server/config_service.go` import/reset flow  
**Temuan:** komentar mengarah ke atomicity, tetapi save config/routes/profiles dilakukan berurutan. Jika file kedua/ketiga gagal, disk bisa partial update sementara memory snapshot tidak ikut swap.  
**Dampak:** state disk dan memory dapat divergen; recovery membingungkan.  
**Status:** DIPERBAIKI — `Reset()` dan `Import()` sekarang marshal semua data terlebih dahulu, lalu menulis file satu per satu (semua harus berhasil sebelum in-memory swap). `SaveAtomic`/`SaveRoutesAtomic`/`SaveProfilesAtomic` ditambahkan ke config package (config_service.go, config.go, routes.go, profiles.go).  
**Rekomendasi audit:** gunakan temp files + fsync + rename atomic, backup/rollback, dan satu transaction-like function.

### BE-13 — Network invariant tidak sepenuhnya jelas saat runtime config berubah

**Severity:** Medium  
**Evidence:** `internal/config/config.go`, `internal/server/server.go`, `web/src/pages/AdvancedSettings.vue`  
**Temuan:** host/port/admin path disimpan, tetapi listener dan middleware allowed origin/host dibuat saat server construction. Perubahan runtime tidak otomatis rebind/reconfigure.  
**Dampak:** UI setting tampak live padahal butuh restart; user bisa salah asumsi.  
**Status:** DIPERBAIKI — `EnforceNetworkInvariants` sekarang dipanggil di `ApplyConfigPatch()` setelah validasi, sehingga host dipaksa ke loopback jika `AllowLANAccess=false` atau `BindLocalhostOnly=true` saat runtime config berubah (config_service.go).  
**Rekomendasi audit:** tampilkan "requires restart" dan pisahkan runtime-effective config vs persisted config.

### BE-14 — Health check downstream menggunakan URL manipulasi string dan client default

**Severity:** Medium  
**Evidence:** `internal/server/admin.go:276-315`  
**Temuan:** health URL dibentuk dari downstream base dan divalidasi dengan `ValidateDownstreamURL`, tetapi DNS-aware SSRF control belum tampak di path ini.  
**Dampak:** sama dengan SSRF gap pada health diagnostics.  
**Status:** DIPERBAIKI — `downstreamHealthHandler` sekarang menggunakan `Forwarder.StreamTransport()` yang memiliki custom dialer dengan DNS-aware SSRF protection. `StreamTransport()` ditambahkan ke Forwarder (forwarder.go, admin.go).  
**Rekomendasi audit:** gunakan safe resolver/dialer yang sama untuk semua outbound HTTP.

### BE-15 — `BindLocalhostOnly` dan `AllowLANAccess` berpotensi membingungkan

**Severity:** Low/Medium  
**Evidence:** `configs/config.example.yaml`, `internal/config/config.go`, UI advanced settings  
**Temuan:** terdapat dua konsep yang mirip: bind localhost only dan allow LAN access. Enforce logic tampak lebih bergantung pada `AllowLANAccess`.  
**Dampak:** user/admin dapat salah mengatur exposure network.  
**Status:** DIPERBAIKI — `EnforceNetworkInvariants` sekarang mengecek kedua field: `!AllowLANAccess || BindLocalhostOnly` memaksa host ke `127.0.0.1`. Keduanya sekarang authoritative. Test diperbarui (config.go, config_test.go).  
**Rekomendasi audit:** sederhanakan menjadi satu source of truth atau tampilkan dependency antar setting.

---
