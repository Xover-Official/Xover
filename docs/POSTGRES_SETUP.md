# Talos PostgreSQL Setup Guide

## Prerequisites

1. **Install PostgreSQL** (if not already installed)

   ```bash
   # Windows (using Chocolatey)
   choco install postgresql
   
   # Or download from: https://www.postgresql.org/download/windows/
   ```

2. **Start PostgreSQL Service**

   ```bash
   # Windows
   net start postgresql-x64-14
   ```

## Database Setup

### 1. Create Database and User

```sql
-- Connect to PostgreSQL as admin
psql -U postgres

-- Create database
CREATE DATABASE talos;

-- Create user
CREATE USER talos_user WITH PASSWORD 'your_secure_password';

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE talos TO talos_user;

-- Connect to talos database
\c talos

-- Grant schema privileges
GRANT ALL ON SCHEMA public TO talos_user;
```

### 2. Run Migrations

```bash
# Set connection string
$env:DATABASE_URL="postgres://talos_user:your_secure_password@localhost:5432/talos?sslmode=disable"

# Run migration
go run cmd/migrate/main.go up
```

### 3. Verify Setup

```bash
# Check migration status
go run cmd/migrate/main.go status
```

Expected output:

```
✅ Connected to PostgreSQL
📋 Database Status:

Existing tables:
  ✓ actions
  ✓ ai_decisions
  ✓ audit_log
  ✓ organizations
  ✓ resources
  ✓ savings_events
  ✓ token_usage
  ✓ users

Total: 8 tables
```

## Configuration

### Update config.yaml

```yaml
database:
  type: "postgres"  # Change from "sqlite"
  host: "localhost"
  port: 5432
  database: "talos"
  user: "talos_user"
  password: "your_secure_password"
  ssl_mode: "disable"  # Use "require" in production
```

### Environment Variables (Alternative)

```bash
# Windows PowerShell
$env:DATABASE_URL="postgres://talos_user:password@localhost:5432/talos?sslmode=disable"
$env:DB_TYPE="postgres"

# Linux/Mac
export DATABASE_URL="postgres://talos_user:password@localhost:5432/talos?sslmode=disable"
export DB_TYPE="postgres"
```

## Running Talos with PostgreSQL

```bash
# Run Guardian
go run cmd/atlas/main.go

# Run Dashboard
go run cmd/dashboard/main.go
```

## Production Deployment

### Docker Compose

```yaml
version: '3.8'
services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: talos
      POSTGRES_USER: talos_user
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
  
  talos:
    build: .
    depends_on:
      - postgres
    environment:
      DATABASE_URL: postgres://talos_user:${DB_PASSWORD}@postgres:5432/talos?sslmode=disable
    ports:
      - "8080:8080"

volumes:
  postgres_data:
```

### Security Best Practices

1. **Use SSL in Production**

   ```yaml
   ssl_mode: "require"
   ```

2. **Strong Passwords**
   - Use 32+ character passwords
   - Store in environment variables or secrets manager

3. **Connection Pooling**
   - Default: 4 connections per worker
   - Adjust based on load

4. **Backup Strategy**

   ```bash
   # Daily backups
   pg_dump -U talos_user talos > backup_$(date +%Y%m%d).sql
   ```

## Troubleshooting

### Connection Refused

```bash
# Check PostgreSQL is running
pg_isready -h localhost -p 5432

# Check firewall
netsh advfirewall firewall add rule name="PostgreSQL" dir=in action=allow protocol=TCP localport=5432
```

### Permission Denied

```sql
-- Grant all privileges
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO talos_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO talos_user;
```

### Migration Errors

```bash
# Drop and recreate (DEV ONLY)
psql -U postgres -c "DROP DATABASE talos;"
psql -U postgres -c "CREATE DATABASE talos;"
go run cmd/migrate/main.go up
```

## Performance Tuning

### Indexes

All critical indexes are created automatically:

- `idx_actions_status` - Fast pending action queries
- `idx_actions_checksum` - Idempotency lookups
- `idx_ai_decisions_resource` - Historical context retrieval

### Query Optimization

```sql
-- Analyze query performance
EXPLAIN ANALYZE SELECT * FROM actions WHERE status = 'PENDING';

-- Update statistics
ANALYZE actions;
```

## Monitoring

### Connection Stats

```sql
SELECT * FROM pg_stat_activity WHERE datname = 'talos';
```

### Table Sizes

```sql
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```
