# AtlasBridge — Daftar Audit Per Kategori

File ini memecah laporan audit utama menjadi bagian-bagian yang lebih mudah dibaca dan dijadikan backlog kerja. Urutan baca yang disarankan: mulai dari ringkasan, lanjut P0 security/backend/CI, lalu frontend, performa, testing, hygiene, dan roadmap.

## Daftar File

| No | File | Fokus |
|---:|---|---|
| 00 | `00_RINGKASAN_EKSEKUTIF_DAN_VALIDASI.md` | Gambaran umum, metodologi, lingkungan validasi, hasil command, inventaris, dan prioritas tertinggi. |
| 01 | `01_SECURITY.md` | Password hashing, login hardening, SSRF, token storage, HostGuard, error leakage, logging, dan CSP. |
| 02 | `02_BACKEND_API_DAN_ARSITEKTUR.md` | Admin path, forwarder, streaming, validation, lifecycle, config atomicity, dan network invariant. |
| 03 | `03_FRONTEND_UX_DAN_UI_STATE.md` | Route salah, halaman zombie, fake control, dummy runtime history, setup wizard, logo, responsiveness, dan error UI. |
| 04 | `04_PERFORMANCE.md` | Bundle, observability, memory/body reading, dan perf gate. |
| 05 | `05_TESTING_CI_DAN_RELEASE.md` | Go validation, CI risk, Playwright, unit/component test, release workflow, signing/SBOM. |
| 06 | `06_REPOSITORY_HYGIENE_ZOMBIE_CODE_DAN_DOKUMENTASI.md` | `.git`, binary sisa, test-results, audit lama, typo folder, `.qwen`, working tree, dan asset duplikasi. |
| 07 | `07_HALAMAN_DAN_KONTROL_UI.md` | Ringkasan halaman/kontrol yang kurang, berlebih, palsu, atau tidak berguna. |
| 08 | `08_PRIORITAS_ROADMAP_READINESS_DAN_CHECKLIST.md` | P0/P1/P2, readiness verdict, batasan audit, checklist developer, dan lampiran evidence. |

## Pembagian Sprint yang Disarankan

### Sprint 1 — P0 Release Blocker
Fokus: security P0, SSRF DNS-aware, login hardening, password hashing, CI Go build/test, dan pembersihan `.git`/binary dari paket distribusi.

### Sprint 2 — P1 UX/API Consistency
Fokus: route `/logs`, tombol Start/Stop palsu, setup wizard, health status contract, `admin_path`, dummy runtime history, dan error response admin.

### Sprint 3 — Maintainability & Observability
Fokus: structured logging + redaction, frontend tests, OpenAPI/shared types, responsive UI, asset registry, dokumentasi authoritative, dan release provenance.

## Prinsip Eksekusi

1. Jangan menggabungkan banyak kategori besar dalam satu PR kecuali ada dependensi langsung.
2. Setiap temuan P0 sebaiknya memiliki test atau minimal verification command.
3. UI control harus memenuhi salah satu dari dua pilihan: benar-benar terhubung API atau dihapus dari tampilan.
4. Config field harus jelas: live-effective, requires restart, deprecated, atau tidak tersedia.
5. Dokumentasi lama perlu diberi status superseded agar tidak menyesatkan developer berikutnya.
