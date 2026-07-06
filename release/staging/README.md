# Smart AI Proxy

OpenAI-compatible intelligent routing proxy with local Web UI control panel.

## Overview

Smart AI Proxy sits between your AI coding assistant (OpenCode, Cursor, Cline, Continue) and 9Router. It analyzes requests, classifies tasks, selects the best route profile, and forwards requests to 9Router for provider-level execution.

- **One endpoint**: `http://127.0.0.1:20127/v1`
- **One model alias**: `smart-auto`
- **One config panel**: `http://127.0.0.1:20127/admin`

## Architecture

```
AI Coding Assistant → Smart AI Proxy → 9Router → AI Providers
```

Smart AI Proxy is a **decision layer only**. It never touches provider failover, load balancing, account rotation, fallback, rate limits, or credentials. That is 9Router's responsibility.

## Default Ports

| Service | Port | URL |
|---|---|---|
| Smart AI Proxy API | `20127` | `http://127.0.0.1:20127/v1` |
| Smart AI Proxy Admin | `20127` | `http://127.0.0.1:20127/admin` |
| 9Router (downstream) | `20128` | `http://127.0.0.1:20128/v1` |

## Tech Stack

- **Backend**: Go, `net/http`, `chi` router
- **Frontend**: Vue 3, TypeScript, Vite, Tailwind CSS, DaisyUI
- **Config**: Local YAML/JSON
- **Target**: Windows x64 first, macOS/Linux later

## Quick Start (Development)

### Prerequisites

- Go 1.21+ (developed with Go 1.25.5)
- Node.js 18+
- 9Router running on port `20128`

### 1. Run Go backend

```bash
go run ./cmd/smart-ai-proxy
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
# {"status":"ok","service":"smart-ai-proxy"}
```

## Build

### Go backend only

```bash
go build -o smart-ai-proxy.exe ./cmd/smart-ai-proxy
```

### Frontend

```bash
cd web
npm install
npm run build
```

## Project Structure

```
smart-ai-proxy/
├── cmd/smart-ai-proxy/      # Application entry point
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
│   ├── logging/             # Metadata logging
│   ├── observability/       # Metrics and diagnostics
│   ├── security/            # Secret redaction, admin auth
│   ├── startup/             # Auto-start registration (Phase 4)
│   ├── tray/                # System tray icon (Phase 4)
│   └── runtime/             # Start/stop/restart controls
├── web/                     # Vue 3 admin UI
├── configs/                 # Example configuration files
├── docs/                    # Documentation
├── scripts/                 # Build and dev scripts
├── testdata/                # Test fixtures
└── packaging/               # Distribution packaging
```

## Configuration

Config is stored locally per platform:

- **Windows**: `%APPDATA%/SmartAIProxy/`
- **macOS**: `~/Library/Application Support/SmartAIProxy/`
- **Linux**: `~/.config/smart-ai-proxy/`

See `configs/` for example configuration files.

## License

TBD
