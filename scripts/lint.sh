#!/bin/bash
# Smart AI Proxy - Lint and Format Script

set -e

FAILED=0

echo "=== Smart AI Proxy Lint ==="

echo "[1/4] go vet..."
if ! go vet ./...; then
    echo "FAIL: go vet"
    FAILED=1
fi

echo "[2/4] gofmt (format)..."
gofmt -w ./internal/ ./cmd/

echo "[3/4] gofmt (check remaining)..."
if [ -n "$(gofmt -l ./internal/ ./cmd/)" ]; then
    echo "FAIL: gofmt check"
    FAILED=1
fi

echo "[4/4] prettier (format)..."
npx --yes prettier --write "src/**/*.{ts,vue,js}" 2>/dev/null

echo
if [ "$FAILED" -eq 0 ]; then
    echo "All lint checks passed."
else
    echo "Lint checks failed!"
    exit 1
fi
