#!/bin/bash

# Talos Enterprise Cloud Optimization - Complete Packaging Script
# This script packages the entire application for cloud deployment

set -e

# Configuration
VERSION=${VERSION:-"1.0.0"}
BUILD_DATE=$(date +%Y-%m-%d-%H%M%S)
PACKAGE_NAME="talos-enterprise-$VERSION-$BUILD_DATE"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# Create package directory
create_package_structure() {
    log_step "Creating package structure..."
    
    rm -rf "$PACKAGE_NAME"
    mkdir -p "$PACKAGE_NAME"
    
    # Create subdirectories
    mkdir -p "$PACKAGE_NAME/{src,k8s,deploy,monitoring,scripts,docs,config}"
    
    log_info "Package structure created: $PACKAGE_NAME"
}

# Copy source code
copy_source_code() {
    log_step "Copying source code..."
    
    # Copy Go source files
    cp -r cmd "$PACKAGE_NAME/src/"
    cp -r internal "$PACKAGE_NAME/src/"
    cp -r pkg "$PACKAGE_NAME/src/"
    cp go.mod "$PACKAGE_NAME/src/"
    cp go.sum "$PACKAGE_NAME/src/"
    cp Dockerfile.production "$PACKAGE_NAME/src/"
    cp Dockerfile "$PACKAGE_NAME/src/"
    
    # Copy configuration files
    cp config.yaml "$PACKAGE_NAME/config/"
    cp docker-compose.production.yml "$PACKAGE_NAME/config/"
    
    log_info "Source code copied successfully"
}

# Copy Kubernetes manifests
copy_k8s_manifests() {
    log_step "Copying Kubernetes manifests..."
    
    cp k8s/*.yaml "$PACKAGE_NAME/k8s/"
    
    log_info "Kubernetes manifests copied successfully"
}

# Copy deployment scripts
copy_deployment_scripts() {
    log_step "Copying deployment scripts..."
    
    cp -r deploy/* "$PACKAGE_NAME/deploy/"
    
    # Make scripts executable
    chmod +x "$PACKAGE_NAME/deploy"/*.sh
    
    log_info "Deployment scripts copied successfully"
}

# Copy monitoring configuration
copy_monitoring() {
    log_step "Copying monitoring configuration..."
    
    if [ -d "monitoring" ]; then
        cp -r monitoring/* "$PACKAGE_NAME/monitoring/"
    fi
    
    log_info "Monitoring configuration copied successfully"
}

# Copy documentation
copy_documentation() {
    log_step "Copying documentation..."
    
    if [ -d "docs" ]; then
        cp -r docs/* "$PACKAGE_NAME/docs/"
    fi
    
    # Copy README
    cp README.md "$PACKAGE_NAME/"
    
    log_info "Documentation copied successfully"
}

# Create deployment scripts
create_deployment_scripts() {
    log_step "Creating deployment scripts..."
    
    # Create master deployment script
    cat > "$PACKAGE_NAME/scripts/deploy.sh" << 'EOF'
#!/bin/bash

# Talos Enterprise - Master Deployment Script
# Choose your cloud provider and deploy

set -e

echo "🚀 Talos Enterprise Cloud Optimization - Deployment"
echo "=================================================="
echo ""
echo "Select cloud provider:"
echo "1) AWS (ECS)"
echo "2) Google Cloud Platform (GKE)"
echo "3) Microsoft Azure (AKS)"
echo "4) Docker Compose (Local)"
echo ""
read -p "Enter your choice (1-4): " choice

case $choice in
    1)
        echo "🔧 Deploying to AWS..."
        ./deploy/aws/deploy.sh
        ;;
    2)
        echo "🔧 Deploying to Google Cloud..."
        ./deploy/gcp/deploy.sh
        ;;
    3)
        echo "🔧 Deploying to Azure..."
        ./deploy/azure/deploy.sh
        ;;
    4)
        echo "🔧 Deploying with Docker Compose..."
        docker-compose -f config/docker-compose.production.yml up -d
        ;;
    *)
        echo "❌ Invalid choice. Exiting."
        exit 1
        ;;
esac

echo ""
echo "✅ Deployment completed!"
echo "📊 Access your dashboard at: http://localhost:8080"
EOF

    # Create environment setup script
    cat > "$PACKAGE_NAME/scripts/setup-env.sh" << 'EOF'
#!/bin/bash

# Talos Enterprise - Environment Setup Script

set -e

echo "🔧 Setting up environment for Talos Enterprise..."

# Check if .env file exists
if [ ! -f ".env" ]; then
    echo "📝 Creating .env file..."
    cat > .env << 'ENVEOF'
# Talos Enterprise Environment Configuration
# Copy this file to .env and update with your actual values

# Database Configuration
POSTGRES_PASSWORD=your-secure-postgres-password
REDIS_PASSWORD=your-secure-redis-password

# AI API Keys
OPENROUTER_API_KEY=your-openrouter-api-key
GEMINI_API_KEY=your-gemini-api-key
DEVIN_API_KEY=your-devin-api-key

# JWT Configuration
JWT_SECRET=your-jwt-secret-key

# Monitoring
GRAFANA_USER=admin
GRAFANA_PASSWORD=your-grafana-password

# Cloud Provider Specific
# AWS
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your-aws-access-key
AWS_SECRET_ACCESS_KEY=your-aws-secret-key

# GCP
PROJECT_ID=your-gcp-project-id
GOOGLE_APPLICATION_CREDENTIALS=path/to/service-account.json

# Azure
RESOURCE_GROUP=talos-rg
LOCATION=eastus
AZURE_SUBSCRIPTION_ID=your-azure-subscription-id
ENVEOF

    echo "✅ .env file created. Please update it with your actual values."
else
    echo "✅ .env file already exists."
fi

echo ""
echo "🔑 Remember to:"
echo "1. Update .env with your actual API keys and passwords"
echo "2. Ensure your cloud provider credentials are configured"
echo "3. Run 'source .env' to load environment variables"
echo ""
echo "🚀 Ready to deploy!"
EOF

    # Make scripts executable
    chmod +x "$PACKAGE_NAME/scripts"/*.sh
    
    log_info "Deployment scripts created successfully"
}

# Create build script
create_build_script() {
    log_step "Creating build script..."
    
    cat > "$PACKAGE_NAME/scripts/build.sh" << 'EOF'
#!/bin/bash

# Talos Enterprise - Build Script

set -e

echo "🔨 Building Talos Enterprise..."

# Build Docker image
docker build -f Dockerfile.production -t talos-enterprise:latest .

echo "✅ Build completed!"
echo "🐳 Image: talos-enterprise:latest"
EOF

    chmod +x "$PACKAGE_NAME/scripts/build.sh"
    
    log_info "Build script created successfully"
}

# Create health check script
create_health_check() {
    log_step "Creating health check script..."
    
    cat > "$PACKAGE_NAME/scripts/health-check.sh" << 'EOF'
#!/bin/bash

# Talos Enterprise - Health Check Script

set -e

echo "🏥 Checking Talos Enterprise health..."

# Check if application is running
if curl -f http://localhost:8080/health > /dev/null 2>&1; then
    echo "✅ Application is healthy"
else
    echo "❌ Application is not responding"
    exit 1
fi

# Check database connection
if docker exec talos-postgres pg_isready -U talos > /dev/null 2>&1; then
    echo "✅ Database is healthy"
else
    echo "❌ Database is not responding"
    exit 1
fi

# Check Redis connection
if docker exec talos-redis redis-cli ping > /dev/null 2>&1; then
    echo "✅ Redis is healthy"
else
    echo "❌ Redis is not responding"
    exit 1
fi

echo "🎉 All services are healthy!"
EOF

    chmod +x "$PACKAGE_NAME/scripts/health-check.sh"
    
    log_info "Health check script created successfully"
}

# Create README for package
create_package_readme() {
    log_step "Creating package README..."
    
    cat > "$PACKAGE_NAME/README.md" << 'EOF'
# Talos Enterprise Cloud Optimization

## 📋 Overview

Talos Enterprise is a comprehensive cloud resource optimization platform powered by AI. This package contains everything you need to deploy Talos to any major cloud provider or run it locally.

## 🚀 Quick Start

### 1. Environment Setup

```bash
# Make scripts executable
chmod +x scripts/*.sh

# Set up environment
./scripts/setup-env.sh
```

### 2. Choose Your Deployment Method

#### Option A: Cloud Deployment
```bash
# Run the master deployment script
./scripts/deploy.sh
```

#### Option B: Local Docker Deployment
```bash
# Build the application
./scripts/build.sh

# Deploy with Docker Compose
docker-compose -f config/docker-compose.production.yml up -d
```

### 3. Verify Deployment

```bash
# Check health status
./scripts/health-check.sh
```

## 📁 Package Structure

```
talos-enterprise/
├── src/                 # Source code
├── k8s/                 # Kubernetes manifests
├── deploy/              # Cloud deployment scripts
│   ├── aws/           # AWS ECS deployment
│   ├── gcp/           # GCP GKE deployment
│   └── azure/         # Azure AKS deployment
├── monitoring/          # Monitoring configuration
├── scripts/             # Utility scripts
├── docs/                # Documentation
└── config/              # Configuration files
```

## 🌩️ Cloud Deployment

### AWS (ECS)
```bash
./deploy/aws/deploy.sh
```

### Google Cloud Platform (GKE)
```bash
./deploy/gcp/deploy.sh
```

### Microsoft Azure (AKS)
```bash
./deploy/azure/deploy.sh
```

## 🐳 Local Development

### Docker Compose
```bash
docker-compose -f config/docker-compose.production.yml up -d
```

### Individual Services
```bash
# Start PostgreSQL
docker run -d --name talos-postgres -e POSTGRES_PASSWORD=password -p 5432:5432 postgres:15

# Start Redis
docker run -d --name talos-redis -p 6379:6379 redis:7-alpine

# Start Talos
docker run -d --name talos-app -p 8080:8080 --link talos-postgres:postgres --link talos-redis:redis talos-enterprise:latest
```

## 📊 Access Points

Once deployed, you can access:

- **Main Application**: http://localhost:8080
- **Dashboard**: http://localhost:8081
- **API Documentation**: http://localhost:8080/docs
- **Grafana**: http://localhost:3000
- **Prometheus**: http://localhost:9090

## 🔧 Configuration

### Environment Variables

Key environment variables to configure:

```bash
# Database
POSTGRES_PASSWORD=your-password
REDIS_PASSWORD=your-password

# AI Services
OPENROUTER_API_KEY=your-key
GEMINI_API_KEY=your-key
DEVIN_API_KEY=your-key

# Security
JWT_SECRET=your-secret
```

### Configuration Files

- `config.yaml` - Main application configuration
- `docker-compose.production.yml` - Docker Compose setup
- `k8s/` - Kubernetes manifests

## 🏥 Health Monitoring

### Health Check Endpoints

- `/health` - Application health status
- `/ready` - Readiness probe
- `/metrics` - Prometheus metrics

### Monitoring Stack

- **Prometheus**: Metrics collection
- **Grafana**: Visualization dashboards
- **Jaeger**: Distributed tracing
- **CloudWatch/Azure Monitor/GCP Monitoring**: Cloud-specific metrics

## 📚 Documentation

- [API Documentation](docs/API_DOCUMENTATION.md)
- [Architecture Guide](docs/ARCHITECTURE.md)
- [Deployment Guide](docs/DEPLOYMENT.md)
- [Troubleshooting](docs/TROUBLESHOOTING.md)

## 🆘 Support

For support and questions:

- 📧 Email: support@talos.io
- 📖 Documentation: https://docs.talos.io
- 🐛 Issues: https://github.com/project-atlas/atlas/issues

## 📄 License

This software is released under the MIT License. See [LICENSE](LICENSE) for details.

---

**Version**: 1.0.0  
**Build Date**: $(date)  
**Platform**: Cross-platform
EOF

    log_info "Package README created successfully"
}

# Create version info
create_version_info() {
    log_step "Creating version information..."
    
    cat > "$PACKAGE_NAME/VERSION" << EOF
Package: Talos Enterprise Cloud Optimization
Version: $VERSION
Build Date: $BUILD_DATE
Git Commit: $(git rev-parse HEAD 2>/dev/null || echo "unknown")
Platform: $(uname -s)
Architecture: $(uname -m)
EOF

    log_info "Version information created"
}

# Create archive
create_archive() {
    log_step "Creating deployment archive..."
    
    tar -czf "$PACKAGE_NAME.tar.gz" "$PACKAGE_NAME"
    
    log_info "Archive created: $PACKAGE_NAME.tar.gz"
}

# Generate checksum
generate_checksum() {
    log_step "Generating checksum..."
    
    sha256sum "$PACKAGE_NAME.tar.gz" > "$PACKAGE_NAME.tar.gz.sha256"
    
    log_info "Checksum generated: $PACKAGE_NAME.tar.gz.sha256"
}

# Main packaging function
main() {
    log_info "📦 Packaging Talos Enterprise v$VERSION"
    log_info "====================================="
    
    create_package_structure
    copy_source_code
    copy_k8s_manifests
    copy_deployment_scripts
    copy_monitoring
    copy_documentation
    create_deployment_scripts
    create_build_script
    create_health_check
    create_package_readme
    create_version_info
    create_archive
    generate_checksum
    
    echo ""
    log_info "🎉 Packaging completed successfully!"
    log_info "📦 Package: $PACKAGE_NAME.tar.gz"
    log_info "🔐 Checksum: $PACKAGE_NAME.tar.gz.sha256"
    log_info "📊 Size: $(du -h "$PACKAGE_NAME.tar.gz" | cut -f1)"
    echo ""
    log_info "🚀 Ready for cloud deployment!"
    log_info "💡 Upload $PACKAGE_NAME.tar.gz to your preferred cloud provider"
    log_info "📖 See $PACKAGE_NAME/README.md for deployment instructions"
}

# Handle script interruption
trap 'log_error "Packaging interrupted"' EXIT

# Run main function
main "$@"
