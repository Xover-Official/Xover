# Talos Enterprise Cloud Optimization - Windows Packaging Script

param(
    [string]$Version = "1.0.0",
    [string]$OutputDir = "."
)

$ErrorActionPreference = "Stop"

# Colors for output
function Write-ColorOutput($ForegroundColor) {
    $fc = $host.UI.RawUI.ForegroundColor
    $host.UI.RawUI.ForegroundColor = $ForegroundColor
    if ($args) {
        Write-Output $args
    }
    $host.UI.RawUI.ForegroundColor = $fc
}

function Write-Info($Message) {
    Write-ColorOutput Green "[INFO] $Message"
}

function Write-Warn($Message) {
    Write-ColorOutput Yellow "[WARN] $Message"
}

function Write-Error($Message) {
    Write-ColorOutput Red "[ERROR] $Message"
}

function Write-Step($Message) {
    Write-ColorOutput Cyan "[STEP] $Message"
}

# Main packaging function
function Package-TalosEnterprise {
    $BuildDate = Get-Date -Format "yyyy-MM-dd-HHmmss"
    $PackageName = "talos-enterprise-$Version-$BuildDate"
    
    Write-Info "📦 Packaging Talos Enterprise v$Version"
    Write-Info "====================================="
    
    # Create package structure
    Write-Step "Creating package structure..."
    if (Test-Path $PackageName) {
        Remove-Item -Recurse -Force $PackageName
    }
    New-Item -ItemType Directory -Path $PackageName | Out-Null
    New-Item -ItemType Directory -Path "$PackageName\src" | Out-Null
    New-Item -ItemType Directory -Path "$PackageName\k8s" | Out-Null
    New-Item -ItemType Directory -Path "$PackageName\deploy" | Out-Null
    New-Item -ItemType Directory -Path "$PackageName\monitoring" | Out-Null
    New-Item -ItemType Directory -Path "$PackageName\scripts" | Out-Null
    New-Item -ItemType Directory -Path "$PackageName\docs" | Out-Null
    New-Item -ItemType Directory -Path "$PackageName\config" | Out-Null
    
    # Copy source code
    Write-Step "Copying source code..."
    Copy-Item -Recurse -Force "cmd" "$PackageName\src\"
    Copy-Item -Recurse -Force "internal" "$PackageName\src\"
    Copy-Item -Recurse -Force "pkg" "$PackageName\src\"
    Copy-Item -Force "go.mod" "$PackageName\src\"
    Copy-Item -Force "go.sum" "$PackageName\src\"
    Copy-Item -Force "Dockerfile.production" "$PackageName\src\"
    Copy-Item -Force "Dockerfile" "$PackageName\src\"
    
    # Copy configuration files
    Copy-Item -Force "config.yaml" "$PackageName\config\"
    Copy-Item -Force "docker-compose.production.yml" "$PackageName\config\"
    
    # Copy Kubernetes manifests
    Write-Step "Copying Kubernetes manifests..."
    Get-ChildItem "k8s\*.yaml" | Copy-Item -Destination "$PackageName\k8s\"
    
    # Copy deployment scripts
    Write-Step "Copying deployment scripts..."
    Copy-Item -Recurse -Force "deploy\*" "$PackageName\deploy\"
    
    # Copy monitoring configuration
    Write-Step "Copying monitoring configuration..."
    if (Test-Path "monitoring") {
        Copy-Item -Recurse -Force "monitoring\*" "$PackageName\monitoring\"
    }
    
    # Copy documentation
    Write-Step "Copying documentation..."
    if (Test-Path "docs") {
        Copy-Item -Recurse -Force "docs\*" "$PackageName\docs\"
    }
    Copy-Item -Force "README.md" "$PackageName\"
    
    # Create deployment scripts
    Write-Step "Creating deployment scripts..."
    
    $DeployScript = @"
@echo off
REM Talos Enterprise - Master Deployment Script
echo 🚀 Talos Enterprise Cloud Optimization - Deployment
echo ==================================================
echo.
echo Select cloud provider:
echo 1) AWS (ECS)
echo 2) Google Cloud Platform (GKE)
echo 3) Microsoft Azure (AKS)
echo 4) Docker Compose (Local)
echo.
set /p choice="Enter your choice (1-4): "

if "%choice%"=="1" (
    echo 🔧 Deploying to AWS...
    call deploy\aws\deploy.sh
) else if "%choice%"=="2" (
    echo 🔧 Deploying to Google Cloud...
    call deploy\gcp\deploy.sh
) else if "%choice%"=="3" (
    echo 🔧 Deploying to Azure...
    call deploy\azure\deploy.sh
) else if "%choice%"=="4" (
    echo 🔧 Deploying with Docker Compose...
    docker-compose -f config\docker-compose.production.yml up -d
) else (
    echo ❌ Invalid choice. Exiting.
    exit /b 1
)

echo.
echo ✅ Deployment completed!
echo 📊 Access your dashboard at: http://localhost:8080
pause
"@
    
    $DeployScript | Out-File -FilePath "$PackageName\scripts\deploy.bat" -Encoding ASCII
    
    $BuildScript = @"
@echo off
REM Talos Enterprise - Build Script
echo 🔨 Building Talos Enterprise...
docker build -f Dockerfile.production -t talos-enterprise:latest .
echo ✅ Build completed!
echo 🐳 Image: talos-enterprise:latest
pause
"@
    
    $BuildScript | Out-File -FilePath "$PackageName\scripts\build.bat" -Encoding ASCII
    
    $HealthCheckScript = @"
@echo off
REM Talos Enterprise - Health Check Script
echo 🏥 Checking Talos Enterprise health...

REM Check if application is running
curl -f http://localhost:8080/health >nul 2>&1
if %errorlevel% equ 0 (
    echo ✅ Application is healthy
) else (
    echo ❌ Application is not responding
    exit /b 1
)

echo 🎉 All services are healthy!
pause
"@
    
    $HealthCheckScript | Out-File -FilePath "$PackageName\scripts\health-check.bat" -Encoding ASCII
    
    # Create version info
    Write-Step "Creating version information..."
    $VersionInfo = @"
Package: Talos Enterprise Cloud Optimization
Version: $Version
Build Date: $BuildDate
Platform: Windows
Architecture: $env:PROCESSOR_ARCHITECTURE
"@
    
    $VersionInfo | Out-File -FilePath "$PackageName\VERSION" -Encoding ASCII
    
    # Create archive
    Write-Step "Creating deployment archive..."
    Compress-Archive -Path $PackageName -DestinationPath "$PackageName.zip" -Force
    
    # Generate checksum
    Write-Step "Generating checksum..."
    $FileHash = Get-FileHash "$PackageName.zip" -Algorithm SHA256
    $FileHash.Hash | Out-File -FilePath "$PackageName.zip.sha256" -Encoding ASCII
    
    Write-Info "🎉 Packaging completed successfully!"
    Write-Info "📦 Package: $PackageName.zip"
    Write-Info "🔐 Checksum: $PackageName.zip.sha256"
    Write-Info "📊 Size: $((Get-Item "$PackageName.zip").Length / 1MB) MB)"
    Write-Info ""
    Write-Info "🚀 Ready for cloud deployment!"
    Write-Info "💡 Upload $PackageName.zip to your preferred cloud provider"
    Write-Info "📖 See $PackageName\README.md for deployment instructions"
}

# Run the packaging function
Package-TalosEnterprise
