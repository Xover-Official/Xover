# 🏢 Talos Enterprise Deployment Guide

## Overview

Talos Enterprise transforms the single-node guardian into a distributed, scalable cloud optimization platform suitable for enterprise environments.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Talos Enterprise Stack                  │
├─────────────────────────────────────────────────────────────┤
│  Load Balancer (Nginx)                                      │
│  ┌─────────────┬─────────────┬─────────────┐               │
│  │   Manager   │   Worker-1  │   Worker-N  │               │
│  │   (API)     │  (Scanner)  │ (Optimizer) │               │
│  └─────────────┴─────────────┴─────────────┘               │
├─────────────────────────────────────────────────────────────┤
│  Redis (Task Queue & Caching)                               │
│  PostgreSQL (Distributed State)                              │
├─────────────────────────────────────────────────────────────┤
│  Prometheus (Metrics) │ Grafana (Dashboards) │ Jaeger (Tracing)│
└─────────────────────────────────────────────────────────────┘
```

## Quick Start

### 1. Environment Setup

Create a `.env` file with your credentials:

```bash
# Database
POSTGRES_PASSWORD=your_secure_postgres_password

# Redis
REDIS_PASSWORD=your_secure_redis_password

# AI API Keys
OPENROUTER_API_KEY=sk-or-v1-your-openrouter-key
DEVIN_API_KEY=apk-your-devin-key
OPENAI_API_KEY=sk-your-openai-key

# Security
JWT_SECRET=your_jwt_secret_key_here

# Monitoring
GRAFANA_USER=admin
GRAFANA_PASSWORD=your_grafana_password
```

### 2. Deploy Enterprise Stack

```bash
# Deploy all services
docker-compose -f docker-compose.enterprise.yml up -d

# Scale workers as needed
docker-compose -f docker-compose.enterprise.yml up -d --scale talos-worker=5
```

### 3. Verify Deployment

```bash
# Check all services
docker-compose -f docker-compose.enterprise.yml ps

# View logs
docker-compose -f docker-compose.enterprise.yml logs -f talos-manager

# Test API
curl http://localhost:8080/health
```

## Services

### Manager (talos-manager)
- **Port**: 8080 (API), 9090 (Metrics)
- **Purpose**: Task scheduling, worker coordination, API gateway
- **Endpoints**:
  - `GET /health` - Health check
  - `GET /api/v1/workers` - List active workers
  - `POST /api/v1/tasks` - Create new task
  - `GET /api/v1/metrics` - System metrics

### Workers (talos-worker)
- **Purpose**: Distributed task execution
- **Scalability**: Horizontal scaling via Docker replicas
- **Task Types**:
  - `scan` - Cloud resource discovery
  - `analyze` - AI-powered analysis
  - `optimize` - Cost optimization actions

### PostgreSQL
- **Port**: 5432
- **Purpose**: Persistent state, audit logs, multi-tenancy
- **Features**:
  - UUID primary keys
  - JSONB for flexible metadata
  - Audit trails
  - Multi-organization support

### Redis
- **Port**: 6379
- **Purpose**: Task queuing, caching, worker coordination
- **Features**:
  - Priority queues
  - Worker heartbeats
  - Distributed locking

## Monitoring

### Grafana Dashboard
- **URL**: http://localhost:3000
- **Credentials**: admin / your_grafana_password
- **Dashboards**:
  - System Overview
  - Worker Performance
  - Cost Optimization Metrics
  - AI Token Usage

### Prometheus Metrics
- **URL**: http://localhost:9091
- **Key Metrics**:
  - `talos_tasks_processed_total`
  - `talos_workers_active`
  - `talos_cost_savings_usd`
  - `talos_ai_tokens_used`

### Jaeger Tracing
- **URL**: http://localhost:16686
- **Purpose**: Distributed tracing for task execution

## Configuration

### Enterprise Config (`config.enterprise.yaml`)

```yaml
guardian:
  mode: "enterprise"
  risk_threshold: 5.0
  dry_run: false  # Production mode

worker:
  concurrency: 10    # Tasks per worker
  timeout: 300       # Task timeout in seconds

redis:
  addr: "redis:6379"
  password: "${REDIS_PASSWORD}"
  db: 0

database:
  type: "postgres"
  host: "postgres"
  port: 5432
  database: "talos_ledger"
  user: "talos"
  password: "${POSTGRES_PASSWORD}"
```

## Scaling

### Horizontal Worker Scaling

```bash
# Scale to 10 workers
docker-compose -f docker-compose.enterprise.yml up -d --scale talos-worker=10

# Scale down to 3 workers
docker-compose -f docker-compose.enterprise.yml up -d --scale talos-worker=3
```

### Manager High Availability

```bash
# Deploy multiple managers behind load balancer
docker-compose -f docker-compose.enterprise.yml up -d --scale talos-manager=3
```

## Security

### Network Security
- All services in isolated Docker network
- SSL/TLS termination at Nginx
- Internal service communication only

### Authentication
- JWT-based API authentication
- Role-based access control (RBAC)
- Multi-organization isolation

### Audit Trail
- All actions logged to PostgreSQL
- Immutable audit logs
- User activity tracking

## Production Considerations

### Resource Requirements

**Minimum (Small Team)**:
- CPU: 4 cores
- RAM: 8GB
- Storage: 50GB SSD

**Recommended (Enterprise)**:
- CPU: 16 cores
- RAM: 32GB
- Storage: 500GB SSD
- Network: 1Gbps

### Backup Strategy

```bash
# PostgreSQL backup
docker exec talos-postgres pg_dump -U talos talos_ledger > backup.sql

# Redis backup
docker exec talos-redis redis-cli BGSAVE
```

### Monitoring Alerts

Set up alerts in Grafana/Prometheus for:
- Worker failures
- High task queue depth
- Database connection issues
- AI API rate limits

## Troubleshooting

### Common Issues

**Workers not connecting**:
```bash
# Check Redis connection
docker exec talos-redis redis-cli ping

# Check worker logs
docker-compose -f docker-compose.enterprise.yml logs talos-worker
```

**Database connection errors**:
```bash
# Check PostgreSQL status
docker exec talos-postgres pg_isready -U talos

# View connection logs
docker-compose -f docker-compose.enterprise.yml logs talos-postgres
```

**High memory usage**:
```bash
# Monitor Redis memory
docker exec talos-redis redis-cli info memory

# Check worker memory
docker stats talos-worker
```

## Migration from Single-Node

### Data Migration

1. Export SQLite data:
```bash
sqlite3 atlas_ledger.db .dump > backup.sql
```

2. Import to PostgreSQL:
```bash
docker exec -i talos-postgres psql -U talos talos_ledger < backup.sql
```

3. Update configuration to use PostgreSQL

### Configuration Migration

Replace `config.yaml` with `config.enterprise.yaml` and update:
- Database connection settings
- Redis configuration
- Worker concurrency settings

## Support

For enterprise support:
- Documentation: `/docs`
- Monitoring: Grafana dashboards
- Logs: `/logs` directory
- Health checks: `/health` endpoint
