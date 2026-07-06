#!/bin/bash
# Smart AI Proxy - Build Script for Unix

set -e

echo "=== Building Smart AI Proxy ==="

echo "[1/3] Building frontend..."
(cd web && npm run build)

echo "[2/3] Building Go backend..."
go build -o smart-ai-proxy ./cmd/smart-ai-proxy

echo "[3/3] Build complete!"
echo "Binary: ./smart-ai-proxy"
