# AtlasBridge — Ringkasan Eksekutif & Validasi

**Catatan:** File ini adalah pecahan laporan audit utama agar temuan dapat dikerjakan per kategori. Tidak ada perubahan source code yang dilakukan.

---

## 1. Ringkasan Eksekutif

AtlasBridge sudah memiliki pondasi teknis yang cukup jelas: backend Go, frontend Vue 3/Vite/TypeScript, struktur konfigurasi YAML, routing engine, middleware keamanan dasar, admin UI, CI, release workflow, serta dokumentasi yang cukup banyak. Namun, setelah renovasi besar, proyek belum layak dianggap release-ready karena masih ada gap serius antara klaim dokumentasi, konfigurasi, UI, dan perilaku runtime.

**Kesimpulan utama:** proyek dapat dibangun pada sisi frontend dan `npm audit` tidak menemukan vulnerability high-level pada paket frontend, tetapi validasi backend Go tidak dapat dijalankan di environment audit karena proyek mensyaratkan Go `1.25.5`, sedangkan environment audit hanya memiliki Go `1.23.2` dan tidak dapat mengunduh toolchain. Selain itu, terdapat beberapa blocker penting pada keamanan, routing admin UI, SSRF protection, password hashing, login hardening, CI/build dependency, serta beberapa UI yang terlihat aktif tetapi tidak terhubung ke API nyata.

**Status umum:**

| Area | Status | Catatan utama |
|---|---:|---|
| Frontend build | Lulus | `npm ci --ignore-scripts`, `npm run build`, dan `npm audit --audit-level=high` berhasil. |
| Backend test | Tidak tervalidasi | Gagal sebelum test karena mismatch Go toolchain. |
| Security | Perlu perbaikan serius | Password masih SHA-256 unsalted; login tanpa rate limit; SSRF DNS gap. |
| Arsitektur | Cukup baik, tetapi belum konsisten | Ada service layer dan snapshot, tetapi beberapa konfigurasi tidak benar-benar berlaku. |
| UI/UX | Banyak gap pasca-renovasi | Ada route salah, halaman zombie, kontrol palsu, dummy runtime history. |
| Testing | Belum memadai | E2E mock-heavy, tidak ada unit/component frontend test, Playwright lokal timeout. |
| Repository hygiene | Bermasalah | ZIP membawa `.git`, binary `.exe~`, test-results, dan dokumen audit lama. |

---

## 2. Metodologi Audit

Audit dilakukan dengan pendekatan berikut:

1. **Static code review** atas backend Go, frontend Vue/TypeScript, konfigurasi, CI, release workflow, dokumentasi, dan struktur repository.
2. **Build validation** pada frontend menggunakan `npm ci`, `npm run build`, dan `npm audit`.
3. **Attempted backend validation** menggunakan `go test ./...`, tetapi gagal karena constraint toolchain.
4. **Security review** terhadap autentikasi, password handling, token handling, SSRF protection, header policy, CORS/origin/host guard, error handling, dan logging.
5. **Architecture review** terhadap pemisahan modul, service layer, konfigurasi, state snapshot, forwarder, runtime, dan lifecycle app.
6. **Frontend/UI review** terhadap routing, navigasi, state management, loading/error state, dummy data, dead page, dan kontrol yang tidak terhubung API.
7. **Repository/package hygiene review** terhadap artefak build, binary, `.git`, test-results, dokumen usang, dan file besar.

Standar pembanding yang digunakan: OWASP ASVS, OWASP Top 10, dan prinsip Core Web Vitals untuk performa web.

---

## 3. Lingkungan Validasi

| Komponen | Nilai |
|---|---|
| OS/container audit | Linux container |
| Go tersedia | `go1.23.2 linux/amd64` |
| Go yang diminta proyek | `go 1.25.5` pada `go.mod` |
| Node | `v22.16.0` |
| npm | `10.9.2` |
| Frontend package manager | npm, `package-lock.json` tersedia |
| Backend module | `github.com/atlasbridge/atlasbridge` |
| Frontend framework | Vue 3 + Vite + TypeScript + Pinia + Vue Router + Tailwind/DaisyUI |

---

## 4. Hasil Validasi Perintah

| Perintah | Hasil | Catatan |
|---|---:|---|
| `npm ci --ignore-scripts` pada `web/` | Lulus | 87 packages installed; audit awal menunjukkan 0 vulnerability. |
| `npm run build` pada `web/` | Lulus | `vue-tsc -b` dan `vite build` berhasil. |
| `npm audit --audit-level=high` | Lulus | `found 0 vulnerabilities`. |
| `go test ./...` | Tidak dapat dijalankan | Local Go `1.23.2`, proyek membutuhkan `1.25.5`; auto-download toolchain gagal karena network/DNS blocked. |
| `npx playwright test` | Gagal validasi lokal | Timeout menunggu `webServer` pada konfigurasi Playwright. Perlu investigasi terpisah. |
| `staticcheck` | Tidak dijalankan | Tool tidak tersedia lokal; CI meng-install `@latest`. |
| `govulncheck` | Tidak dijalankan | Tool tidak tersedia lokal; CI meng-install `@latest`. |

**Catatan penting:** frontend `web/dist` tidak ada di ZIP asli, tetapi tercipta setelah audit menjalankan `npm run build`. Ini penting karena backend menggunakan `//go:embed dist` pada `web/assets.go`, sehingga job Go yang tidak membangun frontend lebih dulu berisiko gagal compile pada environment bersih.

---

## 5. Inventaris Proyek

| Item | Temuan |
|---|---|
| Jumlah file non-`.git`, non-`node_modules`, non-`dist` | 186 file |
| Ekstensi dominan | `.go` 55, `.png` 29, `.vue` 27, `.md` 22, `.ts` 15 |
| File terbesar di ZIP asli | `.git` pack files, `atlasbridge.exe~` 12.56 MB, `docs/Atlas Bridge.png` 1.13 MB |
| Binary dalam source package | `atlasbridge.exe~` terdeteksi sebagai PE32+ Windows executable |
| `.git` dalam ZIP | Ada; ukuran besar dan membawa riwayat internal |
| `web/test-results/` dalam ZIP | Ada; artefak test sebaiknya tidak ikut source distribution |
| Dokumen audit lama | Ada beberapa audit report lama yang berpotensi menyebabkan documentation drift |

---

## 6. Temuan Prioritas Tertinggi

| Prioritas | Temuan | Dampak |
|---:|---|---|
| P0 | Password admin memakai SHA-256 unsalted | Password user-chosen mudah diserang jika hash bocor. |
| P0 | Login admin tanpa rate limiting/backoff | Brute force attack terhadap password admin menjadi realistis. |
| P0 | SSRF protection tidak melakukan DNS/IP validation saat koneksi | Hostname publik dapat resolve ke private/link-local address. |
| P0 | CI Go job berisiko gagal karena `web/dist` tidak ada tetapi di-embed | Pipeline Go test/lint/race/govulncheck bisa gagal dari checkout bersih. |
| P1 | `/logs` route salah diarahkan ke `PrivacySettings.vue` | Halaman `Logs.vue` menjadi zombie/dead page. |
| P1 | `server.admin_path` tidak benar-benar mengubah route backend | UI/config memberi ilusi bisa mengubah admin path, tetapi router hardcoded `/admin`. |
| P1 | Topbar Start/Stop Proxy hanya mengubah state lokal | UI memberi status palsu dan tidak memanggil API runtime. |
| P1 | Setup wizard unauthenticated memanggil endpoint protected | First-run flow membingungkan/berpotensi gagal. |
| P1 | Error handling admin mengembalikan `err.Error()` mentah | Potensi leak detail internal/path/konfigurasi. |
| P1 | Stream idle timeout tidak efektif saat `ReadBytes` block | Streaming dapat menggantung lebih lama dari konfigurasi. |

---
