# AtlasBridge — Halaman/Kontrol Kurang, Berlebih, atau Tidak Berguna

**Catatan:** File ini adalah pecahan laporan audit utama agar temuan dapat dikerjakan per kategori. Tidak ada perubahan source code yang dilakukan.

---

## 13. Halaman/Kontrol yang Kurang, Berlebih, atau Tidak Berguna

| Item | Status audit | Evidence | Rekomendasi |
|---|---|---|---|
| `Logs.vue` | Ada tetapi tidak terakses via `/logs` | `router/index.ts:47-50` | Route ke `Logs.vue` atau hapus. |
| `/runtime` | Ada tetapi tidak ada link langsung | `router/index.ts:32-35`, `Layout.vue:49-56` | Tambah submenu atau gabung benar dengan Startup. |
| Topbar Start/Stop | Kontrol palsu | `Layout.vue:203-210` | Hubungkan API atau hapus. |
| Proxy Auth Mode | Tidak fungsional | `AdvancedSettings.vue:67-73` | Tambah binding/schema atau hapus. |
| Dry-run routing | Tidak fungsional | `AdvancedSettings.vue:116-122` | Implementasikan atau hapus. |
| Expose internal headers | Tidak fungsional | `AdvancedSettings.vue:123-129` | Implementasikan atau hapus. |
| Runtime history | Dummy/hardcoded | `StartupSettings.vue:172-177` | Ganti event log nyata. |
| Logo UI | Inkonsisten dengan asset | `Layout.vue:6-17`, `web/public/*` | Pakai asset resmi. |
| Setup wizard | Flow belum konsisten | `router/index.ts:69-78`, `SetupWizard.vue` | Definisikan first-run bootstrap. |

---
