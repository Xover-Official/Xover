# ❓ Talos Troubleshooting Guide

## Quick Diagnostics

Run this command to check system health:

```bash
cd ~/.talos && docker-compose ps
```

Expected output: All services should show "Up"

---

## Common Issues & Solutions

### 1. Installation Fails

**Symptom**: Installer script errors out

**Possible Causes**:

- Docker not installed
- Docker Compose not installed
- Insufficient permissions
- Port conflicts

**Solutions**:

```bash
# Check Docker
docker --version
# If not installed: https://docs.docker.com/get-docker/

# Check Docker Compose
docker-compose --version
# If not installed: https://docs.docker.com/compose/install/

# Check ports
netstat -an | grep -E '8080|5432|6379|3000|9090'
# If ports are in use, edit docker-compose.yml to use different ports

# Fix permissions (Linux/Mac)
sudo usermod -aG docker $USER
newgrp docker
```

---

### 2. Dashboard Not Loading

**Symptom**: <http://localhost:8080> shows error or doesn't load

**Solutions**:

```bash
# Check if dashboard is running
docker-compose ps dashboard

# Check logs
docker-compose logs dashboard

# Restart dashboard
docker-compose restart dashboard

# If still not working, rebuild
docker-compose up -d --build dashboard
```

**Common errors**:

- "Connection refused" → Dashboard container not running
- "502 Bad Gateway" → Backend not responding
- Blank page → Check browser console for JavaScript errors

---

### 3. No Optimizations Detected

**Symptom**: Talos runs but finds no optimizations

**Possible Causes**:

- Cloud credentials not configured
- No resources in cloud account
- Dry-run mode enabled
- Resources already optimized

**Solutions**:

```bash
# Check config
cat ~/.talos/config.yaml

# Verify API keys are set
grep -E "openrouter_key|devin_key|gpt_5_key" ~/.talos/config.yaml

# Check dry-run mode
grep "dry_run" ~/.talos/config.yaml
# Should be "false" for actual optimizations

# Check logs for errors
docker-compose logs guardian | grep -i error

# Test cloud connection
docker-compose exec guardian /app/talos test-connection
```

---

### 4. Database Connection Errors

**Symptom**: "Failed to connect to database" errors

**Solutions**:

```bash
# Check PostgreSQL is running
docker-compose ps postgres

# Check PostgreSQL logs
docker-compose logs postgres

# Test connection
docker-compose exec postgres psql -U talos_user -d talos -c "SELECT 1;"

# Reset database (WARNING: deletes all data)
docker-compose down -v
docker-compose up -d
docker-compose exec guardian /app/migrate up
```

---

### 5. Redis Cache Issues

**Symptom**: Slow performance or cache errors

**Solutions**:

```bash
# Check Redis is running
docker-compose ps redis

# Test Redis connection
docker-compose exec redis redis-cli ping
# Should return "PONG"

# Clear cache
docker-compose exec redis redis-cli FLUSHALL

# Restart Redis
docker-compose restart redis
```

---

### 6. AI API Errors

**Symptom**: "AI request failed" or high error rates

**Possible Causes**:

- Invalid API keys
- Rate limits exceeded
- API service outage
- Network issues

**Solutions**:

```bash
# Check API key configuration
docker-compose exec guardian env | grep -E "OPENROUTER|DEVIN|GPT"

# Test OpenRouter connection
curl -H "Authorization: Bearer YOUR_KEY" \
  https://openrouter.ai/api/v1/models

# Check rate limits
docker-compose logs guardian | grep -i "rate limit"

# Switch to different AI tier
# Edit config.yaml and set higher risk_threshold
```

---

### 7. High Memory Usage

**Symptom**: System running slow, high memory consumption

**Solutions**:

```bash
# Check container memory usage
docker stats

# Reduce worker pool size
# Edit config.yaml:
guardian:
  worker_pool_size: 3  # Default is 5

# Restart with new config
docker-compose restart guardian

# Increase Docker memory limit
# Docker Desktop → Settings → Resources → Memory
```

---

### 8. Prometheus/Grafana Not Working

**Symptom**: Monitoring dashboards not accessible

**Solutions**:

```bash
# Check services
docker-compose ps prometheus grafana

# Access Prometheus
open http://localhost:9090

# Access Grafana
open http://localhost:3000
# Default login: admin/admin

# Check Prometheus targets
# Go to http://localhost:9090/targets
# All targets should be "UP"

# Restart monitoring stack
docker-compose restart prometheus grafana
```

---

### 9. Helm Deployment Issues

**Symptom**: Kubernetes deployment fails

**Solutions**:

```bash
# Check Helm chart syntax
helm lint ./helm/talos

# Dry-run install
helm install talos ./helm/talos --dry-run --debug

# Check pod status
kubectl get pods -n talos

# Check pod logs
kubectl logs -n talos deployment/talos-guardian

# Describe pod for events
kubectl describe pod -n talos <pod-name>

# Common fixes:
# - Image pull errors: Check image repository and credentials
# - CrashLoopBackOff: Check logs for application errors
# - Pending: Check resource requests vs available capacity
```

---

### 10. Slow Performance

**Symptom**: Dashboard slow, optimizations take too long

**Solutions**:

```bash
# Check system resources
docker stats

# Check database performance
docker-compose exec postgres psql -U talos_user -d talos -c "
  SELECT query, mean_exec_time, calls 
  FROM pg_stat_statements 
  ORDER BY mean_exec_time DESC 
  LIMIT 10;
"

# Enable Redis caching
# Edit config.yaml:
cache:
  redis_addr: "redis:6379"
  enabled: true

# Optimize database
docker-compose exec postgres psql -U talos_user -d talos -c "VACUUM ANALYZE;"

# Check AI API latency
docker-compose logs guardian | grep "AI latency"
```

---

## Logs & Debugging

### View All Logs

```bash
docker-compose logs -f
```

### View Specific Service Logs

```bash
docker-compose logs -f guardian
docker-compose logs -f dashboard
docker-compose logs -f postgres
docker-compose logs -f redis
```

### Export Logs for Support

```bash
docker-compose logs > talos-logs-$(date +%Y%m%d).txt
```

### Enable Debug Mode

```yaml
# config.yaml
guardian:
  debug: true
  log_level: "debug"
```

---

## Health Checks

### System Health Check Script

```bash
#!/bin/bash
# health-check.sh

echo "🏥 Talos Health Check"
echo "===================="

# Check Docker
if docker --version > /dev/null 2>&1; then
    echo "✅ Docker installed"
else
    echo "❌ Docker not found"
fi

# Check services
cd ~/.talos
services=("postgres" "redis" "guardian" "dashboard")
for service in "${services[@]}"; do
    if docker-compose ps $service | grep -q "Up"; then
        echo "✅ $service running"
    else
        echo "❌ $service not running"
    fi
done

# Check ports
ports=(8080 5432 6379 3000 9090)
for port in "${ports[@]}"; do
    if nc -z localhost $port 2>/dev/null; then
        echo "✅ Port $port open"
    else
        echo "⚠️  Port $port not accessible"
    fi
done

# Check disk space
disk_usage=$(df -h ~/.talos | tail -1 | awk '{print $5}' | sed 's/%//')
if [ $disk_usage -lt 80 ]; then
    echo "✅ Disk space OK ($disk_usage% used)"
else
    echo "⚠️  Disk space low ($disk_usage% used)"
fi

echo ""
echo "Health check complete!"
```

---

## Getting Help

### Self-Service Resources

1. **Documentation**: <https://docs.talos.dev>
2. **FAQ**: <https://docs.talos.dev/faq>
3. **Community Forum**: <https://community.talos.dev>
4. **GitHub Issues**: <https://github.com/your-org/talos/issues>

### Support Channels

**Solo License**:

- Email: <support@talos.dev> (48h response)
- Community forum

**Team License**:

- Email: <support@talos.dev> (24h response)
- Priority Slack channel
- Monthly check-in calls

**Enterprise License**:

- 24/7 phone support
- Dedicated Slack channel
- Assigned customer success manager
- SLA guarantees

### Reporting Bugs

Include this information:

```
1. Talos version: [run: docker-compose exec guardian /app/talos version]
2. Operating system: [e.g., Ubuntu 22.04, macOS 13, Windows 11]
3. Deployment method: [Docker Compose, Kubernetes, etc.]
4. Steps to reproduce:
5. Expected behavior:
6. Actual behavior:
7. Logs: [attach logs]
8. Screenshots: [if applicable]
```

---

## Emergency Procedures

### Complete Reset (Nuclear Option)

```bash
# WARNING: This deletes ALL data
cd ~/.talos
docker-compose down -v
rm -rf data/
docker-compose up -d
docker-compose exec guardian /app/migrate up
```

### Backup Before Reset

```bash
# Backup database
docker-compose exec postgres pg_dump -U talos_user talos > backup.sql

# Backup config
cp config.yaml config.yaml.backup

# Restore later
docker-compose exec -T postgres psql -U talos_user talos < backup.sql
```

---

> **Still stuck?** Email <support@talos.dev> with your logs and we'll help you out!
