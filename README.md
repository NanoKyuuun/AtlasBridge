# AtlasBridge

AtlasBridge AI Proxy

OpenAI-compatible intelligent routing proxy with local Web UI control panel.

## Overview

AtlasBridge sits between your AI coding assistant (OpenCode, Cursor, Cline, Continue) and 9Router. It analyzes requests, classifies tasks, selects the best route profile, and forwards requests to 9Router for provider-level execution.

- **One endpoint**: `http://127.0.0.1:20127/v1`
- **One model alias**: `atlas-auto` (smart-auto also works)
- **One config panel**: `http://127.0.0.1:20127/admin`

## Architecture

```
AI Coding Assistant → AtlasBridge → 9Router → AI Providers
```

AtlasBridge is a **decision layer only**. It never touches provider failover, load balancing, account rotation, fallback, rate limits, or credentials. That is 9Router's responsibility.

## Default Ports

| Service | Port | URL |
|---|---|---|
| AtlasBridge API | `20127` | `http://127.0.0.1:20127/v1` |
| AtlasBridge Admin | `20127` | `http://127.0.0.1:20127/admin` |
| 9Router (downstream) | `20128` | `http://127.0.0.1:20128/v1` |

## Security Features (Enforced)

AtlasBridge enforces the following security measures by default:

| Feature | Status |
|---|---|
| Admin auth enabled by default | Enforced |
| Body size limits (16 MiB chat, 1 MiB admin, 8 MiB import) | Enforced |
| Concurrency bulkhead with weighted semaphore | Enforced |
| Immutable config snapshot with atomic swap | Enforced |
| Downstream URL validation (no credentials, no fragments, blocked IPs) | Enforced |
| Redirect policy (max 5 hops, revalidation on each hop) | Enforced |
| Origin guard on admin API | Enforced |
| Content-Type enforcement on state-changing requests | Enforced |
| Security headers (CSP, X-Frame-Options, Referrer-Policy) | Enforced |
| DNS rebinding protection (HostGuard) | Enforced |
| Error response sanitization (generic messages + correlation IDs) | Enforced |
| Response header allowlist for downstream | Enforced |
| Structured JSON logging with redaction | Enforced |
| Atomic config file writes with backup | Enforced |
| OS-level single-instance lock (flock) | Enforced |
| Graceful shutdown with timeout | Enforced |
| Constant-time token comparison | Enforced |
| Request ID validation (max 128 chars, safe charset) | Enforced |
| CI quality gate (tests, vet, race, lint, vuln scan, npm audit) | Enforced |

## Planned Features

| Feature | Status |
|---|---|
| TLS / reverse-proxy guidance for LAN | Planned |
| Data-plane API key for LAN mode | Planned |
| Per-client rate limiting | Planned |
| Circuit breaker for downstream connectivity | Planned |
| Prometheus/OpenTelemetry metrics | Planned |
| Threat model documentation | Planned |
| SBOM generation and artifact signing | Planned |

## Tech Stack

- **Backend**: Go 1.25, `net/http`, `chi` router
- **Frontend**: Vue 3, TypeScript, Vite, Tailwind CSS, DaisyUI
- **Config**: Local YAML/JSON
- **Target**: Windows x64 first, macOS/Linux later

## Quick Start (Development)

### Prerequisites

- Go 1.25+
- Node.js 20+
- 9Router running on port `20128`

### 1. Run Go backend

```bash
go run ./cmd/atlasbridge
```

### 2. Run frontend dev server

```bash
cd web
npm install
npm run dev
```

Frontend dev server runs on `http://localhost:5173` and proxies API calls to Go backend.

### 3. Verify

```bash
curl http://127.0.0.1:20127/health
# {"status":"ok","service":"atlasbridge"}
```

## Build

### Go backend only

```bash
go build -o atlasbridge.exe ./cmd/atlasbridge
```

### Frontend

```bash
cd web
npm install
npm run build
```

## CI Pipeline

Pull requests automatically run:

- `go vet ./...`
- `staticcheck ./...`
- `go test ./...`
- `go test -race ./...`
- `govulncheck ./...`
- `npm ci && vue-tsc -b && npm run build`
- `npm audit --audit-level=high`
- Secret scan (Gitleaks)
- Dependency review (on PRs)

All GitHub Actions are pinned to full commit SHAs for supply-chain security.

## Project Structure

```
atlasbridge/
├── cmd/atlasbridge/      # Application entry point
├── internal/
│   ├── app/                 # Application bootstrap
│   ├── server/              # HTTP server with chi router
│   ├── proxy/               # Core proxy engine (Phase 1)
│   ├── analyzer/            # Request analysis (Phase 2)
│   ├── classifier/          # Task classification (Phase 2)
│   ├── routing/             # Route profile selection (Phase 2)
│   ├── forwarder/           # 9Router forwarding (Phase 1)
│   ├── config/              # Configuration loading and validation
│   ├── storage/             # Local file persistence
│   ├── logging/             # Structured logging
│   ├── observability/       # Metrics and diagnostics
│   ├── security/            # Admin auth, token management
│   ├── redactor/            # Log redaction
│   ├── startup/             # Auto-start and single-instance lock
│   ├── tray/                # System tray icon
│   └── runtime/             # Start/stop/restart controls
├── web/                     # Vue 3 admin UI
├── configs/                 # Example configuration files
├── docs/                    # Documentation and audit reports
├── scripts/                 # Build and dev scripts
├── testdata/                # Test fixtures
├── packaging/               # Distribution packaging
└── npm-wrapper/             # npm CLI wrapper
```

## Configuration

Config is stored locally per platform:

- **Windows**: `%APPDATA%/AtlasBridge/`
- **macOS**: `~/Library/Application Support/AtlasBridge/`
- **Linux**: `~/.config/AtlasBridge/`

See `configs/` for example configuration files.

## Security Deployment Guide

### Localhost Mode (Default)

Default configuration is secure for local use:

- Binds to `127.0.0.1` only
- Admin auth enabled by default (bcrypt password)
- No prompt/API key logging
- Session tokens expire after 24 hours
- Login rate limiting with lockout

### LAN Mode

If you need to access AtlasBridge from other machines on your network:

1. **Enable LAN access** in Advanced Settings (requires admin auth)
2. **Set up TLS** — AtlasBridge runs plain HTTP; use a reverse proxy (nginx/Caddy) with TLS termination in front
3. **Enable admin auth** — mandatory when LAN is active
4. **Set strong password** (12+ characters)
5. **Restrict network** — use firewall to limit which IPs can connect

⚠️ **Never expose AtlasBridge to the internet.** It is designed as a local proxy only.

### Data Plane Authentication

When LAN access is enabled, the `/v1` data plane requires Bearer token authentication. The token is the same admin session token used for the Web UI.

## Trust Boundary

AtlasBridge operates as a **local proxy**. Key trust boundaries:

- **AtlasBridge → 9Router**: HTTP on loopback. AtlasBridge trusts 9Router as the execution backend. No auth on this link by default.
- **AI Coding Assistant → AtlasBridge**: HTTP on loopback. The data plane (`/v1`) has no auth on localhost by default.
- **Admin UI → AtlasBridge**: HTTP on loopback. Protected by admin Bearer token (enabled by default).
- **LAN access**: Not recommended without TLS. If enabled, admin auth is mandatory.

## License

MIT
