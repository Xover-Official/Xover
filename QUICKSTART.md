# Talos: Quick Start Guide

## 🚀 Get Started in 5 Minutes

### Prerequisites

- Go 1.21+
- PostgreSQL 15+ (optional, SQLite works for dev)
- Redis (optional, for caching)

### Installation

```bash
# 1. Clone repository
git clone https://github.com/your-org/talos.git
cd talos

# 2. Install dependencies
go mod download

# 3. Configure
cp config.yaml.example config.yaml
# Edit config.yaml with your API keys
```

### Running Talos

#### Development Mode (SQLite)

```bash
# Run Guardian
go run cmd/atlas/main.go

# Run Dashboard (separate terminal)
go run cmd/dashboard/main.go

# Open browser
open http://localhost:8080
```

#### Production Mode (PostgreSQL)

```bash
# 1. Setup database
createdb talos
go run cmd/migrate/main.go up

# 2. Update config.yaml
database:
  type: "postgres"

# 3. Run
docker-compose up --build
```

## 📋 Features

### ✅ Implemented

- **5-Tier AI Swarm** - Gemini, Claude, GPT-5, Devin
- **Multi-Cloud** - AWS, Azure, GCP support
- **PostgreSQL** - Production-grade persistence
- **Redis Caching** - 30-50% cost reduction
- **Vault Integration** - Secure secrets management
- **Cost Forecasting** - Predictive analytics
- **ROI Tracking** - Real-time token cost monitoring
- **Dashboard** - Glassmorphic UI with live charts

### 🔄 In Progress

- Vector database for AI memory
- RBAC & multi-tenancy
- Advanced ML models

## 🎯 Common Tasks

### View Token Usage

```bash
curl http://localhost:8080/api/roi
```

### Run Migration

```bash
go run cmd/migrate/main.go up
```

### Check Database Status

```bash
go run cmd/migrate/main.go status
```

### Enable Dry Run Mode

```yaml
# config.yaml
guardian:
  dry_run: true  # Simulate without executing
```

## 📚 Documentation

- [Implementation Plan](./brain/implementation_plan.md)
- [PostgreSQL Setup](./docs/POSTGRES_SETUP.md)
- [Architecture](./docs/ARCHITECTURE.md)

## 🆘 Troubleshooting

### Build Errors

```bash
go mod tidy
go clean -cache
```

### Database Connection

```bash
# Test PostgreSQL
psql -U talos_user -d talos -h localhost

# Test Redis
redis-cli ping
```

### API Keys

Set environment variables:

```bash
export OPENROUTER_KEY="sk-or-v1-..."
export DEVIN_KEY="apk_..."
export GPT5_KEY="sk-..."
```

## 🎉 Success

If you see:

```
⚔️  Guardian Active. Entering OODA Loop...
💰 Token Tracker initialized. Current ROI: 0.0%
```

Talos is running! Visit <http://localhost:8080> to see the dashboard.
