param (
    [string]$Mode = "Development",
    [string]$Command = "none"
)

if ($Mode -eq "Production") {
    $env:GOGC = "50"
    Write-Host "--- ATLAS ENGINE: PRODUCTION MODE (32GB OPTIMIZED) ---"
} else {
    Write-Host "--- ATLAS ENGINE: DEVELOPMENT MODE ---"
}

if ($Command -eq "run-engine") {
    Write-Host "Launching Atlas Engine..."
    .\atlas.exe --config config.yaml --scan-all
} 
elseif ($Command -eq "clean") {
    Write-Host "Cleaning artifacts..."
    Remove-Item -Path ".\atlas.exe", ".\demo_risk.exe" -ErrorAction SilentlyContinue
    Write-Host "Clean complete."
}
else {
    Write-Host "Environment ready. Use -Command run-engine to start."
}
