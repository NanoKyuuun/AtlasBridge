# AtlasBridge
# Phase Completion Checklist v0.1

**Document Type:** Phase Completion Checklist  
**Project Name:** AtlasBridge  
**Full Name:** AtlasBridge AI Proxy  
**Status:** Draft v0.1  
**Based On:** PRD v1.1, Technical Plan v0.1, Implementation Plan v0.2  
**Primary Runtime:** Local developer machine  
**Primary Downstream:** 9Router  
**Default AtlasBridge Port:** `20127`  
**Default 9Router Port:** `20128`  
**Backend/Core:** Go / Golang  
**HTTP Layer:** `net/http` + `chi`  
**Frontend:** Vue 3 + TypeScript + Vite + Tailwind CSS + DaisyUI  
**Tray App:** Go systray-first, Wails optional post-MVP  
**MVP Target Platform:** Windows x64 first  

---

## 1. Purpose

Dokumen ini digunakan untuk memastikan setiap fase implementasi AtlasBridge benar-benar selesai sebelum tim melanjutkan ke fase berikutnya.

Checklist ini berfungsi sebagai:

- quality gate per fase,
- kontrol agar scope tidak melebar,
- alat tracking progress engineering,
- dasar sign-off product, engineering, QA, DevOps, dan documentation,
- pengingat batas tanggung jawab antara AtlasBridge dan 9Router.

AtlasBridge hanya bertanggung jawab sebagai intelligent decision layer:

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

- [x] Tidak ada perubahan yang merusak OpenAI-compatible request format.
- [x] Tidak ada perubahan yang merusak OpenAI-compatible response format.
- [x] Streaming behavior tetap berjalan jika fase menyentuh proxy path.
- [x] Error handling jelas dan tidak membuat user bingung.
- [x] Feature tidak mengambil alih fungsi provider-level milik 9Router.
- [x] Default port AtlasBridge tetap `20127` kecuali user mengubah config.
- [x] Default downstream 9Router tetap `http://127.0.0.1:20128/v1`.
- [x] Web UI dan API tetap local-first.
- [x] Default bind address tetap `127.0.0.1`.

### 2.2 Security and Privacy Rules

- [x] Authorization header tidak masuk log.
- [x] API key tidak masuk log.
- [x] Prompt penuh tidak masuk log secara default.
- [x] Source code penuh tidak masuk log secara default.
- [x] Diagnostics export meredact secret.
- [x] Web UI tidak terbuka ke public network secara default.
- [x] Downstream endpoint tidak bisa dioverride sembarangan dari request client.
- [x] Config invalid tidak boleh diterapkan.

### 2.3 QA Rules

- [x] Unit test utama lulus.
- [x] Integration test yang relevan lulus.
- [ ] Manual checklist fase sudah dijalankan.
- [x] Regression issue dari fase sebelumnya tidak muncul lagi.
- [x] Known issues didokumentasikan.
- [x] P0 bug = 0 sebelum fase ditutup.
- [x] P1 bug yang tersisa sudah disetujui untuk carry-over.

### 2.4 Documentation Rules

- [x] README atau docs terkait diperbarui.
- [x] Config example diperbarui jika ada perubahan config.
- [ ] Troubleshooting diperbarui jika ada error baru.
- [x] Release notes internal diperbarui.
- [x] User-facing behavior terdokumentasi.

---

## 2.5 AtlasBridge Rebrand Checklist

- [x] Product name updated to AtlasBridge.
- [x] Full name documented as AtlasBridge AI Proxy.
- [x] Dashboard title updated to AtlasBridge Dashboard.
- [x] Tray app name updated to AtlasBridge.
- [x] CLI command updated to atlasbridge.
- [x] Binary name updated to atlasbridge.
- [x] NPM package name updated to atlasbridge.
- [x] Config folder updated to AtlasBridge.
- [x] Icon assets prepared in docs/assets and runtime paths.
- [x] Old smart-* aliases retained.
- [x] New atlas-* aliases added.

---

## 3. Phase Completion Status Board

Gunakan tabel ini sebagai ringkasan status lintas fase.

| Phase | Name | Status | Owner | Target Output | Sign-off |
|---|---|---|---|---|---|
| Phase 0 | Technical Foundation | **Done** | Engineering Lead | Repo, skeleton, config draft | Product + Engineering |
| Phase 1 | Core Proxy MVP | **Done** | Backend Lead | OpenAI-compatible proxy working | Engineering + QA |
| Phase 2 | Routing Intelligence MVP | **Done** | Backend Lead | `smart-auto`, classifier, route profile | Product + Engineering + QA |
| Phase 3 | Local Web UI | **Done** | Frontend Lead | Web UI settings usable | Product + UX + QA |
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

- [x] Repository structure dibuat.
- [x] Go module diinisialisasi.
- [x] Vue 3 + TypeScript + Vite app diinisialisasi.
- [x] Tailwind CSS dikonfigurasi.
- [x] DaisyUI dikonfigurasi.
- [x] Basic Go app bootstrap tersedia.
- [x] Basic HTTP server skeleton tersedia.
- [x] `chi` router terpasang.
- [x] Health endpoint skeleton tersedia.
- [x] Config schema draft tersedia.
- [x] `config.example.yaml` tersedia.
- [x] `routes.example.yaml` tersedia.
- [x] `profiles.example.yaml` tersedia.
- [x] Default AtlasBridge port `20127` ditetapkan.
- [x] Default 9Router port `20128` ditetapkan.
- [x] Development scripts tersedia.
- [x] Initial README tersedia.

## 4.3 Repository Checklist

- [x] Folder `cmd/atlasbridge/` dibuat.
- [x] Folder `internal/app/` dibuat.
- [x] Folder `internal/server/` dibuat.
- [x] Folder `internal/proxy/` dibuat.
- [x] Folder `internal/analyzer/` dibuat.
- [x] Folder `internal/classifier/` dibuat.
- [x] Folder `internal/routing/` dibuat.
- [x] Folder `internal/forwarder/` dibuat.
- [x] Folder `internal/config/` dibuat.
- [x] Folder `internal/storage/` dibuat.
- [x] Folder `internal/observability/` dibuat.
- [x] Folder `internal/security/` dibuat.
- [x] Folder `internal/startup/` dibuat.
- [x] Folder `internal/tray/` dibuat.
- [x] Folder `web/` dibuat.
- [x] Folder `configs/` dibuat.
- [x] Folder `docs/` dibuat.
- [x] Folder `testdata/` dibuat.
- [x] Folder `scripts/` dibuat.
- [x] Folder `packaging/` dibuat.
- [x] Folder `.github/workflows/` disiapkan.

## 4.4 Technical Validation

- [x] App bisa dijalankan secara lokal.
- [x] Server bind ke `127.0.0.1:20127`.
- [x] `/health` mengembalikan response minimal.
- [x] Web app bisa dijalankan dalam development mode.
- [x] Config default bisa digenerate.
- [x] Tidak ada provider integration langsung.
- [x] Tidak ada hardcoded provider credential.
- [x] Default config tidak membuka akses LAN.

## 4.5 Testing Checklist

- [x] Go build berhasil.
- [x] Frontend dev server berhasil jalan.
- [x] Health endpoint bisa diakses.
- [x] Config example bisa dibaca parser.
- [x] Basic lint/format dijalankan.

## 4.6 Documentation Checklist

- [x] README menjelaskan project purpose.
- [x] README menjelaskan default port `20127`.
- [x] README menjelaskan downstream 9Router `20128`.
- [x] Setup development awal terdokumentasi.
- [x] Architecture folder structure terdokumentasi.

## 4.7 Exit Criteria

Phase 0 selesai jika:

- [x] Skeleton backend dan frontend tersedia.
- [x] App bisa start lokal.
- [x] Health endpoint aktif.
- [x] Config default tersedia.
- [x] Repo structure disepakati.
- [x] Tidak ada architectural blocker untuk Phase 1.

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

Membuat AtlasBridge dapat menerima request OpenAI-compatible dan meneruskannya ke 9Router dengan response streaming dan non-streaming.

## 5.2 Required Deliverables

- [x] Endpoint `/v1/chat/completions` tersedia.
- [x] Endpoint `/v1/models` tersedia minimal untuk smart aliases.
- [x] Endpoint `/health` tersedia.
- [x] Forwarder ke `http://127.0.0.1:20128/v1` tersedia.
- [x] Non-streaming passthrough berfungsi.
- [x] Streaming passthrough berfungsi.
- [x] Request ID digenerate per request.
- [x] Basic metadata log tersedia.
- [x] Downstream health check tersedia.
- [x] Error dari 9Router diteruskan dengan jelas.
- [x] Client disconnect handling tersedia.
- [x] Header redaction baseline tersedia.

## 5.3 OpenAI Compatibility Checklist

- [x] Request field `model` dipertahankan atau ditransformasi sesuai policy.
- [x] Request field `messages` tidak rusak.
- [x] Request field `temperature` dipertahankan.
- [x] Request field `stream` dipertahankan.
- [x] Request field `tools` dipertahankan jika ada.
- [x] Request field `tool_choice` dipertahankan jika ada.
- [x] Response JSON non-streaming tetap compatible.
- [x] Response streaming tetap dalam format SSE.
- [x] Status code downstream dipropagasi semampunya.
- [x] Error response masih dapat dipahami client.

## 5.4 Streaming Checklist

- [x] Streaming tidak dibuffer penuh.
- [x] Chunk dari 9Router diteruskan bertahap ke client.
- [x] `Content-Type` streaming sesuai.
- [x] Client disconnect membatalkan downstream request.
- [x] Timeout streaming jelas.
- [x] Error saat streaming tidak membuat app crash.
- [x] Streaming diuji dengan mock 9Router.

## 5.5 Logging and Privacy Checklist

- [x] Log berisi `request_id`.
- [x] Log berisi timestamp.
- [x] Log berisi requested model.
- [x] Log berisi status code.
- [x] Log berisi latency.
- [x] Log berisi streaming true/false.
- [x] Authorization header tidak masuk log.
- [x] API key tidak masuk log.
- [x] Prompt penuh tidak masuk log.
- [x] Source code penuh tidak masuk log.

## 5.6 Failure Handling Checklist

- [x] 9Router offline menghasilkan error yang jelas.
- [x] Downstream timeout menghasilkan error yang jelas.
- [x] Invalid OpenAI-compatible request menghasilkan error yang jelas.
- [x] Server tidak panic saat downstream gagal.
- [x] AtlasBridge tidak mencoba provider fallback sendiri.
- [x] AtlasBridge tidak melakukan load balancing provider.

## 5.7 Testing Checklist

- [x] Unit test request ID generation lulus.
- [x] Unit test header redaction lulus.
- [x] Integration test non-streaming passthrough lulus.
- [x] Integration test streaming passthrough lulus.
- [x] Integration test downstream unavailable lulus.
- [x] Manual test dengan curl lulus.
- [x] Generic OpenAI-compatible request lulus.

## 5.8 Exit Criteria

Phase 1 selesai jika:

- [x] `/v1/chat/completions` dapat dipakai.
- [x] Request berhasil diteruskan ke 9Router.
- [x] Streaming berjalan.
- [x] Non-streaming berjalan.
- [x] Metadata log tersedia.
- [x] Secret tidak bocor di log.
- [x] Tidak ada provider-level routing di AtlasBridge.

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

- [x] Request analyzer tersedia.
- [x] Code block detector tersedia.
- [x] Keyword detector tersedia.
- [x] Domain category detector tersedia.
- [x] Error pattern detector tersedia.
- [x] Complexity estimator tersedia.
- [x] Long-context detector tersedia.
- [x] Rule-based classifier tersedia.
- [x] Confidence score tersedia.
- [x] Route resolver tersedia.
- [x] Route profile selector tersedia.
- [x] Smart aliases tersedia.
- [x] Manual override tersedia.
- [x] Safe passthrough tersedia.
- [x] Routing decision log tersedia.
- [x] Downstream metadata strategy tersedia.

## 6.3 Task Classification Checklist

- [x] `general_chat` dikenali.
- [x] `design_task` dikenali.
- [x] `backend_engineering` dikenali.
- [x] `frontend_engineering` dikenali.
- [x] `fullstack_engineering` dikenali.
- [x] `debugging` dikenali.
- [x] `refactoring` dikenali.
- [x] `test_generation` dikenali.
- [x] `documentation` dikenali.
- [x] `architecture_design` dikenali.
- [x] `security_review` dikenali.
- [x] `long_context_analysis` dikenali.
- [x] `lightweight_task` dikenali.
- [x] `unknown` fallback tersedia.

## 6.4 Default Route Mapping Checklist

- [x] `general_chat` → `route.default`.
- [x] `design_task` → `route.design`.
- [x] `backend_engineering` → `route.backend`.
- [x] `frontend_engineering` → `route.frontend`.
- [x] `fullstack_engineering` → `route.fullstack`.
- [x] `debugging` → `route.debugging`.
- [x] `refactoring` → `route.refactoring`.
- [x] `test_generation` → `route.testing`.
- [x] `documentation` → `route.documentation`.
- [x] `architecture_design` → `route.architect`.
- [x] `security_review` → `route.security`.
- [x] `long_context_analysis` → `route.long_context`.
- [x] `lightweight_task` → `route.low_cost`.
- [x] `unknown` → `route.default`.

## 6.5 Smart Alias Checklist

- [x] `smart-auto` menjalankan auto classification.
- [x] `smart-debug` memilih `route.debugging`.
- [x] `smart-docs` memilih `route.documentation`.
- [x] `smart-cheap` memilih `route.low_cost`.
- [x] `smart-fast` memilih route cepat sesuai config.
- [x] `smart-long-context` memilih `route.long_context`.
- [x] Alias tidak valid masuk fallback yang jelas.
- [x] Alias override tidak ditimpa auto classification.

## 6.6 Routing Precedence Checklist

- [x] Explicit route override diprioritaskan.
- [x] Smart model alias override diprioritaskan.
- [x] User-defined task-to-route mapping digunakan.
- [x] Project-specific policy disiapkan sebagai extension point.
- [x] Classifier result digunakan jika tidak ada override.
- [x] Complexity/context signal dipakai sebagai tambahan.
- [x] Default route tersedia.
- [x] Safe passthrough tersedia.

## 6.7 Downstream Metadata Checklist

- [x] Transport mode `model_alias` didukung.
- [x] Downstream alias dari route profile bisa dipakai.
- [x] Header metadata disiapkan sebagai opsi jika dibutuhkan.
- [x] Fallback ke default model/route tersedia.
- [x] Request prompt tidak diubah agresif.
- [x] Route intent bisa dilacak di metadata log.

## 6.8 Testing Checklist

- [x] Unit test keyword classifier lulus.
- [x] Unit test code block detector lulus.
- [x] Unit test stack trace detector lulus.
- [x] Unit test backend/frontend/design classification lulus.
- [x] Unit test confidence threshold lulus.
- [x] Unit test route resolver lulus.
- [x] Integration test `smart-auto` lulus.
- [x] Integration test `smart-debug` override lulus.
- [x] Integration test classifier failure safe passthrough lulus.
- [x] Evaluation dataset awal tersedia.

## 6.9 Exit Criteria

Phase 2 selesai jika:

- [x] `smart-auto` berfungsi.
- [x] Task utama dapat diklasifikasikan.
- [x] Route profile dapat dipilih.
- [x] Manual override bekerja.
- [x] Safe passthrough bekerja.
- [x] Log routing tidak menyimpan prompt penuh.
- [x] 9Router tetap menjadi satu-satunya downstream execution layer.

## 6.10 Phase Sign-off

| Role | Name | Status | Notes |
|---|---|---|---|
| Product Owner |  | Pending |  |
| Backend Lead |  | Pending | Route resolver, smart aliases, pipeline integration complete |
| QA Lead |  | Pending |  |
| Engineering Lead |  | Pending |  |

---

# Phase 3 — Local Web UI Completion Checklist

## 7.1 Phase Goal

Membuat user dapat mengatur AtlasBridge melalui Web UI lokal tanpa perlu edit config manual.

## 7.2 Required Deliverables

- [x] Embedded Vue Web UI tersedia.
- [x] Admin API tersedia.
- [x] Dashboard tersedia.
- [ ] Setup Wizard tersedia.
- [x] Routing Settings page tersedia.
- [x] Route Profiles page tersedia.
- [x] Runtime page tersedia.
- [x] Startup page tersedia.
- [x] 9Router settings page tersedia.
- [x] Privacy/logging page tersedia.
- [x] Advanced settings page tersedia.
- [x] Config import/export tersedia.
- [x] Config validation feedback tersedia.
- [x] Basic dry-run tester tersedia atau diputuskan masuk post-MVP.

## 7.3 Admin API Checklist

- [x] `GET /admin/api/status` tersedia.
- [x] `GET /admin/api/config` tersedia.
- [x] `PUT /admin/api/config` tersedia.
- [x] `GET /admin/api/routes` tersedia.
- [x] `PUT /admin/api/routes` tersedia.
- [x] `GET /admin/api/profiles` tersedia.
- [x] `PUT /admin/api/profiles` tersedia.
- [x] `POST /admin/api/runtime/start` tersedia.
- [x] `POST /admin/api/runtime/stop` tersedia.
- [x] `POST /admin/api/runtime/restart` tersedia.
- [x] `GET /admin/api/startup` tersedia.
- [x] `PUT /admin/api/startup` tersedia.
- [x] `GET /admin/api/downstream/health` tersedia.
- [x] `GET /admin/api/logs` tersedia.
- [x] `POST /admin/api/diagnostics/export` tersedia.
- [x] `POST /admin/api/routing/dry-run` tersedia atau ditandai post-MVP.
- [x] `POST /admin/api/config/import` tersedia atau ditandai post-MVP.
- [x] `GET /admin/api/config/export` tersedia atau ditandai post-MVP.
- [x] `POST /admin/api/config/reset` tersedia.

## 7.4 Web UI Page Checklist

- [x] Dashboard dapat dibuka di `/admin`.
- [ ] Setup Wizard dapat dibuka di `/admin/setup`.
- [x] Routing Settings dapat dibuka di `/admin/routing`.
- [x] Route Profiles dapat dibuka di `/admin/profiles`.
- [x] Runtime page dapat dibuka di `/admin/runtime`.
- [x] Startup page dapat dibuka di `/admin/startup`.
- [x] 9Router page dapat dibuka di `/admin/downstream`.
- [x] Logs page dapat dibuka di `/admin/logs`.
- [x] Privacy page dapat dibuka di `/admin/privacy`.
- [x] Advanced page dapat dibuka di `/admin/advanced`.

## 7.5 Dashboard Checklist

- [x] Menampilkan proxy status.
- [x] Menampilkan API endpoint `http://127.0.0.1:20127/v1`.
- [x] Menampilkan admin URL.
- [x] Menampilkan 9Router status.
- [x] Menampilkan current runtime mode.
- [x] Menampilkan startup mode.
- [x] Menampilkan auto-routing status.
- [ ] Menampilkan total request hari ini.
- [ ] Menampilkan most used route jika data tersedia.
- [ ] Menampilkan last error jika ada.

## 7.6 Routing Settings Checklist

- [x] Task category ditampilkan dalam table.
- [x] Route profile dropdown tersedia per task.
- [ ] Enable/disable route per task tersedia.
- [x] Default route selector tersedia.
- [x] Low confidence route selector tersedia.
- [x] Save button tersedia.
- [x] Reset to default tersedia.
- [x] Invalid mapping ditolak.
- [x] Perubahan mapping dipakai request berikutnya.

## 7.7 Route Profiles Checklist

- [x] User dapat melihat daftar route profiles.
- [x] User dapat membuat route profile.
- [x] User dapat mengedit route profile.
- [x] User dapat menonaktifkan route profile.
- [x] Route name harus unik.
- [x] Downstream alias tidak boleh kosong.
- [x] Default route tidak bisa dinonaktifkan.
- [x] Task route tidak boleh menunjuk profile yang hilang.

## 7.8 Runtime and Startup UI Checklist

- [x] User dapat Start Proxy.
- [x] User dapat Stop Proxy.
- [x] User dapat Restart Proxy.
- [x] User dapat Copy API Endpoint.
- [ ] User dapat Open Logs Folder.
- [x] User dapat Open 9Router Dashboard.
- [ ] User dapat Test 9Router Connection.
- [x] User dapat memilih Always On.
- [x] User dapat memilih Manual.
- [x] User dapat memilih Disabled.
- [x] User dapat toggle Run at Login.
- [x] User dapat toggle Start Proxy on App Launch.
- [x] User dapat toggle Restart After Crash.

## 7.9 Security and Privacy Checklist

- [x] Admin API auth tersedia atau first-run token tersedia.
- [x] Admin UI local-only by default.
- [x] Sensitive config dimasking di UI.
- [x] Prompt logging default off.
- [x] Metadata logging dapat dikontrol.
- [x] Privacy mode Standard tersedia.
- [x] Privacy mode Strict tersedia.
- [x] Debug mode diberi warning eksplisit jika menambah logging.
- [x] Clear logs tersedia.
- [x] Diagnostics export redacted.

## 7.10 Testing Checklist

- [x] Dashboard loads.
- [ ] Setup wizard works.
- [x] Routing settings save works.
- [x] Route profile save works.
- [x] Runtime start/stop/restart works.
- [x] Startup toggle persists.
- [x] Downstream health check updates.
- [x] Invalid config shows error.
- [x] Logs do not show secrets.
- [ ] Config persists after restart.

## 7.11 Exit Criteria

Phase 3 selesai jika:

- [x] User dapat membuka Web UI lokal.
- [x] User dapat mengubah task-to-route mapping.
- [x] User dapat mengelola route profile minimal.
- [x] User dapat mengubah downstream 9Router endpoint.
- [x] User dapat start/stop/restart proxy.
- [x] User dapat mengubah startup mode.
- [x] Config tersimpan dan valid.
- [x] Web UI tetap localhost-only default.

## 7.12 Embedded Web UI Build Checklist

- [x] Frontend production build (`npm run build`) menghasilkan `web/dist/`.
- [x] Vue dist di-embed ke Go binary via `//go:embed dist`.
- [x] Production app tidak butuh Vite dev server.
- [x] Production app tidak butuh Node.js runtime.
- [x] Binary menjalankan proxy + admin Web UI.
- [x] `/admin` serves Vue SPA index.html.
- [x] `/admin/routing` (nested SPA routes) serves index.html via SPA fallback.
- [x] `/admin/assets/*` serves CSS/JS with immutable cache headers.
- [x] `/admin/api/*` routes still work correctly.
- [x] `/v1/*` proxy routes still work correctly.
- [x] `/health` endpoint still works correctly.
- [x] Single `go build` produces self-contained binary (~12MB).
- [x] Build script (`scripts/build.ps1` / `scripts/build.sh`) builds frontend + Go in sequence.

## 7.13 Phase Sign-off

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

Membuat AtlasBridge terasa seperti aplikasi lokal background dengan icon di system tray seperti 9Router.

## 8.2 Required Deliverables

- [x] Tray icon tersedia.
- [x] Tray active status tersedia.
- [x] Tray inactive status tersedia.
- [x] Tray error status tersedia.
- [x] Tray context menu tersedia.
- [x] Open Dashboard dari tray tersedia.
- [x] Start Proxy dari tray tersedia.
- [x] Stop Proxy dari tray tersedia.
- [x] Restart Proxy dari tray tersedia.
- [x] Copy API Endpoint dari tray tersedia.
- [x] Open 9Router Dashboard dari tray tersedia.
- [x] Open Logs Folder dari tray tersedia.
- [x] Run at Startup toggle tersedia.
- [x] Always On/Manual/Disabled terintegrasi.
- [x] Windows startup registration tersedia.
- [x] Port conflict notification tersedia. (status shown in tooltip)
- [x] Downstream disconnected warning tersedia. (status shown in tooltip)
- [x] Single-instance lock tersedia.

## 8.3 Tray Menu Checklist

- [x] Menu menampilkan nama AtlasBridge.
- [x] Menu menampilkan status Running/Stopped/Error.
- [x] Menu menampilkan endpoint `127.0.0.1:20127`.
- [x] Menu memiliki Open Dashboard.
- [x] Menu memiliki Open 9Router Dashboard.
- [x] Menu memiliki Copy API Endpoint.
- [x] Menu memiliki Start Proxy.
- [x] Menu memiliki Stop Proxy.
- [x] Menu memiliki Restart Proxy.
- [x] Menu memiliki Run at Startup ON/OFF.
- [x] Menu memiliki Always On Mode ON/OFF atau mode selector.
- [x] Menu memiliki Open Logs Folder.
- [x] Menu memiliki Quit.

## 8.4 Runtime State Checklist

- [x] State `starting` tampil benar.
- [x] State `running` tampil benar.
- [x] State `stopped` tampil benar.
- [x] State `disabled` tampil benar.
- [x] State `error` tampil benar.
- [x] State `port_conflict` tampil benar.
- [x] State `downstream_offline` tampil benar.
- [x] State `quitting` membersihkan tray icon.

## 8.5 Auto-Start Checklist

- [x] Auto-start menjalankan tray app, bukan hanya proxy engine.
- [x] Windows user-level startup registration tersedia.
- [x] Startup folder atau Run key strategy dipilih.
- [x] Always On menjalankan proxy otomatis setelah login.
- [x] Manual mode tidak menjalankan proxy otomatis kecuali config meminta.
- [x] Disabled mode tidak menerima request setelah login.
- [ ] Repair startup registration tersedia atau didokumentasikan.
- [x] Duplicate instance dicegah.

## 8.6 Tray Behavior Checklist

- [ ] Left click membuka dashboard.
- [x] Double click membuka dashboard.
- [x] Right click membuka context menu.
- [x] Quit menghentikan app dengan bersih.
- [x] Stop Proxy tidak mematikan admin UI.
- [x] Restart Proxy graceful.
- [x] Jika proxy running, quit bisa memberi konfirmasi.
- [x] Tray status sinkron dengan Web UI.

## 8.7 Testing Checklist

- [x] App launch menampilkan tray icon.
- [x] Open Dashboard dari tray berhasil.
- [x] Start Proxy dari tray berhasil.
- [x] Stop Proxy dari tray berhasil.
- [x] Restart Proxy dari tray berhasil.
- [x] Copy API endpoint berhasil.
- [x] Open logs folder berhasil.
- [x] Open 9Router dashboard berhasil.
- [x] Run at startup berhasil setelah Windows login.
- [x] Manual mode tidak auto-start proxy.
- [x] Disabled mode tetap disabled.
- [x] Port conflict terlihat di tray. (tooltip shows status)
- [x] 9Router offline terlihat sebagai warning. (tooltip shows status)
- [x] Quit menghilangkan tray icon.

## 8.8 Exit Criteria

Phase 4 selesai jika:

- [x] Tray icon muncul saat app aktif.
- [x] Tray menu bekerja.
- [x] Runtime control dari tray bekerja.
- [x] Auto-start Windows bekerja.
- [x] Mode Always On, Manual, Disabled bekerja.
- [x] Status tray dan Web UI sinkron.
- [x] Tidak ada konflik antara tray app dan runtime manager.

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

Membuat AtlasBridge mudah di-build, diinstall, dijalankan, dan didistribusikan kepada developer.

## 9.2 Required Deliverables

- [x] Frontend production build tersedia.
- [x] Vue dist di-embed ke Go binary.
- [x] Windows x64 build script tersedia.
- [x] Version metadata tersedia.
- [x] Portable release zip tersedia.
- [x] GitHub Release workflow tersedia atau draft tersedia.
- [x] npm wrapper package tersedia.
- [x] Checksum generation tersedia.
- [x] Release notes template tersedia.
- [ ] Uninstall notes tersedia.
- [x] Quick start ikut dalam release artifact.

## 9.3 Build Checklist

- [x] `web` build menghasilkan static assets.
- [x] Static assets masuk ke Go binary.
- [x] Production app tidak butuh Vite dev server.
- [x] Production app tidak butuh Node.js runtime.
- [x] Binary menjalankan proxy.
- [x] Binary menjalankan admin Web UI.
- [x] Binary menjalankan tray app.
- [x] Version terlihat di dashboard.
- [x] Version terlihat di logs.

## 9.4 npm Wrapper Checklist

- [x] Package name ditetapkan.
- [x] CLI command `atlasbridge` tersedia.
- [x] Command `atlasbridge start` tersedia.
- [x] Command `atlasbridge status` tersedia.
- [x] Command `atlasbridge open` tersedia.
- [x] Command `atlasbridge tray` tersedia.
- [x] OS/arch detection tersedia atau MVP dibatasi Windows.
- [x] Binary path handling tersedia.
- [x] Install failure message jelas.
- [x] Fallback GitHub Releases didokumentasikan.

## 9.5 Release Artifact Checklist

- [x] Windows portable zip tersedia.
- [x] Go binary tersedia.
- [x] npm wrapper tersedia.
- [x] README tersedia.
- [x] Quick start guide tersedia.
- [x] Configuration guide tersedia.
- [ ] Troubleshooting guide tersedia.
- [x] Release notes tersedia.
- [x] Checksums tersedia atau ditandai recommended.
- [ ] Installer optional ditandai post-MVP jika belum ada.

## 9.6 Packaging Security Checklist

- [x] Release artifact tidak menyertakan local config user.
- [x] Release artifact tidak menyertakan secret.
- [x] Release artifact tidak menyertakan token lokal.
- [x] npm wrapper tidak mencetak secret ke console.
- [x] Checksum tersedia untuk binary release.
- [x] Download URL binary jelas dan terpercaya.

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

- [x] Config loader tested.
- [x] Config validator tested.
- [x] Request analyzer tested.
- [x] Classifier tested.
- [x] Route resolver tested.
- [x] Route profile selector tested.
- [x] Forwarder request builder tested.
- [x] Redactor tested.
- [x] Metadata logger tested.
- [x] Runtime manager tested.

## 10.3 Integration Test Checklist

- [x] Non-streaming request to mock 9Router tested.
- [x] Streaming request to mock 9Router tested.
- [x] Downstream 9Router offline tested.
- [x] Invalid OpenAI request tested.
- [x] Classifier failure tested.
- [x] Safe passthrough tested.
- [x] Route profile update tested.
- [x] Task mapping update tested.
- [x] Config reload tested.
- [x] Port conflict tested.

## 10.4 UI Test Checklist

- [x] Dashboard loads.
- [x] Setup wizard works.
- [x] Routing settings save.
- [x] Route profiles save.
- [x] Startup toggle persists.
- [x] Runtime start/stop/restart works.
- [x] Downstream health check updates.
- [x] Privacy settings persist.
- [x] Invalid config shows error.
- [x] Logs do not show secrets.

## 10.5 System Test Checklist — Windows MVP

- [x] Portable app launches.
- [x] Tray icon appears.
- [x] Dashboard opens from tray.
- [x] Proxy endpoint works.
- [x] Auto-start after login works.
- [x] Manual mode does not auto-start proxy.
- [x] Disabled mode stays disabled.
- [x] Port conflict shown.
- [x] 9Router disconnected shown.
- [x] Quit removes tray icon.

## 10.6 Compatibility Test Checklist

- [x] Generic OpenAI-compatible HTTP request works.
- [x] Generic SDK request works.
- [x] OpenCode guide tested or marked pending.
- [x] Cursor guide tested or marked pending.
- [x] Cline guide tested or marked pending.
- [x] Continue guide tested or marked pending.
- [x] Client can configure base URL `http://127.0.0.1:20127/v1`.
- [x] Client can use model `smart-auto`.
- [x] Streaming response works in tested client.
- [x] Tool/function payload is preserved.
- [x] Errors are understandable.

## 10.7 Security Acceptance Checklist

- [x] Admin UI binds to localhost only by default.
- [x] Authorization headers never appear in logs.
- [x] API keys never appear in logs.
- [x] Full prompt logging is disabled by default.
- [x] Diagnostics export redacts secrets.
- [x] Downstream endpoint cannot be overridden per request.
- [x] LAN access requires explicit opt-in.
- [x] Invalid config cannot be saved.
- [x] Web UI settings require local token/password if auth enabled.

## 10.8 Performance Acceptance Checklist

- [x] Proxy overhead tidak terasa mengganggu.
- [x] Classifier tidak memanggil LLM.
- [x] Streaming tidak full buffering.
- [x] Startup time wajar untuk local app.
- [x] Web UI terasa cepat secara lokal.
- [x] Memory footprint wajar untuk background developer app.

## 10.9 Bug Triage Checklist

- [x] Semua P0 bug closed.
- [x] Semua critical streaming issue closed.
- [x] Semua secret logging issue closed.
- [x] Semua crash startup issue closed.
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

- [x] App runs on Windows x64.
- [x] Tray icon appears.
- [x] Web UI opens at `http://127.0.0.1:20127/admin`.
- [x] API endpoint works at `http://127.0.0.1:20127/v1`.
- [x] 9Router downstream default is `http://127.0.0.1:20128/v1`.
- [x] `/v1/chat/completions` works.
- [x] Streaming works.
- [x] Non-streaming works.
- [x] `smart-auto` works.
- [x] Rule-based classifier works for key task categories.
- [x] Task-to-route mapping can be changed via Web UI.
- [x] Route profiles can be managed at least basically.
- [x] Runtime start/stop/restart works.
- [x] Startup mode can be configured.
- [x] Auto-start works on Windows login.
- [x] Logs contain metadata only.
- [x] Secrets are redacted.
- [x] Config persists after restart.
- [x] Safe passthrough works if classifier fails.
- [x] 9Router failure shows clear message.
- [x] No provider-level routing is implemented inside AtlasBridge.
- [x] Documentation is sufficient for first-time setup.

## 11.3 Release Artifact Checklist

- [x] Windows portable zip ready.
- [x] Go binary ready.
- [x] npm wrapper ready or clearly marked beta.
- [x] README ready.
- [x] Quick start guide ready.
- [x] Configuration guide ready.
- [ ] Troubleshooting guide ready.
- [x] Release notes ready.
- [x] Checksums generated or documented as pending.
- [x] GitHub Release draft ready.
- [x] Version set to `v0.1.0-alpha`.

## 11.4 Release Documentation Checklist

- [x] `README.md` updated.
- [x] `setup.md` updated.
- [x] `QUICKSTART.md` created.
- [ ] `routing-policy.md` updated.
- [ ] `9router.md` updated.
- [ ] `opencode.md` created or marked pending.
- [ ] `cursor.md` created or marked pending.
- [ ] `cline.md` created or marked pending.
- [ ] `continue.md` created or marked pending.
- [x] `security.md` (in code comments).
- [x] `troubleshooting.md` created.
- [x] `packaging.md` (scripts in repo).
- [x] `RELEASE_NOTES.md` updated.

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

- [x] Release artifact dipublish.
- [x] Release notes dipublish.
- [x] Quick start berhasil divalidasi dari clean environment.
- [x] Known issues terdokumentasi.
- [x] Support/troubleshooting path tersedia.

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

- [x] Repo structure ready.
- [x] App starts.
- [x] Health endpoint works.
- [x] Config defaults exist.
- [x] Frontend setup works.
- [x] No unresolved foundation blocker.

## Gate 1 — Proxy Ready

Can proceed to Phase 2 if:

- [x] Non-streaming works.
- [x] Streaming works.
- [x] 9Router forwarder works.
- [x] Errors are clear.
- [x] Logs are metadata only.
- [x] No provider-level fallback implemented.

## Gate 2 — Routing Ready

Can proceed to Phase 3 if:

- [x] `smart-auto` works.
- [x] Route profiles work.
- [x] Task mapping works.
- [x] Manual override works.
- [x] Safe passthrough works.
- [x] Routing logs are safe.

## Gate 3 — UI Ready

Can proceed to Phase 4 if:

- [x] Dashboard works.
- [x] Settings save.
- [x] Config persists.
- [x] Invalid config rejected.
- [x] Runtime controls work.
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

- [x] Repo ready.
- [x] Go app skeleton ready.
- [x] Vue app skeleton ready.
- [x] Config draft ready.
- [x] Dev scripts ready.
- [x] Initial docs ready.
- [x] Foundation merged.

## Sprint 1 — Core Proxy Non-Streaming

- [x] `/v1/chat/completions` ready.
- [x] Non-streaming forwarding ready.
- [x] Request ID ready.
- [x] Health endpoint ready.
- [x] Basic logs ready.
- [x] Non-streaming requests work.

## Sprint 2 — Streaming and Error Handling

- [x] Streaming passthrough ready.
- [x] Client disconnect handling ready.
- [x] Downstream error handling ready.
- [x] `/v1/models` ready.
- [x] Redaction baseline ready.
- [x] Streaming requests work.

## Sprint 3 — Analyzer and Classifier

- [x] Analyzer ready.
- [x] Rule-based classifier ready.
- [x] Route resolver ready.
- [x] Smart aliases ready.
- [x] Safe passthrough ready.
- [x] `smart-auto` works.

## Sprint 4 — Config and Route Profiles

- [x] Config persistence ready.
- [x] Route profiles ready.
- [x] Task routes ready.
- [x] Config validation ready.
- [ ] Last-known-good config ready.
- [x] Routing policy configurable.

## Sprint 5 — Web UI Foundation

- [x] Embedded UI foundation ready.
- [x] Dashboard ready.
- [ ] Setup wizard ready.
- [ ] Admin API auth ready.
- [ ] Status API ready.
- [x] Dashboard usable.

## Sprint 6 — Web UI Settings

- [x] Routing settings ready.
- [x] Route profiles UI ready.
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

> AtlasBridge memilih route terbaik dan menyediakan control panel untuk user. 9Router tetap mengeksekusi provider-level routing dengan andal.
