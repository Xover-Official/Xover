param (
    [string]$Command = "help"
)

$ErrorActionPreference = "Stop"

function Show-Help {
    Write-Host "Usage: .\manage.ps1 [command]"
    Write-Host "Commands:"
    Write-Host "  build       - Build the application binaries"
    Write-Host "  run         - Build and run the application"
    Write-Host "  run-engine  - Run the engine directly (go run)"
    Write-Host "  test        - Run unit tests"
    Write-Host "  clean       - Clean build artifacts"
}

if ($Command -eq "build") {
    Write-Host "🔨 Building Talos..."
    if (-not (Test-Path "bin")) { New-Item -ItemType Directory -Path "bin" | Out-Null }
    
    $env:CGO_ENABLED="0"
    go build -o bin/talos.exe ./cmd/atlas
    if (Test-Path "./cmd/talos-cli") {
        go build -o bin/talos-cli.exe ./cmd/talos-cli
    }
    Write-Host "✅ Build complete."
}
elseif ($Command -eq "run") {
    Write-Host "🔥 Starting Talos..."
    & .\manage.ps1 build
    ./bin/talos.exe
}
elseif ($Command -eq "run-engine") {
    Write-Host "🔥 Starting Talos Engine (Dev Mode)..."
    go run ./cmd/atlas/main.go
}
elseif ($Command -eq "test") {
    Write-Host "🧪 Running Unit Tests..."
    go test -v ./internal/... ./cmd/...
}
elseif ($Command -eq "clean") {
    Write-Host "🧹 Cleaning up..."
    if (Test-Path "bin") { Remove-Item -Recurse -Force "bin" }
    Write-Host "✅ Clean complete."
}
else {
    Show-Help
}