# AtlasBridge Static Technical Audit

**Jenis audit:** Static code review, architecture review, security design review, business-flow review, dan CI/CD review  
**Snapshot:** `9ba10da`  
**Status:** Conditional Alpha Readiness  
**Batasan:** Pengujian Go, frontend build, Playwright, integration test, dan benchmark belum dijalankan penuh pada lingkungan audit.

## Kesimpulan singkat

AtlasBridge memiliki fondasi arsitektur yang cukup baik untuk proxy lokal satu pengguna. Pemisahan package, immutable configuration snapshot, body limit, bulkhead, autentikasi admin, dan pemeriksaan dasar CI sudah tersedia.

Namun, terdapat empat masalah utama yang harus dipisahkan berdasarkan konteks deployment:

| ID | Temuan | Prioritas | Keputusan |
|---|---|---:|---|
| C-01 | Reset/import dapat menghapus kontinuitas credential admin | P0 | Wajib diperbaiki sebelum alpha harian |
| C-02 | `atlas-auto` dan `smart-auto` dapat mengabaikan custom route | P0 | Wajib diperbaiki sebelum alpha harian |
| C-03 | Mode LAN membuka `/v1` tanpa autentikasi data plane | P0 untuk LAN | LAN tetap NO-GO |
| C-04 | Validator SSRF belum dipakai pada jalur outbound produksi | P0 untuk LAN/remote | Remote/LAN tetap NO-GO |

## Temuan prioritas

### C-01 — Risiko admin lockout setelah reset/import

**Dampak:** pengguna dapat kehilangan akses ke seluruh endpoint admin karena autentikasi aktif tetapi hash token kosong.

**Perbaikan wajib:**
- Pertahankan token hash aktif saat reset/import; atau buat token baru secara atomik.
- Jangan menerima `AdminTokenHash` dari public import.
- Tolak konfigurasi dengan `AdminAuthEnabled=true` dan hash kosong.
- Tambahkan integration test reset dan export-import.

**Selesai jika:**
- Token lama tetap bekerja, atau token baru diterima UI satu kali.
- `/admin/api/status` tetap dapat diakses setelah reset/import.

### C-02 — Smart alias mengabaikan route aktif

**Dampak:** perubahan pada Routing Settings tidak selalu memengaruhi keputusan `atlas-auto`.

**Perbaikan wajib:**
- Teruskan active routes ke `resolveSmartAlias()`.
- Hapus penggunaan `DefaultRoutesConfig()` pada runtime request.
- Tambahkan regression test untuk custom route.

**Selesai jika:**
- Perubahan route langsung memengaruhi `atlas-auto` dan `smart-auto` tanpa restart.

### C-03 — Data plane LAN tanpa autentikasi

**Dampak:** perangkat lain di jaringan dapat menggunakan `/v1`, menghabiskan kuota, memenuhi bulkhead, atau meningkatkan biaya downstream.

**Perbaikan wajib:**
- Nonaktifkan atau tandai LAN sebagai experimental sampai aman.
- Tambahkan data-plane API key yang berbeda dari admin token.
- Tambahkan rate limit dan concurrency limit per client.
- Gunakan HTTPS melalui reverse proxy atau native TLS.

**Selesai jika:**
- Request LAN ke `/v1` tanpa credential selalu ditolak.

### C-04 — Proteksi SSRF belum terhubung

**Dampak:** helper validasi tersedia, tetapi outbound request aktual masih dapat melewati network policy.

**Perbaikan wajib:**
- Gunakan satu `ValidatedDownstreamTransport` untuk forwarder, health check, dan combo test.
- Validasi IPv4 dan IPv6.
- Validasi ulang setiap redirect.
- Terapkan explicit allow policy untuk loopback 9Router.
- Batasi remote destination dengan allowlist.

**Selesai jika:**
- Tidak ada outbound HTTP request yang dapat melewati network policy.

## Temuan P1 penting

1. Kontrak konfigurasi backend, file contoh, UI, dan runtime belum sepenuhnya sinkron.
2. Import/reset multi-file belum transaksional.
3. Request body dapat dibaca dan disalin beberapa kali.
4. Response non-stream belum memiliki size limit.
5. Stream belum memiliki idle timeout dan maximum lifetime yang jelas.
6. Privacy, logging, redaction, dan retention belum menjadi satu policy engine.
7. Security headers belum diterapkan ke seluruh admin SPA.
8. Playwright belum benar-benar dijalankan di CI.
9. Transparansi OpenAI-compatible proxy dapat terganggu pada header, content type, dan stream error.
10. Observability belum merekam hasil akhir request secara lengkap.

## Catatan classifier

Evaluasi statis awal menunjukkan kualitas classifier perlu diperbaiki. Substring pendek seperti `ui`, `api`, dan `test` berpotensi menimbulkan false positive.

**Arah perbaikan:**
- Gunakan token boundary.
- Terapkan weighted evidence, bukan first-match sederhana.
- Buat confusion matrix.
- Tambahkan adversarial dan holdout dataset.
- Gunakan accuracy, precision, recall, dan dampak biaya route sebagai quality gate.

## Keputusan audit statis

| Skenario | Status |
|---|---|
| Development localhost satu pengguna | Conditional GO setelah C-01 dan C-02 |
| Penggunaan lokal harian | Conditional GO setelah seluruh P0 lokal dan integration test lulus |
| LAN rumah/kantor | NO-GO |
| Multi-user | NO-GO |
| Production gateway | NO-GO |

## Catatan metodologis

Dokumen ini adalah audit statis. Hasil runtime, performa, keamanan dinamis, dan kelayakan rilis harus dibuktikan pada dokumen pengujian terpisah.
