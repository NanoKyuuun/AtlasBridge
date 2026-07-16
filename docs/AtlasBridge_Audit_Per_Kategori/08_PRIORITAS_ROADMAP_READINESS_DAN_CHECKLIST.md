# AtlasBridge — Prioritas, Roadmap, Readiness, dan Checklist

**Catatan:** File ini adalah pecahan laporan audit utama agar temuan dapat dikerjakan per kategori. Tidak ada perubahan source code yang dilakukan.

---

## 14. Rekomendasi Prioritas Tanpa Mengubah Source

### P0 — Harus dibereskan sebelum release/public testing

1. Ganti password hashing dari SHA-256 ke Argon2id/bcrypt/scrypt.
2. Tambahkan rate limiting/backoff/lockout untuk login admin.
3. Terapkan SSRF protection DNS-aware pada semua outbound path, termasuk redirect dan health check.
4. Perbaiki CI Go job agar `web/dist` tersedia sebelum compile, atau refactor embed strategy.
5. Pastikan `go test ./...`, `go test -race`, `go vet`, `staticcheck`, dan `govulncheck` benar-benar lulus di environment bersih.
6. Hapus binary `atlasbridge.exe~` dan `.git` dari paket distribusi.

### P1 — Harus dibereskan sebelum user acceptance testing

1. Perbaiki route `/logs` dan tentukan nasib `Logs.vue`.
2. Hubungkan Start/Stop Proxy ke API nyata atau hapus dari topbar.
3. Samakan kontrak downstream health: `connected` vs `ok`.
4. Perjelas `admin_path`: implementasikan route dinamis atau hapus setting.
5. Hapus dummy runtime history.
6. Perbaiki setup wizard/first-run auth flow.
7. Standarisasi error response admin agar tidak leak detail internal.
8. Tambahkan frontend unit/component tests untuk route, store, dan UI controls critical.

### P2 — Peningkatan kualitas dan maintainability

1. Tambahkan API schema/OpenAPI atau shared TypeScript types dari backend contract.
2. Terapkan structured logger + redactor secara menyeluruh.
3. Tambahkan responsive layout untuk admin UI.
4. Rapikan dokumentasi lama dan tandai audit yang sudah superseded.
5. Pin versi staticcheck/govulncheck.
6. Tambahkan signing/SBOM/provenance pada release.
7. Buat asset registry logo/icon.
8. Tambahkan load/perf smoke test sederhana.

---

## 15. Readiness Verdict

| Kategori | Verdict |
|---|---|
| Development continuation | Layak dilanjutkan |
| Internal demo terbatas | Bisa, dengan catatan security belum aman |
| User acceptance testing | Belum disarankan sebelum P0/P1 selesai |
| Public/LAN exposure | Tidak disarankan saat ini |
| Production release | Belum layak |

**Verdict akhir:** AtlasBridge sudah melewati fase “prototype berantakan”, tetapi belum mencapai fase “release-ready”. Renovasi sudah menghasilkan struktur yang cukup kuat, namun meninggalkan gap yang khas setelah refactor besar: route mati, kontrol UI semu, dokumentasi drift, security claim yang belum sepenuhnya enforce, dan pipeline yang belum bisa menjamin build/test bersih dari checkout baru.

---

## 16. Batasan Audit

1. Audit ini tidak memperbaiki source code sesuai permintaan user.
2. Backend Go tests tidak bisa dijalankan karena toolchain mismatch dan network toolchain download tidak tersedia.
3. E2E Playwright tidak tervalidasi karena timeout local preview server pada environment audit.
4. Tidak dilakukan dynamic penetration test terhadap server berjalan.
5. Tidak dilakukan fuzzing, load testing nyata, atau runtime memory profiling.
6. Tidak dilakukan audit historis penuh terhadap semua commit di `.git`.
7. Temuan berbasis static review dapat berubah setelah proyek dijalankan di environment native Windows/Go 1.25.5.

---

## 17. Checklist Audit Cepat untuk Tim Developer

- [ ] Jalankan `git archive` dari commit bersih untuk paket source.
- [ ] Pastikan `web/dist` strategy tidak mematahkan Go compile/test.
- [ ] Jalankan `go test ./...` dengan Go 1.25.5.
- [ ] Jalankan `go test -race ./...`.
- [ ] Jalankan `govulncheck ./...`.
- [ ] Jalankan `staticcheck ./...` dengan versi pinned.
- [ ] Jalankan `npm ci && npm run build && npm audit --audit-level=high`.
- [ ] Jalankan Playwright di CI dan satu integration suite dengan backend nyata.
- [ ] Review semua UI controls: harus save ke backend atau dihapus.
- [ ] Review semua config fields: harus efektif atau diberi label requires restart/deprecated.
- [ ] Implementasikan security P0 sebelum LAN/public exposure.

---

## 18. Lampiran — File Evidence Utama

| Area | File penting |
|---|---|
| Router backend/admin API | `internal/server/server.go` |
| Middleware security | `internal/server/middleware.go` |
| Auth/token/password | `internal/security/security.go`, `internal/server/admin.go` |
| SSRF/downstream validation | `internal/netutil/ssrf.go`, `internal/server/downstream_validate.go`, `internal/forwarder/forwarder.go` |
| Config service | `internal/server/config_service.go`, `internal/config/config.go` |
| Admin UI embed | `web/assets.go`, `internal/server/admin_ui.go` |
| Frontend routing | `web/src/router/index.ts` |
| Layout/navigation | `web/src/components/Layout.vue` |
| Advanced controls | `web/src/pages/AdvancedSettings.vue` |
| Dashboard health status | `web/src/pages/Dashboard.vue` |
| Startup/runtime | `web/src/pages/StartupSettings.vue`, `web/src/pages/Runtime.vue` |
| Auth store | `web/src/stores/auth.ts` |
| CI | `.github/workflows/ci.yml` |
| Release | `.github/workflows/release.yml` |

---

**Catatan penutup:** laporan ini sengaja mempertahankan bentuk audit tanpa patch. Urutan prioritas di atas dapat dipakai sebagai backlog teknis untuk sprint perbaikan berikutnya.
