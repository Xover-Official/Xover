$ErrorActionPreference = "Stop"

Write-Host "🔨 Building Talos..." -ForegroundColor Cyan

$buildDir = "bin"
if (-not (Test-Path $buildDir)) {
    New-Item -ItemType Directory -Force -Path $buildDir | Out-Null
}

Write-Host "📦 Installing Dependencies..."
go mod tidy

Write-Host "  -> Building Core Service..."
go build -o "$buildDir/talos.exe" ./cmd/atlas

Write-Host "  -> Building CLI..."
go build -o "$buildDir/talos-cli.exe" ./cmd/talos-cli

Write-Host "✅ Build complete. Binaries are in '$buildDir/'" -ForegroundColor Green

Write-Host "🧪 Running Tests..." -ForegroundColor Cyan
go test -v ./internal/... ./cmd/...

Write-Host "🚀 Running E2E Tests..." -ForegroundColor Cyan
go test -v ./tests/e2e/...
