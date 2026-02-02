# 🚀 TALOS Deployment Guide

TALOS is a distributed, production-ready cloud optimization engine. This guide covers deployment across major environments.

## ⚡ Quick Start (5 Commands or Less)

```bash
docker-compose up -d                   # Start the backend (Postgres, Redis, Workers)
go build -o dashboard ./cmd/dashboard  # Build the dashboard
./dashboard                             # Start the dashboard
go build -o atlas ./cmd/atlas           # Build the CLI engine
./atlas run                             # Start the first OODA cycle
```

## 🌩️ Multi-Cloud Production Deployment

### 1. AWS (ECS/EKS)

- **Infrastructure**: Use the provided CloudFormation/Terraform templates in `deploy/aws`.
- **Database**: Amazon RDS for PostgreSQL.
- **Cache**: Amazon ElastiCache for Redis.
- **Compute**: ECS Fargate or EKS Cluster.
- **Deployment**: `kubectl apply -f k8s/aws/`

### 2. Google Cloud (GKE)

- **Infrastructure**: GKE Autopilot or Standard cluster.
- **Database**: Cloud SQL for PostgreSQL.
- **Cache**: Memorystore for Redis.
- **Deployment**: `kubectl apply -f k8s/gcp/`

### 3. Azure (AKS)

- **Infrastructure**: Azure Kubernetes Service.
- **Database**: Azure Database for PostgreSQL.
- **Cache**: Azure Cache for Redis.
- **Deployment**: `kubectl apply -f k8s/azure/`

## 🔑 Environment Variables Reference

| Variable | Description | Default |
|----------|-------------|---------|
| `OPENROUTER_API_KEY` | Key for AI Swarm Orchestration | Required |
| `DATABASE_DSN` | PostgreSQL connection string | Required |
| `REDIS_PASSWORD` | Password for Redis cache | Optional |
| `JWT_SECRET_KEY` | Key for signing session tokens | Required |
| `CLOUD_PROVIDER` | `aws`, `azure`, or `gcp` | `aws` |
| `DRY_RUN` | If `true`, TALOS will only propose changes | `true` |

## 🏥 Health Check Endpoints

TALOS provides enterprise-standard health checks for load balancers and orchestrators:

- **Liveness**: `GET /health` (Returns 200 OK if service is up)
- **Readiness**: `GET /ready` (Returns 200 OK if DB and Redis connections are active)
- **Metrics**: `GET /metrics` (Prometheus-compatible metrics endpoint)

## 🐳 Docker Deployment

The production `Dockerfile` uses a multi-stage build to minimize image size and security surface area.

```bash
docker build -t talos-enterprise:latest .
docker run --env-file .env talos-enterprise
```

---

🛡️ **TALOS**: Production-ready. Autonomous. Safe.
