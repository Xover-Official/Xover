#!/bin/bash
# Talos Easy Installer for Solo Users
# Usage: curl -fsSL https://get.talos.dev | bash

set -e

echo "🚀 Talos Autonomous Cloud Guardian - Easy Installer"
echo "=================================================="
echo ""

# Check prerequisites
echo "📋 Checking prerequisites..."

# Check Docker
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first:"
    echo "   https://docs.docker.com/get-docker/"
    exit 1
fi
echo "✅ Docker found"

# Check Docker Compose
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose first:"
    echo "   https://docs.docker.com/compose/install/"
    exit 1
fi
echo "✅ Docker Compose found"

# Create installation directory
INSTALL_DIR="${HOME}/.talos"
echo ""
echo "📁 Installing to: $INSTALL_DIR"
mkdir -p "$INSTALL_DIR"
cd "$INSTALL_DIR"

# Download latest release
echo ""
echo "⬇️  Downloading Talos..."
LATEST_VERSION=$(curl -s https://api.github.com/repos/your-org/talos/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
curl -L "https://github.com/your-org/talos/archive/${LATEST_VERSION}.tar.gz" -o talos.tar.gz
tar -xzf talos.tar.gz --strip-components=1
rm talos.tar.gz

# Create config file
echo ""
echo "⚙️  Creating configuration..."
cat > config.yaml <<EOF
guardian:
  mode: "personal"
  risk_threshold: 5.0
  indie_force: true
  dry_run: true  # Start in safe mode

ai:
  openrouter_key: ""  # Add your API key here
  devin_key: ""
  gpt_5_key: ""

database:
  type: "sqlite"  # Solo users can use SQLite

network:
  dashboard_port: 8080
  enable_sse: true
EOF

echo "✅ Configuration created at: $INSTALL_DIR/config.yaml"

# Create .env file
cat > .env <<EOF
DB_PASSWORD=talos_secure_$(openssl rand -hex 16)
GRAFANA_PASSWORD=admin
VAULT_TOKEN=root
EOF

echo "✅ Environment file created"

# Pull Docker images
echo ""
echo "🐳 Pulling Docker images..."
docker-compose pull

# Start services
echo ""
echo "🚀 Starting Talos..."
docker-compose up -d

# Wait for services
echo ""
echo "⏳ Waiting for services to start..."
sleep 10

# Check health
if docker-compose ps | grep -q "Up"; then
    echo ""
    echo "✅ Talos is running!"
    echo ""
    echo "🎉 Installation Complete!"
    echo ""
    echo "📊 Dashboard: http://localhost:8080"
    echo "📈 Grafana: http://localhost:3000 (admin/admin)"
    echo "🔍 Prometheus: http://localhost:9090"
    echo ""
    echo "📝 Next steps:"
    echo "   1. Edit $INSTALL_DIR/config.yaml and add your API keys"
    echo "   2. Restart: cd $INSTALL_DIR && docker-compose restart"
    echo "   3. View logs: cd $INSTALL_DIR && docker-compose logs -f"
    echo ""
    echo "📚 Documentation: https://docs.talos.dev"
    echo "💬 Support: https://discord.gg/talos"
else
    echo "❌ Something went wrong. Check logs:"
    echo "   cd $INSTALL_DIR && docker-compose logs"
    exit 1
fi
