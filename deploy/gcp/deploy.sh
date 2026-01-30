#!/bin/bash

# Talos Enterprise Cloud Optimization - GCP Deployment Script
# This script deploys Talos to Google Cloud Platform

set -e

# Configuration
PROJECT_ID=${PROJECT_ID:-"your-gcp-project-id"}
REGION=${REGION:-"us-central1"}
ZONE=${ZONE:-"us-central1-a"}
CLUSTER_NAME=${CLUSTER_NAME:-"talos-cluster"}
NAMESPACE="talos"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check gcloud CLI
    if ! command -v gcloud &> /dev/null; then
        log_error "Google Cloud CLI is not installed. Please install it first."
        exit 1
    fi
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed. Please install it first."
        exit 1
    fi
    
    # Check kubectl
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed. Please install it first."
        exit 1
    fi
    
    # Set project
    gcloud config set project "$PROJECT_ID"
    
    log_info "Prerequisites check passed!"
}

# Enable required APIs
enable_apis() {
    log_info "Enabling required GCP APIs..."
    
    gcloud services enable \
        container.googleapis.com \
        cloudbuild.googleapis.com \
        artifactregistry.googleapis.com \
        run.googleapis.com \
        monitoring.googleapis.com \
        logging.googleapis.com \
        sqladmin.googleapis.com \
        redis.googleapis.com
    
    log_info "APIs enabled successfully"
}

# Build and push to Artifact Registry
build_and_push() {
    log_info "Building and pushing to Artifact Registry..."
    
    # Configure Docker for GCP
    gcloud auth configure-docker "$REGION-docker.pkg.dev"
    
    # Build image
    docker build -f Dockerfile.production -t "$REGION-docker.pkg.dev/$PROJECT_ID/talos/talos:latest" .
    
    # Push image
    docker push "$REGION-docker.pkg.dev/$PROJECT_ID/talos/talos:latest"
    
    log_info "Image pushed to Artifact Registry"
}

# Create GKE cluster
create_gke_cluster() {
    log_info "Creating GKE cluster..."
    
    gcloud container clusters create "$CLUSTER_NAME" \
        --region "$REGION" \
        --node-locations "$ZONE" \
        --num-nodes 3 \
        --machine-type "e2-standard-2" \
        --enable-autoscaling \
        --min-nodes 1 \
        --max-nodes 10 \
        --enable-autorepair \
        --enable-autoupgrade \
        --enable-ip-alias \
        --enable-stackdriver-kubernetes \
        --enable-cloud-logging \
        --enable-cloud-monitoring \
        --enable-autoscaling \
        --workload-pool "$PROJECT_ID.svc.id.goog" \
        --cluster-version "latest" || log_warn "Cluster might already exist"
    
    # Get cluster credentials
    gcloud container clusters get-credentials "$CLUSTER_NAME" --region "$REGION"
    
    log_info "GKE cluster created and credentials configured"
}

# Deploy to GKE
deploy_to_gke() {
    log_info "Deploying to GKE..."
    
    # Create namespace
    kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -
    
    # Apply secrets
    kubectl apply -f k8s/talos-secrets.yaml -n "$NAMESPACE"
    
    # Deploy PostgreSQL
    kubectl apply -f k8s/postgres-deployment.yaml -n "$NAMESPACE"
    
    # Deploy Redis
    kubectl apply -f k8s/redis-deployment.yaml -n "$NAMESPACE"
    
    # Deploy Talos application
    sed "s|image: talos:latest|image: $REGION-docker.pkg.dev/$PROJECT_ID/talos/talos:latest|g" k8s/talos-deployment.yaml | kubectl apply -f - -n "$NAMESPACE"
    
    # Wait for deployments
    kubectl wait --for=condition=available --timeout=300s deployment/postgres -n "$NAMESPACE"
    kubectl wait --for=condition=available --timeout=300s deployment/redis -n "$NAMESPACE"
    kubectl wait --for=condition=available --timeout=300s deployment/talos-app -n "$NAMESPACE"
    
    log_info "Application deployed successfully"
}

# Set up Cloud SQL
setup_cloud_sql() {
    log_info "Setting up Cloud SQL..."
    
    # Create Cloud SQL instance
    gcloud sql instances create talos-postgres \
        --database-version=POSTGRES_15 \
        --tier=db-custom-2-7680 \
        --region="$REGION" \
        --storage-size=100GB \
        --storage-type=SSD \
        --backup-start-time=02:00 \
        --retained-backups-count=7 \
        --retained-transaction-log-days=7 \
        --enable-bin-log \
        --authorized-networks=0.0.0.0/0 || log_warn "SQL instance might already exist"
    
    # Create database
    gcloud sql databases create talos --instance=talos-postgres || log_warn "Database might already exist"
    
    # Create database user
    gcloud sql users create talos --instance=talos-postgres --password=$(openssl rand -base64 32) || log_warn "User might already exist"
    
    log_info "Cloud SQL set up successfully"
}

# Set up Memorystore (Redis)
setup_memorystore() {
    log_info "Setting up Memorystore for Redis..."
    
    gcloud redis instances create talos-redis \
        --region="$REGION" \
        --tier=STANDARD_HA \
        --memory-size-gb=4 \
        --replica-count=2 \
        --redis-version=redis_7_0 \
        --display-name="Talos Redis" || log_warn "Redis instance might already exist"
    
    log_info "Memorystore set up successfully"
}

# Set up Cloud Run alternative
deploy_cloud_run() {
    log_info "Deploying to Cloud Run (alternative)..."
    
    gcloud run deploy talos \
        --image="$REGION-docker.pkg.dev/$PROJECT_ID/talos/talos:latest" \
        --region="$REGION" \
        --platform=managed \
        --allow-unauthenticated \
        --memory=512Mi \
        --cpu=1 \
        --max-instances=10 \
        --min-instances=1 \
        --set-env-vars="GIN_MODE=release,TALOS_ENV=production" \
        --set-secrets="OPENROUTER_API_KEY=talos-secrets:OPENROUTER_API_KEY" \
        --port=8080 \
        --timeout=300s
    
    log_info "Cloud Run deployment completed"
}

# Set up monitoring
setup_monitoring() {
    log_info "Setting up monitoring..."
    
    # Create Cloud Monitoring dashboard
    gcloud monitoring dashboards create --config-from-file=monitoring/gcp-dashboard.json || log_warn "Dashboard creation failed"
    
    # Create alert policies
    gcloud alpha monitoring policies create --policy-from-file=monitoring/gcp-alerts.json || log_warn "Alert policies creation failed"
    
    # Enable Log-based metrics
    gcloud logging metrics create talos_errors \
        --description="Count of Talos application errors" \
        --log-filter='resource.type="k8s_container" AND resource.labels.namespace_name="talos" AND severity="ERROR"' \
        --bucket-name=_Default
    
    log_info "Monitoring set up successfully"
}

# Get external IP
get_external_ip() {
    log_info "Getting external IP..."
    
    # Wait for external IP
    kubectl wait --for=condition=ready --timeout=300s service/talos-service -n "$NAMESPACE"
    
    # Get external IP
    EXTERNAL_IP=$(kubectl get service talos-service -n "$NAMESPACE" -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
    
    if [ -z "$EXTERNAL_IP" ]; then
        log_warn "External IP not available yet. Please check manually."
    else
        log_info "🎉 Talos is accessible at: http://$EXTERNAL_IP"
    fi
}

# Main deployment function
main() {
    log_info "Starting Talos Enterprise deployment to GCP..."
    
    check_prerequisites
    enable_apis
    build_and_push
    create_gke_cluster
    setup_cloud_sql
    setup_memorystore
    deploy_to_gke
    setup_monitoring
    get_external_ip
    
    log_info "🎉 Talos Enterprise deployment to GCP completed successfully!"
}

# Handle script interruption
trap 'log_info "Deployment interrupted"' EXIT

# Run main function
main "$@"
