# Smart AI Proxy
# Implementation Plan v0.2

**Document Type:** Implementation Plan  
**Project Name:** Smart AI Proxy  
**Status:** Draft v0.2  
**Based On:** PRD v1.1 and Technical Plan v0.1  
**Primary Runtime:** Local developer machine  
**Primary Downstream:** 9Router  
**Default Smart AI Proxy Port:** `20127`  
**Default 9Router Port:** `20128`  
**Backend/Core:** Go / Golang  
**HTTP Layer:** `net/http` + `chi`  
**Frontend:** Vue 3 + TypeScript + Vite + Tailwind CSS + DaisyUI  
**Tray App:** Go systray-first, Wails optional post-MVP  
**Storage MVP:** Local YAML/JSON + JSONL metadata logs  
**Storage Post-MVP:** SQLite optional  
**Distribution:** Portable binary, GitHub Releases, npm wrapper, optional native installer  

---

## 1. Executive Summary

Implementation Plan ini menjelaskan langkah implementasi Smart AI Proxy secara terstruktur dari foundation sampai rilis MVP dan post-MVP.

Smart AI Proxy akan dibangun sebagai aplikasi lokal untuk developer yang berjalan di background. Aplikasi ini menyediakan endpoint OpenAI-compatible di `http://127.0.0.1:20127/v1`, menyediakan Web UI lokal di `http://127.0.0.1:20127/admin`, memiliki tray icon seperti 9Router saat aktif, mampu auto-start saat laptop/komputer restart, dan meneruskan request ke 9Router default di `http://127.0.0.1:20128/v1`.

Smart AI Proxy hanya bertanggung jawab sebagai decision layer:

- menerima request OpenAI-compatible,
- menganalisis request,
- mengklasifikasikan task,
- memilih route profile,
- meneruskan request ke 9Router,
- menyediakan konfigurasi routing melalui Web UI,
- menyediakan runtime control,
- menyediakan tray icon,
- menyediakan auto-start,
- mencatat metadata non-sensitif.

Smart AI Proxy tidak boleh mengambil alih fungsi 9Router seperti:

- provider failover,
- load balancing,
- rotasi akun,
- fallback model,
- provider credential management,
- rate limit handling,
- provider-level execution.

---

## 2. Implementation Principles

### 2.1 No Code in Planning Documents

Dokumen ini tidak berisi kode implementasi. Dokumen ini hanya menjelaskan rencana kerja, fase, modul, dependency, acceptance criteria, testing strategy, release strategy, dan delivery plan.

### 2.2 Boundary First

Setiap fase implementasi wajib menjaga batas tanggung jawab antara Smart AI Proxy dan 9Router.

Smart AI Proxy:

```text
Analyze request → classify task → select route profile → forward to 9Router
```

9Router:

```text
Provider routing → failover → fallback → load balancing → account rotation → rate limit handling
```

### 2.3 Local-First and Privacy-First

Smart AI Proxy harus aman untuk berjalan di local machine developer.

Default behavior:

- bind ke `127.0.0.1`,
- tidak membuka Web UI ke public network,
- tidak log prompt penuh,
- tidak log API key,
- tidak log source code penuh,
- downstream endpoint hanya dari config,
- metadata logging saja secara default.

### 2.4 Developer Experience First

Target utama adalah developer yang menggunakan OpenCode, Cursor, Cline, Continue, atau aplikasi OpenAI-compatible lain.

Prioritas UX:

- setup mudah,
- port tidak bentrok dengan port umum developer,
- satu endpoint untuk semua client,
- Web UI untuk konfigurasi,
- tray icon untuk kontrol cepat,
- auto-start opsional,
- error message jelas.

### 2.5 Incremental Delivery

Implementasi dilakukan bertahap agar setiap fase menghasilkan deliverable yang bisa diuji.

Urutan umum:

```text
Foundation
↓
Core Proxy
↓
Routing Intelligence
↓
Web UI
↓
Tray + Auto-start
↓
Packaging
↓
Compatibility + QA
↓
MVP Release
↓
Post-MVP Improvements
```

---

## 3. Source Alignment

Implementation Plan ini mengikuti dua dokumen utama:

### 3.1 PRD v1.1 Alignment

PRD v1.1 menetapkan Smart AI Proxy sebagai OpenAI-compatible intelligent routing proxy dengan Local Web Control Panel. Fitur utama yang harus diimplementasikan mencakup:

- OpenAI-compatible proxy.
- Streaming and non-streaming passthrough.
- `smart-auto` model alias.
- Rule-based task classifier.
- Route profile selection.
- Forwarding ke 9Router.
- Local persistent configuration.
- Web UI lokal.
- Routing settings page.
- Startup settings.
- Runtime control.
- Privacy/logging settings.
- Config import/export sederhana.
- Always On, Manual, dan Disabled mode.

### 3.2 Technical Plan v0.1 Alignment

Technical Plan v0.1 menetapkan keputusan teknis berikut:

- Go sebagai core backend.
- `net/http` + `chi` sebagai HTTP foundation.
- Vue 3 + TypeScript + Vite sebagai frontend.
- Tailwind CSS + DaisyUI sebagai UI layer.
- Default Smart AI Proxy port `20127`.
- Default 9Router port `20128`.
- Satu port untuk API dan Web UI.
- Local YAML/JSON untuk config MVP.
- JSONL metadata logs.
- Tray app dengan Go systray atau Wails sebagai opsi lanjutan.
- Single binary dengan embedded Web UI.
- Windows x64 sebagai target MVP awal.
- macOS/Linux sebagai target berikutnya.

---

## 4. Target MVP Definition

### 4.1 MVP Goal

MVP dianggap berhasil jika user dapat:

1. menjalankan Smart AI Proxy di local machine,
2. melihat tray icon saat aplikasi aktif,
3. membuka Web UI lokal,
4. mengatur downstream 9Router endpoint,
5. mengatur task-to-route mapping,
6. menggunakan endpoint `http://127.0.0.1:20127/v1`,
7. memilih model virtual `smart-auto`,
8. mengirim request dari OpenAI-compatible client,
9. menerima streaming/non-streaming response dari 9Router,
10. melihat metadata routing dasar,
11. start/stop/restart proxy dari Web UI/tray,
12. mengaktifkan atau menonaktifkan auto-start,
13. menjaga prompt/source code/API key tidak masuk log default.

### 4.2 MVP Includes

MVP mencakup:

- Go app bootstrap.
- Config system.
- HTTP server.
- `/v1/chat/completions`.
- `/v1/models`.
- `/health`.
- Admin API.
- Embedded Vue Web UI.
- Task classifier rule-based.
- Task-to-route mapping.
- Route profiles.
- 9Router forwarder.
- Streaming passthrough.
- Non-streaming passthrough.
- Request ID.
- Metadata logger.
- Secret redactor.
- Safe passthrough.
- Runtime manager.
- Startup manager.
- Tray icon.
- Windows x64 portable build.
- npm wrapper basic.
- Documentation.

### 4.3 MVP Excludes

MVP tidak mencakup:

- direct provider integration,
- provider failover,
- provider load balancing,
- provider account rotation,
- provider fallback,
- provider credential manager,
- full desktop shell,
- enterprise RBAC,
- cloud sync,
- repository indexing,
- ML-based classifier,
- prompt storage by default,
- team workspace,
- advanced analytics,
- auto-update,
- hosted cloud dashboard.

---

## 5. Delivery Strategy

### 5.1 Recommended Delivery Model

Gunakan delivery model berbasis milestone dan sprint.

Rekomendasi:

```text
Phase 0  : Technical Foundation
Phase 1  : Core Proxy MVP
Phase 2  : Routing Intelligence MVP
Phase 3  : Local Web UI
Phase 4  : Tray and Auto-Start
Phase 5  : Packaging and Distribution
Phase 6  : QA, Compatibility, and Hardening
Phase 7  : MVP Release
Phase 8  : Post-MVP Improvements
```

### 5.2 Suggested Timeline

Estimasi ini mengasumsikan tim kecil 1 sampai 3 developer.

| Phase | Estimated Duration | Output |
|---|---:|---|
| Phase 0 | 3-5 hari | Repository, architecture skeleton, config draft |
| Phase 1 | 1-2 minggu | Core OpenAI-compatible proxy working |
| Phase 2 | 1-2 minggu | Classifier + route profile selection |
| Phase 3 | 2-3 minggu | Web UI settings usable |
| Phase 4 | 1-2 minggu | Tray icon + auto-start |
| Phase 5 | 1 minggu | Build, npm wrapper, portable release |
| Phase 6 | 1-2 minggu | Compatibility + security + QA hardening |
| Phase 7 | 2-3 hari | MVP release candidate |
| Phase 8 | Continuous | Post-MVP roadmap |

Total estimasi MVP realistis:

```text
8 - 12 minggu
```

Untuk solo developer, estimasi realistis:

```text
10 - 16 minggu
```

---

## 6. Workstreams

### 6.1 Backend/Core Workstream

Fokus:

- Go app bootstrap.
- HTTP server.
- proxy endpoint.
- forwarder to 9Router.
- streaming passthrough.
- request analysis.
- classifier.
- routing policy engine.
- config manager.
- observability.
- runtime manager.

### 6.2 Frontend/Web UI Workstream

Fokus:

- Vue 3 + TypeScript + Vite setup.
- Tailwind CSS + DaisyUI setup.
- dashboard page.
- routing settings.
- route profiles.
- runtime controls.
- startup settings.
- privacy/logging.
- advanced settings.
- API integration.

### 6.3 Desktop/Tray Workstream

Fokus:

- tray icon.
- tray status.
- context menu.
- open dashboard.
- start/stop/restart.
- run at startup toggle.
- quit behavior.

### 6.4 DevOps/Packaging Workstream

Fokus:

- build scripts.
- embedded Web UI.
- cross-platform build.
- GitHub Releases.
- npm package wrapper.
- portable binary.
- Windows release artifacts.
- installer plan.

### 6.5 QA/Compatibility Workstream

Fokus:

- unit tests.
- integration tests.
- streaming tests.
- config validation tests.
- UI tests.
- security tests.
- compatibility with OpenCode/Cursor/Cline/Continue.
- manual Windows startup tests.

### 6.6 Documentation Workstream

Fokus:

- README.
- quick start.
- setup guide.
- 9Router integration guide.
- OpenCode/Cursor/Cline/Continue guide.
- routing policy guide.
- troubleshooting.
- security/privacy guide.

---

## 7. Phase 0 — Technical Foundation

### 7.1 Goal

Menyiapkan struktur project, keputusan teknis, build environment, dan skeleton aplikasi sebelum fitur utama dibangun.

### 7.2 Deliverables

- Repository structure.
- Go module initialized.
- Vue/Vite app initialized.
- Tailwind CSS + DaisyUI configured.
- Config schema draft.
- Default port `20127`.
- Downstream default `20128`.
- Basic app bootstrap.
- Basic logging.
- Basic health endpoint skeleton.
- Development scripts.
- Initial documentation.

### 7.3 Tasks

| ID | Task | Owner | Priority | Output |
|---|---|---|---|---|
| P0-001 | Create repository structure | Backend | P0 | Folder sesuai technical plan |
| P0-002 | Initialize Go app | Backend | P0 | Go module and app entry |
| P0-003 | Add HTTP server skeleton | Backend | P0 | Local server can start |
| P0-004 | Add chi router | Backend | P0 | Routes separated by group |
| P0-005 | Initialize Vue TS Vite app | Frontend | P0 | Web folder ready |
| P0-006 | Add Tailwind CSS + DaisyUI | Frontend | P0 | UI style foundation |
| P0-007 | Define config file layout | Backend | P0 | `config.yaml`, `routes.yaml`, `profiles.yaml` draft |
| P0-008 | Define default ports | Backend | P0 | `20127` and `20128` constants/config |
| P0-009 | Add dev scripts | DevOps | P1 | Build/run scripts |
| P0-010 | Add initial docs | Docs | P1 | README draft |

### 7.4 Acceptance Criteria

- App can start locally.
- Server binds to `127.0.0.1:20127`.
- Health endpoint returns status.
- Web UI dev server can run separately during development.
- Config default files can be generated.
- No provider integration exists.
- Repository structure is documented.

### 7.5 Risks

| Risk | Mitigation |
|---|---|
| Scope terlalu cepat melebar | Lock MVP folder and module boundary |
| Config terlalu kompleks sejak awal | Start with minimal schema |
| Frontend/backend integration belum jelas | Define Admin API contract early |

---

## 8. Phase 1 — Core Proxy MVP

### 8.1 Goal

Membuat Smart AI Proxy dapat menerima request OpenAI-compatible dan meneruskannya ke 9Router.

### 8.2 Deliverables

- `/v1/chat/completions`.
- `/v1/models`.
- Basic request validation.
- Forwarder to 9Router.
- Streaming passthrough.
- Non-streaming passthrough.
- Request ID.
- Basic metadata logs.
- Health check.
- Downstream health check.
- Safe error propagation.

### 8.3 Tasks

| ID | Task | Owner | Priority | Output |
|---|---|---|---|---|
| P1-001 | Implement `/v1/chat/completions` route | Backend | P0 | Main proxy endpoint |
| P1-002 | Implement 9Router forwarder | Backend | P0 | Requests forwarded to `20128/v1` |
| P1-003 | Preserve request body | Backend | P0 | OpenAI-compatible body remains valid |
| P1-004 | Preserve response body | Backend | P0 | Client receives compatible response |
| P1-005 | Support non-streaming response | Backend | P0 | Full JSON passthrough |
| P1-006 | Support streaming response | Backend | P0 | SSE passthrough |
| P1-007 | Handle client disconnect | Backend | P0 | Downstream context cancelled |
| P1-008 | Generate request ID | Backend | P0 | Traceable request |
| P1-009 | Add metadata log | Backend | P0 | JSONL metadata |
| P1-010 | Add `/v1/models` | Backend | P1 | Smart aliases visible |
| P1-011 | Add downstream health check | Backend | P1 | Detect 9Router available/unavailable |
| P1-012 | Add basic error mapping | Backend | P1 | Clear error response |

### 8.4 Acceptance Criteria

- A generic OpenAI-compatible request can be sent to `http://127.0.0.1:20127/v1/chat/completions`.
- Request is forwarded to `http://127.0.0.1:20128/v1/chat/completions`.
- Non-streaming response works.
- Streaming response works.
- Request ID appears in logs.
- Authorization header is not logged.
- If 9Router is offline, user receives clear downstream error.
- Smart AI Proxy does not attempt provider fallback.

### 8.5 Testing

| Test Type | Scenario |
|---|---|
| Unit | Request ID generation |
| Unit | Header redaction |
| Integration | Mock 9Router non-streaming |
| Integration | Mock 9Router streaming |
| Integration | Downstream unavailable |
| Compatibility | Generic OpenAI SDK request |
| Manual | Use curl/Postman-like request |

---

## 9. Phase 2 — Routing Intelligence MVP

### 9.1 Goal

Menambahkan kemampuan task-aware routing dengan classifier ringan dan route profile selection.

### 9.2 Deliverables

- Request analyzer.
- Rule-based task classifier.
- Task categories.
- Confidence score.
- Task-to-route mapping.
- Route profile selector.
- Smart model aliases.
- Manual override.
- Safe passthrough.
- Routing metadata to 9Router.

### 9.3 Task Categories MVP

| Task Category | Default Route Profile |
|---|---|
| `general_chat` | `route.default` |
| `design_task` | `route.design` |
| `backend_engineering` | `route.backend` |
| `frontend_engineering` | `route.frontend` |
| `fullstack_engineering` | `route.fullstack` |
| `debugging` | `route.debugging` |
| `refactoring` | `route.refactoring` |
| `test_generation` | `route.testing` |
| `documentation` | `route.documentation` |
| `architecture_design` | `route.architect` |
| `security_review` | `route.security` |
| `long_context_analysis` | `route.long_context` |
| `lightweight_task` | `route.low_cost` |
| `unknown` | `route.default` |

### 9.4 Tasks

| ID | Task | Owner | Priority | Output |
|---|---|---|---|---|
| P2-001 | Build request analyzer | Backend | P0 | Extract prompt signals |
| P2-002 | Detect code blocks | Backend | P0 | Code signal |
| P2-003 | Detect keywords | Backend | P0 | Debug/docs/refactor/etc. |
| P2-004 | Detect domain category | Backend | P0 | Design/backend/frontend/fullstack |
| P2-005 | Detect error patterns | Backend | P0 | Stack trace/error signal |
| P2-006 | Estimate complexity | Backend | P1 | low/medium/high |
| P2-007 | Detect long context | Backend | P1 | long-context route signal |
| P2-008 | Implement rule-based classifier | Backend | P0 | task type + confidence |
| P2-009 | Implement route resolver | Backend | P0 | task → route profile |
| P2-010 | Implement route profile selector | Backend | P0 | route → downstream alias |
| P2-011 | Implement smart aliases | Backend | P0 | `smart-auto`, `smart-debug`, etc. |
| P2-012 | Implement manual override | Backend | P1 | alias/route override |
| P2-013 | Add routing decision log | Backend | P0 | reason + confidence |
| P2-014 | Add safe passthrough | Backend | P0 | default route on failure |
| P2-015 | Define downstream metadata strategy | Backend | P0 | model/header/metadata mode |

### 9.5 Routing Metadata Strategy

MVP harus mendukung minimal satu mekanisme yang stabil untuk mengirim route intent ke 9Router.

Recommended order:

1. Use model alias compatible with 9Router.
2. Use configurable downstream alias.
3. Use header metadata if 9Router supports it.
4. Fallback to default model/route.

Config example concept:

```yaml
route_profiles:
  route.backend:
    downstream_alias: combo.backend
    transport_mode: model_alias
```

### 9.6 Acceptance Criteria

- `smart-auto` triggers classification.
- `smart-debug` bypasses auto classification and selects debugging route.
- Debugging prompt selects `route.debugging`.
- Backend prompt selects `route.backend`.
- Frontend prompt selects `route.frontend`.
- Design prompt selects `route.design`.
- Documentation prompt selects `route.documentation`.
- Low confidence request uses `route.default`.
- Classifier error does not block request.
- Routing logs do not store full prompt.

### 9.7 Testing

| Test Type | Scenario |
|---|---|
| Unit | Keyword classifier |
| Unit | Code block detector |
| Unit | Stack trace detector |
| Unit | Backend/frontend/design rules |
| Unit | Confidence threshold |
| Unit | Route resolver |
| Integration | `smart-auto` to route profile |
| Integration | `smart-debug` override |
| Integration | Classifier failure safe passthrough |
| Regression | Evaluation prompt dataset |

---

## 10. Phase 3 — Local Web UI

### 10.1 Goal

Membuat user dapat mengatur Smart AI Proxy melalui browser lokal tanpa edit config manual.

### 10.2 Deliverables

- Embedded Vue Web UI.
- Admin API.
- Dashboard.
- Setup Wizard.
- Routing Settings.
- Route Profiles.
- Runtime page.
- Startup page.
- 9Router settings.
- Privacy/logging page.
- Advanced settings.
- Config import/export.
- Config validation feedback.
- Basic dry-run tester.

### 10.3 Admin API Endpoints

| Endpoint | Method | Purpose |
|---|---|---|
| `/admin/api/status` | GET | Runtime status |
| `/admin/api/config` | GET/PUT | Read/update main config |
| `/admin/api/routes` | GET/PUT | Task-to-route mapping |
| `/admin/api/profiles` | GET/PUT | Route profiles |
| `/admin/api/runtime/start` | POST | Start proxy engine |
| `/admin/api/runtime/stop` | POST | Stop proxy engine |
| `/admin/api/runtime/restart` | POST | Restart proxy engine |
| `/admin/api/startup` | GET/PUT | Startup settings |
| `/admin/api/downstream/health` | GET | 9Router status |
| `/admin/api/logs` | GET | Metadata logs |
| `/admin/api/diagnostics/export` | POST | Export diagnostics |
| `/admin/api/routing/dry-run` | POST | Test classifier without forwarding |
| `/admin/api/config/import` | POST | Import config |
| `/admin/api/config/export` | GET | Export config |
| `/admin/api/config/reset` | POST | Reset config |

### 10.4 Web UI Pages

| Page | Path | Priority | Purpose |
|---|---|---|---|
| Dashboard | `/admin` | P0 | Status ringkas |
| Setup Wizard | `/admin/setup` | P0 | First-run setup |
| Routing Settings | `/admin/routing` | P0 | Task-to-route mapping |
| Route Profiles | `/admin/profiles` | P0 | Kelola route profiles |
| Runtime | `/admin/runtime` | P0 | Start/stop/restart |
| Startup | `/admin/startup` | P0 | Always On/Manual/Disabled |
| 9Router | `/admin/downstream` | P0 | Endpoint 9Router |
| Logs | `/admin/logs` | P1 | Metadata logs |
| Privacy | `/admin/privacy` | P1 | Logging/privacy |
| Advanced | `/admin/advanced` | P1 | Port/import/export/debug |

### 10.5 Frontend Tasks

| ID | Task | Owner | Priority | Output |
|---|---|---|---|---|
| P3-001 | Build layout shell | Frontend | P0 | Sidebar + navbar |
| P3-002 | Build dashboard cards | Frontend | P0 | Status cards |
| P3-003 | Build setup wizard | Frontend | P0 | First-run flow |
| P3-004 | Build routing table | Frontend | P0 | Task-route mapping |
| P3-005 | Build route profile editor | Frontend | P0 | CRUD UI |
| P3-006 | Build runtime controls | Frontend | P0 | Start/stop/restart |
| P3-007 | Build startup settings | Frontend | P0 | Always On/Manual/Disabled |
| P3-008 | Build downstream settings | Frontend | P0 | 9Router endpoint |
| P3-009 | Build privacy/logging page | Frontend | P1 | Privacy modes |
| P3-010 | Build advanced settings | Frontend | P1 | Port/import/export |
| P3-011 | Build toast/alert system | Frontend | P0 | User feedback |
| P3-012 | Add theme support | Frontend | P2 | DaisyUI themes |
| P3-013 | Add dry-run tester UI | Frontend | P1 | Classification preview |

### 10.6 Backend Tasks

| ID | Task | Owner | Priority | Output |
|---|---|---|---|---|
| P3-101 | Implement Admin API auth | Backend | P0 | Local token/session |
| P3-102 | Implement config read/write | Backend | P0 | Persistent settings |
| P3-103 | Implement config validation | Backend | P0 | Save guard |
| P3-104 | Implement route profile API | Backend | P0 | CRUD profiles |
| P3-105 | Implement task route API | Backend | P0 | Mapping update |
| P3-106 | Implement runtime API | Backend | P0 | Start/stop/restart |
| P3-107 | Implement startup API | Backend | P0 | run_at_login setting |
| P3-108 | Implement logs API | Backend | P1 | Metadata logs view |
| P3-109 | Implement diagnostics export | Backend | P1 | Redacted export |
| P3-110 | Embed Web UI build | Backend | P0 | Single binary static assets |

### 10.7 Acceptance Criteria

- User can open `http://127.0.0.1:20127/admin`.
- Dashboard shows proxy status.
- Dashboard shows downstream 9Router status.
- User can edit task-to-route mapping.
- User can edit route profiles.
- User can change startup mode.
- User can start/stop/restart proxy.
- User can update 9Router endpoint.
- Config persists after app restart.
- Invalid config is rejected with clear message.
- Prompt logging remains disabled by default.
- Web UI is localhost-only by default.

---

## 11. Phase 4 — Tray and Auto-Start

### 11.1 Goal

Membuat Smart AI Proxy terasa seperti aplikasi lokal background dengan icon di system tray seperti 9Router.

### 11.2 Deliverables

- Tray icon.
- Tray status.
- Tray context menu.
- Open dashboard from tray.
- Start/stop/restart from tray.
- Copy API endpoint.
- Open 9Router dashboard.
- Open logs folder.
- Run at startup toggle.
- Always On/Manual/Disabled integration.
- Windows startup registration.
- Port conflict notification.
- Downstream disconnected warning.

### 11.3 Tray Menu MVP

```text
Smart AI Proxy
Status: Running / Stopped / Error
Endpoint: 127.0.0.1:20127
Open Dashboard
Open 9Router Dashboard
Copy API Endpoint
Start Proxy
Stop Proxy
Restart Proxy
Run at Startup: ON/OFF
Always On Mode: ON/OFF
Open Logs Folder
Quit
```

### 11.4 Runtime States

| State | Meaning | Tray/UI Behavior |
|---|---|---|
| `starting` | App/proxy starting | Show starting tooltip |
| `running` | Proxy accepts `/v1` requests | Active icon |
| `stopped` | App alive, proxy stopped | Inactive icon |
| `disabled` | Proxy intentionally disabled | Disabled badge |
| `error` | Runtime error | Warning icon |
| `port_conflict` | `20127` unavailable | Error tooltip |
| `downstream_offline` | 9Router unavailable | Warning badge |
| `quitting` | App exiting | Remove tray icon |

### 11.5 Auto-Start Strategy

Auto-start should run the user-level tray app, not only a system service.

Reason:

- Tray icon must appear.
- User needs quick control.
- User should see status.
- Web UI should be easy to open.
- Startup behavior should follow user session.

### 11.6 Platform Priority

MVP:

```text
Windows x64 first
```

Post-MVP:

```text
macOS arm64/x64
Linux x64
```

### 11.7 Tasks

| ID | Task | Owner | Priority | Output |
|---|---|---|---|---|
| P4-001 | Add tray library integration | Desktop | P0 | Icon appears |
| P4-002 | Define tray icon assets | Design/Desktop | P0 | active/inactive/error icon |
| P4-003 | Implement tray menu | Desktop | P0 | Context actions |
| P4-004 | Open dashboard from tray | Desktop | P0 | Browser opens admin UI |
| P4-005 | Start/stop/restart from tray | Desktop | P0 | Runtime actions |
| P4-006 | Copy endpoint action | Desktop | P1 | Clipboard integration |
| P4-007 | Open logs folder action | Desktop | P1 | File explorer opens |
| P4-008 | Open 9Router dashboard | Desktop | P1 | Browser opens `20128/dashboard` |
| P4-009 | Implement run-at-login Windows | Desktop | P0 | Startup registration |
| P4-010 | Sync tray status with runtime | Desktop | P0 | Status accurate |
| P4-011 | Show port conflict warning | Desktop | P0 | Clear tray error |
| P4-012 | Show downstream warning | Desktop | P1 | 9Router offline visible |
| P4-013 | Implement quit behavior | Desktop | P0 | Graceful shutdown |

### 11.8 Acceptance Criteria

- Tray icon appears when app starts.
- Left click or double click opens dashboard.
- Right click opens menu.
- Start/stop/restart works from tray.
- Run at Startup toggle updates startup registration.
- App can start automatically after Windows login.
- Always On starts proxy automatically.
- Manual mode does not auto-start proxy unless configured.
- Disabled mode keeps proxy inactive.
- Quit removes tray icon and stops app cleanly.

### 11.9 Risks

| Risk | Mitigation |
|---|---|
| Tray library issue on Linux | Windows-first MVP |
| Windows security blocks startup | Add repair startup button |
| Tray app and service conflict | Avoid Windows Service for MVP |
| User quits app accidentally | Confirmation if proxy running |
| Auto-start creates duplicate app instances | Single-instance lock |

---

## 12. Phase 5 — Packaging and Distribution

### 12.1 Goal

Membuat Smart AI Proxy mudah diinstall, dijalankan, dan didistribusikan untuk developer.

### 12.2 Distribution Channels

MVP distribution:

1. Portable Windows binary.
2. GitHub Releases.
3. npm global package wrapper.

Post-MVP distribution:

1. Windows installer.
2. macOS DMG.
3. Linux AppImage/deb/rpm.
4. Platform-specific npm packages.
5. Auto-update mechanism.

### 12.3 Build Model

Frontend:

```text
Vue/Vite source
↓
Static dist
↓
Embedded in Go binary
```

Backend:

```text
Go source
↓
Single executable
↓
Runs proxy + admin UI + tray
```

Distribution:

```text
Go binary
↓
GitHub Releases
↓
npm wrapper downloads or launches binary
```

### 12.4 npm Distribution Strategy

Because target users are developers, npm is a good distribution channel even though the core app is Go.

Recommended stages:

#### Stage 1 — MVP npm wrapper

Package name concept:

```text
smart-ai-proxy
```

Behavior:

```text
npm install -g smart-ai-proxy
smart-ai-proxy start
smart-ai-proxy status
smart-ai-proxy open
smart-ai-proxy tray
```

The npm package acts as installer/launcher for Go binary.

#### Stage 2 — GitHub Release Downloader

The npm package downloads the correct binary from GitHub Releases based on OS and architecture.

Requirements:

- detect OS/arch,
- download matching binary,
- verify checksum,
- store binary in npm package cache or user app data,
- expose CLI wrapper.

#### Stage 3 — Platform-Specific Packages

Long-term package layout:

```text
smart-ai-proxy
@smart-ai-proxy/win32-x64
@smart-ai-proxy/win32-arm64
@smart-ai-proxy/darwin-x64
@smart-ai-proxy/darwin-arm64
@smart-ai-proxy/linux-x64
@smart-ai-proxy/linux-arm64
```

### 12.5 Tasks

| ID | Task | Owner | Priority | Output |
|---|---|---|---|---|
| P5-001 | Add frontend production build | DevOps | P0 | Vue dist generated |
| P5-002 | Embed frontend into Go binary | Backend | P0 | Single app serves UI |
| P5-003 | Add Windows x64 build script | DevOps | P0 | `.exe` generated |
| P5-004 | Add version metadata | DevOps | P0 | version visible in UI/logs |
| P5-005 | Add portable release zip | DevOps | P0 | downloadable zip |
| P5-006 | Add GitHub Release workflow | DevOps | P1 | release artifacts |
| P5-007 | Add npm wrapper package | DevOps | P1 | install via npm |
| P5-008 | Add checksum generation | DevOps | P1 | binary verification |
| P5-009 | Add installer research | DevOps | P2 | MSI/NSIS plan |
| P5-010 | Add release notes template | Docs | P1 | changelog |
| P5-011 | Add uninstall notes | Docs | P1 | cleanup guide |

### 12.6 Acceptance Criteria

- A Windows user can download portable zip, run the app, see tray icon, and open dashboard.
- A developer can install via npm wrapper and launch the binary.
- Build contains embedded Web UI.
- No separate Node.js runtime is needed to run production app.
- Version appears in dashboard and logs.
- Release artifact includes README/quick start.
- Release does not include secrets or local config.

---

## 13. Phase 6 — QA, Compatibility, and Hardening

### 13.1 Goal

Memastikan MVP stabil, kompatibel dengan client target, aman secara default, dan tidak merusak OpenAI-compatible behavior.

### 13.2 Test Matrix

| Category | Required Tests |
|---|---|
| Unit | analyzer, classifier, route resolver, config validator |
| Integration | proxy forwarding, streaming, downstream offline |
| UI | settings pages, save config, runtime controls |
| Security | header redaction, local-only binding, admin auth |
| Privacy | no prompt logging, diagnostics redaction |
| Compatibility | OpenCode, Cursor, Cline, Continue, generic SDK |
| System | tray icon, startup, port conflict, reboot |
| Packaging | portable binary, npm wrapper, embedded UI |

### 13.3 Compatibility Targets

MVP must validate at least:

- Generic OpenAI-compatible HTTP request.
- Generic OpenAI SDK-like client.
- One AI coding assistant.
- Streaming request.
- Non-streaming request.
- Tool/function calling passthrough if supported by client payload.

Post-MVP expands to:

- OpenCode.
- Cursor.
- Cline.
- Continue.
- Custom scripts.

### 13.4 Security Acceptance Criteria

MVP security is acceptable if:

- Admin UI binds to localhost only by default.
- Authorization headers never appear in logs.
- API keys never appear in logs.
- Full prompt logging is disabled by default.
- Diagnostics export redacts secrets.
- Downstream endpoint cannot be overridden per request.
- LAN access requires explicit opt-in.
- Invalid config cannot be saved.
- Web UI settings require local token/password if auth enabled.

### 13.5 Performance Acceptance Criteria

| Metric | MVP Target |
|---|---|
| Proxy overhead | Low enough to not feel disruptive |
| Classifier overhead | Lightweight, no LLM call |
| Streaming behavior | No full buffering |
| Startup time | Fast for local app |
| Web UI responsiveness | Feels instant locally |
| Memory footprint | Suitable for background developer app |

### 13.6 Tasks

| ID | Task | Owner | Priority | Output |
|---|---|---|---|---|
| P6-001 | Add unit test suite | QA/Backend | P0 | Core modules tested |
| P6-002 | Add integration test with mock 9Router | QA/Backend | P0 | Forwarding verified |
| P6-003 | Add streaming test | QA/Backend | P0 | SSE verified |
| P6-004 | Add config validation tests | QA/Backend | P0 | Invalid config rejected |
| P6-005 | Add redaction tests | QA/Security | P0 | Secrets hidden |
| P6-006 | Add UI manual QA checklist | QA/Frontend | P1 | Pages tested |
| P6-007 | Add Windows tray manual tests | QA/Desktop | P0 | Tray stable |
| P6-008 | Add startup/reboot test | QA/Desktop | P0 | Auto-start works |
| P6-009 | Add compatibility test guide | QA/Docs | P1 | Client validation |
| P6-010 | Run release candidate test pass | QA | P0 | MVP RC signoff |

---

## 14. Phase 7 — MVP Release

### 14.1 Goal

Merilis versi MVP yang dapat digunakan developer secara realistis, dengan batasan fitur yang jelas.

### 14.2 MVP Release Criteria

MVP siap dirilis jika:

1. App runs on Windows x64.
2. Tray icon appears.
3. Web UI opens at `http://127.0.0.1:20127/admin`.
4. API endpoint works at `http://127.0.0.1:20127/v1`.
5. 9Router downstream default is `http://127.0.0.1:20128/v1`.
6. `/v1/chat/completions` works.
7. Streaming works.
8. Non-streaming works.
9. `smart-auto` works.
10. Rule-based classifier works for key task categories.
11. Task-to-route mapping can be changed via Web UI.
12. Route profiles can be managed at least basically.
13. Runtime start/stop/restart works.
14. Startup mode can be configured.
15. Auto-start works on Windows login.
16. Logs contain metadata only.
17. Secrets are redacted.
18. Config persists after restart.
19. Safe passthrough works if classifier fails.
20. 9Router failure shows clear message.
21. No provider-level routing is implemented inside Smart AI Proxy.
22. Documentation is sufficient for first-time setup.

### 14.3 Release Artifacts

| Artifact | Required |
|---|---|
| Windows portable zip | Yes |
| Go binary | Yes |
| npm wrapper | Yes |
| README | Yes |
| Quick start guide | Yes |
| Configuration guide | Yes |
| Troubleshooting guide | Yes |
| Release notes | Yes |
| Checksums | Recommended |
| Installer | Optional MVP |

### 14.4 Release Naming

Recommended version:

```text
v0.1.0-alpha
```

Then:

```text
v0.2.0-beta
v0.3.0-rc.1
v1.0.0
```

### 14.5 Post-Release Monitoring

After MVP release, track:

- install issues,
- port conflict reports,
- 9Router connection issues,
- streaming compatibility issues,
- classifier wrong-route reports,
- tray icon issues,
- startup not working,
- UI confusion,
- config corruption,
- packaging/npm install issues.

---

## 15. Phase 8 — Post-MVP Improvements

### 15.1 Routing Quality

- More classifier rules.
- Multi-label classification.
- Better confidence scoring.
- Evaluation dataset.
- Dry-run improvements.
- User feedback on route quality.
- Project-specific routing.
- Route templates.

### 15.2 Observability

- Better dashboard metrics.
- Request distribution charts.
- Route usage analytics.
- Safe passthrough trend.
- Downstream latency trend.
- Exportable reports.
- SQLite metadata storage.

### 15.3 Packaging

- Native Windows installer.
- macOS DMG.
- Linux AppImage/deb/rpm.
- platform-specific npm packages.
- automatic update check.
- signed binaries.

### 15.4 Desktop Experience

- Wails desktop shell evaluation.
- Native settings window.
- Better tray states.
- Notification messages.
- startup repair tool.
- first-run onboarding.

### 15.5 Team Mode

- Shared policy file.
- workspace profiles.
- team route presets.
- config export/import templates.
- audit log.
- policy lock.

### 15.6 Advanced Intelligence

- cost-aware routing.
- latency-aware routing.
- historical performance analysis.
- learning-assisted classifier.
- prompt normalization.
- opt-in evaluation samples.

---

## 16. Epics and Backlog

### Epic E1 — Core Application Foundation

| Story ID | User Story | Priority |
|---|---|---|
| E1-S1 | As a developer, I can start Smart AI Proxy locally. | P0 |
| E1-S2 | As a developer, I can access health status. | P0 |
| E1-S3 | As a developer, I can use default config files. | P0 |
| E1-S4 | As a maintainer, I can build frontend and backend. | P0 |
| E1-S5 | As a maintainer, I can produce a Windows binary. | P1 |

### Epic E2 — OpenAI-Compatible Proxy

| Story ID | User Story | Priority |
|---|---|---|
| E2-S1 | As a user, I can send chat completion requests to Smart AI Proxy. | P0 |
| E2-S2 | As a user, I can receive non-streaming responses. | P0 |
| E2-S3 | As a user, I can receive streaming responses. | P0 |
| E2-S4 | As a user, I can use generic OpenAI-compatible client. | P0 |
| E2-S5 | As an admin, I can see downstream errors clearly. | P1 |

### Epic E3 — Routing Intelligence

| Story ID | User Story | Priority |
|---|---|---|
| E3-S1 | As a user, I can use `smart-auto`. | P0 |
| E3-S2 | As a user, debugging prompts route to debugging route. | P0 |
| E3-S3 | As a user, backend prompts route to backend route. | P0 |
| E3-S4 | As a user, frontend prompts route to frontend route. | P0 |
| E3-S5 | As a user, design prompts route to design route. | P0 |
| E3-S6 | As a power user, I can override routing via alias. | P1 |
| E3-S7 | As a user, request still works when classification fails. | P0 |

### Epic E4 — Web UI

| Story ID | User Story | Priority |
|---|---|---|
| E4-S1 | As a user, I can open the local dashboard. | P0 |
| E4-S2 | As a user, I can change task-to-route mapping. | P0 |
| E4-S3 | As a user, I can manage route profiles. | P0 |
| E4-S4 | As a user, I can start/stop/restart proxy. | P0 |
| E4-S5 | As a user, I can configure 9Router endpoint. | P0 |
| E4-S6 | As a user, I can configure startup mode. | P0 |
| E4-S7 | As a user, I can import/export config. | P1 |
| E4-S8 | As a user, I can view privacy-safe logs. | P1 |

### Epic E5 — Tray and Startup

| Story ID | User Story | Priority |
|---|---|---|
| E5-S1 | As a user, I see tray icon when app is active. | P0 |
| E5-S2 | As a user, I can open dashboard from tray. | P0 |
| E5-S3 | As a user, I can start/stop/restart from tray. | P0 |
| E5-S4 | As a user, I can enable run at startup. | P0 |
| E5-S5 | As a user, Smart AI Proxy starts after login in Always On mode. | P0 |
| E5-S6 | As a user, I can copy API endpoint from tray. | P1 |

### Epic E6 — Packaging and Distribution

| Story ID | User Story | Priority |
|---|---|---|
| E6-S1 | As a user, I can download a portable binary. | P0 |
| E6-S2 | As a developer, I can install via npm. | P1 |
| E6-S3 | As a maintainer, I can publish GitHub Releases. | P1 |
| E6-S4 | As a user, I can read quick start docs. | P0 |
| E6-S5 | As a user, I can troubleshoot common errors. | P1 |

### Epic E7 — Security and Privacy

| Story ID | User Story | Priority |
|---|---|---|
| E7-S1 | As a user, my API keys are not logged. | P0 |
| E7-S2 | As a user, my prompt is not logged by default. | P0 |
| E7-S3 | As a user, Web UI is local-only by default. | P0 |
| E7-S4 | As a user, diagnostics export is redacted. | P1 |
| E7-S5 | As a user, invalid config is rejected. | P0 |

---

## 17. Traceability Matrix

### 17.1 PRD Requirement to Implementation Phase

| PRD Area | Implementation Phase |
|---|---|
| OpenAI-compatible proxy | Phase 1 |
| Streaming passthrough | Phase 1 |
| Non-streaming passthrough | Phase 1 |
| Request analysis | Phase 2 |
| Task classification | Phase 2 |
| Design/backend/frontend classification | Phase 2 |
| Task-to-route mapping | Phase 2, Phase 3 |
| Route profile selection | Phase 2, Phase 3 |
| Manual override | Phase 2 |
| Web UI | Phase 3 |
| Runtime control | Phase 3, Phase 4 |
| Startup mode | Phase 3, Phase 4 |
| Auto-start on boot/login | Phase 4 |
| Tray icon | Phase 4 |
| Local config persistence | Phase 0, Phase 3 |
| Observability/logging | Phase 1, Phase 6 |
| Privacy/security | Phase 1, Phase 3, Phase 6 |
| Config import/export | Phase 3 |
| Packaging | Phase 5 |
| Compatibility testing | Phase 6 |

### 17.2 Technical Decision to Implementation Phase

| Technical Decision | Implementation Phase |
|---|---|
| Go core | Phase 0 onward |
| `net/http` + `chi` | Phase 0, Phase 1 |
| Vue TS Vite | Phase 0, Phase 3 |
| Tailwind CSS + DaisyUI | Phase 0, Phase 3 |
| Port `20127` | Phase 0 |
| 9Router `20128` | Phase 0, Phase 1 |
| Same port UI/API | Phase 3 |
| YAML/JSON config | Phase 0, Phase 3 |
| JSONL logs | Phase 1 |
| Go systray | Phase 4 |
| Embedded Web UI | Phase 5 |
| npm wrapper | Phase 5 |
| Windows-first MVP | Phase 4, Phase 5, Phase 6 |

---

## 18. Suggested Sprint Plan

### Sprint 0 — Planning and Foundation

Duration: 3-5 days

Goals:

- repo ready,
- Go app skeleton,
- Vue app skeleton,
- config draft,
- dev scripts,
- initial docs.

Output:

- foundation merged.

### Sprint 1 — Core Proxy Non-Streaming

Duration: 1 week

Goals:

- `/v1/chat/completions`,
- non-streaming forwarding,
- request ID,
- health endpoint,
- basic logs.

Output:

- non-streaming requests work.

### Sprint 2 — Streaming and Error Handling

Duration: 1 week

Goals:

- streaming passthrough,
- client disconnect handling,
- downstream error handling,
- `/v1/models`,
- redaction baseline.

Output:

- streaming requests work.

### Sprint 3 — Analyzer and Classifier

Duration: 1 week

Goals:

- analyzer,
- rule-based classifier,
- route resolver,
- smart aliases,
- safe passthrough.

Output:

- `smart-auto` works.

### Sprint 4 — Config and Route Profiles

Duration: 1 week

Goals:

- config persistence,
- route profiles,
- task routes,
- config validation,
- last-known-good config.

Output:

- routing policy is configurable.

### Sprint 5 — Web UI Foundation

Duration: 1 week

Goals:

- embedded UI foundation,
- dashboard,
- setup wizard,
- admin API auth,
- status API.

Output:

- dashboard usable.

### Sprint 6 — Web UI Settings

Duration: 1-2 weeks

Goals:

- routing settings,
- route profiles,
- downstream settings,
- runtime controls,
- startup settings,
- privacy settings.

Output:

- Web UI can configure main product behavior.

### Sprint 7 — Tray and Startup

Duration: 1-2 weeks

Goals:

- tray icon,
- tray menu,
- open dashboard,
- start/stop/restart,
- Windows run at login,
- status sync.

Output:

- app behaves like background tray app.

### Sprint 8 — Packaging and npm

Duration: 1 week

Goals:

- embedded build,
- Windows portable release,
- GitHub Release workflow,
- npm wrapper,
- checksums.

Output:

- installable/distributable MVP candidate.

### Sprint 9 — QA and Compatibility

Duration: 1-2 weeks

Goals:

- test suite,
- mock 9Router tests,
- streaming compatibility,
- redaction tests,
- manual AI coding assistant tests,
- documentation.

Output:

- release candidate.

### Sprint 10 — MVP Release

Duration: 2-3 days

Goals:

- release artifacts,
- release notes,
- quick start validation,
- known issues,
- publish.

Output:

- `v0.1.0-alpha`.

---

## 19. Repository Plan

Recommended repository structure:

```text
smart-ai-proxy/
├── cmd/
│   └── smart-ai-proxy/
│       └── main.go
├── internal/
│   ├── app/
│   ├── server/
│   ├── proxy/
│   ├── analyzer/
│   ├── classifier/
│   ├── routing/
│   ├── forwarder/
│   ├── config/
│   ├── storage/
│   ├── observability/
│   ├── security/
│   ├── startup/
│   └── tray/
├── web/
│   ├── src/
│   │   ├── app/
│   │   ├── pages/
│   │   ├── components/
│   │   ├── stores/
│   │   ├── router/
│   │   └── api/
│   ├── public/
│   ├── package.json
│   ├── vite.config.ts
│   ├── tailwind.config.ts
│   └── index.html
├── configs/
│   ├── config.example.yaml
│   ├── routes.example.yaml
│   └── profiles.example.yaml
├── docs/
│   ├── prd.md
│   ├── technical-plan.md
│   ├── implementation-plan.md
│   ├── setup.md
│   ├── routing-policy.md
│   ├── security.md
│   ├── troubleshooting.md
│   └── integrations/
│       ├── opencode.md
│       ├── cursor.md
│       ├── cline.md
│       └── continue.md
├── testdata/
│   ├── classification/
│   ├── openai-compatible/
│   └── streaming/
├── scripts/
│   ├── build.sh
│   ├── build.ps1
│   ├── package.ps1
│   └── release.sh
├── packaging/
│   ├── npm/
│   ├── windows/
│   ├── macos/
│   └── linux/
├── .github/
│   └── workflows/
├── README.md
└── CHANGELOG.md
```

---

## 20. Configuration Implementation Plan

### 20.1 Config Files

MVP config files:

```text
config.yaml
routes.yaml
profiles.yaml
logs.jsonl
```

Recommended location:

| OS | Location |
|---|---|
| Windows | `%APPDATA%/SmartAIProxy/` |
| macOS | `~/Library/Application Support/SmartAIProxy/` |
| Linux | `~/.config/smart-ai-proxy/` |

### 20.2 Config Lifecycle

```text
App starts
↓
Find config directory
↓
Load config files
↓
If missing, generate defaults
↓
Validate config
↓
If valid, use config
↓
If invalid, use last-known-good
↓
If no valid config, enter setup required state
```

### 20.3 Config Save Flow

```text
User updates config in Web UI
↓
Admin API validates payload
↓
Config validator checks references and security
↓
Write temp config
↓
Backup current config
↓
Replace active config
↓
Reload runtime if needed
↓
Update UI status
```

### 20.4 Last-Known-Good Strategy

Maintain:

```text
config.yaml
config.last-good.yaml
config.backup.yaml
```

If config breaks startup:

- app should not crash silently,
- restore last-known-good if safe,
- show error in UI/tray,
- allow reset configuration.

---

## 21. Runtime Control Plan

### 21.1 Runtime Components

| Component | Responsibility |
|---|---|
| App process | Owns tray, UI, config, runtime manager |
| HTTP server | Serves `/v1`, `/admin`, `/health` |
| Proxy engine | Accepts or rejects `/v1` requests depending on mode |
| Runtime manager | Start/stop/restart state |
| Startup manager | Run-at-login registration |
| Tray controller | Quick user controls |

### 21.2 Start Behavior

If mode is Always On:

```text
App launch → start proxy engine → show tray running
```

If mode is Manual:

```text
App launch → show tray → proxy state follows start_proxy_on_app_launch config
```

If mode is Disabled:

```text
App launch → show tray disabled → proxy does not accept requests
```

### 21.3 Stop Behavior

Stop proxy should:

- stop accepting new `/v1` requests,
- let current request finish if possible,
- update runtime state,
- update tray and Web UI,
- keep admin UI available.

### 21.4 Restart Behavior

Restart proxy should:

- validate config,
- stop accepting new proxy requests,
- gracefully wait for current active requests up to timeout,
- restart proxy engine,
- update status.

---

## 22. Routing Implementation Plan

### 22.1 Routing Flow

```text
OpenAI-compatible request
↓
Request ID assigned
↓
Model field inspected
↓
Explicit override checked
↓
Messages analyzed
↓
Task classified
↓
Task-to-route mapping resolved
↓
Route profile selected
↓
Downstream alias resolved
↓
Request transformed minimally
↓
Forwarded to 9Router
```

### 22.2 Override Flow

```text
model = smart-debug
↓
Recognize smart alias
↓
Select route.debugging
↓
Skip auto classification for route decision
↓
Forward to 9Router
```

### 22.3 Auto Flow

```text
model = smart-auto
↓
Analyze request
↓
Classify task
↓
Select route based on Web UI mapping
↓
Forward to 9Router
```

### 22.4 Passthrough Flow

```text
model = real model / unknown model
↓
If auto-routing disabled or not smart alias
↓
Passthrough or default policy depending config
↓
Forward to 9Router
```

### 22.5 Failure Flow

```text
Classifier error
↓
Log classification_status = failed
↓
Select route.default
↓
Forward to 9Router
```

---

## 23. Web UI UX Implementation Plan

### 23.1 First-Run Setup Wizard

Steps:

1. Welcome.
2. Confirm Smart AI Proxy endpoint.
3. Configure 9Router endpoint.
4. Test 9Router connection.
5. Configure default route.
6. Configure basic task-to-route mapping.
7. Configure startup mode.
8. Configure privacy mode.
9. Finish and show copy endpoint button.

### 23.2 Dashboard

Dashboard widgets:

- Proxy status.
- API endpoint.
- Admin URL.
- 9Router status.
- Current runtime mode.
- Startup mode.
- Auto-routing status.
- Total requests today.
- Most used route.
- Last error.

### 23.3 Routing Settings

Main table:

| Task Category | Route Profile | Enabled | Notes |
|---|---|---|---|
| Design | Dropdown | Toggle | UI/UX/design |
| Backend | Dropdown | Toggle | API/db/server |
| Frontend | Dropdown | Toggle | UI/component/browser |
| Debugging | Dropdown | Toggle | error/fix |
| Documentation | Dropdown | Toggle | README/docs |
| Architecture | Dropdown | Toggle | system design |

### 23.4 Route Profile Editor

Fields:

- route name,
- label,
- description,
- downstream alias,
- priority mode,
- enabled,
- notes.

Validation:

- route name must be unique,
- downstream alias cannot be empty,
- default route cannot be disabled,
- task route cannot point to missing route profile.

### 23.5 Runtime Page

Controls:

- Start Proxy.
- Stop Proxy.
- Restart Proxy.
- Copy API Endpoint.
- Open Logs Folder.
- Open 9Router Dashboard.
- Test 9Router Connection.

### 23.6 Startup Page

Controls:

- Run at login.
- Start proxy on app launch.
- Always On.
- Manual.
- Disabled.
- Restart after crash.
- Repair startup registration.

### 23.7 Privacy Page

Controls:

- privacy mode: Standard / Strict / Debug,
- metadata logging enabled,
- prompt logging disabled by default,
- redact secrets,
- retention days,
- clear logs,
- export diagnostics.

### 23.8 Advanced Page

Controls:

- host,
- port,
- downstream URL,
- timeout,
- config import,
- config export,
- reset config,
- debug mode,
- LAN access opt-in.

---

## 24. Security Implementation Plan

### 24.1 Default Security Posture

| Area | Default |
|---|---|
| Bind address | `127.0.0.1` |
| Web UI exposure | Local only |
| Prompt logging | Off |
| API key logging | Never |
| Authorization logging | Redacted |
| Downstream override by request | Not allowed |
| LAN access | Off |
| Admin auth | Enabled or first-run local token |
| Diagnostics export | Redacted |

### 24.2 Redaction Targets

Must redact:

- Authorization headers,
- bearer tokens,
- API keys,
- passwords,
- private keys,
- `.env`-like secrets,
- provider credentials,
- raw request prompt in default mode.

### 24.3 Admin Auth MVP

Acceptable MVP options:

1. Local token generated on first run.
2. Password set in first-run setup.
3. Localhost-only token cookie.

Recommendation:

```text
First-run generated admin token + optional user password later
```

### 24.4 LAN Access

If user enables LAN access:

- show warning,
- require admin auth,
- require confirmation,
- log setting change,
- show visible badge in UI.

---

## 25. Observability Implementation Plan

### 25.1 Metadata Logs

Use JSONL for MVP.

Fields:

```text
timestamp
request_id
client_name
requested_model
task_type
classification_confidence
selected_route
downstream_alias
routing_reason
status_code
proxy_latency_ms
downstream_latency_ms
streaming
error_class
```

### 25.2 Metrics MVP

Compute from JSONL:

- total requests,
- requests by route,
- requests by task type,
- streaming vs non-streaming,
- safe passthrough count,
- manual override count,
- error count,
- downstream unavailable count,
- average latency,
- port conflict count.

### 25.3 Diagnostics Export

Export bundle should include:

- app version,
- OS info,
- config summary redacted,
- route profiles,
- task routes,
- recent metadata logs,
- runtime status,
- downstream status.

Never include:

- full prompt,
- source code,
- API keys,
- Authorization headers,
- raw credentials.

---

## 26. Testing Plan

### 26.1 Unit Tests

Modules:

- config loader,
- config validator,
- request analyzer,
- classifier,
- route resolver,
- route profile selector,
- forwarder request builder,
- redactor,
- metadata logger,
- runtime manager.

### 26.2 Integration Tests

Scenarios:

- non-streaming request to mock 9Router,
- streaming request to mock 9Router,
- downstream 9Router offline,
- invalid OpenAI request,
- classifier failure,
- safe passthrough,
- route profile update,
- task mapping update,
- config reload,
- port conflict.

### 26.3 UI Tests

Manual MVP checklist:

- dashboard loads,
- setup wizard works,
- routing settings save,
- route profiles save,
- startup toggle persists,
- runtime start/stop/restart works,
- downstream health check updates,
- privacy settings persist,
- invalid config shows error,
- logs do not show secrets.

### 26.4 System Tests

Windows MVP:

- portable app launches,
- tray icon appears,
- dashboard opens from tray,
- proxy endpoint works,
- auto-start after login works,
- Manual mode does not auto-start proxy,
- Disabled mode stays disabled,
- port conflict shown,
- 9Router disconnected shown,
- quit removes tray icon.

### 26.5 Compatibility Tests

Test clients:

- generic OpenAI-compatible HTTP request,
- generic SDK request,
- OpenCode,
- Cursor,
- Cline,
- Continue.

Minimum pass criteria:

- client can configure base URL `http://127.0.0.1:20127/v1`,
- client can use model `smart-auto`,
- streaming response works if client uses streaming,
- tool/function payload is preserved,
- errors are understandable.

---

## 27. CI/CD Plan

### 27.1 CI Checks

On pull request:

- Go formatting.
- Go unit tests.
- Go static checks.
- Frontend type check.
- Frontend build.
- Config example validation.
- Redaction tests.
- Basic integration tests with mock 9Router.

### 27.2 Release Workflow

On tag:

```text
Create version tag
↓
Build Vue app
↓
Embed Web UI
↓
Build Go binaries
↓
Generate checksums
↓
Package portable zip
↓
Publish GitHub Release draft
↓
Publish npm package if release approved
```

### 27.3 Versioning

Use semantic versioning:

```text
v0.1.0-alpha
v0.2.0-beta
v0.3.0-rc.1
v1.0.0
```

### 27.4 Branching Strategy

Recommended:

```text
main       → stable
develop    → active integration
feature/*  → feature work
release/*  → release candidate stabilization
hotfix/*   → urgent fixes
```

For solo developer, simpler strategy:

```text
main + feature branches + tagged releases
```

---

## 28. Documentation Plan

### 28.1 Required MVP Docs

| Doc | Purpose |
|---|---|
| README.md | Project overview and quick start |
| setup.md | Installation and first-run setup |
| routing-policy.md | Task-to-route and route profile explanation |
| 9router.md | 9Router integration guide |
| opencode.md | OpenCode setup |
| cursor.md | Cursor setup |
| cline.md | Cline setup |
| continue.md | Continue setup |
| security.md | Privacy and security defaults |
| troubleshooting.md | Common issues |
| packaging.md | npm/binary distribution |
| release-notes.md | Version changes |

### 28.2 Troubleshooting Topics

Must include:

- port `20127` already used,
- 9Router not running,
- wrong downstream URL,
- streaming not working,
- tray icon not appearing,
- auto-start not working,
- config invalid,
- npm install failed,
- firewall/security warning,
- AI client cannot connect,
- `smart-auto` not routing as expected.

---

## 29. Risk Register

| Risk | Probability | Impact | Mitigation |
|---|---:|---:|---|
| Port `20127` already used | Low | Medium | Clear error and custom port option |
| 9Router not running | High | Medium | Dashboard health status and clear error |
| Streaming passthrough broken | Medium | High | Integration tests with SSE |
| Classifier inaccurate | High | Medium | Manual override and safe fallback |
| Web UI config bug | Medium | High | Validation and last-known-good config |
| Prompt accidentally logged | Low | High | Redaction tests and logging defaults |
| Tray unstable on Linux | Medium | Low for MVP | Windows-first MVP |
| Auto-start blocked by OS | Medium | Medium | Repair startup flow |
| npm wrapper download issue | Medium | Medium | GitHub Release fallback instructions |
| Scope creep into provider routing | Medium | High | Architecture review gate |
| 9Router route metadata mismatch | Medium | High | Configurable transport mode |
| UI exposed to network accidentally | Low | High | Bind localhost by default |
| Duplicate app instances | Medium | Medium | Single-instance lock |
| Config corruption | Medium | Medium | Backup + last-known-good |

---

## 30. Decision Gates

### Gate 0 — Foundation Ready

Can proceed if:

- repo structure ready,
- app starts,
- health endpoint works,
- config defaults exist,
- frontend setup works.

### Gate 1 — Proxy Ready

Can proceed if:

- non-streaming works,
- streaming works,
- 9Router forwarder works,
- errors are clear,
- logs are metadata only.

### Gate 2 — Routing Ready

Can proceed if:

- `smart-auto` works,
- route profiles work,
- task mapping works,
- manual override works,
- safe passthrough works.

### Gate 3 — UI Ready

Can proceed if:

- dashboard works,
- settings save,
- config persists,
- invalid config rejected,
- runtime controls work.

### Gate 4 — Tray Ready

Can proceed if:

- tray icon appears,
- menu works,
- startup works,
- status sync works.

### Gate 5 — Release Candidate

Can release if:

- all P0 tests pass,
- Windows portable build works,
- npm wrapper works or documented as beta,
- docs are ready,
- no secret leaks,
- no provider-level boundary violation.

---

## 31. Definition of Ready

A task is ready for implementation if:

- requirement is clear,
- acceptance criteria are defined,
- owner is assigned,
- dependencies are known,
- affected modules are identified,
- test approach is clear,
- security/privacy impact is understood,
- no unresolved architectural question blocks it.

---

## 32. Definition of Done

A task is done if:

- implementation meets acceptance criteria,
- no OpenAI-compatible behavior is broken,
- streaming behavior is preserved if relevant,
- config validation is applied if config is touched,
- logs do not leak sensitive data,
- tests are added or manual checklist updated,
- documentation is updated if user-facing,
- feature works with default port `20127`,
- downstream remains 9Router,
- no provider-level routing responsibility is added to Smart AI Proxy.

---

## 33. MVP Success Metrics

### 33.1 Product Success

- User can complete first-run setup.
- User can connect at least one AI coding assistant.
- User can use `smart-auto`.
- User can change routing settings from Web UI.
- User can enable auto-start.
- User can control app from tray.

### 33.2 Technical Success

- Streaming success rate high in tested clients.
- Proxy overhead low.
- Config persists across restart.
- No full prompt logs by default.
- Safe passthrough works.
- 9Router errors are clear.
- Portable build works.

### 33.3 Quality Success

- P0 bugs closed before MVP release.
- No known secret logging issue.
- No known critical streaming issue.
- No known startup crash issue.
- Documentation covers main setup path.

---

## 34. Initial Open Questions to Resolve During Implementation

1. What exact mechanism should be used to pass route intent to 9Router: model alias, custom header, or metadata?
2. Should Web UI auth be mandatory on localhost from MVP or optional?
3. Should tray run by default after `npm install -g`, or only after explicit `smart-ai-proxy tray`?
4. Should npm wrapper download binary from GitHub Releases or bundle Windows binary first?
5. Should MVP support macOS/Linux tray, or document Windows-only tray initially?
6. Should `/v1/models` list only smart aliases or also fetch downstream models from 9Router?
7. Should stop proxy disable only `/v1/*` while keeping `/admin/*` alive?
8. Should Manual mode start proxy on app launch by default or wait for user click?
9. Should config import require full validation before applying?
10. Should debug mode allow prompt logging only with explicit warning?

---

## 35. Recommended MVP Cut Line

If time becomes limited, keep these as mandatory:

- Go core app.
- `/v1/chat/completions`.
- streaming and non-streaming.
- forward to 9Router.
- `smart-auto`.
- task classifier basic.
- route profiles basic.
- Web UI dashboard.
- routing settings.
- startup settings.
- runtime controls.
- tray icon basic.
- metadata-only logs.
- Windows portable build.
- setup documentation.

Move these to post-MVP if needed:

- advanced logs page,
- diagnostics export,
- config import/export,
- dry-run tester,
- npm platform-specific packages,
- macOS/Linux tray,
- native installer,
- SQLite,
- project-specific routing,
- team mode.

---

## 36. Final Implementation Positioning

Smart AI Proxy MVP should be implemented as:

```text
A Go-based local background app with:
- OpenAI-compatible proxy endpoint,
- intelligent task-to-route selection,
- Vue TS DaisyUI Web UI,
- system tray control,
- auto-start support,
- local config persistence,
- privacy-safe metadata logging,
- 9Router as the only downstream execution layer.
```

The first release should prove the core product promise:

> Developer cukup memakai satu endpoint dan satu model virtual seperti `smart-auto`; Smart AI Proxy memilih route terbaik berdasarkan task dan setting user; 9Router tetap mengeksekusi provider-level routing secara andal.

