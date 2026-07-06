@echo off
REM Smart AI Proxy - Run Script for development (Windows)

echo === Starting Smart AI Proxy (dev mode) ===
echo API:    http://127.0.0.1:20127/v1
echo Admin:  http://127.0.0.1:20127/admin
echo Health: http://127.0.0.1:20127/health
echo.

go run ./cmd/smart-ai-proxy
