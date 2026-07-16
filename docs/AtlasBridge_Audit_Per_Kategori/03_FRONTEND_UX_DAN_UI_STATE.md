# AtlasBridge — Frontend, UX, dan UI State

**Catatan:** File ini adalah pecahan laporan audit utama agar temuan dapat dikerjakan per kategori. Tidak ada perubahan source code yang dilakukan.

---

## 9. Temuan Detail — Frontend, UX, dan UI State

### FE-01 — Route `/logs` salah diarahkan ke Privacy page

**Severity:** High  
**Evidence:** `web/src/router/index.ts:47-50`, `web/src/pages/Logs.vue`  
**Temuan:** path `/logs` memuat `PrivacySettings.vue`, padahal `Logs.vue` tersedia.  
**Dampak:** `Logs.vue` menjadi dead/zombie page. User tidak bisa membuka halaman log yang sebenarnya dari navigasi.  
**Rekomendasi audit:** mapping route `/logs` harus diarahkan ke `Logs.vue` atau hapus file tersebut jika tidak dipakai.

### FE-02 — `PrivacySettings.vue` diroute dua kali

**Severity:** Low/Medium  
**Evidence:** `web/src/router/index.ts:47-54`  
**Temuan:** `/logs` dan `/privacy` sama-sama memuat PrivacySettings.  
**Dampak:** navigasi “Privacy & Logs” tidak sesuai isi halaman.  
**Rekomendasi audit:** pisahkan privacy dan logs, atau gabungkan konten dengan nama route yang benar.

### FE-03 — Tombol Start/Stop Proxy di topbar hanya fake local toggle

**Severity:** High  
**Evidence:** `web/src/components/Layout.vue:121-127`, `158`, `203-210`  
**Temuan:** tombol mengubah `proxyRunning` local ref dan menampilkan toast, tetapi tidak memanggil `/runtime/start` atau `/runtime/stop`.  
**Dampak:** user melihat status palsu; risiko operasional tinggi karena proxy mungkin tetap berjalan/berhenti berbeda dari UI.  
**Rekomendasi audit:** hilangkan tombol atau hubungkan ke API runtime dengan optimistic state yang benar.

### FE-04 — Runtime page ada tetapi tidak punya nav link langsung

**Severity:** Medium  
**Evidence:** `web/src/router/index.ts:32-35`, `web/src/components/Layout.vue:49-56`  
**Temuan:** route `/runtime` ada, tetapi sidebar hanya menuju `/startup`. Active state mencakup `/runtime`, namun user tidak diarahkan ke sana.  
**Dampak:** halaman runtime menjadi semi-zombie/tersembunyi.  
**Rekomendasi audit:** tambah nav tab/submenu atau hapus route jika tidak digunakan.

### FE-05 — Dashboard menilai downstream sehat hanya jika status `ok`

**Severity:** Medium  
**Evidence:** `web/src/pages/Dashboard.vue:241`, `internal/server/admin.go:309`  
**Temuan:** backend health sukses mengembalikan `status: "connected"`, sedangkan dashboard mengecek `status === "ok"`.  
**Dampak:** downstream sehat dapat ditampilkan sebagai disconnected.  
**Rekomendasi audit:** samakan kontrak status API dan frontend type.

### FE-06 — Advanced Settings berisi kontrol tidak fungsional

**Severity:** High  
**Evidence:** `web/src/pages/AdvancedSettings.vue:67-73`, `116-129`  
**Temuan:** `Proxy Auth Mode` tidak memiliki `v-model`/save binding; `Dry-run routing` dan `Expose internal headers` hanya mengubah flag `dirty`, tidak terlihat terhubung ke config.  
**Dampak:** user mengira mengubah security/runtime behavior, padahal tidak tersimpan.  
**Rekomendasi audit:** hapus kontrol placeholder atau hubungkan ke schema backend.

### FE-07 — Runtime history dummy/hardcoded

**Severity:** Medium  
**Evidence:** `web/src/pages/StartupSettings.vue:172-177`  
**Temuan:** halaman startup berisi `runtimeHistory` hardcoded tanggal 2025.  
**Dampak:** data palsu pasca-renovasi; user bisa salah percaya bahwa event benar-benar terjadi.  
**Rekomendasi audit:** ganti dengan API event log atau tampilkan empty state.

### FE-08 — Setup wizard unauthenticated tetapi menyimpan ke protected endpoint

**Severity:** High  
**Evidence:** `web/src/router/index.ts:69-78`, `web/src/pages/SetupWizard.vue:210-212`, `web/src/stores/config.ts`  
**Temuan:** route `/setup` boleh dibuka tanpa token, tetapi save config memanggil endpoint admin protected.  
**Dampak:** first-run onboarding berpotensi gagal atau membingungkan, terutama karena login UI meminta password yang mungkin belum diset.  
**Rekomendasi audit:** definisikan bootstrap flow: setup token one-time, password creation first, atau setup only after auth.

### FE-09 — Login copy menyebut password, tetapi sistem masih punya legacy token behavior

**Severity:** Medium  
**Evidence:** `web/src/pages/Login.vue`, `internal/server/admin.go:582-588`  
**Temuan:** UI hanya meminta password, sedangkan backend login juga menerima legacy admin token as password.  
**Dampak:** behavior membingungkan dan dapat memperluas attack surface.  
**Rekomendasi audit:** pisahkan endpoint/token migration atau hapus legacy compatibility setelah migrasi.

### FE-10 — Logo asset tersedia tetapi Layout masih memakai inline SVG/AB circle

**Severity:** Low  
**Evidence:** `web/public/atlasbridge-logo-mark-512.png`, `docs/assets/logo/*`, `web/src/components/Layout.vue:6-17`, `86-92`  
**Temuan:** asset logo sudah ada, tetapi UI utama masih menggunakan icon inline dan “AB” circle.  
**Dampak:** branding tidak konsisten; komentar user tentang logo belum ada mungkin berasal dari UI yang tidak memakai asset final.  
**Rekomendasi audit:** gunakan single source logo asset dan hapus variasi yang tidak dipakai.

### FE-11 — Responsiveness masih lemah

**Severity:** Medium  
**Evidence:** `web/src/components/Layout.vue:4`, grid fixed di beberapa page, `web/src/style.css` tidak terlihat memiliki media query signifikan  
**Temuan:** sidebar fixed `w-[260px]`, banyak grid `grid-cols-3/4`, dan sedikit breakpoint mobile.  
**Dampak:** admin UI berpotensi buruk di laptop sempit/tablet/mobile.  
**Rekomendasi audit:** lakukan responsive audit 360/768/1024 px dan tambahkan mobile sidebar/collapse.

### FE-12 — Error handling frontend banyak console-only

**Severity:** Medium  
**Evidence:** marker `console.error` pada beberapa halaman dan API calls  
**Temuan:** beberapa catch path hanya mencatat error ke console tanpa user-facing feedback yang jelas.  
**Dampak:** user tidak tahu apakah save/load gagal.  
**Rekomendasi audit:** standarisasi toast/error banner dan retry action.

### FE-13 — Banyak inline style dan dynamic style tersebar

**Severity:** Low/Medium  
**Evidence:** `web/src/components/Layout.vue`, `web/src/pages/Dashboard.vue`, `web/src/style.css`  
**Temuan:** style tersebar antara Tailwind utility, CSS custom property, inline style, dan dynamic style.  
**Dampak:** sulit menjaga konsistensi tema dan CSP.  
**Rekomendasi audit:** konsolidasikan design token dan hindari inline style.

---
