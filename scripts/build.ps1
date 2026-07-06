# Smart AI Proxy - Build Script for Windows
# Produces portable release with version info

$ErrorActionPreference = "Stop"

$VERSION = "0.1.0"
$ARTIFACT_DIR = "release"
$BINARY_NAME = "smart-ai-proxy.exe"

Write-Host "=== Building Smart AI Proxy v$VERSION ==="
Write-Host ""

# Create release directory
if (Test-Path $ARTIFACT_DIR) {
    Remove-Item -Recurse -Force $ARTIFACT_DIR
}
New-Item -ItemType Directory -Path $ARTIFACT_DIR | Out-Null

# Step 1: Build frontend
Write-Host "[1/4] Building frontend..."
Push-Location web
try {
    npm run build
    if ($LASTEXITCODE -ne 0) {
        throw "Frontend build failed!"
    }
}
finally {
    Pop-Location
}
Write-Host "      Frontend built successfully"

# Step 2: Build Go backend with version info
Write-Host "[2/4] Building Go backend..."
$ldflags = "-X github.com/smart-ai-proxy/smart-ai-proxy/internal/server.Version=$VERSION"
go build -ldflags $ldflags -o "$ARTIFACT_DIR/$BINARY_NAME" ./cmd/smart-ai-proxy
if ($LASTEXITCODE -ne 0) {
    throw "Go build failed!"
}
Write-Host "      Binary built: $ARTIFACT_DIR/$BINARY_NAME"

# Step 3: Create portable zip
Write-Host "[3/4] Creating portable distribution..."
$zipName = "smart-ai-proxy-$VERSION-windows-x64-portable.zip"
$content = @(
    "$BINARY_NAME",
    "README.md",
    "QUICKSTART.md"
)

# Copy files to temp staging area
$staging = "$ARTIFACT_DIR/staging"
New-Item -ItemType Directory -Path $staging -Force | Out-Null
Copy-Item "$ARTIFACT_DIR/$BINARY_NAME" "$staging/"
Copy-Item "README.md" "$staging/"
if (Test-Path "QUICKSTART.md") {
    Copy-Item "QUICKSTART.md" "$staging/"
}

# Create zip
Compress-Archive -Path "$staging/*" -DestinationPath "$ARTIFACT_DIR/$zipName" -Force
Write-Host "      Created: $ARTIFACT_DIR/$zipName"

# Step 4: Generate checksums
Write-Host "[4/4] Generating checksums..."
$checksumFile = "$ARTIFACT_DIR/checksums.txt"
Get-ChildItem "$ARTIFACT_DIR/*.zip" | ForEach-Object {
    $hash = (Get-FileHash $_.FullName -Algorithm SHA256).Hash.ToLower()
    "$hash  $($_.Name)" | Add-Content $checksumFile
}
Get-ChildItem "$ARTIFACT_DIR/*.exe" | ForEach-Object {
    $hash = (Get-FileHash $_.FullName -Algorithm SHA256).Hash.ToLower()
    "$hash  $($_.Name)" | Add-Content $checksumFile
}
Write-Host "      Checksums written to: $checksUMFile"

Write-Host ""
Write-Host "=== Build Complete ==="
Write-Host "Artifacts in: $ARTIFACT_DIR/"
Get-ChildItem $ARTIFACT_DIR | ForEach-Object {
    Write-Host "  - $($_.Name)"
}