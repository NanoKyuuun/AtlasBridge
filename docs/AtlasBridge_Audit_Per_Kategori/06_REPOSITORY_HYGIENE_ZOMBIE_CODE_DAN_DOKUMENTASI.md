# AtlasBridge — Repository Hygiene, Zombie Code, dan Dokumentasi

**Catatan:** File ini adalah pecahan laporan audit utama agar temuan dapat dikerjakan per kategori. Tidak ada perubahan source code yang dilakukan.

---

## 12. Temuan Detail — Repository Hygiene, Zombie Code, dan Dokumentasi

### HYGIENE-01 — ZIP berisi `.git`

**Severity:** High  
**Evidence:** ZIP asli berisi `AtlasBridge/.git/`, termasuk pack files besar  
**Temuan:** source package membawa histori Git.  
**Dampak:** membocorkan history, metadata remote, branch, commit lama, dan memperbesar distribusi.  
**Rekomendasi audit:** jangan kirim `.git` pada paket audit/release kecuali memang diminta.

### HYGIENE-02 — ZIP berisi binary `atlasbridge.exe~`

**Severity:** High  
**Evidence:** `atlasbridge.exe~` 12.56 MB, file type PE32+ Windows executable  
**Temuan:** binary dengan ekstensi backup `~` ada di source package.  
**Dampak:** risiko supply-chain, malware false positive, ukuran besar, dan ketidakjelasan versi binary.  
**Rekomendasi audit:** hapus dari source package; build binary hanya melalui release workflow terverifikasi.

### HYGIENE-03 — ZIP berisi `web/test-results/`

**Severity:** Low/Medium  
**Evidence:** ZIP asli mengandung `AtlasBridge/web/test-results/`  
**Temuan:** artefak test ikut terkirim.  
**Dampak:** noise, potensi data internal, dan memperbesar package.  
**Rekomendasi audit:** tambahkan ke `.gitignore`/export-ignore dan bersihkan sebelum zip.

### HYGIENE-04 — Dokumen audit lama dan report ganda menyebabkan drift

**Severity:** Medium  
**Evidence:** `docs/AtlasBridge_Audit_Report.md`, `docs/AtlasBridge_Comprehensive_Audit_Report_2026-07-11.md`, `docs/AtlasBridge_Audit_Markdown/*`  
**Temuan:** beberapa audit report lama masih ada di repo. Sebagian temuan mungkin sudah berubah setelah renovasi.  
**Dampak:** pembaca tidak tahu report mana yang authoritative.  
**Rekomendasi audit:** arsipkan audit lama ke folder `docs/archive/` dan tandai superseded.

### HYGIENE-05 — Folder `docs/tamplate/` typo

**Severity:** Low  
**Evidence:** `docs/tamplate/fe.html`  
**Temuan:** nama folder “tamplate” typo.  
**Dampak:** kecil, tetapi menunjukkan kurangnya kerapian pasca-renovasi.  
**Rekomendasi audit:** rename ke `template` atau hapus jika tidak digunakan.

### HYGIENE-06 — `.qwen/skills` ikut dalam ZIP

**Severity:** Low/Medium  
**Evidence:** `.qwen/skills/auto-skill-security-doc-verification/SKILL.md`  
**Temuan:** folder tool/agent internal ikut paket.  
**Dampak:** noise dan potensi bocor proses internal.  
**Rekomendasi audit:** exclude dari source release kecuali memang bagian project.

### HYGIENE-07 — Git working tree terlihat tidak clean setelah ekstraksi

**Severity:** Medium  
**Evidence:** `git status --short` menampilkan banyak `M` pada file utama setelah ekstraksi  
**Temuan:** karena ZIP membawa `.git`, status menunjukkan banyak file berbeda dari index. Ini bisa disebabkan line endings atau working tree memang tidak sinkron.  
**Dampak:** sulit memastikan source yang diaudit sama dengan commit tertentu.  
**Rekomendasi audit:** kirim archive dari commit bersih (`git archive`) atau sertakan commit SHA authoritative.

### HYGIENE-08 — Logo/asset duplikasi

**Severity:** Low  
**Evidence:** `docs/assets/logo/*`, `web/public/*`, `docs/assets/app/*`, `docs/assets/web/*`  
**Temuan:** banyak ukuran/varian logo di beberapa folder. Sebagian mungkin sah, tetapi belum ada asset manifest tunggal.  
**Dampak:** risiko UI memakai logo lama/salah.  
**Rekomendasi audit:** buat asset registry dan hapus duplikat tidak perlu.

---
