# 🚀 Talos Enterprise Cloud Optimization API Documentation

## 📋 Overview

Talos provides a comprehensive RESTful API for cloud resource optimization, AI-powered decision making, and real-time monitoring. This documentation covers all available endpoints, request/response formats, and usage examples.

## 🔐 Authentication

All API endpoints require JWT authentication. Include the token in the Authorization header:

```http
Authorization: Bearer <your-jwt-token>
```

## 📊 API Endpoints

### 🎯 Dashboard & Monitoring

#### Get System Status
```http
GET /api/v1/system/status
```

**Response:**
```json
{
  "status": "operational",
  "mode": "DRY_RUN",
  "version": "v1.0.0-beta",
  "timestamp": "2026-01-28T23:00:00Z",
  "services": {
    "orchestrator": "healthy",
    "cloud_adapters": "healthy",
    "ai_services": "healthy"
  }
}
```

#### Get ROI Metrics
```http
GET /api/v1/analytics/roi
```

**Response:**
```json
{
  "total_cost_savings": 15420.50,
  "roi_percentage": 234.5,
  "optimizations_applied": 127,
  "monthly_trend": [
    {"month": "2024-01", "savings": 1200.00},
    {"month": "2024-02", "savings": 1450.75}
  ]
}
```

#### Get Token Usage Breakdown
```http
GET /api/v1/analytics/tokens
```

**Response:**
```json
{
  "model_breakdown": {
    "gemini-2.0-flash-exp": {
      "tokens_used": 125000,
      "cost_usd": 0.75,
      "requests": 450
    },
    "gpt-4o-mini": {
      "tokens_used": 89000,
      "cost_usd": 0.89,
      "requests": 234
    }
  },
  "total_tokens": 214000,
  "total_cost_usd": 1.64
}
```

#### Get Cloud Resources
```http
GET /api/v1/resources
```

**Query Parameters:**
- `provider` (optional): Filter by cloud provider (aws, azure, gcp, ibm, oracle)
- `type` (optional): Filter by resource type (ec2, rds, vm, storage)
- `region` (optional): Filter by region

**Response:**
```json
{
  "resources": [
    {
      "id": "i-1234567890abcdef0",
      "type": "ec2",
      "provider": "aws",
      "region": "us-east-1",
      "state": "running",
      "cpu_usage": 45.2,
      "memory_usage": 67.8,
      "cost_per_month": 125.00,
      "optimization_score": 78.5,
      "rightsizing_recommendation": "t3.large",
      "estimated_savings": 45.50,
      "tags": {
        "Environment": "production",
        "Application": "web-server"
      }
    }
  ],
  "total_count": 1,
  "total_monthly_cost": 125.00
}
```

### 🤖 AI Optimization

#### Analyze Resource
```http
POST /api/v1/ai/analyze
```

**Request Body:**
```json
{
  "resource_id": "i-1234567890abcdef0",
  "analysis_type": "optimization",
  "risk_tolerance": "medium",
  "include_cost_analysis": true
}
```

**Response:**
```json
{
  "analysis_id": "analysis_123456",
  "recommendations": [
    {
      "type": "resize",
      "action": "downsize",
      "target_instance": "t3.large",
      "confidence": 0.87,
      "estimated_savings": 45.50,
      "risk_level": "low",
      "reasoning": "CPU and memory usage consistently below 70% for 30 days"
    }
  ],
  "ai_model_used": "gemini-pro",
  "analysis_timestamp": "2026-01-28T23:00:00Z"
}
```

#### Apply Optimization
```http
POST /api/v1/optimizations/{resource_id}/apply
```

**Request Body:**
```json
{
  "action": "resize",
  "target_type": "t3.large",
  "schedule": "immediate",
  "confirmation_required": true
}
```

**Response:**
```json
{
  "optimization_id": "opt_789012",
  "status": "scheduled",
  "estimated_completion": "2026-01-28T23:05:00Z",
  "rollback_available": true,
  "confirmation_token": "abc123def456"
}
```

### 📈 Performance Metrics

#### Get Resource Metrics
```http
GET /api/v1/metrics/{resource_id}
```

**Query Parameters:**
- `start_time` (optional): ISO 8601 start time
- `end_time` (optional): ISO 8601 end time
- `granularity` (optional): 5m, 1h, 1d

**Response:**
```json
{
  "resource_id": "i-1234567890abcdef0",
  "metrics": {
    "cpu_utilization": [
      {"timestamp": "2026-01-28T22:00:00Z", "value": 45.2},
      {"timestamp": "2026-01-28T22:05:00Z", "value": 47.8}
    ],
    "memory_utilization": [
      {"timestamp": "2026-01-28T22:00:00Z", "value": 67.8},
      {"timestamp": "2026-01-28T22:05:00Z", "value": 69.1}
    ],
    "network_in": [
      {"timestamp": "2026-01-28T22:00:00Z", "value": 1024.5},
      {"timestamp": "2026-01-28T22:05:00Z", "value": 1156.2}
    ]
  },
  "period": {
    "start": "2026-01-28T22:00:00Z",
    "end": "2026-01-28T23:00:00Z"
  }
}
```

### 🔧 Configuration Management

#### Get Configuration
```http
GET /api/v1/config
```

**Response:**
```json
{
  "guardian": {
    "mode": "DRY_RUN",
    "approval_required": true,
    "auto_optimization": false
  },
  "ai": {
    "default_model": "gemini-pro",
    "risk_threshold": 0.7,
    "max_concurrent_requests": 10
  },
  "monitoring": {
    "scanning_interval": "5m",
    "metrics_retention": "30d",
    "alert_thresholds": {
      "cpu": 80.0,
      "memory": 85.0
    }
  }
}
```

#### Update Configuration
```http
PUT /api/v1/config
```

**Request Body:**
```json
{
  "guardian.mode": "PRODUCTION",
  "ai.risk_threshold": 0.8,
  "monitoring.scanning_interval": "10m"
}
```

### 🚨 Alerts & Notifications

#### Get Active Alerts
```http
GET /api/v1/alerts
```

**Response:**
```json
{
  "alerts": [
    {
      "id": "alert_123",
      "severity": "warning",
      "resource_id": "i-1234567890abcdef0",
      "type": "high_cpu",
      "message": "CPU usage above 80% for 15 minutes",
      "timestamp": "2026-01-28T22:45:00Z",
      "acknowledged": false
    }
  ]
}
```

#### Acknowledge Alert
```http
POST /api/v1/alerts/{alert_id}/acknowledge
```

## 🔄 WebSocket API

### Real-time Updates
Connect to WebSocket for real-time updates:

```javascript
const ws = new WebSocket('ws://localhost:8080/ws/updates');

ws.onmessage = function(event) {
  const data = JSON.parse(event.data);
  console.log('Update:', data);
};
```

**Message Types:**
- `resource_update`: Resource state changes
- `optimization_complete`: Optimization job completion
- `alert_triggered`: New alert generation
- `metrics_update`: Real-time metrics

## 📝 Error Handling

All API errors follow consistent format:

```json
{
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "Resource with ID i-1234567890abcdef0 not found",
    "details": {
      "resource_id": "i-1234567890abcdef0",
      "timestamp": "2026-01-28T23:00:00Z"
    }
  }
}
```

**Common Error Codes:**
- `UNAUTHORIZED`: Invalid or missing authentication
- `FORBIDDEN`: Insufficient permissions
- `RESOURCE_NOT_FOUND`: Resource does not exist
- `VALIDATION_ERROR`: Invalid request parameters
- `RATE_LIMIT_EXCEEDED`: Too many requests
- `INTERNAL_ERROR`: Server-side error

## 🚀 Rate Limiting

API requests are rate-limited:
- **Standard tier**: 100 requests/minute
- **Premium tier**: 1000 requests/minute
- **Enterprise tier**: Unlimited

Rate limit headers are included in responses:
```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1640758800
```

## 📚 SDK & Client Libraries

### Go Client
```go
import "github.com/project-atlas/atlas/client"

client := client.New("https://api.talos.io", "your-api-key")
resources, err := client.Resources.List()
```

### Python Client
```python
from talos_client import TalosClient

client = TalosClient(api_key="your-api-key")
resources = client.resources.list()
```

### JavaScript Client
```javascript
import { TalosClient } from '@talos/client';

const client = new TalosClient('your-api-key');
const resources = await client.resources.list();
```

## 🔍 Examples & Use Cases

### 1. Automated Resource Optimization
```bash
# Get all underutilized EC2 instances
curl -H "Authorization: Bearer $TOKEN" \
  "https://api.talos.io/v1/resources?type=ec2&optimization_score=lt:50"

# Apply optimization recommendations
curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"action": "resize", "target_type": "t3.large"}' \
  "https://api.talos.io/v1/optimizations/i-1234567890abcdef0/apply"
```

### 2. Cost Analysis & Reporting
```bash
# Get monthly cost breakdown
curl -H "Authorization: Bearer $TOKEN" \
  "https://api.talos.io/v1/analytics/costs?period=monthly"

# Export optimization report
curl -H "Authorization: Bearer $TOKEN" \
  "https://api.talos.io/v1/reports/optimizations?format=csv&period=30d"
```

### 3. Real-time Monitoring
```javascript
// WebSocket connection for live updates
const ws = new WebSocket('wss://api.talos.io/ws/updates');

ws.onmessage = (event) => {
  const update = JSON.parse(event.data);
  if (update.type === 'alert_triggered') {
    showAlert(update.data);
  }
};
```

## 🎯 Best Practices

1. **Authentication**: Always use HTTPS and rotate API keys regularly
2. **Rate Limiting**: Implement exponential backoff for failed requests
3. **Error Handling**: Always check response status and handle errors gracefully
4. **Pagination**: Use pagination for large resource lists
5. **Caching**: Cache configuration and static data to reduce API calls
6. **Webhooks**: Use webhooks for real-time notifications instead of polling

## 📞 Support

- **Documentation**: https://docs.talos.io
- **API Reference**: https://api.talos.io/docs
- **Support**: support@talos.io
- **Status Page**: https://status.talos.io

---

*Last updated: January 28, 2026*
