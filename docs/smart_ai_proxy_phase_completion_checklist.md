# Smart AI Proxy
# Phase Completion Checklist v0.1

**Document Type:** Phase Completion Checklist  
**Project Name:** Smart AI Proxy  
**Status:** Draft v0.1  
**Based On:** PRD v1.1, Technical Plan v0.1, Implementation Plan v0.2  
**Primary Runtime:** Local developer machine  
**Primary Downstream:** 9Router  
**Default Smart AI Proxy Port:** `20127`  
**Default 9Router Port:** `20128`  
**Backend/Core:** Go / Golang  
**HTTP Layer:** `net/http` + `chi`  
**Frontend:** Vue 3 + TypeScript + Vite + Tailwind CSS + DaisyUI  
**Tray App:** Go systray-first, Wails optional post-MVP  
**MVP Target Platform:** Windows x64 first  

---

## 1. Purpose

Dokumen ini digunakan untuk memastikan setiap fase implementasi Smart AI Proxy benar-benar selesai sebelum tim melanjutkan ke fase berikutnya.

Checklist ini berfungsi sebagai:

- quality gate per fase,
- kontrol agar scope tidak melebar,
- alat tracking progress engineering,
- dasar sign-off product, engineering, QA, DevOps, dan documentation,
- pengingat batas tanggung jawab antara Smart AI Proxy dan 9Router.

Smart AI Proxy hanya bertanggung jawab sebagai intelligent decision layer:

```text
Receive OpenAI-compatible request
↓
Analyze request
↓
Classify task
↓
Select route profile
↓
Forward request to 9Router
```

9Router tetap bertanggung jawab untuk:

```text
Provider routing
Failover
Fallback model
Load balancing
Account rotation
Rate limit handling
Provider credential handling
```

---

## 2. Global Completion Rules

Sebelum sebuah fase dinyatakan selesai, seluruh aturan global berikut harus dipenuhi.

### 2.1 Engineering Rules

- [ ] Tidak ada perubahan yang merusak OpenAI-compatible request format.
- [ ] Tidak ada perubahan yang merusak OpenAI-compatible response format.
- [ ] Streaming behavior tetap berjalan jika fase menyentuh proxy path.
- [ ] Error handling jelas dan tidak membuat user bingung.
- [ ] Feature tidak mengambil alih fungsi provider-level milik 9Router.
- [ ] Default port Smart AI Proxy tetap `20127` kecuali user mengubah config.
- [ ] Default downstream 9Router tetap `http://127.0.0.1:20128/v1`.
- [ ] Web UI dan API tetap local-first.
- [ ] Default bind address tetap `127.0.0.1`.

### 2.2 Security and Privacy Rules

- [ ] Authorization header tidak masuk log.
- [ ] API key tidak masuk log.
- [ ] Prompt penuh tidak masuk log secara default.
- [ ] Source code penuh tidak masuk log secara default.
- [ ] Diagnostics export meredact secret.
- [ ] Web UI tidak terbuka ke public network secara default.
- [ ] Downstream endpoint tidak bisa dioverride sembarangan dari request client.
- [ ] Config invalid tidak boleh diterapkan.

### 2.3 QA Rules

- [ ] Unit test utama lulus.
- [ ] Integration test yang relevan lulus.
- [ ] Manual checklist fase sudah dijalankan.
- [ ] Regression issue dari fase sebelumnya tidak muncul lagi.
- [ ] Known issues didokumentasikan.
- [ ] P0 bug = 0 sebelum fase ditutup.
- [ ] P1 bug yang tersisa sudah disetujui untuk carry-over.

### 2.4 Documentation Rules

- [ ] README atau docs terkait diperbarui.
- [ ] Config example diperbarui jika ada perubahan config.
- [ ] Troubleshooting diperbarui jika ada error baru.
- [ ] Release notes internal diperbarui.
- [ ] User-facing behavior terdokumentasi.

---

## 3. Phase Completion Status Board

Gunakan tabel ini sebagai ringkasan status lintas fase.

| Phase | Name | Status | Owner | Target Output | Sign-off |
|---|---|---|---|---|---|
| Phase 0 | Technical Foundation | Not Started / In Progress / Done | Engineering Lead | Repo, skeleton, config draft | Product + Engineering |
| Phase 1 | Core Proxy MVP | Not Started / In Progress / Done | Backend Lead | OpenAI-compatible proxy working | Engineering + QA |
| Phase 2 | Routing Intelligence MVP | Not Started / In Progress / Done | Backend Lead | `smart-auto`, classifier, route profile | Product + Engineering + QA |
| Phase 3 | Local Web UI | Not Started / In Progress / Done | Frontend Lead | Web UI settings usable | Product + UX + QA |
| Phase 4 | Tray and Auto-Start | Not Started / In Progress / Done | Desktop Lead | Tray app + startup behavior | Product + Engineering + QA |
| Phase 5 | Packaging and Distribution | Not Started / In Progress / Done | DevOps Lead | Portable build + npm wrapper | DevOps + QA |
| Phase 6 | QA, Compatibility, and Hardening | Not Started / In Progress / Done | QA Lead | Stable release candidate | QA + Security + Product |
| Phase 7 | MVP Release | Not Started / In Progress / Done | Release Manager | `v0.1.0-alpha` | Product + Engineering |
| Phase 8 | Post-MVP Improvements | Not Started / In Progress / Done | Product + Engineering | Roadmap execution | Product |

---

# Phase 0 — Technical Foundation Completion Checklist

## 4.1 Phase Goal

Menyiapkan struktur project, skeleton aplikasi, keputusan teknis final, config draft, dan development workflow awal.

## 4.2 Required Deliverables

- [ ] Repository structure dibuat.
- [ ] Go module diinisialisasi.
- [ ] Vue 3 + TypeScript + Vite app diinisialisasi.
- [ ] Tailwind CSS dikonfigurasi.
- [ ] DaisyUI dikonfigurasi.
- [ ] Basic Go app bootstrap tersedia.
- [ ] Basic HTTP server skeleton tersedia.
- [ ] `chi` router terpasang.
- [ ] Health endpoint skeleton tersedia.
- [ ] Config schema draft tersedia.
- [ ] `config.example.yaml` tersedia.
- [ ] `routes.example.yaml` tersedia.
- [ ] `profiles.example.yaml` tersedia.
- [ ] Default Smart AI Proxy port `20127` ditetapkan.
- [ ] Default 9Router port `20128` ditetapkan.
- [ ] Development scripts tersedia.
- [ ] Initial README tersedia.

## 4.3 Repository Checklist

- [ ] Folder `cmd/smart-ai-proxy/` dibuat.
- [ ] Folder `internal/app/` dibuat.
- [ ] Folder `internal/server/` dibuat.
- [ ] Folder `internal/proxy/` dibuat.
- [ ] Folder `internal/analyzer/` dibuat.
- [ ] Folder `internal/classifier/` dibuat.
- [ ] Folder `internal/routing/` dibuat.
- [ ] Folder `internal/forwarder/` dibuat.
- [ ] Folder `internal/config/` dibuat.
- [ ] Folder `internal/storage/` dibuat.
- [ ] Folder `internal/observability/` dibuat.
- [ ] Folder `internal/security/` dibuat.
- [ ] Folder `internal/startup/` dibuat.
- [ ] Folder `internal/tray/` dibuat.
- [ ] Folder `web/` dibuat.
- [ ] Folder `configs/` dibuat.
- [ ] Folder `docs/` dibuat.
- [ ] Folder `testdata/` dibuat.
- [ ] Folder `scripts/` dibuat.
- [ ] Folder `packaging/` dibuat.
- [ ] Folder `.github/workflows/` disiapkan.

## 4.4 Technical Validation

- [ ] App bisa dijalankan secara lokal.
- [ ] Server bind ke `127.0.0.1:20127`.
- [ ] `/health` mengembalikan response minimal.
- [ ] Web app bisa dijalankan dalam development mode.
- [ ] Config default bisa digenerate.
- [ ] Tidak ada provider integration langsung.
- [ ] Tidak ada hardcoded provider credential.
- [ ] Default config tidak membuka akses LAN.

## 4.5 Testing Checklist

- [ ] Go build berhasil.
- [ ] Frontend dev server berhasil jalan.
- [ ] Health endpoint bisa diakses.
- [ ] Config example bisa dibaca parser.
- [ ] Basic lint/format dijalankan.

## 4.6 Documentation Checklist

- [ ] README menjelaskan project purpose.
- [ ] README menjelaskan default port `20127`.
- [ ] README menjelaskan downstream 9Router `20128`.
- [ ] Setup development awal terdokumentasi.
- [ ] Architecture folder structure terdokumentasi.

## 4.7 Exit Criteria

Phase 0 selesai jika:

- [ ] Skeleton backend dan frontend tersedia.
- [ ] App bisa start lokal.
- [ ] Health endpoint aktif.
- [ ] Config default tersedia.
- [ ] Repo structure disepakati.
- [ ] Tidak ada architectural blocker untuk Phase 1.

## 4.8 Phase Sign-off

| Role | Name | Status | Notes |
|---|---|---|---|
| Product Owner |  | Pending |  |
| Engineering Lead |  | Pending |  |
| Backend Lead |  | Pending |  |
| Frontend Lead |  | Pending |  |

---

# Phase 1 — Core Proxy MVP Completion Checklist

## 5.1 Phase Goal

Membuat Smart AI Proxy dapat menerima request OpenAI-compatible dan meneruskannya ke 9Router dengan response streaming dan non-streaming.

## 5.2 Required Deliverables

- [ ] Endpoint `/v1/chat/completions` tersedia.
- [ ] Endpoint `/v1/models` tersedia minimal untuk smart aliases.
- [ ] Endpoint `/health` tersedia.
- [ ] Forwarder ke `http://127.0.0.1:20128/v1` tersedia.
- [ ] Non-streaming passthrough berfungsi.
- [ ] Streaming passthrough berfungsi.
- [ ] Request ID digenerate per request.
- [ ] Basic metadata log tersedia.
- [ ] Downstream health check tersedia.
- [ ] Error dari 9Router diteruskan dengan jelas.
- [ ] Client disconnect handling tersedia.
- [ ] Header redaction baseline tersedia.

## 5.3 OpenAI Compatibility Checklist

- [ ] Request field `model` dipertahankan atau ditransformasi sesuai policy.
- [ ] Request field `messages` tidak rusak.
- [ ] Request field `temperature` dipertahankan.
- [ ] Request field `stream` dipertahankan.
- [ ] Request field `tools` dipertahankan jika ada.
- [ ] Request field `tool_choice` dipertahankan jika ada.
- [ ] Response JSON non-streaming tetap compatible.
- [ ] Response streaming tetap dalam format SSE.
- [ ] Status code downstream dipropagasi semampunya.
- [ ] Error response masih dapat dipahami client.

## 5.4 Streaming Checklist

- [ ] Streaming tidak dibuffer penuh.
- [ ] Chunk dari 9Router diteruskan bertahap ke client.
- [ ] `Content-Type` streaming sesuai.
- [ ] Client disconnect membatalkan downstream request.
- [ ] Timeout streaming jelas.
- [ ] Error saat streaming tidak membuat app crash.
- [ ] Streaming diuji dengan mock 9Router.

## 5.5 Logging and Privacy Checklist

- [ ] Log berisi `request_id`.
- [ ] Log berisi timestamp.
- [ ] Log berisi requested model.
- [ ] Log berisi status code.
- [ ] Log berisi latency.
- [ ] Log berisi streaming true/false.
- [ ] Authorization header tidak masuk log.
- [ ] API key tidak masuk log.
- [ ] Prompt penuh tidak masuk log.
- [ ] Source code penuh tidak masuk log.

## 5.6 Failure Handling Checklist

- [ ] 9Router offline menghasilkan error yang jelas.
- [ ] Downstream timeout menghasilkan error yang jelas.
- [ ] Invalid OpenAI-compatible request menghasilkan error yang jelas.
- [ ] Server tidak panic saat downstream gagal.
- [ ] Smart AI Proxy tidak mencoba provider fallback sendiri.
- [ ] Smart AI Proxy tidak melakukan load balancing provider.

## 5.7 Testing Checklist

- [ ] Unit test request ID generation lulus.
- [ ] Unit test header redaction lulus.
- [ ] Integration test non-streaming passthrough lulus.
- [ ] Integration test streaming passthrough lulus.
- [ ] Integration test downstream unavailable lulus.
- [ ] Manual test dengan curl lulus.
- [ ] Generic OpenAI-compatible request lulus.

## 5.8 Exit Criteria

Phase 1 selesai jika:

- [ ] `/v1/chat/completions` dapat dipakai.
- [ ] Request berhasil diteruskan ke 9Router.
- [ ] Streaming berjalan.
- [ ] Non-streaming berjalan.
- [ ] Metadata log tersedia.
- [ ] Secret tidak bocor di log.
- [ ] Tidak ada provider-level routing di Smart AI Proxy.

## 5.9 Phase Sign-off

| Role | Name | Status | Notes |
|---|---|---|---|
| Backend Lead |  | Pending |  |
| QA Lead |  | Pending |  |
| Security Reviewer |  | Pending |  |
| Engineering Lead |  | Pending |  |

---

# Phase 2 — Routing Intelligence MVP Completion Checklist

## 6.1 Phase Goal

Menambahkan request analyzer, rule-based classifier, route resolver, smart aliases, manual override, dan safe passthrough.

## 6.2 Required Deliverables

- [ ] Request analyzer tersedia.
- [ ] Code block detector tersedia.
- [ ] Keyword detector tersedia.
- [ ] Domain category detector tersedia.
- [ ] Error pattern detector tersedia.
- [ ] Complexity estimator tersedia.
- [ ] Long-context detector tersedia.
- [ ] Rule-based classifier tersedia.
- [ ] Confidence score tersedia.
- [ ] Route resolver tersedia.
- [ ] Route profile selector tersedia.
- [ ] Smart aliases tersedia.
- [ ] Manual override tersedia.
- [ ] Safe passthrough tersedia.
- [ ] Routing decision log tersedia.
- [ ] Downstream metadata strategy tersedia.

## 6.3 Task Classification Checklist

- [ ] `general_chat` dikenali.
- [ ] `design_task` dikenali.
- [ ] `backend_engineering` dikenali.
- [ ] `frontend_engineering` dikenali.
- [ ] `fullstack_engineering` dikenali.
- [ ] `debugging` dikenali.
- [ ] `refactoring` dikenali.
- [ ] `test_generation` dikenali.
- [ ] `documentation` dikenali.
- [ ] `architecture_design` dikenali.
- [ ] `security_review` dikenali.
- [ ] `long_context_analysis` dikenali.
- [ ] `lightweight_task` dikenali.
- [ ] `unknown` fallback tersedia.

## 6.4 Default Route Mapping Checklist

- [ ] `general_chat` → `route.default`.
- [ ] `design_task` → `route.design`.
- [ ] `backend_engineering` → `route.backend`.
- [ ] `frontend_engineering` → `route.frontend`.
- [ ] `fullstack_engineering` → `route.fullstack`.
- [ ] `debugging` → `route.debugging`.
- [ ] `refactoring` → `route.refactoring`.
- [ ] `test_generation` → `route.testing`.
- [ ] `documentation` → `route.documentation`.
- [ ] `architecture_design` → `route.architect`.
- [ ] `security_review` → `route.security`.
- [ ] `long_context_analysis` → `route.long_context`.
- [ ] `lightweight_task` → `route.low_cost`.
- [ ] `unknown` → `route.default`.

## 6.5 Smart Alias Checklist

- [ ] `smart-auto` menjalankan auto classification.
- [ ] `smart-debug` memilih `route.debugging`.
- [ ] `smart-docs` memilih `route.documentation`.
- [ ] `smart-cheap` memilih `route.low_cost`.
- [ ] `smart-fast` memilih route cepat sesuai config.
- [ ] `smart-long-context` memilih `route.long_context`.
- [ ] Alias tidak valid masuk fallback yang jelas.
- [ ] Alias override tidak ditimpa auto classification.

## 6.6 Routing Precedence Checklist

- [ ] Explicit route override diprioritaskan.
- [ ] Smart model alias override diprioritaskan.
- [ ] User-defined task-to-route mapping digunakan.
- [ ] Project-specific policy disiapkan sebagai extension point.
- [ ] Classifier result digunakan jika tidak ada override.
- [ ] Complexity/context signal dipakai sebagai tambahan.
- [ ] Default route tersedia.
- [ ] Safe passthrough tersedia.

## 6.7 Downstream Metadata Checklist

- [ ] Transport mode `model_alias` didukung.
- [ ] Downstream alias dari route profile bisa dipakai.
- [ ] Header metadata disiapkan sebagai opsi jika dibutuhkan.
- [ ] Fallback ke default model/route tersedia.
- [ ] Request prompt tidak diubah agresif.
- [ ] Route intent bisa dilacak di metadata log.

## 6.8 Testing Checklist

- [ ] Unit test keyword classifier lulus.
- [ ] Unit test code block detector lulus.
- [ ] Unit test stack trace detector lulus.
- [ ] Unit test backend/frontend/design classification lulus.
- [ ] Unit test confidence threshold lulus.
- [ ] Unit test route resolver lulus.
- [ ] Integration test `smart-auto` lulus.
- [ ] Integration test `smart-debug` override lulus.
- [ ] Integration test classifier failure safe passthrough lulus.
- [ ] Evaluation dataset awal tersedia.

## 6.9 Exit Criteria

Phase 2 selesai jika:

- [ ] `smart-auto` berfungsi.
- [ ] Task utama dapat diklasifikasikan.
- [ ] Route profile dapat dipilih.
- [ ] Manual override bekerja.
- [ ] Safe passthrough bekerja.
- [ ] Log routing tidak menyimpan prompt penuh.
- [ ] 9Router tetap menjadi satu-satunya downstream execution layer.

## 6.10 Phase Sign-off

| Role | Name | Status | Notes |
|---|---|---|---|
| Product Owner |  | Pending |  |
| Backend Lead |  | Pending |  |
| QA Lead |  | Pending |  |
| Engineering Lead |  | Pending |  |

---

# Phase 3 — Local Web UI Completion Checklist

## 7.1 Phase Goal

Membuat user dapat mengatur Smart AI Proxy melalui Web UI lokal tanpa perlu edit config manual.

## 7.2 Required Deliverables

- [ ] Embedded Vue Web UI tersedia.
- [ ] Admin API tersedia.
- [ ] Dashboard tersedia.
- [ ] Setup Wizard tersedia.
- [ ] Routing Settings page tersedia.
- [ ] Route Profiles page tersedia.
- [ ] Runtime page tersedia.
- [ ] Startup page tersedia.
- [ ] 9Router settings page tersedia.
- [ ] Privacy/logging page tersedia.
- [ ] Advanced settings page tersedia.
- [ ] Config import/export tersedia.
- [ ] Config validation feedback tersedia.
- [ ] Basic dry-run tester tersedia atau diputuskan masuk post-MVP.

## 7.3 Admin API Checklist

- [ ] `GET /admin/api/status` tersedia.
- [ ] `GET /admin/api/config` tersedia.
- [ ] `PUT /admin/api/config` tersedia.
- [ ] `GET /admin/api/routes` tersedia.
- [ ] `PUT /admin/api/routes` tersedia.
- [ ] `GET /admin/api/profiles` tersedia.
- [ ] `PUT /admin/api/profiles` tersedia.
- [ ] `POST /admin/api/runtime/start` tersedia.
- [ ] `POST /admin/api/runtime/stop` tersedia.
- [ ] `POST /admin/api/runtime/restart` tersedia.
- [ ] `GET /admin/api/startup` tersedia.
- [ ] `PUT /admin/api/startup` tersedia.
- [ ] `GET /admin/api/downstream/health` tersedia.
- [ ] `GET /admin/api/logs` tersedia.
- [ ] `POST /admin/api/diagnostics/export` tersedia.
- [ ] `POST /admin/api/routing/dry-run` tersedia atau ditandai post-MVP.
- [ ] `POST /admin/api/config/import` tersedia atau ditandai post-MVP.
- [ ] `GET /admin/api/config/export` tersedia atau ditandai post-MVP.
- [ ] `POST /admin/api/config/reset` tersedia.

## 7.4 Web UI Page Checklist

- [ ] Dashboard dapat dibuka di `/admin`.
- [ ] Setup Wizard dapat dibuka di `/admin/setup`.
- [ ] Routing Settings dapat dibuka di `/admin/routing`.
- [ ] Route Profiles dapat dibuka di `/admin/profiles`.
- [ ] Runtime page dapat dibuka di `/admin/runtime`.
- [ ] Startup page dapat dibuka di `/admin/startup`.
- [ ] 9Router page dapat dibuka di `/admin/downstream`.
- [ ] Logs page dapat dibuka di `/admin/logs`.
- [ ] Privacy page dapat dibuka di `/admin/privacy`.
- [ ] Advanced page dapat dibuka di `/admin/advanced`.

## 7.5 Dashboard Checklist

- [ ] Menampilkan proxy status.
- [ ] Menampilkan API endpoint `http://127.0.0.1:20127/v1`.
- [ ] Menampilkan admin URL.
- [ ] Menampilkan 9Router status.
- [ ] Menampilkan current runtime mode.
- [ ] Menampilkan startup mode.
- [ ] Menampilkan auto-routing status.
- [ ] Menampilkan total request hari ini.
- [ ] Menampilkan most used route jika data tersedia.
- [ ] Menampilkan last error jika ada.

## 7.6 Routing Settings Checklist

- [ ] Task category ditampilkan dalam table.
- [ ] Route profile dropdown tersedia per task.
- [ ] Enable/disable route per task tersedia.
- [ ] Default route selector tersedia.
- [ ] Low confidence route selector tersedia.
- [ ] Save button tersedia.
- [ ] Reset to default tersedia.
- [ ] Invalid mapping ditolak.
- [ ] Perubahan mapping dipakai request berikutnya.

## 7.7 Route Profiles Checklist

- [ ] User dapat melihat daftar route profiles.
- [ ] User dapat membuat route profile.
- [ ] User dapat mengedit route profile.
- [ ] User dapat menonaktifkan route profile.
- [ ] Route name harus unik.
- [ ] Downstream alias tidak boleh kosong.
- [ ] Default route tidak bisa dinonaktifkan.
- [ ] Task route tidak boleh menunjuk profile yang hilang.

## 7.8 Runtime and Startup UI Checklist

- [ ] User dapat Start Proxy.
- [ ] User dapat Stop Proxy.
- [ ] User dapat Restart Proxy.
- [ ] User dapat Copy API Endpoint.
- [ ] User dapat Open Logs Folder.
- [ ] User dapat Open 9Router Dashboard.
- [ ] User dapat Test 9Router Connection.
- [ ] User dapat memilih Always On.
- [ ] User dapat memilih Manual.
- [ ] User dapat memilih Disabled.
- [ ] User dapat toggle Run at Login.
- [ ] User dapat toggle Start Proxy on App Launch.
- [ ] User dapat toggle Restart After Crash.

## 7.9 Security and Privacy Checklist

- [ ] Admin API auth tersedia atau first-run token tersedia.
- [ ] Admin UI local-only by default.
- [ ] Sensitive config dimasking di UI.
- [ ] Prompt logging default off.
- [ ] Metadata logging dapat dikontrol.
- [ ] Privacy mode Standard tersedia.
- [ ] Privacy mode Strict tersedia.
- [ ] Debug mode diberi warning eksplisit jika menambah logging.
- [ ] Clear logs tersedia.
- [ ] Diagnostics export redacted.

## 7.10 Testing Checklist

- [ ] Dashboard loads.
- [ ] Setup wizard works.
- [ ] Routing settings save works.
- [ ] Route profile save works.
- [ ] Runtime start/stop/restart works.
- [ ] Startup toggle persists.
- [ ] Downstream health check updates.
- [ ] Invalid config shows error.
- [ ] Logs do not show secrets.
- [ ] Config persists after restart.

## 7.11 Exit Criteria

Phase 3 selesai jika:

- [ ] User dapat membuka Web UI lokal.
- [ ] User dapat mengubah task-to-route mapping.
- [ ] User dapat mengelola route profile minimal.
- [ ] User dapat mengubah downstream 9Router endpoint.
- [ ] User dapat start/stop/restart proxy.
- [ ] User dapat mengubah startup mode.
- [ ] Config tersimpan dan valid.
- [ ] Web UI tetap localhost-only default.

## 7.12 Phase Sign-off

| Role | Name | Status | Notes |
|---|---|---|---|
| Product Owner |  | Pending |  |
| Frontend Lead |  | Pending |  |
| Backend Lead |  | Pending |  |
| UX Reviewer |  | Pending |  |
| QA Lead |  | Pending |  |

---

# Phase 4 — Tray and Auto-Start Completion Checklist

## 8.1 Phase Goal

Membuat Smart AI Proxy terasa seperti aplikasi lokal background dengan icon di system tray seperti 9Router.

## 8.2 Required Deliverables

- [ ] Tray icon tersedia.
- [ ] Tray active status tersedia.
- [ ] Tray inactive status tersedia.
- [ ] Tray error status tersedia.
- [ ] Tray context menu tersedia.
- [ ] Open Dashboard dari tray tersedia.
- [ ] Start Proxy dari tray tersedia.
- [ ] Stop Proxy dari tray tersedia.
- [ ] Restart Proxy dari tray tersedia.
- [ ] Copy API Endpoint dari tray tersedia.
- [ ] Open 9Router Dashboard dari tray tersedia.
- [ ] Open Logs Folder dari tray tersedia.
- [ ] Run at Startup toggle tersedia.
- [ ] Always On/Manual/Disabled terintegrasi.
- [ ] Windows startup registration tersedia.
- [ ] Port conflict notification tersedia.
- [ ] Downstream disconnected warning tersedia.
- [ ] Single-instance lock tersedia.

## 8.3 Tray Menu Checklist

- [ ] Menu menampilkan nama Smart AI Proxy.
- [ ] Menu menampilkan status Running/Stopped/Error.
- [ ] Menu menampilkan endpoint `127.0.0.1:20127`.
- [ ] Menu memiliki Open Dashboard.
- [ ] Menu memiliki Open 9Router Dashboard.
- [ ] Menu memiliki Copy API Endpoint.
- [ ] Menu memiliki Start Proxy.
- [ ] Menu memiliki Stop Proxy.
- [ ] Menu memiliki Restart Proxy.
- [ ] Menu memiliki Run at Startup ON/OFF.
- [ ] Menu memiliki Always On Mode ON/OFF atau mode selector.
- [ ] Menu memiliki Open Logs Folder.
- [ ] Menu memiliki Quit.

## 8.4 Runtime State Checklist

- [ ] State `starting` tampil benar.
- [ ] State `running` tampil benar.
- [ ] State `stopped` tampil benar.
- [ ] State `disabled` tampil benar.
- [ ] State `error` tampil benar.
- [ ] State `port_conflict` tampil benar.
- [ ] State `downstream_offline` tampil benar.
- [ ] State `quitting` membersihkan tray icon.

## 8.5 Auto-Start Checklist

- [ ] Auto-start menjalankan tray app, bukan hanya proxy engine.
- [ ] Windows user-level startup registration tersedia.
- [ ] Startup folder atau Run key strategy dipilih.
- [ ] Always On menjalankan proxy otomatis setelah login.
- [ ] Manual mode tidak menjalankan proxy otomatis kecuali config meminta.
- [ ] Disabled mode tidak menerima request setelah login.
- [ ] Repair startup registration tersedia atau didokumentasikan.
- [ ] Duplicate instance dicegah.

## 8.6 Tray Behavior Checklist

- [ ] Left click membuka dashboard.
- [ ] Double click membuka dashboard.
- [ ] Right click membuka context menu.
- [ ] Quit menghentikan app dengan bersih.
- [ ] Stop Proxy tidak mematikan admin UI.
- [ ] Restart Proxy graceful.
- [ ] Jika proxy running, quit bisa memberi konfirmasi.
- [ ] Tray status sinkron dengan Web UI.

## 8.7 Testing Checklist

- [ ] App launch menampilkan tray icon.
- [ ] Open Dashboard dari tray berhasil.
- [ ] Start Proxy dari tray berhasil.
- [ ] Stop Proxy dari tray berhasil.
- [ ] Restart Proxy dari tray berhasil.
- [ ] Copy API endpoint berhasil.
- [ ] Open logs folder berhasil.
- [ ] Open 9Router dashboard berhasil.
- [ ] Run at startup berhasil setelah Windows login.
- [ ] Manual mode tidak auto-start proxy.
- [ ] Disabled mode tetap disabled.
- [ ] Port conflict terlihat di tray.
- [ ] 9Router offline terlihat sebagai warning.
- [ ] Quit menghilangkan tray icon.

## 8.8 Exit Criteria

Phase 4 selesai jika:

- [ ] Tray icon muncul saat app aktif.
- [ ] Tray menu bekerja.
- [ ] Runtime control dari tray bekerja.
- [ ] Auto-start Windows bekerja.
- [ ] Mode Always On, Manual, Disabled bekerja.
- [ ] Status tray dan Web UI sinkron.
- [ ] Tidak ada konflik antara tray app dan runtime manager.

## 8.9 Phase Sign-off

| Role | Name | Status | Notes |
|---|---|---|---|
| Product Owner |  | Pending |  |
| Desktop Lead |  | Pending |  |
| Backend Lead |  | Pending |  |
| QA Lead |  | Pending |  |

---

# Phase 5 — Packaging and Distribution Completion Checklist

## 9.1 Phase Goal

Membuat Smart AI Proxy mudah di-build, diinstall, dijalankan, dan didistribusikan kepada developer.

## 9.2 Required Deliverables

- [ ] Frontend production build tersedia.
- [ ] Vue dist di-embed ke Go binary.
- [ ] Windows x64 build script tersedia.
- [ ] Version metadata tersedia.
- [ ] Portable release zip tersedia.
- [ ] GitHub Release workflow tersedia atau draft tersedia.
- [ ] npm wrapper package tersedia.
- [ ] Checksum generation tersedia.
- [ ] Release notes template tersedia.
- [ ] Uninstall notes tersedia.
- [ ] Quick start ikut dalam release artifact.

## 9.3 Build Checklist

- [ ] `web` build menghasilkan static assets.
- [ ] Static assets masuk ke Go binary.
- [ ] Production app tidak butuh Vite dev server.
- [ ] Production app tidak butuh Node.js runtime.
- [ ] Binary menjalankan proxy.
- [ ] Binary menjalankan admin Web UI.
- [ ] Binary menjalankan tray app.
- [ ] Version terlihat di dashboard.
- [ ] Version terlihat di logs.

## 9.4 npm Wrapper Checklist

- [ ] Package name ditetapkan.
- [ ] CLI command `smart-ai-proxy` tersedia.
- [ ] Command `smart-ai-proxy start` tersedia.
- [ ] Command `smart-ai-proxy status` tersedia.
- [ ] Command `smart-ai-proxy open` tersedia.
- [ ] Command `smart-ai-proxy tray` tersedia.
- [ ] OS/arch detection tersedia atau MVP dibatasi Windows.
- [ ] Binary path handling tersedia.
- [ ] Install failure message jelas.
- [ ] Fallback GitHub Releases didokumentasikan.

## 9.5 Release Artifact Checklist

- [ ] Windows portable zip tersedia.
- [ ] Go binary tersedia.
- [ ] npm wrapper tersedia.
- [ ] README tersedia.
- [ ] Quick start guide tersedia.
- [ ] Configuration guide tersedia.
- [ ] Troubleshooting guide tersedia.
- [ ] Release notes tersedia.
- [ ] Checksums tersedia atau ditandai recommended.
- [ ] Installer optional ditandai post-MVP jika belum ada.

## 9.6 Packaging Security Checklist

- [ ] Release artifact tidak menyertakan local config user.
- [ ] Release artifact tidak menyertakan secret.
- [ ] Release artifact tidak menyertakan token lokal.
- [ ] npm wrapper tidak mencetak secret ke console.
- [ ] Checksum tersedia untuk binary release.
- [ ] Download URL binary jelas dan terpercaya.

## 9.7 Testing Checklist

- [ ] Portable zip dapat diekstrak dan dijalankan.
- [ ] App dari portable zip menampilkan tray icon.
- [ ] Dashboard bisa dibuka dari portable build.
- [ ] API endpoint bekerja dari portable build.
- [ ] npm install global berhasil pada environment target.
- [ ] npm command dapat menjalankan binary.
- [ ] Build repeatable di CI atau lokal.

## 9.8 Exit Criteria

Phase 5 selesai jika:

- [ ] Windows portable build siap digunakan.
- [ ] Embedded Web UI bekerja.
- [ ] npm wrapper basic bekerja atau status beta terdokumentasi.
- [ ] Release artifact memiliki docs utama.
- [ ] Tidak ada secret dalam artifact.
- [ ] Build process terdokumentasi.

## 9.9 Phase Sign-off

| Role | Name | Status | Notes |
|---|---|---|---|
| DevOps Lead |  | Pending |  |
| Engineering Lead |  | Pending |  |
| QA Lead |  | Pending |  |
| Documentation Owner |  | Pending |  |

---

# Phase 6 — QA, Compatibility, and Hardening Completion Checklist

## 10.1 Phase Goal

Memastikan MVP stabil, kompatibel dengan client target, aman secara default, dan tidak merusak OpenAI-compatible behavior.

## 10.2 Unit Test Checklist

- [ ] Config loader tested.
- [ ] Config validator tested.
- [ ] Request analyzer tested.
- [ ] Classifier tested.
- [ ] Route resolver tested.
- [ ] Route profile selector tested.
- [ ] Forwarder request builder tested.
- [ ] Redactor tested.
- [ ] Metadata logger tested.
- [ ] Runtime manager tested.

## 10.3 Integration Test Checklist

- [ ] Non-streaming request to mock 9Router tested.
- [ ] Streaming request to mock 9Router tested.
- [ ] Downstream 9Router offline tested.
- [ ] Invalid OpenAI request tested.
- [ ] Classifier failure tested.
- [ ] Safe passthrough tested.
- [ ] Route profile update tested.
- [ ] Task mapping update tested.
- [ ] Config reload tested.
- [ ] Port conflict tested.

## 10.4 UI Test Checklist

- [ ] Dashboard loads.
- [ ] Setup wizard works.
- [ ] Routing settings save.
- [ ] Route profiles save.
- [ ] Startup toggle persists.
- [ ] Runtime start/stop/restart works.
- [ ] Downstream health check updates.
- [ ] Privacy settings persist.
- [ ] Invalid config shows error.
- [ ] Logs do not show secrets.

## 10.5 System Test Checklist — Windows MVP

- [ ] Portable app launches.
- [ ] Tray icon appears.
- [ ] Dashboard opens from tray.
- [ ] Proxy endpoint works.
- [ ] Auto-start after login works.
- [ ] Manual mode does not auto-start proxy.
- [ ] Disabled mode stays disabled.
- [ ] Port conflict shown.
- [ ] 9Router disconnected shown.
- [ ] Quit removes tray icon.

## 10.6 Compatibility Test Checklist

- [ ] Generic OpenAI-compatible HTTP request works.
- [ ] Generic SDK request works.
- [ ] OpenCode guide tested or marked pending.
- [ ] Cursor guide tested or marked pending.
- [ ] Cline guide tested or marked pending.
- [ ] Continue guide tested or marked pending.
- [ ] Client can configure base URL `http://127.0.0.1:20127/v1`.
- [ ] Client can use model `smart-auto`.
- [ ] Streaming response works in tested client.
- [ ] Tool/function payload is preserved.
- [ ] Errors are understandable.

## 10.7 Security Acceptance Checklist

- [ ] Admin UI binds to localhost only by default.
- [ ] Authorization headers never appear in logs.
- [ ] API keys never appear in logs.
- [ ] Full prompt logging is disabled by default.
- [ ] Diagnostics export redacts secrets.
- [ ] Downstream endpoint cannot be overridden per request.
- [ ] LAN access requires explicit opt-in.
- [ ] Invalid config cannot be saved.
- [ ] Web UI settings require local token/password if auth enabled.

## 10.8 Performance Acceptance Checklist

- [ ] Proxy overhead tidak terasa mengganggu.
- [ ] Classifier tidak memanggil LLM.
- [ ] Streaming tidak full buffering.
- [ ] Startup time wajar untuk local app.
- [ ] Web UI terasa cepat secara lokal.
- [ ] Memory footprint wajar untuk background developer app.

## 10.9 Bug Triage Checklist

- [ ] Semua P0 bug closed.
- [ ] Semua critical streaming issue closed.
- [ ] Semua secret logging issue closed.
- [ ] Semua crash startup issue closed.
- [ ] P1 bug yang tersisa memiliki owner dan target fix.
- [ ] Known issues didokumentasikan.

## 10.10 Exit Criteria

Phase 6 selesai jika:

- [ ] Test suite utama lulus.
- [ ] Manual Windows checklist lulus.
- [ ] Security acceptance lulus.
- [ ] Compatibility minimum lulus.
- [ ] No P0 bug.
- [ ] Release candidate siap dibuat.

## 10.11 Phase Sign-off

| Role | Name | Status | Notes |
|---|---|---|---|
| QA Lead |  | Pending |  |
| Security Reviewer |  | Pending |  |
| Engineering Lead |  | Pending |  |
| Product Owner |  | Pending |  |

---

# Phase 7 — MVP Release Completion Checklist

## 11.1 Phase Goal

Merilis versi MVP yang dapat digunakan developer secara realistis dengan batasan fitur yang jelas.

## 11.2 MVP Release Criteria

- [ ] App runs on Windows x64.
- [ ] Tray icon appears.
- [ ] Web UI opens at `http://127.0.0.1:20127/admin`.
- [ ] API endpoint works at `http://127.0.0.1:20127/v1`.
- [ ] 9Router downstream default is `http://127.0.0.1:20128/v1`.
- [ ] `/v1/chat/completions` works.
- [ ] Streaming works.
- [ ] Non-streaming works.
- [ ] `smart-auto` works.
- [ ] Rule-based classifier works for key task categories.
- [ ] Task-to-route mapping can be changed via Web UI.
- [ ] Route profiles can be managed at least basically.
- [ ] Runtime start/stop/restart works.
- [ ] Startup mode can be configured.
- [ ] Auto-start works on Windows login.
- [ ] Logs contain metadata only.
- [ ] Secrets are redacted.
- [ ] Config persists after restart.
- [ ] Safe passthrough works if classifier fails.
- [ ] 9Router failure shows clear message.
- [ ] No provider-level routing is implemented inside Smart AI Proxy.
- [ ] Documentation is sufficient for first-time setup.

## 11.3 Release Artifact Checklist

- [ ] Windows portable zip ready.
- [ ] Go binary ready.
- [ ] npm wrapper ready or clearly marked beta.
- [ ] README ready.
- [ ] Quick start guide ready.
- [ ] Configuration guide ready.
- [ ] Troubleshooting guide ready.
- [ ] Release notes ready.
- [ ] Checksums generated or documented as pending.
- [ ] GitHub Release draft ready.
- [ ] Version set to `v0.1.0-alpha`.

## 11.4 Release Documentation Checklist

- [ ] `README.md` updated.
- [ ] `setup.md` updated.
- [ ] `routing-policy.md` updated.
- [ ] `9router.md` updated.
- [ ] `opencode.md` created or marked pending.
- [ ] `cursor.md` created or marked pending.
- [ ] `cline.md` created or marked pending.
- [ ] `continue.md` created or marked pending.
- [ ] `security.md` updated.
- [ ] `troubleshooting.md` updated.
- [ ] `packaging.md` updated.
- [ ] `release-notes.md` updated.

## 11.5 Post-Release Monitoring Checklist

Track after MVP release:

- [ ] Install issues.
- [ ] Port conflict reports.
- [ ] 9Router connection issues.
- [ ] Streaming compatibility issues.
- [ ] Classifier wrong-route reports.
- [ ] Tray icon issues.
- [ ] Startup not working reports.
- [ ] UI confusion reports.
- [ ] Config corruption reports.
- [ ] npm install issues.

## 11.6 Exit Criteria

Phase 7 selesai jika:

- [ ] Release artifact dipublish.
- [ ] Release notes dipublish.
- [ ] Quick start berhasil divalidasi dari clean environment.
- [ ] Known issues terdokumentasi.
- [ ] Support/troubleshooting path tersedia.

## 11.7 Phase Sign-off

| Role | Name | Status | Notes |
|---|---|---|---|
| Product Owner |  | Pending |  |
| Release Manager |  | Pending |  |
| Engineering Lead |  | Pending |  |
| QA Lead |  | Pending |  |
| Documentation Owner |  | Pending |  |

---

# Phase 8 — Post-MVP Improvements Completion Checklist

## 12.1 Phase Goal

Melanjutkan peningkatan kualitas routing, observability, packaging, desktop experience, team mode, dan intelligence setelah MVP terbukti.

## 12.2 Routing Quality Checklist

- [ ] More classifier rules added.
- [ ] Multi-label classification evaluated.
- [ ] Confidence scoring improved.
- [ ] Evaluation dataset expanded.
- [ ] Dry-run tester improved.
- [ ] User feedback on route quality captured.
- [ ] Project-specific routing designed.
- [ ] Route templates added.

## 12.3 Observability Checklist

- [ ] Dashboard metrics improved.
- [ ] Request distribution charts added.
- [ ] Route usage analytics added.
- [ ] Safe passthrough trend added.
- [ ] Downstream latency trend added.
- [ ] Exportable reports added.
- [ ] SQLite metadata storage evaluated or implemented.

## 12.4 Packaging Checklist

- [ ] Native Windows installer evaluated or implemented.
- [ ] macOS DMG evaluated or implemented.
- [ ] Linux AppImage/deb/rpm evaluated or implemented.
- [ ] Platform-specific npm packages evaluated or implemented.
- [ ] Auto-update mechanism evaluated.
- [ ] Binary signing evaluated.

## 12.5 Desktop Experience Checklist

- [ ] Wails desktop shell evaluated.
- [ ] Native settings window evaluated.
- [ ] Better tray states added.
- [ ] Notification messages added.
- [ ] Startup repair tool improved.
- [ ] First-run onboarding improved.

## 12.6 Team Mode Checklist

- [ ] Shared policy file designed.
- [ ] Workspace profiles designed.
- [ ] Team route presets designed.
- [ ] Config export/import templates added.
- [ ] Audit log designed.
- [ ] Policy lock evaluated.

## 12.7 Advanced Intelligence Checklist

- [ ] Cost-aware routing designed.
- [ ] Latency-aware routing designed.
- [ ] Historical performance analysis designed.
- [ ] Learning-assisted classifier evaluated.
- [ ] Prompt normalization evaluated.
- [ ] Opt-in evaluation samples evaluated.

## 12.8 Exit Criteria

Phase 8 item selesai jika:

- [ ] Fitur memiliki requirement yang jelas.
- [ ] Fitur tidak melanggar boundary dengan 9Router.
- [ ] Fitur memiliki acceptance criteria.
- [ ] Fitur memiliki migration/upgrade consideration jika menyentuh config.
- [ ] Fitur terdokumentasi.
- [ ] Fitur memiliki testing plan.

---

# 13. Decision Gate Checklist

## Gate 0 — Foundation Ready

Can proceed to Phase 1 if:

- [ ] Repo structure ready.
- [ ] App starts.
- [ ] Health endpoint works.
- [ ] Config defaults exist.
- [ ] Frontend setup works.
- [ ] No unresolved foundation blocker.

## Gate 1 — Proxy Ready

Can proceed to Phase 2 if:

- [ ] Non-streaming works.
- [ ] Streaming works.
- [ ] 9Router forwarder works.
- [ ] Errors are clear.
- [ ] Logs are metadata only.
- [ ] No provider-level fallback implemented.

## Gate 2 — Routing Ready

Can proceed to Phase 3 if:

- [ ] `smart-auto` works.
- [ ] Route profiles work.
- [ ] Task mapping works.
- [ ] Manual override works.
- [ ] Safe passthrough works.
- [ ] Routing logs are safe.

## Gate 3 — UI Ready

Can proceed to Phase 4 if:

- [ ] Dashboard works.
- [ ] Settings save.
- [ ] Config persists.
- [ ] Invalid config rejected.
- [ ] Runtime controls work.
- [ ] Web UI local-only default.

## Gate 4 — Tray Ready

Can proceed to Phase 5 if:

- [ ] Tray icon appears.
- [ ] Tray menu works.
- [ ] Startup works.
- [ ] Status sync works.
- [ ] Single-instance lock works.

## Gate 5 — Release Candidate Ready

Can proceed to release if:

- [ ] All P0 tests pass.
- [ ] Windows portable build works.
- [ ] npm wrapper works or documented as beta.
- [ ] Docs are ready.
- [ ] No secret leaks.
- [ ] No provider-level boundary violation.

---

# 14. Sprint Completion Checklist

## Sprint 0 — Planning and Foundation

- [ ] Repo ready.
- [ ] Go app skeleton ready.
- [ ] Vue app skeleton ready.
- [ ] Config draft ready.
- [ ] Dev scripts ready.
- [ ] Initial docs ready.
- [ ] Foundation merged.

## Sprint 1 — Core Proxy Non-Streaming

- [ ] `/v1/chat/completions` ready.
- [ ] Non-streaming forwarding ready.
- [ ] Request ID ready.
- [ ] Health endpoint ready.
- [ ] Basic logs ready.
- [ ] Non-streaming requests work.

## Sprint 2 — Streaming and Error Handling

- [ ] Streaming passthrough ready.
- [ ] Client disconnect handling ready.
- [ ] Downstream error handling ready.
- [ ] `/v1/models` ready.
- [ ] Redaction baseline ready.
- [ ] Streaming requests work.

## Sprint 3 — Analyzer and Classifier

- [ ] Analyzer ready.
- [ ] Rule-based classifier ready.
- [ ] Route resolver ready.
- [ ] Smart aliases ready.
- [ ] Safe passthrough ready.
- [ ] `smart-auto` works.

## Sprint 4 — Config and Route Profiles

- [ ] Config persistence ready.
- [ ] Route profiles ready.
- [ ] Task routes ready.
- [ ] Config validation ready.
- [ ] Last-known-good config ready.
- [ ] Routing policy configurable.

## Sprint 5 — Web UI Foundation

- [ ] Embedded UI foundation ready.
- [ ] Dashboard ready.
- [ ] Setup wizard ready.
- [ ] Admin API auth ready.
- [ ] Status API ready.
- [ ] Dashboard usable.

## Sprint 6 — Web UI Settings

- [ ] Routing settings ready.
- [ ] Route profiles UI ready.
- [ ] Downstream settings ready.
- [ ] Runtime controls ready.
- [ ] Startup settings ready.
- [ ] Privacy settings ready.
- [ ] Main product behavior configurable from Web UI.

## Sprint 7 — Tray and Startup

- [ ] Tray icon ready.
- [ ] Tray menu ready.
- [ ] Open dashboard ready.
- [ ] Start/stop/restart ready.
- [ ] Windows run at login ready.
- [ ] Status sync ready.
- [ ] Background tray app behavior ready.

## Sprint 8 — Packaging and npm

- [ ] Embedded build ready.
- [ ] Windows portable release ready.
- [ ] GitHub Release workflow ready or draft.
- [ ] npm wrapper ready.
- [ ] Checksums ready.
- [ ] Distributable MVP candidate ready.

## Sprint 9 — QA and Compatibility

- [ ] Test suite ready.
- [ ] Mock 9Router tests ready.
- [ ] Streaming compatibility tested.
- [ ] Redaction tests passed.
- [ ] AI coding assistant manual tests run.
- [ ] Documentation ready.
- [ ] Release candidate ready.

## Sprint 10 — MVP Release

- [ ] Release artifacts ready.
- [ ] Release notes ready.
- [ ] Quick start validation done.
- [ ] Known issues listed.
- [ ] Release published.
- [ ] Version `v0.1.0-alpha` available.

---

# 15. Evidence Checklist

Setiap phase completion harus menyimpan evidence berikut.

| Evidence Type | Required? | Example |
|---|---|---|
| PR/commit list | Yes | Link commit atau PR |
| Test result | Yes | Unit/integration/manual result |
| Screenshot | For UI/tray phases | Dashboard, tray menu, startup page |
| Log sample | Yes | Metadata log redacted |
| Config sample | If config changed | Example valid config |
| Build artifact | For packaging/release | Portable zip, binary, npm package |
| Known issues | Yes | Bugs accepted for next phase |
| Sign-off note | Yes | Product/Engineering/QA approval |

---

# 16. Final MVP Go/No-Go Checklist

MVP boleh dirilis hanya jika seluruh item berikut terpenuhi.

## Product Go/No-Go

- [ ] User bisa setup dari awal.
- [ ] User bisa membuka Web UI.
- [ ] User bisa mengatur 9Router endpoint.
- [ ] User bisa mengatur task-to-route mapping.
- [ ] User bisa memakai `smart-auto`.
- [ ] User bisa melihat tray icon.
- [ ] User bisa start/stop/restart dari UI/tray.
- [ ] User bisa mengaktifkan auto-start.
- [ ] User mendapat error yang jelas jika 9Router offline.

## Engineering Go/No-Go

- [ ] Proxy path stabil.
- [ ] Streaming stabil.
- [ ] Non-streaming stabil.
- [ ] Config persistence stabil.
- [ ] Runtime manager stabil.
- [ ] Tray app stabil pada Windows.
- [ ] Packaging build reproducible.
- [ ] No provider-level routing added.

## Security Go/No-Go

- [ ] No API key logs.
- [ ] No Authorization header logs.
- [ ] No full prompt logs by default.
- [ ] Web UI local-only default.
- [ ] Diagnostics redacted.
- [ ] Invalid config rejected.
- [ ] LAN access opt-in only.

## QA Go/No-Go

- [ ] P0 bug count = 0.
- [ ] Critical streaming bug count = 0.
- [ ] Critical startup bug count = 0.
- [ ] Critical secret leak bug count = 0.
- [ ] Manual Windows test passed.
- [ ] Minimum compatibility test passed.

## Documentation Go/No-Go

- [ ] README complete.
- [ ] Quick start complete.
- [ ] 9Router guide complete.
- [ ] Routing policy guide complete.
- [ ] Troubleshooting complete.
- [ ] Security/privacy docs complete.
- [ ] Release notes complete.

---

# 17. Final Notes

Checklist ini harus dipakai sebagai dokumen hidup. Setiap kali PRD, Technical Plan, atau Implementation Plan berubah, checklist ini juga harus diperbarui.

Prinsip paling penting:

> Smart AI Proxy memilih route terbaik dan menyediakan control panel untuk user. 9Router tetap mengeksekusi provider-level routing dengan andal.
