# AtlasBridge — Testing, CI, dan Release

**Catatan:** File ini adalah pecahan laporan audit utama agar temuan dapat dikerjakan per kategori. Tidak ada perubahan source code yang dilakukan.

---

## 11. Temuan Detail — Testing, CI, dan Release

### TEST-01 — Go validation tidak dapat dijalankan pada environment audit

**Severity:** Validation blocker  
**Evidence:** `go.mod:3`, local `go version go1.23.2`, command `go test ./...`  
**Temuan:** proyek memerlukan `go 1.25.5`; environment audit memiliki `1.23.2`. Auto-download toolchain gagal karena network/DNS blocked.  
**Dampak:** hasil audit tidak dapat menyatakan backend tests lulus.  
**Rekomendasi audit:** jalankan ulang di environment dengan Go 1.25.5 atau sediakan toolchain lock/CI artifact.

### TEST-02 — CI Go jobs berisiko gagal pada checkout bersih karena `web/dist` tidak tersedia

**Severity:** Critical  
**Evidence:** `web/assets.go:9-16`, `.gitignore` mengabaikan `web/dist/`, ZIP asli tidak berisi `web/dist/`, `.github/workflows/ci.yml:14-77` Go jobs tidak membangun frontend terlebih dahulu  
**Temuan:** package Go meng-embed `dist`, tetapi `dist` generated dan ignored. CI Go lint/test/race/govulncheck berjalan di job terpisah sebelum/tanpa frontend build.  
**Dampak:** CI Go dapat gagal compile pada environment bersih.  
**Rekomendasi audit:** build frontend sebelum Go job, commit minimal placeholder embed, gunakan build tags, atau pisahkan embed package untuk release only.

### TEST-03 — CI menggunakan tool `@latest` untuk staticcheck/govulncheck

**Severity:** Medium  
**Evidence:** `.github/workflows/ci.yml:28-32`, `73-77`  
**Temuan:** actions dipin ke commit SHA, tetapi Go security/lint tools di-install dengan `@latest`.  
**Dampak:** CI dapat berubah tanpa perubahan source; reproducibility turun.  
**Rekomendasi audit:** pin versi staticcheck dan govulncheck.

### TEST-04 — Frontend tidak memiliki unit/component test runner

**Severity:** Medium  
**Evidence:** `web/package.json` scripts hanya `dev`, `build`, `preview`, `test:e2e`  
**Temuan:** tidak ada Vitest/Jest/component test.  
**Dampak:** bug kontrak UI/state sederhana dapat lolos sampai E2E/manual.  
**Rekomendasi audit:** tambah Vitest + Vue Test Utils untuk store, route guard, dan komponen critical.

### TEST-05 — E2E tests menggunakan mocked API sehingga tidak menguji integrasi backend nyata

**Severity:** Medium  
**Evidence:** `web/e2e/fixtures.ts` mengintercept `/admin/api` request  
**Temuan:** mock membantu UI test, tetapi tidak mendeteksi mismatch backend/frontend seperti `connected` vs `ok`.  
**Dampak:** false confidence pada integrasi.  
**Rekomendasi audit:** tambahkan minimal satu suite integration E2E dengan backend Go nyata.

### TEST-06 — Playwright gagal dijalankan lokal pada audit

**Severity:** Validation warning  
**Evidence:** `npx playwright test --reporter=list` timeout menunggu `webServer` 30 detik  
**Temuan:** test runner tidak berhasil start/observe preview server pada environment audit.  
**Dampak:** E2E tidak tervalidasi dalam audit ini.  
**Rekomendasi audit:** jalankan di CI atau debug `webServer.url/baseURL`, host binding, dan readiness path.

### TEST-07 — Release workflow tidak menjalankan test sebelum upload asset

**Severity:** High  
**Evidence:** `.github/workflows/release.yml:25-56`  
**Temuan:** release workflow build frontend, build Go binary, zip, checksum, upload draft. Tidak terlihat menjalankan `go test`, `npm audit`, E2E, govulncheck, atau signing/SBOM.  
**Dampak:** release asset dapat dibuat dari kode yang gagal test.  
**Rekomendasi audit:** jadikan release bergantung pada CI sukses, atau ulang minimal smoke/security gates sebelum asset dibuat.

### TEST-08 — Release tidak memiliki signing/SBOM/provenance

**Severity:** Medium  
**Evidence:** `.github/workflows/release.yml:44-56`  
**Temuan:** hanya checksum SHA256; belum ada signing, SBOM, atau build provenance.  
**Dampak:** supply-chain trust belum kuat untuk binary Windows.  
**Rekomendasi audit:** tambah cosign/minisign, SBOM, dan provenance attestation.

---
