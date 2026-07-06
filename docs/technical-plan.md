# AtlasBridge
# Technical Plan v0.1

**Document Type:** Technical Plan  
**Project Name:** AtlasBridge  
**Full Name:** AtlasBridge AI Proxy  
**Status:** Draft v0.1  
**Primary Runtime:** Local developer machine  
**Primary Downstream:** 9Router  
**Default AtlasBridge Port:** `20127`  
**Default 9Router Port:** `20128`  
**Frontend:** Vue 3 + TypeScript + Vite + Tailwind CSS + DaisyUI  
**Backend/Core:** Go / Golang  
**Primary Goal:** OpenAI-compatible intelligent routing proxy with local Web UI and tray icon control  

---

## 1. Executive Technical Summary

AtlasBridge akan dibangun sebagai aplikasi lokal untuk developer yang berjalan di background, menyediakan endpoint OpenAI-compatible, melakukan analisis request, mengklasifikasikan task, memilih route profile, lalu meneruskan request ke 9Router.

AtlasBridge tidak menggantikan 9Router. 9Router tetap bertanggung jawab terhadap failover, load balancing, rotasi akun, fallback model, rate limit handling, provider abstraction, dan provider credential handling.

AtlasBridge bertanggung jawab terhadap:

- OpenAI-compatible proxy layer.
- Request analysis.
- Task classification.
- Task-to-route mapping.
- Route profile selection.
- Forwarding ke 9Router.
- Local Web UI untuk konfigurasi.
- System tray icon saat aplikasi aktif.
- Auto-start saat laptop/komputer restart.
- Runtime control: start, stop, restart, quit.
- Local configuration persistence.
- Metadata logging tanpa menyimpan prompt penuh secara default.

Keputusan teknologi utama:

```text
Core Language       : Go
HTTP Layer          : net/http
Router              : chi
Frontend            : Vue 3 + TypeScript + Vite
UI Framework        : Tailwind CSS + DaisyUI
Default Proxy Port  : 20127
Default 9Router     : 20128
Storage MVP         : Local YAML/JSON
Storage Post-MVP    : SQLite
Tray App            : Go systray or Wails-based tray shell
Packaging           : Single local app with embedded Web UI
```

---

## 2. Confirmed Technical Decisions

## 2.1 Language Decision

AtlasBridge menggunakan **Go** sebagai bahasa utama.

Alasan:

- Cocok untuk HTTP proxy dan network service.
- Performa ringan.
- Cocok untuk streaming.
- Concurrency model sederhana.
- Bisa dipaketkan sebagai binary lokal.
- Mudah dijalankan sebagai background app.
- Cocok untuk Windows, macOS, dan Linux.
- Standard library kuat, terutama untuk HTTP server/client.

Alternatif seperti Node.js, Python, dan Rust tidak dipilih sebagai core utama.

Node.js cocok untuk UI tooling tetapi kurang ideal untuk background proxy single-binary. Python bagus untuk prototyping, tetapi kurang ideal untuk local proxy production-style. Rust sangat kuat, tetapi complexity lebih tinggi untuk MVP.

## 2.2 Backend Framework Decision

Backend menggunakan:

```text
Go net/http + chi
```

`net/http` digunakan sebagai fondasi HTTP server/client. `chi` digunakan sebagai lightweight router untuk memisahkan endpoint API, admin API, health check, dan static Web UI.

## 2.3 Frontend Decision

Frontend menggunakan:

```text
Vue 3 + TypeScript + Vite + Tailwind CSS + DaisyUI
```

Alasan:

- Vue cocok untuk dashboard settings yang interaktif.
- TypeScript menjaga maintainability.
- Vite cepat untuk development dan build.
- DaisyUI mempercepat pembuatan komponen dashboard.
- Tailwind + DaisyUI mudah dibuat clean, modern, dan konsisten.
- Build frontend bisa di-embed ke binary Go.

## 2.4 Port Decision

Default port:

```text
AtlasBridge API  : http://127.0.0.1:20127/v1
AtlasBridge UI   : http://127.0.0.1:20127/admin
Smart Health Check  : http://127.0.0.1:20127/health
9Router API         : http://127.0.0.1:20128/v1
9Router Dashboard   : http://127.0.0.1:20128/dashboard
```

Alasan memilih `20127`:

- Tidak memakai port umum developer seperti `3000`, `5173`, `8000`, `8080`, dan `8888`.
- Mudah diingat karena tepat sebelum 9Router `20128`.
- Tetap port non-privileged, sehingga tidak butuh permission khusus seperti port di bawah `1024`.
- Memisahkan AtlasBridge dari 9Router dengan jelas.

## 2.5 UI and API Same Port Decision

AtlasBridge menggunakan satu port untuk API dan Web UI.

```text
/v1/*       → OpenAI-compatible API
/admin/*    → Web UI
/admin/api  → Admin API
/health     → Health check
```

Alasan:

- Setup lebih sederhana.
- Tidak perlu dua server.
- Tidak perlu CORS kompleks untuk local UI.
- Cocok untuk single binary.
- Mudah dipahami user.

## 2.6 Tray Icon Decision

AtlasBridge harus memiliki **system tray icon** saat aktif.

Behavior:

```text
Aplikasi aktif
↓
Icon muncul di system tray
↓
Klik kiri membuka Web UI
↓
Klik kanan menampilkan quick menu
```

Tray menu minimum:

```text
AtlasBridge
Status: Running / Stopped
Open Dashboard
Start Proxy
Stop Proxy
Restart Proxy
Always On: ON/OFF
Run at Startup: ON/OFF
Open Logs
Quit
```

Catatan penting:

Untuk mendukung tray icon, AtlasBridge sebaiknya berjalan sebagai **user-level background application**, bukan Windows Service murni. Service murni biasanya tidak cocok untuk UI/tray interaction. Untuk MVP personal developer, user-level background app lebih sesuai.

---

## 3. System Architecture

## 3.1 High-Level Architecture

```text
OpenCode / Cursor / Cline / Continue
        ↓
OpenAI-Compatible Request
        ↓
AtlasBridge : 127.0.0.1:20127
        ↓
Request Analyzer
        ↓
Task Classifier
        ↓
Routing Policy Engine
        ↓
Route Profile Selector
        ↓
9Router Forwarder
        ↓
9Router : 127.0.0.1:20128
        ↓
Multiple AI Providers
```

## 3.2 Local Control Architecture

```text
User
 ↓
System Tray Icon
 ↓
AtlasBridge Runtime Controller
 ├── Start Proxy
 ├── Stop Proxy
 ├── Restart Proxy
 ├── Open Dashboard
 ├── Toggle Auto-start
 └── Quit App

Browser
 ↓
http://127.0.0.1:20127/admin
 ↓
Vue + DaisyUI Web UI
 ↓
Admin API
 ↓
Config Manager / Runtime Controller
```

## 3.3 Component Diagram

```text
AtlasBridge
├── App Bootstrap
├── Runtime Manager
├── Tray Controller
├── Startup Manager
├── HTTP Server
│   ├── OpenAI-Compatible API
│   ├── Admin API
│   ├── Health API
│   └── Embedded Web UI
├── Proxy Core
│   ├── Request Analyzer
│   ├── Task Classifier
│   ├── Routing Policy Engine
│   ├── Route Profile Selector
│   ├── Request Transformer
│   └── 9Router Forwarder
├── Config System
│   ├── Config Loader
│   ├── Config Validator
│   ├── Config Writer
│   └── Config Watcher
├── Observability
│   ├── Metadata Logger
│   ├── Metrics Collector
│   ├── Diagnostics Exporter
│   └── Privacy Redactor
└── Storage
    ├── config.yaml
    ├── routes.yaml
    ├── profiles.yaml
    └── optional atlasbridge.db
```

---

## 4. Runtime Modes

AtlasBridge harus mendukung tiga runtime mode.

## 4.1 Always On Mode

Proxy otomatis aktif saat user login atau komputer menyala.

Behavior:

- Tray icon muncul.
- Proxy engine otomatis start.
- Web UI tersedia.
- Endpoint `/v1` aktif.
- Auto-start enabled.

Use case:

Developer yang menggunakan AI coding assistant setiap hari.

## 4.2 Manual Mode

Aplikasi tidak auto-start. User menjalankan aplikasi secara manual.

Behavior:

- Tray icon muncul setelah user menjalankan app.
- Proxy bisa start otomatis saat app dibuka atau menunggu user menekan Start.
- Auto-start disabled.

Use case:

Developer yang hanya sesekali memakai AtlasBridge.

## 4.3 Disabled Mode

Proxy engine tidak menerima request, meskipun aplikasi/tray bisa tetap terbuka untuk konfigurasi.

Behavior:

- Tray icon menunjukkan status inactive.
- Web UI tetap bisa dibuka.
- Endpoint `/v1` mengembalikan error yang jelas atau tidak aktif sesuai konfigurasi.
- User bisa mengaktifkan kembali dari tray atau Web UI.

Use case:

User ingin mematikan sementara proxy tanpa uninstall.

---

## 5. Port and Networking Plan

## 5.1 Default Port Mapping

| Component | Default Port | URL |
|---|---:|---|
| AtlasBridge API | `20127` | `http://127.0.0.1:20127/v1` |
| AtlasBridge Web UI | `20127` | `http://127.0.0.1:20127/admin` |
| AtlasBridge Admin API | `20127` | `http://127.0.0.1:20127/admin/api` |
| AtlasBridge Health | `20127` | `http://127.0.0.1:20127/health` |
| 9Router API | `20128` | `http://127.0.0.1:20128/v1` |
| 9Router Dashboard | `20128` | `http://127.0.0.1:20128/dashboard` |

## 5.2 Avoided Ports

AtlasBridge tidak menggunakan port berikut sebagai default:

```text
3000   → common Node/Next/React dev server
5173   → common Vite dev server
8000   → common Python/dev server
8080   → common Java/Go/proxy server
8888   → common notebook/dev tooling
5000   → common Flask/dev server
5432   → PostgreSQL
6379   → Redis
27017  → MongoDB
```

## 5.3 Port Conflict Behavior

Jika port `20127` sudah dipakai:

1. Aplikasi tidak boleh silent switch tanpa memberi tahu user.
2. Tray icon menunjukkan status error.
3. Web UI, jika masih bisa terbuka melalui fallback admin port, menampilkan error.
4. CLI/log menampilkan pesan yang jelas.
5. User bisa memilih port baru melalui config atau startup prompt.
6. Sistem menyarankan alternatif seperti `20126`, `20129`, atau custom port.

## 5.4 Binding Policy

Default binding:

```text
127.0.0.1 only
```

Tidak boleh bind ke `0.0.0.0` secara default.

Alasan:

- Mengurangi risiko akses dari jaringan luar.
- Cocok untuk aplikasi local developer.
- Lebih aman untuk Web UI dan API key.

Opsi advanced:

```text
allow_lan_access: false
```

Jika user mengaktifkan LAN access, UI harus memberi peringatan keamanan.

---

## 6. OpenAI-Compatible API Plan

## 6.1 MVP Endpoints

| Endpoint | Method | Purpose |
|---|---|---|
| `/v1/chat/completions` | POST | Main OpenAI-compatible chat completion proxy |
| `/v1/models` | GET | Menampilkan smart aliases dan/atau passthrough models |
| `/health` | GET | Health check |
| `/admin` | GET | Web UI |
| `/admin/api/status` | GET | Status runtime |
| `/admin/api/config` | GET/PUT | Read/update config |
| `/admin/api/routes` | GET/PUT | Read/update task-to-route mapping |
| `/admin/api/profiles` | GET/PUT | Read/update route profiles |
| `/admin/api/runtime/start` | POST | Start proxy engine |
| `/admin/api/runtime/stop` | POST | Stop proxy engine |
| `/admin/api/runtime/restart` | POST | Restart proxy engine |
| `/admin/api/startup` | GET/PUT | Auto-start settings |
| `/admin/api/logs` | GET | Read metadata logs |
| `/admin/api/diagnostics/export` | POST | Export diagnostics |

## 6.2 Request Handling Flow

```text
Receive request
↓
Assign request ID
↓
Validate OpenAI-compatible body
↓
Read model field
↓
Detect explicit override
↓
Analyze messages
↓
Classify task
↓
Resolve route policy
↓
Select route profile
↓
Transform model/header/metadata for 9Router
↓
Forward to 9Router
↓
Stream or return response
↓
Write metadata log
```

## 6.3 Streaming Plan

Streaming harus menjadi first-class requirement.

Rules:

- Jangan buffer seluruh response.
- Forward chunk dari 9Router ke client secepat mungkin.
- Preserve Server-Sent Events format.
- Propagate client disconnect ke downstream.
- Timeout harus jelas.
- Error saat streaming harus diteruskan semampunya tanpa merusak client.

## 6.4 Non-Streaming Plan

Non-streaming request diteruskan ke 9Router dan response dikembalikan sebagai JSON utuh.

Rules:

- Preserve status code jika memungkinkan.
- Preserve OpenAI-compatible response structure.
- Jangan mengubah content response.
- Tambahkan internal metadata hanya di log, bukan response utama, kecuali debug mode aktif.

---

## 7. Routing and Classification Plan

## 7.1 Route Profile Concept

AtlasBridge memilih **route profile**, bukan provider final.

Contoh:

```text
route.design
route.backend
route.frontend
route.debugging
route.refactoring
route.documentation
route.reasoning
route.low_cost
route.long_context
route.security
route.default
```

9Router tetap menjalankan provider-level routing.

## 7.2 Task Categories

MVP task categories:

| Task Category | Route Profile |
|---|---|
| Design / UI / UX | `route.design` |
| Backend Engineering | `route.backend` |
| Frontend Engineering | `route.frontend` |
| Fullstack Engineering | `route.fullstack` |
| Debugging | `route.debugging` |
| Refactoring | `route.refactoring` |
| Testing | `route.testing` |
| Documentation | `route.documentation` |
| Architecture Planning | `route.reasoning` |
| Security Review | `route.security` |
| Simple / Lightweight Task | `route.low_cost` |
| Long Context Analysis | `route.long_context` |
| Unknown | `route.default` |

## 7.3 User Custom Routing Examples

User bisa mengatur routing seperti:

```text
Design      → DeepSeek route via 9Router
Backend     → Gemini route via 9Router
Frontend    → Claude route via 9Router
Debugging   → Reasoning/debug route via 9Router
Docs        → Low-cost writing route via 9Router
```

Namun di AtlasBridge, konfigurasi sebaiknya tetap memakai abstraksi:

```text
Design      → route.design
Backend     → route.backend
Frontend    → route.frontend
```

Lalu `route.*` diterjemahkan menjadi downstream alias untuk 9Router.

## 7.4 Routing Decision Precedence

Urutan prioritas:

1. Explicit route override.
2. Smart model alias override.
3. User-defined task-to-route mapping.
4. Project-specific route policy.
5. Task classifier result.
6. Complexity/context signal.
7. Default route.
8. Safe passthrough.

## 7.5 Classification Strategy MVP

MVP menggunakan rule-based dan heuristic classifier.

Input signal:

- Keyword: debug, fix, error, stack trace, refactor, test, docs, README, architecture, backend, frontend, UI, design.
- Presence of code block.
- File path patterns: `.go`, `.ts`, `.vue`, `.tsx`, `.py`, `.java`, etc.
- Framework keywords: Vue, React, Laravel, Express, Gin, Echo, Next, Nest, etc.
- Error pattern: stack trace, exception, panic, traceback.
- Prompt length.
- Tool call presence.
- System prompt context.
- Model alias used by client.

Output:

```text
task_type
confidence
complexity
route_candidate
routing_reason
```

## 7.6 Safe Passthrough

Jika classifier gagal:

- Request tetap diteruskan ke 9Router.
- Gunakan `route.default`.
- Log `classification_status: failed`.
- Jangan blokir user kecuali request invalid secara format.

---

## 8. Configuration Plan

## 8.1 Storage Format MVP

MVP menggunakan file lokal:

```text
config.yaml
routes.yaml
profiles.yaml
```

Lokasi rekomendasi:

Windows:

```text
%APPDATA%/AtlasBridge/
```

macOS:

```text
~/Library/Application Support/AtlasBridge/
```

Linux:

```text
~/.config/AtlasBridge/
```

## 8.2 Main Config Model

Konsep konfigurasi:

```yaml
app:
  name: AtlasBridge
  mode: always_on
  first_run_completed: false

server:
  host: 127.0.0.1
  port: 20127
  admin_path: /admin
  api_base_path: /v1

downstream:
  type: 9router
  base_url: http://127.0.0.1:20128/v1
  timeout_seconds: 120

security:
  admin_auth_enabled: true
  admin_token_hash: null
  bind_localhost_only: true
  allow_lan_access: false

startup:
  run_at_login: false
  start_proxy_on_app_launch: true
  restart_after_crash: true

routing:
  auto_routing: true
  default_route: route.default
  low_confidence_route: route.default
  explicit_override_enabled: true
  confidence_threshold: 0.55

logging:
  level: info
  privacy_mode: standard
  prompt_logging_enabled: false
  metadata_logging_enabled: true
  retention_days: 7
```

## 8.3 Task Routes Model

```yaml
task_routes:
  design: route.design
  backend: route.backend
  frontend: route.frontend
  fullstack: route.fullstack
  debugging: route.debugging
  refactoring: route.refactoring
  testing: route.testing
  documentation: route.documentation
  architecture_design: route.reasoning
  security_review: route.security
  lightweight_task: route.low_cost
  long_context_analysis: route.long_context
  unknown: route.default
```

## 8.4 Route Profiles Model

```yaml
route_profiles:
  route.design:
    label: Design
    description: UI, UX, design system, layout, visual thinking
    downstream_alias: combo.design
    priority: quality
    enabled: true

  route.backend:
    label: Backend
    description: API, database, service, architecture, server logic
    downstream_alias: combo.backend
    priority: balanced
    enabled: true

  route.frontend:
    label: Frontend
    description: Vue, React, UI component, client-side implementation
    downstream_alias: combo.frontend
    priority: balanced
    enabled: true

  route.default:
    label: Default
    description: Default fallback route
    downstream_alias: combo.default
    priority: balanced
    enabled: true
```

## 8.5 Config Validation

Saat startup dan setiap config update, sistem harus validasi:

- Port valid.
- Host valid.
- Downstream URL valid.
- Route profile yang dirujuk task route tersedia.
- Default route tersedia.
- Disabled route tidak dipakai sebagai default.
- Admin token valid jika auth enabled.
- Logging privacy mode valid.
- Startup mode valid.

Jika config invalid:

- Jangan crash tanpa pesan.
- Tampilkan error di tray/Web UI/log.
- Gunakan last-known-good config jika tersedia.
- Sediakan reset config.

---

## 9. Web UI Technical Plan

## 9.1 Frontend Stack

```text
Vue 3
TypeScript
Vite
Tailwind CSS
DaisyUI
Pinia
Vue Router
```

Pinia digunakan untuk state management ringan.

Vue Router digunakan untuk halaman admin.

## 9.2 Web UI Pages

| Page | Path | Purpose |
|---|---|---|
| Dashboard | `/admin` | Status ringkas proxy |
| Setup Wizard | `/admin/setup` | First-time setup |
| Routing Settings | `/admin/routing` | Task-to-route mapping |
| Route Profiles | `/admin/profiles` | Kelola route profiles |
| Runtime | `/admin/runtime` | Start/stop/restart |
| Startup | `/admin/startup` | Auto-start settings |
| 9Router | `/admin/downstream` | Konfigurasi endpoint 9Router |
| Logs | `/admin/logs` | Metadata logs |
| Privacy | `/admin/privacy` | Logging/privacy/security |
| Advanced | `/admin/advanced` | Port, debug, import/export |

## 9.3 Dashboard Content

Dashboard menampilkan:

- Proxy status: Running / Stopped / Error.
- Current endpoint: `http://127.0.0.1:20127/v1`.
- Admin UI URL.
- 9Router endpoint: `http://127.0.0.1:20128/v1`.
- 9Router connection status.
- Auto-routing ON/OFF.
- Run at startup ON/OFF.
- Total request today.
- Most used task type.
- Last error jika ada.

## 9.4 Routing Settings UI

Komponen:

- Table task category.
- Dropdown route profile.
- Toggle enable/disable route per task.
- Default route selector.
- Low confidence route selector.
- Save button.
- Reset to default button.
- Dry-run tester.

Dry-run tester:

User memasukkan contoh prompt, lalu UI menampilkan:

```text
Detected Task       : backend
Confidence          : 0.78
Selected Route      : route.backend
Downstream Alias    : combo.backend
Reason              : Detected API/database/backend keywords
```

## 9.5 Route Profiles UI

Komponen:

- Profile list.
- Add profile.
- Edit profile.
- Disable profile.
- Downstream alias input.
- Priority selector: speed, cost, quality, balanced.
- Description.
- Test route button.

## 9.6 Startup UI

Komponen:

- Run at login toggle.
- Start proxy when app launches toggle.
- Always On mode.
- Manual mode.
- Disabled mode.
- Restart after crash toggle.
- Current startup registration status.

## 9.7 DaisyUI Design Direction

Gunakan style dashboard modern:

- Sidebar navigation.
- Top status bar.
- Cards untuk status.
- Badge untuk Running/Stopped/Error.
- Toggle untuk setting.
- Modal untuk confirm destructive action.
- Alert untuk port conflict atau downstream disconnected.
- Theme default: light dengan optional dark mode.

Komponen DaisyUI yang cocok:

```text
navbar
drawer
menu
card
badge
alert
modal
toggle
select
tabs
table
toast
steps
dropdown
```

---

## 10. Tray App Technical Plan

## 10.1 Tray Requirements

Tray icon muncul saat AtlasBridge app aktif.

Status icon:

| State | Tray Behavior |
|---|---|
| Running | Icon normal/active |
| Stopped | Icon inactive/grey |
| Error | Icon warning |
| Starting | Tooltip shows starting |
| Port conflict | Tooltip shows port error |
| Downstream disconnected | Tooltip shows warning |

## 10.2 Tray Menu

Minimum menu:

```text
AtlasBridge
Status: Running
Endpoint: 127.0.0.1:20127
Open Dashboard
Start Proxy
Stop Proxy
Restart Proxy
Run at Startup: ON/OFF
Always On Mode: ON/OFF
Open 9Router Dashboard
Open Logs Folder
Copy API Endpoint
Quit
```

## 10.3 Tray Click Behavior

| Action | Behavior |
|---|---|
| Left click | Open dashboard |
| Right click | Open context menu |
| Double click | Open dashboard |
| Quit | Stop proxy and exit app |
| Stop Proxy | Stop accepting `/v1` requests |
| Start Proxy | Start accepting `/v1` requests |
| Restart Proxy | Graceful restart proxy engine |

## 10.4 Tray Implementation Options

### Option A: Go + systray

Best for MVP if we want lightweight tray app.

Pros:

- Simple.
- Go-native.
- Cross-platform support.
- Good enough for tray menu.

Cons:

- May require cgo.
- Linux tray behavior may vary by desktop environment.
- UI remains browser-based.

### Option B: Wails

Best for post-MVP if we want desktop shell with Vue UI.

Pros:

- Go + Web UI integration.
- Native window.
- System tray APIs.
- Better desktop app experience.

Cons:

- More scope.
- Packaging more involved.
- Larger architecture surface.

Recommendation:

```text
MVP      : Go + systray + browser-based Web UI
Post-MVP : Evaluate Wails if desktop shell becomes important
```

---

## 11. Auto-Start Technical Plan

## 11.1 Auto-Start Strategy

Auto-start should run the **tray app**, not only the proxy engine.

Reason:

- User needs tray icon.
- User needs quick control.
- User can see status.
- User can open Web UI quickly.

## 11.2 Platform Strategy

### Windows

Recommended startup approach:

- User-level startup registration.
- Use Windows Run key or Startup folder shortcut.
- Avoid Windows Service for MVP because tray UI needs user session.

### macOS

Recommended startup approach:

- LaunchAgent for user login.
- App opens in background.
- Tray/menu bar icon appears.

### Linux

Recommended startup approach:

- XDG autostart `.desktop` entry.
- App starts when user logs in.
- Tray support depends on desktop environment.

## 11.3 Auto-Start Settings

Config:

```yaml
startup:
  run_at_login: true
  start_proxy_on_app_launch: true
  restart_after_crash: true
```

UI states:

```text
Run at Startup: ON / OFF
Start Proxy on App Launch: ON / OFF
Restart After Crash: ON / OFF
```

## 11.4 Startup Failure Handling

If auto-start fails:

- Show warning in dashboard.
- Show tray tooltip warning.
- Log failure reason.
- Provide “Repair Startup Registration” button.
- Provide manual instruction if OS blocks registration.

---

## 12. Security Plan

## 12.1 Localhost-Only by Default

AtlasBridge binds to:

```text
127.0.0.1
```

Not:

```text
0.0.0.0
```

Unless user explicitly enables LAN access.

## 12.2 Admin UI Protection

Admin UI controls routing, port, downstream URL, logs, and runtime. Therefore:

- Admin API should require local admin token if auth enabled.
- First-run setup generates token/password.
- Token stored securely as hash.
- Token never shown fully after creation.
- Session can be local browser cookie or bearer token.

## 12.3 API Key Handling

Rules:

- Do not log Authorization header.
- Do not log API keys.
- Do not expose downstream key in UI.
- Mask secrets in config UI.
- Redact common secret patterns in logs.

## 12.4 Prompt Privacy

Default:

```text
prompt_logging_enabled: false
```

Metadata-only logs:

- request ID
- task type
- selected route
- latency
- status code
- error class
- confidence

No full source code or prompt stored by default.

## 12.5 Downstream Allowlist

MVP only allows downstream endpoint from config.

Client cannot override downstream URL per request.

---

## 13. Observability Plan

## 13.1 Logs

Default metadata log fields:

```text
timestamp
request_id
client_name
requested_model
task_type
classification_confidence
selected_route
downstream_alias
status_code
proxy_latency_ms
downstream_latency_ms
streaming
error_class
```

## 13.2 Metrics

MVP metrics:

- Total requests.
- Requests by task type.
- Requests by route.
- Safe passthrough count.
- Manual override count.
- Error count.
- Average proxy latency.
- Average downstream latency.
- Streaming vs non-streaming count.
- Port conflict count.
- 9Router disconnected count.

## 13.3 Diagnostics Export

Export should include:

- App version.
- OS info.
- Config summary with secrets redacted.
- Recent metadata logs.
- Route profiles.
- Task routes.
- Runtime state.
- Downstream health status.

Never include:

- Full prompts.
- Source code.
- API keys.
- Raw Authorization headers.

---

## 14. Data and Storage Plan

## 14.1 MVP Storage

Use local files:

```text
config.yaml
routes.yaml
profiles.yaml
logs.jsonl
```

## 14.2 Post-MVP Storage

Add SQLite for:

- Request metadata history.
- Metrics summary.
- Feedback.
- Evaluation samples if user opts in.
- Config version history.

## 14.3 Config Backup

Maintain:

```text
config.yaml
config.last-good.yaml
config.backup.yaml
```

On invalid update:

- Reject update.
- Preserve previous config.
- Show validation error.

---

## 15. Build and Packaging Plan

## 15.1 Build Model

Frontend build:

```text
Vue/Vite source
↓
Static dist
↓
Embedded in Go binary
```

Backend build:

```text
Go source
↓
Single executable
↓
Runs proxy + Web UI + tray
```

## 15.2 Distribution Targets

MVP target:

```text
Windows x64
```

Next targets:

```text
macOS arm64/x64
Linux x64
```

## 15.3 App Names

Binary names:

```text
atlasbridge.exe
atlasbridge
```

Config folder:

```text
AtlasBridge
```

Tray name:

```text
AtlasBridge
```

## 15.4 Installer

MVP can start with portable binary.

Post-MVP installer:

- Windows: MSI or NSIS.
- macOS: DMG.
- Linux: AppImage/deb/rpm.

Installer should offer:

- Install location.
- Create desktop shortcut.
- Run at startup.
- Launch after install.

---

## 16. Development Environment

## 16.1 Required Tools

```text
Go
Node.js
pnpm or npm
Vite
Git
```

## 16.2 Local Development Ports

Development mode may use:

```text
Go API        : 20127
Vite dev UI   : 5173
9Router       : 20128
```

Production mode uses one port:

```text
AtlasBridge : 20127
```

## 16.3 Local Dev Flow

```text
Run 9Router on 20128
Run Go backend on 20127
Run Vue dev server on 5173
Proxy admin UI API calls to 20127
Test OpenAI-compatible request via 20127/v1
```

## 16.4 Production Flow

```text
Build Vue app
Embed dist into Go
Build Go binary
Run binary
Open http://127.0.0.1:20127/admin
```

---

## 17. Suggested Repository Structure

```text
atlasbridge/
├── cmd/
│   └── atlasbridge/
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
│   ├── technical-plan.md
│   ├── prd.md
│   ├── setup.md
│   └── routing-policy.md
├── testdata/
│   ├── classification/
│   └── openai-compatible/
├── scripts/
│   ├── build.sh
│   └── build.ps1
└── README.md
```

---

## 18. Testing Plan

## 18.1 Unit Tests

Target modules:

- Request analyzer.
- Task classifier.
- Routing policy engine.
- Config validator.
- Secret redactor.
- Route profile selector.
- Admin API handlers.

## 18.2 Integration Tests

Scenarios:

- Forwarding to mock 9Router.
- Streaming passthrough.
- Non-streaming passthrough.
- Config update.
- Route mapping update.
- Port conflict.
- Downstream unavailable.
- Classifier failure.
- Safe passthrough.

## 18.3 Compatibility Tests

Test with:

- OpenCode.
- Cursor.
- Cline.
- Continue.
- Generic OpenAI SDK client.
- curl/postman-like request.

## 18.4 UI Tests

Minimum:

- Dashboard loads.
- Routing table save works.
- Route profile save works.
- Startup toggle persists.
- Runtime start/stop works.
- Config validation error appears.
- Logs page hides sensitive data.

## 18.5 Manual System Tests

Windows MVP:

- App launches.
- Tray icon appears.
- Click tray opens dashboard.
- Run at startup works after reboot/login.
- Proxy starts on login.
- 9Router downstream connects.
- OpenCode/Cursor can use `http://127.0.0.1:20127/v1`.
- Stop proxy blocks/pauses request clearly.
- Quit removes tray icon.

---

## 19. MVP Scope

MVP includes:

1. Go core proxy.
2. OpenAI-compatible `/v1/chat/completions`.
3. Streaming passthrough.
4. Non-streaming passthrough.
5. 9Router default downstream at `127.0.0.1:20128/v1`.
6. AtlasBridge default port `20127`.
7. Rule-based task classifier.
8. Task-to-route mapping.
9. Route profile selector.
10. Vue TS + DaisyUI Web UI.
11. Dashboard.
12. Routing settings page.
13. Route profiles page.
14. Startup settings page.
15. Runtime start/stop/restart.
16. Tray icon with basic menu.
17. Run at login.
18. Metadata logging.
19. Local config persistence.
20. Health check.
21. Privacy-safe default logging.

MVP excludes:

1. Direct provider integration.
2. Provider failover.
3. Provider account rotation.
4. Provider fallback.
5. Provider credential manager.
6. Full desktop shell.
7. Complex team mode.
8. RBAC.
9. Cloud sync.
10. Prompt storage by default.
11. ML-based classifier.
12. Repository indexing.

---

## 20. Implementation Phases

## Phase 0: Technical Foundation

Goal:

Finalize architecture and setup repository.

Deliverables:

- Repository structure.
- Go app bootstrap.
- Vue/Vite app bootstrap.
- Config format draft.
- Port decision finalized.
- 9Router downstream default set to `20128`.

## Phase 1: Core Proxy MVP

Goal:

Make AtlasBridge work as OpenAI-compatible proxy.

Deliverables:

- `/v1/chat/completions`.
- Forward to 9Router.
- Streaming passthrough.
- Non-streaming passthrough.
- Request ID.
- Health check.
- Basic logs.

## Phase 2: Routing Intelligence MVP

Goal:

Add task-aware routing.

Deliverables:

- Request analyzer.
- Rule-based classifier.
- Task-to-route mapping.
- Route profile selector.
- Manual alias override.
- Safe passthrough.

## Phase 3: Local Web UI

Goal:

Make configuration easy through browser UI.

Deliverables:

- Vue TS + DaisyUI dashboard.
- Routing settings.
- Route profiles.
- Runtime page.
- Startup page.
- Privacy/logs page.
- Admin API.

## Phase 4: Tray and Auto-Start

Goal:

Make app behave like 9Router-style local background app.

Deliverables:

- Tray icon.
- Tray menu.
- Open dashboard from tray.
- Start/stop/restart from tray.
- Run at login.
- Status indicators.
- Port conflict notification.

## Phase 5: Packaging and Installer

Goal:

Make app easy to install and use.

Deliverables:

- Windows binary.
- Portable release.
- Optional installer.
- Documentation.
- Quick start guide.
- Troubleshooting guide.

## Phase 6: Quality and Compatibility

Goal:

Make it stable with developer tools.

Deliverables:

- Test matrix.
- OpenCode guide.
- Cursor guide.
- Cline guide.
- Continue guide.
- Streaming compatibility tests.
- Log redaction tests.

---

## 21. Technical Risks and Mitigations

| Risk | Impact | Mitigation |
|---|---|---|
| Port `20127` already used | App cannot start normally | Clear error, allow custom port, suggest alternatives |
| 9Router not running | Requests fail | Dashboard health status, clear error, retry health check |
| Streaming broken | AI coding tool unusable | Minimal transformation, streaming integration tests |
| Tray icon inconsistent on Linux | UX issue | Windows-first MVP, document Linux limitations |
| Auto-start blocked by OS | App not active after restart | Repair startup button and manual instruction |
| Classifier inaccurate | Wrong route selected | Manual override, confidence threshold, safe fallback |
| Web UI exposed to network | Security risk | Bind localhost by default |
| Prompt logged accidentally | Privacy risk | Metadata-only logs, redaction tests |
| Boundary creep with 9Router | Architecture overlap | Keep provider execution out of AtlasBridge |
| Config invalid | App unstable | Validation, last-known-good config |

---

## 22. Security Acceptance Criteria

MVP security is acceptable if:

- App binds to `127.0.0.1` by default.
- Admin UI is not publicly exposed.
- Prompt logging is disabled by default.
- Authorization header is never logged.
- Downstream URL cannot be overridden by request body.
- Config UI masks secrets.
- Logs are metadata-only by default.
- LAN access requires explicit user action.
- Diagnostics export redacts sensitive data.

---

## 23. Performance Targets

Initial targets:

| Metric | Target |
|---|---|
| Proxy routing overhead | Low enough to not disturb coding workflow |
| Classification overhead | Lightweight, rule-based |
| Streaming buffering | Minimal |
| Startup time | Fast for local app |
| Memory usage | Suitable for always-on developer machine |
| UI load time | Fast on localhost |

Suggested internal targets:

```text
Routing decision overhead: < 20 ms for simple request
Admin UI load: < 1 second locally
Health check response: < 100 ms
Streaming first chunk overhead: minimal, no full buffering
```

---

## 24. Documentation Plan

Required docs:

- Quick Start.
- Install Guide.
- Connect to OpenCode.
- Connect to Cursor.
- Connect to Cline.
- Connect to Continue.
- Configure 9Router.
- Routing Settings Guide.
- Route Profiles Guide.
- Tray Icon Guide.
- Auto-start Guide.
- Security and Privacy Guide.
- Troubleshooting Port Conflict.
- Troubleshooting 9Router Connection.
- Troubleshooting Streaming.

---

## 25. Final Technical Recommendation

Final architecture recommendation:

```text
Build AtlasBridge as a Go-based local background application with:
- OpenAI-compatible proxy endpoint,
- default port 20127,
- downstream 9Router at 20128,
- Vue 3 + TypeScript + Vite Web UI,
- DaisyUI for dashboard components,
- system tray icon when active,
- user-level auto-start at login,
- local YAML/JSON config,
- privacy-safe metadata logging,
- no provider-level execution.
```

This gives the product the same local-app feel as 9Router while keeping the architecture clean:

```text
AtlasBridge chooses the route.
9Router executes provider routing.
```

---

## 26. Reference Notes

Key external facts used for this technical plan:

- 9Router documentation/README shows default dashboard at `http://localhost:20128` and OpenAI-compatible API at `http://localhost:20128/v1`.
- Go is positioned around simple, secure, scalable systems with built-in concurrency and a robust standard library.
- chi is a lightweight Go router compatible with `net/http`.
- DaisyUI is a Tailwind CSS component library that provides semantic component class names.
- systray is a cross-platform Go library for notification area icons and menus.
- Wails provides system tray APIs for cross-platform Go desktop applications.
