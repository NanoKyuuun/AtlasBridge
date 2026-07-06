@echo off
REM AtlasBridge - Lint and Format Script for Windows

setlocal
set FAILED=0

echo === AtlasBridge Lint ===

echo [1/4] go vet...
go vet ./...
if %errorlevel% neq 0 (
    echo FAIL: go vet
    set FAILED=1
)

echo [2/4] gofmt (format)...
gofmt -w ./internal/ ./cmd/
if %errorlevel% neq 0 (
    echo FAIL: gofmt
    set FAILED=1
)

echo [3/4] gofmt (check for remaining issues)...
gofmt -l ./internal/ ./cmd/ >nul 2>&1
if %errorlevel% neq 0 (
    echo FAIL: gofmt check
    set FAILED=1
)

echo [4/4] prettier (format)...
npx --yes prettier --write "src/**/*.{ts,vue,js}" 2>nul
if %errorlevel% neq 0 (
    echo FAIL: prettier
    set FAILED=1
)

echo.
if %FAILED% equ 0 (
    echo All lint checks passed.
) else (
    echo Lint checks failed!
    exit /b 1
)
