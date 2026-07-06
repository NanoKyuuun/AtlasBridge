#!/bin/bash
# AtlasBridge - Run Script for development

set -e

echo "=== Starting AtlasBridge (dev mode) ==="
echo "API:    http://127.0.0.1:20127/v1"
echo "Admin:  http://127.0.0.1:20127/admin"
echo "Health: http://127.0.0.1:20127/health"
echo ""

go run ./cmd/atlasbridge
