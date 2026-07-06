# Smart AI Proxy Setup Guide

## Prerequisites

- Windows x64
- 9Router running on port 20128
- ~50MB disk space

## Installation Options

### Option 1: Portable (Recommended)
1. Download `smart-ai-proxy-v0.1.0-windows-x64-portable.zip`
2. Extract to any folder
3. Run `smart-ai-proxy.exe`

### Option 2: npm Wrapper
```bash
npm install -g smart-ai-proxy
smart-ai-proxy start
```

## First-Time Setup

1. **Ensure 9Router is running** on port 20128

2. **Start Smart AI Proxy**:
   - Portable: Run `smart-ai-proxy.exe`
   - npm: `smart-ai-proxy start`

3. **Tray icon appears** in system tray

4. **Open Web UI**: http://127.0.0.1:20127/admin

5. **Configure routing** (optional):
   - Go to Routing Settings
   - Adjust task-to-route mappings
   - Create route profiles

6. **Select startup mode**:
   - Always On: Proxy starts automatically
   - Manual: Start proxy when needed
   - Disabled: Proxy doesn't accept requests

## Verify Installation

1. **Health check**:
   ```bash
   curl http://127.0.0.1:20127/health
   ```

2. **Test API**:
   ```bash
   curl -X POST http://127.0.0.1:20127/v1/chat/completions \
     -H "Content-Type: application/json" \
     -d '{"model":"smart-auto","messages":[{"role":"user","content":"hello"}]}'
   ```

## Configuration

Config file location: `%APPDATA%/SmartAIProxy/config.yaml`

Key settings:
- `server.port`: API port (default: 20127)
- `downstream.base_url`: 9Router URL
- `startup.run_at_login`: Enable auto-start
- `routing.auto_routing`: Enable smart routing

## Uninstallation

### Portable
Simply delete the extracted folder

### npm
```bash
npm uninstall -g smart-ai-proxy
```