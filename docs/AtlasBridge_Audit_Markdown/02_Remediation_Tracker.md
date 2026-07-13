# AtlasBridge Remediation Tracker

Gunakan dokumen ini untuk mencatat perbaikan setelah audit statis.

## Status

- `OPEN` — belum dikerjakan
- `IN PROGRESS` — sedang dikerjakan
- `READY FOR TEST` — implementasi selesai, belum diverifikasi
- `VERIFIED` — lulus pengujian
- `DEFERRED` — ditunda dengan alasan tertulis

## P0 — Release Blocker

| ID | Tugas | Status | Bukti yang wajib |
|---|---|---|---|
| P0-01 | Pertahankan atau rotasi credential admin saat reset/import | VERIFIED | `TestResetPreservesAdminTokenHash`, `TestImportPreservesAdminTokenHash`, `TestImportPreservesHashWhenImportedConfigHasEmptyHash`, `TestValidateRejectsEmptyHashWithAuthEnabled` — semua lulus |
| P0-02 | Gunakan active routes pada smart alias | VERIFIED | `TestSmartAutoHonorsCustomRoutes` — lulus |
| P0-03 | Tutup mode LAN sampai data-plane auth tersedia | VERIFIED | `TestDataPlaneAuthRejectsNoToken`, `TestDataPlaneAuthAcceptsValidToken`, `TestDataPlaneAuthRejectsWrongToken`, `TestDataPlaneAuthDisabledPassesThrough` — semua lulus |
| P0-04 | Hubungkan SSRF policy ke seluruh outbound request | VERIFIED | `TestIsAllowedIP_*` (7 tests), `TestValidateDownstreamURL_BlocksPrivateIP`, `TestSafeRedirectPolicy_*` (5 tests), `TestSafeRedirectPolicy_IntegrationWithLiveServer` — semua lulus |

## P1 — Stabilitas dan kepercayaan

| ID | Tugas | Status | Bukti yang wajib |
|---|---|---|---|
| P1-01 | Gunakan schema konfigurasi sebagai sumber kebenaran | OPEN | Unknown field ditolak |
| P1-02 | Tampilkan `restart_required` dan actual listener address | OPEN | UI dan integration test |
| P1-03 | Buat import/reset transaksional | OPEN | Fault-injection test |
| P1-04 | Tambahkan schema version dan migration | OPEN | Migration test |
| P1-05 | Baca request body satu kali | OPEN | Benchmark allocation |
| P1-06 | Tambahkan response size limit dan stream timeout | VERIFIED | `MaxResponseBody` 64MB di forwarder, stream idle 5min + max lifetime 30min — tests lulus |
| P1-07 | Satukan logging, privacy, redaction, dan retention | OPEN | Test matrix privacy |
| P1-08 | Terapkan security headers ke seluruh admin SPA | VERIFIED | `SecurityHeaders` middleware dipindah ke `/admin` parent route — test lulus |
| P1-09 | Jalankan Playwright di CI | VERIFIED | CI job Playwright + 42 E2E tests (auth, dashboard, settings, navigation) — dikonfigurasi di `ci.yml` |
| P1-10 | Lengkapi observability request lifecycle | OPEN | Metrics/log verification |

## P2 — Product quality

| ID | Tugas | Status |
|---|---|---|
| P2-01 | Perbaiki prefix header `X-RateLimit-` | OPEN |
| P2-02 | Pertahankan safe downstream content type | OPEN |
| P2-03 | Tangani JSON error pada request streaming | OPEN |
| P2-04 | Hindari reserialization seluruh payload | OPEN |
| P2-05 | Tambahkan noindex pada admin | OPEN |
| P2-06 | Sinkronkan CLI dengan konfigurasi aktual | OPEN |
| P2-07 | Tambahkan accessibility check | OPEN |
| P2-08 | Tambahkan SBOM, checksum, dan attestation | OPEN |

## Aturan penutupan temuan

Sebuah item hanya boleh berstatus `VERIFIED` apabila:
1. implementasi selesai;
2. test otomatis tersedia;
3. test dijalankan dan lulus;
4. bukti hasil dicatat;
5. tidak muncul regresi pada area terkait.
