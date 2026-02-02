package deployment

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// DeploymentManager handles application deployment
type DeploymentManager struct {
	kubeClient *kubernetes.Clientset
	namespace  string
	logger     interface{} // Simplified logger interface
}

// NewDeploymentManager creates a new deployment manager
func NewDeploymentManager(kubeconfigPath, namespace string, logger interface{}) (*DeploymentManager, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return &DeploymentManager{
		kubeClient: clientset,
		namespace:  namespace,
		logger:     logger,
	}, nil
}

// DockerComposeConfig represents Docker Compose configuration
type DockerComposeConfig struct {
	Version  string                   `yaml:"version"`
	Services map[string]ServiceConfig `yaml:"services"`
	Volumes  map[string]VolumeConfig  `yaml:"volumes"`
	Networks map[string]NetworkConfig `yaml:"networks"`
}

// ServiceConfig represents a Docker Compose service
type ServiceConfig struct {
	Image       string            `yaml:"image"`
	Ports       []string          `yaml:"ports"`
	Environment map[string]string `yaml:"environment"`
	Command     []string          `yaml:"command,omitempty"`
	Volumes     []string          `yaml:"volumes,omitempty"`
	DependsOn   []string          `yaml:"depends_on"`
	HealthCheck *HealthCheck      `yaml:"healthcheck"`
	Resources   *ResourceLimits   `yaml:"resources"`
	Restart     string            `yaml:"restart"`
}

// HealthCheck represents health check configuration
type HealthCheck struct {
	Test     []string `yaml:"test"`
	Interval string   `yaml:"interval"`
	Timeout  string   `yaml:"timeout"`
	Retries  int      `yaml:"retries"`
}

// ResourceLimits represents resource limits
type ResourceLimits struct {
	Limits       map[string]string `yaml:"limits"`
	Reservations map[string]string `yaml:"reservations"`
}

// VolumeConfig represents volume configuration
type VolumeConfig struct {
	Driver string `yaml:"driver"`
}

// NetworkConfig represents network configuration
type NetworkConfig struct {
	Driver string `yaml:"driver"`
}

// GenerateDockerCompose generates Docker Compose configuration
func (dm *DeploymentManager) GenerateDockerCompose() (*DockerComposeConfig, error) {
	config := &DockerComposeConfig{
		Version:  "3.8",
		Services: make(map[string]ServiceConfig),
		Volumes: map[string]VolumeConfig{
			"redis_data":     {Driver: "local"},
			"postgres_data":  {Driver: "local"},
			"analytics_data": {Driver: "local"},
		},
		Networks: map[string]NetworkConfig{
			"talos_network": {Driver: "bridge"},
		},
	}

	// Dashboard service
	config.Services["dashboard"] = ServiceConfig{
		Image: "talos/dashboard:latest",
		Ports: []string{"8080:8080"},
		Environment: map[string]string{
			"PORT":           "8080",
			"MODE":           "production",
			"REDIS_ADDRESS":  "redis:6379",
			"DB_HOST":        "postgres",
			"DB_PORT":        "5432",
			"DB_NAME":        "talos",
			"JWT_SECRET":     "${JWT_SECRET}",
			"CLOUD_PROVIDER": "${CLOUD_PROVIDER:-aws}",
			"CLOUD_REGION":   "${CLOUD_REGION:-us-east-1}",
			"CLOUD_DRY_RUN":  "${CLOUD_DRY_RUN:-true}",
		},
		Volumes: []string{
			"./config:/app/config:ro",
			"./logs:/app/logs",
		},
		DependsOn: []string{"redis", "postgres"},
		HealthCheck: &HealthCheck{
			Test:     []string{"CMD", "curl", "-f", "http://localhost:8080/health"},
			Interval: "30s",
			Timeout:  "10s",
			Retries:  3,
		},
		Resources: &ResourceLimits{
			Limits: map[string]string{
				"memory": "512M",
				"cpus":   "0.5",
			},
		},
		Restart: "unless-stopped",
	}

	// Worker service
	config.Services["worker"] = ServiceConfig{
		Image: "talos/worker:latest",
		Environment: map[string]string{
			"REDIS_ADDRESS":  "redis:6379",
			"DB_HOST":        "postgres",
			"DB_PORT":        "5432",
			"DB_NAME":        "talos",
			"WORKER_ID":      "${WORKER_ID:-worker-1}",
			"CLOUD_PROVIDER": "${CLOUD_PROVIDER:-aws}",
			"CLOUD_REGION":   "${CLOUD_REGION:-us-east-1}",
			"CLOUD_DRY_RUN":  "${CLOUD_DRY_RUN:-true}",
		},
		Volumes: []string{
			"./config:/app/config:ro",
			"./logs:/app/logs",
		},
		DependsOn: []string{"redis", "postgres"},
		Resources: &ResourceLimits{
			Limits: map[string]string{
				"memory": "1G",
				"cpus":   "1.0",
			},
		},
		Restart: "unless-stopped",
	}

	// Redis service
	config.Services["redis"] = ServiceConfig{
		Image:   "redis:7-alpine",
		Ports:   []string{"6379:6379"},
		Volumes: []string{"redis_data:/data"},
		HealthCheck: &HealthCheck{
			Test:     []string{"CMD", "redis-cli", "ping"},
			Interval: "30s",
			Timeout:  "10s",
			Retries:  3,
		},
		Resources: &ResourceLimits{
			Limits: map[string]string{
				"memory": "256M",
				"cpus":   "0.25",
			},
		},
		Restart: "unless-stopped",
	}

	// PostgreSQL service
	config.Services["postgres"] = ServiceConfig{
		Image: "postgres:15-alpine",
		Ports: []string{"5432:5432"},
		Environment: map[string]string{
			"POSTGRES_DB":       "talos",
			"POSTGRES_USER":     "talos",
			"POSTGRES_PASSWORD": "${POSTGRES_PASSWORD}",
		},
		Volumes: []string{"postgres_data:/var/lib/postgresql/data"},
		HealthCheck: &HealthCheck{
			Test:     []string{"CMD-SHELL", "pg_isready -U talos -d talos"},
			Interval: "30s",
			Timeout:  "10s",
			Retries:  3,
		},
		Resources: &ResourceLimits{
			Limits: map[string]string{
				"memory": "1G",
				"cpus":   "0.5",
			},
		},
		Restart: "unless-stopped",
	}

	// Prometheus service
	config.Services["prometheus"] = ServiceConfig{
		Image: "prom/prometheus:latest",
		Ports: []string{"9090:9090"},
		Volumes: []string{
			"./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml:ro",
			"prometheus_data:/prometheus",
		},
		Restart: "unless-stopped",
	}

	// Grafana service
	config.Services["grafana"] = ServiceConfig{
		Image: "grafana/grafana:latest",
		Ports: []string{"3000:3000"},
		Environment: map[string]string{
			"GF_SECURITY_ADMIN_PASSWORD": "${GRAFANA_PASSWORD:-admin}",
			"GF_USERS_ALLOW_SIGN_UP":     "false",
		},
		Volumes: []string{
			"grafana_data:/var/lib/grafana",
			"./monitoring/grafana/dashboards:/etc/grafana/provisioning/dashboards:ro",
			"./monitoring/grafana/datasources:/etc/grafana/provisioning/datasources:ro",
		},
		DependsOn: []string{"prometheus"},
		Restart:   "unless-stopped",
	}

	return config, nil
}

// SaveDockerCompose saves Docker Compose configuration to file
func (dm *DeploymentManager) SaveDockerCompose(config *DockerComposeConfig, outputPath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal docker-compose config: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write docker-compose file: %w", err)
	}

	return nil
}

// GenerateKubernetesManifests generates Kubernetes manifests
func (dm *DeploymentManager) GenerateKubernetesManifests() (map[string][]byte, error) {
	manifests := make(map[string][]byte)

	// Namespace
	namespaceData := fmt.Sprintf(`apiVersion: v1
kind: Namespace
metadata:
  name: %s
  labels:
    name: %s
`, dm.namespace, dm.namespace)
	manifests["namespace.yaml"] = []byte(namespaceData)

	// ConfigMap
	configMapData := fmt.Sprintf(`apiVersion: v1
kind: ConfigMap
metadata:
  name: talos-config
  namespace: %s
data:
  MODE: "production"
  REDIS_ADDRESS: "redis:6379"
  DB_HOST: "postgres"
  DB_PORT: "5432"
  DB_NAME: "talos"
  CLOUD_PROVIDER: "aws"
  CLOUD_REGION: "us-east-1"
`, dm.namespace)
	manifests["configmap.yaml"] = []byte(configMapData)

	// Secret
	secretData := fmt.Sprintf(`apiVersion: v1
kind: Secret
metadata:
  name: talos-secrets
  namespace: %s
type: Opaque
data:
  # Base64 encoded values - replace with actual encoded secrets
  jwt-secret: <base64-encoded-jwt-secret>
  postgres-password: <base64-encoded-postgres-password>
  openrouter-api-key: <base64-encoded-openrouter-key>
  gemini-api-key: <base64-encoded-gemini-key>
`, dm.namespace)
	manifests["secret.yaml"] = []byte(secretData)

	// Redis Deployment
	redisDeployment := fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
  namespace: %s
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
      - name: redis
        image: redis:7-alpine
        ports:
        - containerPort: 6379
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "250m"
        livenessProbe:
          exec:
            command:
            - redis-cli
            - ping
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          exec:
            command:
            - redis-cli
            - ping
          initialDelaySeconds: 5
          periodSeconds: 5
`, dm.namespace)
	manifests["redis-deployment.yaml"] = []byte(redisDeployment)

	// Dashboard Deployment
	dashboardDeployment := fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: dashboard
  namespace: %s
spec:
  replicas: 3
  selector:
    matchLabels:
      app: dashboard
  template:
    metadata:
      labels:
        app: dashboard
    spec:
      containers:
      - name: dashboard
        image: talos/dashboard:latest
        ports:
        - containerPort: 8080
        env:
        - name: PORT
          value: "8080"
        - name: REDIS_ADDRESS
          value: "redis:6379"
        envFrom:
        - configMapRef:
            name: talos-config
        - secretRef:
            name: talos-secrets
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
`, dm.namespace)
	manifests["dashboard-deployment.yaml"] = []byte(dashboardDeployment)

	// Services
	redisService := fmt.Sprintf(`apiVersion: v1
kind: Service
metadata:
  name: redis
  namespace: %s
spec:
  selector:
    app: redis
  ports:
  - port: 6379
    targetPort: 6379
  type: ClusterIP
`, dm.namespace)
	manifests["redis-service.yaml"] = []byte(redisService)

	dashboardService := fmt.Sprintf(`apiVersion: v1
kind: Service
metadata:
  name: dashboard
  namespace: %s
spec:
  selector:
    app: dashboard
  ports:
  - port: 8080
    targetPort: 8080
  type: LoadBalancer
`, dm.namespace)
	manifests["dashboard-service.yaml"] = []byte(dashboardService)

	// Horizontal Pod Autoscaler
	hpaData := fmt.Sprintf(`apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: dashboard-hpa
  namespace: %s
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: dashboard
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
`, dm.namespace)
	manifests["hpa.yaml"] = []byte(hpaData)

	return manifests, nil
}

// SaveKubernetesManifests saves Kubernetes manifests to files
func (dm *DeploymentManager) SaveKubernetesManifests(manifests map[string][]byte, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	for filename, data := range manifests {
		path := filepath.Join(outputDir, filename)
		if err := os.WriteFile(path, data, 0644); err != nil {
			return fmt.Errorf("failed to write manifest %s: %w", filename, err)
		}
	}

	return nil
}

// DeployToKubernetes deploys the application to Kubernetes
func (dm *DeploymentManager) DeployToKubernetes(ctx context.Context, manifests map[string][]byte) error {
	// This would implement actual Kubernetes deployment
	// For now, it's a placeholder that would use the kubernetes client
	// to create/update resources in the cluster

	fmt.Println("Deploying to Kubernetes...")
	for filename := range manifests {
		fmt.Printf("  - %s\n", filename)
	}

	return nil
}

// GenerateDeploymentPackage creates a complete deployment package
func (dm *DeploymentManager) GenerateDeploymentPackage(outputDir string) error {
	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate Docker Compose
	dockerCompose, err := dm.GenerateDockerCompose()
	if err != nil {
		return fmt.Errorf("failed to generate docker-compose: %w", err)
	}

	if err := dm.SaveDockerCompose(dockerCompose, filepath.Join(outputDir, "docker-compose.yml")); err != nil {
		return fmt.Errorf("failed to save docker-compose: %w", err)
	}

	// Generate Kubernetes manifests
	manifests, err := dm.GenerateKubernetesManifests()
	if err != nil {
		return fmt.Errorf("failed to generate kubernetes manifests: %w", err)
	}

	k8sDir := filepath.Join(outputDir, "kubernetes")
	if err := dm.SaveKubernetesManifests(manifests, k8sDir); err != nil {
		return fmt.Errorf("failed to save kubernetes manifests: %w", err)
	}

	// Generate environment file
	envFile := `# Environment Configuration
PORT=8080
MODE=production

# Database Configuration
DB_HOST=postgres
DB_PORT=5432
DB_NAME=talos
DB_USER=talos
POSTGRES_PASSWORD=your-secure-password

# Redis Configuration
REDIS_ADDRESS=redis:6379

# Security
JWT_SECRET=your-super-secret-jwt-key-at-least-32-characters-long

# AI Services
OPENROUTER_API_KEY=your-openrouter-api-key
GEMINI_API_KEY=your-gemini-api-key
CLAUDE_API_KEY=your-claude-api-key

# Cloud Configuration
CLOUD_PROVIDER=aws
CLOUD_REGION=us-east-1
CLOUD_DRY_RUN=false

# Monitoring
PROMETHEUS_ENABLED=true
JAEGER_ENABLED=false

# Grafana
GRAFANA_PASSWORD=your-grafana-password
`

	if err := os.WriteFile(filepath.Join(outputDir, ".env"), []byte(envFile), 0600); err != nil {
		return fmt.Errorf("failed to write environment file: %w", err)
	}

	// Generate deployment script
	deployScript := `#!/bin/bash

# Talos Cloud Guardian Deployment Script

set -e

echo "🚀 Deploying Talos Cloud Guardian..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Validate required environment variables
required_vars=("JWT_SECRET" "POSTGRES_PASSWORD")
for var in "${required_vars[@]}"; do
    if [ -z "${!var}" ]; then
        echo "❌ Required environment variable $var is not set"
        exit 1
    fi
done

# Create necessary directories
mkdir -p logs data/analytics

# Deploy with Docker Compose
echo "📦 Deploying with Docker Compose..."
docker-compose up -d

# Wait for services to be ready
echo "⏳ Waiting for services to be ready..."
sleep 30

# Check service health
echo "🏥 Checking service health..."
if curl -f http://localhost:8080/health > /dev/null 2>&1; then
    echo "✅ Dashboard is healthy"
else
    echo "❌ Dashboard is not healthy"
    exit 1
fi

echo "🎉 Deployment completed successfully!"
echo "📊 Dashboard: http://localhost:8080"
echo "📈 Prometheus: http://localhost:9090"
echo "📊 Grafana: http://localhost:3000 (admin/admin)"

echo "📝 Logs: docker-compose logs -f"
echo "🛑 Stop: docker-compose down"
`

	if err := os.WriteFile(filepath.Join(outputDir, "deploy.sh"), []byte(deployScript), 0755); err != nil {
		return fmt.Errorf("failed to write deployment script: %w", err)
	}

	// Generate README
	readme := `# Talos Cloud Guardian - Deployment Package

## Quick Start

### Prerequisites
- Docker and Docker Compose
- At least 4GB RAM and 2 CPU cores
- Valid cloud provider credentials (AWS, Azure, or GCP)

### Deployment

1. **Configure Environment**
   bash
   cp .env.example .env
   # Edit .env with your configuration
   

2. **Deploy**
   bash
   ./deploy.sh

3. **Access Services**
   - Dashboard: http://localhost:8080
   - Prometheus: http://localhost:9090
   - Grafana: http://localhost:3000

### Kubernetes Deployment

For Kubernetes deployment, use the manifests in the kubernetes/ directory:

bash
kubectl apply -f kubernetes/

### Monitoring

- **Prometheus**: Metrics collection and alerting
- **Grafana**: Visualization and dashboards
- **Health Checks**: Built-in health monitoring

### Configuration

Key environment variables:
- CLOUD_PROVIDER: aws, azure, or gcp
- CLOUD_REGION: Your preferred region
- CLOUD_DRY_RUN: Set to false for actual optimizations
- JWT_SECRET: Secure secret for authentication

### Support

For issues and questions:
1. Check logs: docker-compose logs -f
2. Verify health: curl http://localhost:8080/health
3. Review configuration in .env

---

**⚠️ Production Notes:**
- Change default passwords and secrets
- Enable HTTPS in production
- Configure backup strategies
- Set up proper monitoring and alerting
`

	if err := os.WriteFile(filepath.Join(outputDir, "README.md"), []byte(readme), 0644); err != nil {
		return fmt.Errorf("failed to write README: %w", err)
	}

	return nil
}
