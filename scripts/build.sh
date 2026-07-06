#!/bin/bash
# AtlasBridge - Build Script for Unix

set -e

echo "=== Building AtlasBridge ==="

echo "[1/3] Building frontend..."
(cd web && npm run build)

echo "[2/3] Building Go backend..."
go build -o atlasbridge ./cmd/atlasbridge

echo "[3/3] Build complete!"
echo "Binary: ./atlasbridge"
