#!/bin/bash

# Talos Enterprise Cloud Optimization - Azure Deployment Script
# This script deploys Talos to Microsoft Azure

set -e

# Configuration
RESOURCE_GROUP=${RESOURCE_GROUP:-"talos-rg"}
LOCATION=${LOCATION:-"eastus"}
CLUSTER_NAME=${CLUSTER_NAME:-"talos-aks"}
NAMESPACE="talos"
ACR_NAME=${ACR_NAME:-"talosregistry"}

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
    
    # Check Azure CLI
    if ! command -v az &> /dev/null; then
        log_error "Azure CLI is not installed. Please install it first."
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
    
    # Check Azure login
    if ! az account show &> /dev/null; then
        log_error "Azure login required. Please run 'az login'."
        exit 1
    fi
    
    log_info "Prerequisites check passed!"
}

# Create resource group
create_resource_group() {
    log_info "Creating resource group..."
    
    az group create \
        --name "$RESOURCE_GROUP" \
        --location "$LOCATION" \
        --tags "project=talos" "environment=production" || log_warn "Resource group might already exist"
    
    log_info "Resource group created/verified"
}

# Create Azure Container Registry
create_acr() {
    log_info "Creating Azure Container Registry..."
    
    az acr create \
        --resource-group "$RESOURCE_GROUP" \
        --name "$ACR_NAME" \
        --sku Premium \
        --admin-enabled true \
        --tags "project=talos" || log_warn "ACR might already exist"
    
    # Get ACR login server
    ACR_LOGIN_SERVER=$(az acr show \
        --resource-group "$RESOURCE_GROUP" \
        --name "$ACR_NAME" \
        --query loginServer \
        --output tsv)
    
    log_info "ACR Login Server: $ACR_LOGIN_SERVER"
}

# Build and push to ACR
build_and_push() {
    log_info "Building and pushing to ACR..."
    
    # Login to ACR
    az acr login --name "$ACR_NAME"
    
    # Build image
    docker build -f Dockerfile.production -t "$ACR_NAME.azurecr.io/talos:latest" .
    
    # Push image
    docker push "$ACR_NAME.azurecr.io/talos:latest"
    
    log_info "Image pushed to ACR successfully"
}

# Create AKS cluster
create_aks_cluster() {
    log_info "Creating AKS cluster..."
    
    az aks create \
        --resource-group "$RESOURCE_GROUP" \
        --name "$CLUSTER_NAME" \
        --node-count 3 \
        --node-vm-size Standard_B2s \
        --enable-addons monitoring,ingress-appgw,http_application_routing \
        --generate-ssh-keys \
        --enable-cluster-autoscaler \
        --min-count 1 \
        --max-count 5 \
        --network-plugin azure \
        --service-cidr 10.0.0.0/16 \
        --dns-service-ip 10.0.0.10 \
        --docker-bridge-address 172.17.0.1/16 \
        --tags "project=talos" || log_warn "AKS cluster might already exist"
    
    # Get cluster credentials
    az aks get-credentials \
        --resource-group "$RESOURCE_GROUP" \
        --name "$CLUSTER_NAME" \
        --overwrite-existing
    
    log_info "AKS cluster created and credentials configured"
}

# Deploy to AKS
deploy_to_aks() {
    log_info "Deploying to AKS..."
    
    # Create namespace
    kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -
    
    # Apply secrets
    kubectl apply -f k8s/talos-secrets.yaml -n "$NAMESPACE"
    
    # Deploy PostgreSQL
    kubectl apply -f k8s/postgres-deployment.yaml -n "$NAMESPACE"
    
    # Deploy Redis
    kubectl apply -f k8s/redis-deployment.yaml -n "$NAMESPACE"
    
    # Update image in deployment
    sed "s|image: talos:latest|image: $ACR_NAME.azurecr.io/talos:latest|g" k8s/talos-deployment.yaml | kubectl apply -f - -n "$NAMESPACE"
    
    # Wait for deployments
    kubectl wait --for=condition=available --timeout=300s deployment/postgres -n "$NAMESPACE"
    kubectl wait --for=condition=available --timeout=300s deployment/redis -n "$NAMESPACE"
    kubectl wait --for=condition=available --timeout=300s deployment/talos-app -n "$NAMESPACE"
    
    log_info "Application deployed successfully"
}

# Set up Azure Database for PostgreSQL
setup_azure_postgresql() {
    log_info "Setting up Azure Database for PostgreSQL..."
    
    az postgres server create \
        --resource-group "$RESOURCE_GROUP" \
        --name "talos-postgres" \
        --location "$LOCATION" \
        --admin-user "talosadmin" \
        --admin-password "$(openssl rand -base64 32)" \
        --sku-name B_Gen5_2 \
        --version 15 \
        --storage-size 51200 \
        --backup-retention 7 \
        --geo-redundant-backup Enabled \
        --tags "project=talos" || log_warn "PostgreSQL server might already exist"
    
    # Configure firewall rules
    az postgres server firewall-rule create \
        --resource-group "$RESOURCE_GROUP" \
        --server-name "talos-postgres" \
        --name "AllowAllAzureIPs" \
        --start-ip-address 0.0.0.0 \
        --end-ip-address 0.0.0.0 || log_warn "Firewall rule might already exist"
    
    # Create database
    az postgres db create \
        --resource-group "$RESOURCE_GROUP" \
        --server-name "talos-postgres" \
        --name "talos" || log_warn "Database might already exist"
    
    log_info "Azure PostgreSQL set up successfully"
}

# Set up Azure Cache for Redis
setup_azure_redis() {
    log_info "Setting up Azure Cache for Redis..."
    
    az redis create \
        --resource-group "$RESOURCE_GROUP" \
        --name "talos-redis" \
        --location "$LOCATION" \
        --sku Basic \
        --vm-size C0 \
        --tags "project=talos" || log_warn "Redis cache might already exist"
    
    # Get Redis connection string
    REDIS_CONNECTION_STRING=$(az redis show \
        --resource-group "$RESOURCE_GROUP" \
        --name "talos-redis" \
        --query hostName \
        --output tsv)
    
    log_info "Azure Redis set up successfully"
}

# Set up Application Gateway
setup_app_gateway() {
    log_info "Setting up Application Gateway..."
    
    # Create public IP
    az network public-ip create \
        --resource-group "$RESOURCE_GROUP" \
        --name "talos-pip" \
        --location "$LOCATION" \
        --sku Standard \
        --allocation-method Static \
        --tags "project=talos" || log_warn "Public IP might already exist"
    
    # Get public IP address
    PUBLIC_IP=$(az network public-ip show \
        --resource-group "$RESOURCE_GROUP" \
        --name "talos-pip" \
        --query ipAddress \
        --output tsv)
    
    log_info "Application Gateway set up with public IP: $PUBLIC_IP"
}

# Set up Azure Monitor
setup_monitoring() {
    log_info "Setting up Azure Monitor..."
    
    # Create Log Analytics workspace
    az monitor log-analytics workspace create \
        --resource-group "$RESOURCE_GROUP" \
        --workspace-name "talos-law" \
        --location "$LOCATION" \
        --tags "project=talos" || log_warn "Log Analytics workspace might already exist"
    
    # Create Application Insights
    az monitor app-insights component create \
        --resource-group "$RESOURCE_GROUP" \
        --app "talos-appinsights" \
        --location "$LOCATION" \
        --application-type web \
        --tags "project=talos" || log_warn "Application Insights might already exist"
    
    # Create alert rules
    az monitor metrics alert create \
        --resource-group "$RESOURCE_GROUP" \
        --name "TalosHighCPU" \
        --scopes "/subscriptions/$(az account show --query id -o tsv)/resourceGroups/$RESOURCE_GROUP/providers/Microsoft.ContainerService/managedClusters/$CLUSTER_NAME" \
        --condition "avg PercentageCpu > 80" \
        --window-size 5m \
        --evaluation-frequency 1m \
        --description "High CPU usage alert" \
        --severity 2 \
        --tags "project=talos" || log_warn "Alert rule might already exist"
    
    log_info "Azure Monitor set up successfully"
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
    log_info "Starting Talos Enterprise deployment to Azure..."
    
    check_prerequisites
    create_resource_group
    create_acr
    build_and_push
    create_aks_cluster
    setup_azure_postgresql
    setup_azure_redis
    deploy_to_aks
    setup_app_gateway
    setup_monitoring
    get_external_ip
    
    log_info "🎉 Talos Enterprise deployment to Azure completed successfully!"
}

# Handle script interruption
trap 'log_info "Deployment interrupted"' EXIT

# Run main function
main "$@"
