# AtlasBridge Troubleshooting Guide

## Common Issues

### Port 20127 Already in Use
**Symptom**: `port conflict on 127.0.0.1:20127`

**Solution**:
1. Stop any other AtlasBridge instances
2. Or change port in config.yaml: `server.port: 20128`

### 9Router Not Reachable
**Symptom**: `downstream service unavailable` error

**Solution**:
1. Ensure 9Router is running on port 20128
2. Check 9Router health at http://127.0.0.1:20128/health
3. Update downstream URL in config.yaml if needed

### Tray Icon Not Appearing
**Symptom**: App runs but no tray icon

**Solution**:
1. Check system tray settings
2. Try restarting the app
3. Check logs for errors

### Auto-Start Not Working
**Symptom**: App doesn't start after Windows login

**Solution**:
1. Enable "Run at Startup" in tray menu or web UI
2. Check Windows Task Manager startup tab
3. Manually add to Windows startup if needed

### Proxy Returns "proxy is not running"
**Symptom**: 503 error when calling API

**Solution**:
1. In Manual mode, start proxy via tray menu or web UI
2. Or switch to Always On mode for automatic startup

### Web UI Not Loading
**Symptom**: Cannot open http://127.0.0.1:20127/admin

**Solution**:
1. Check app is running
2. Check firewall settings
3. Try http://localhost:20127/admin

## Getting Help

- GitHub Issues: https://github.com/atlasbridge/atlasbridge/issues
- Check logs at: `%APPDATA%/AtlasBridge/logs/`

## Log Location
- Windows: `%APPDATA%/AtlasBridge/logs/`

## Config Location
- Windows: `%APPDATA%/AtlasBridge/config.yaml`