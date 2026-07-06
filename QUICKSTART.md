# AtlasBridge - Quick Start

## Installation

### Option 1: Portable (No Install)
1. Download the portable ZIP from releases
2. Extract to any folder
3. Run `atlasbridge.exe`

### Option 2: npm Wrapper
```bash
npm install -g atlasbridge
atlasbridge start
```

## First Run

1. Ensure 9Router is running on port `20128`
2. Run AtlasBridge:
   - Portable: `./atlasbridge.exe`
   - npm: `atlasbridge start`
3. Tray icon appears in system tray
4. Open dashboard: http://127.0.0.1:20127/admin

## Configuration

- Config file: `%APPDATA%/SmartAIProxy/config.yaml`
- Logs: `%APPDATA%/SmartAIProxy/logs/`
- Default API endpoint: `http://127.0.0.1:20127/v1`
- Default admin: `http://127.0.0.1:20127/admin`

## CLI Commands

| Command | Description |
|---------|-------------|
| `start` | Start the proxy |
| `stop` | Stop the proxy |
| `status` | Show current status |
| `open` | Open dashboard in browser |
| `tray` | Show tray icon |
| `version` | Show version |

## Troubleshooting

- Port 20127 in use: Stop other instances or change port in config
- 9Router not reachable: Ensure 9Router is running on port 20128
- Dashboard not loading: Check firewall settings

## Support

- GitHub: https://github.com/smart-ai-proxy/smart-ai-proxy
- Issues: https://github.com/smart-ai-proxy/smart-ai-proxy/issues