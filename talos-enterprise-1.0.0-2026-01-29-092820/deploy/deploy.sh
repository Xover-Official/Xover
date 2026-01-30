#!/bin/bash

# Talos Enterprise Cloud Optimization - AWS Deployment Script
# This script deploys Talos to AWS ECS with full infrastructure

set -e

# Configuration
AWS_REGION=${AWS_REGION:-"us-east-1"}
ECR_REPOSITORY=${ECR_REPOSITORY:-"talos-enterprise"}
ENVIRONMENT=${ENVIRONMENT:-"production"}
TAG=${TAG:-"latest"}

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
    
    # Check AWS CLI
    if ! command -v aws &> /dev/null; then
        log_error "AWS CLI is not installed. Please install it first."
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
    
    # Check AWS credentials
    if ! aws sts get-caller-identity &> /dev/null; then
        log_error "AWS credentials are not configured. Please run 'aws configure'."
        exit 1
    fi
    
    log_info "Prerequisites check passed!"
}

# Create ECR repository
create_ecr_repository() {
    log_info "Creating ECR repository..."
    
    aws ecr create-repository \
        --repository-name "$ECR_REPOSITORY" \
        --region "$AWS_REGION" \
        --image-scanning-configuration scanOnPush=true \
        --image-tag-mutability MUTABLE || log_warn "Repository might already exist"
    
    # Get repository URI
    ECR_URI=$(aws ecr describe-repositories \
        --repository-names "$ECR_REPOSITORY" \
        --region "$AWS_REGION" \
        --query 'repositories[0].repositoryUri' \
        --output text)
    
    log_info "ECR Repository URI: $ECR_URI"
}

# Build and push Docker image
build_and_push_image() {
    log_info "Building Docker image..."
    
    # Login to ECR
    aws ecr get-login-password --region "$AWS_REGION" | docker login --username AWS --password-stdin "$ECR_URI"
    
    # Build image
    docker build -f Dockerfile.production -t "$ECR_REPOSITORY:$TAG" .
    
    # Tag image
    docker tag "$ECR_REPOSITORY:$TAG" "$ECR_URI:$TAG"
    
    # Push image
    log_info "Pushing image to ECR..."
    docker push "$ECR_URI:$TAG"
    
    log_info "Image pushed successfully: $ECR_URI:$TAG"
}

# Deploy to ECS
deploy_to_ecs() {
    log_info "Deploying to ECS..."
    
    # Create ECS cluster
    aws ecs create-cluster \
        --cluster-name "talos-cluster" \
        --region "$AWS_REGION" \
        --capacity-providers FARGATE,FARGATE_SPOT || log_warn "Cluster might already exist"
    
    # Create task definition
    cat > task-definition.json << EOF
{
    "family": "talos-task",
    "networkMode": "awsvpc",
    "requiresCompatibilities": ["FARGATE"],
    "cpu": "512",
    "memory": "1024",
    "executionRoleArn": "arn:aws:iam::$(aws sts get-caller-identity --query Account --output text):role/ecsTaskExecutionRole",
    "taskRoleArn": "arn:aws:iam::$(aws sts get-caller-identity --query Account --output text):role/ecsTaskRole",
    "containerDefinitions": [
        {
            "name": "talos",
            "image": "$ECR_URI:$TAG",
            "portMappings": [
                {
                    "containerPort": 8080,
                    "protocol": "tcp"
                }
            ],
            "environment": [
                {
                    "name": "GIN_MODE",
                    "value": "release"
                },
                {
                    "name": "TALOS_ENV",
                    "value": "$ENVIRONMENT"
                }
            ],
            "secrets": [
                {
                    "name": "OPENROUTER_API_KEY",
                    "valueFrom": "arn:aws:secretsmanager:$(aws sts get-caller-identity --query Account --output text):secret:talos-secrets:OPENROUTER_API_KEY"
                }
            ],
            "logConfiguration": {
                "logDriver": "awslogs",
                "options": {
                    "awslogs-group": "/ecs/talos",
                    "awslogs-region": "$AWS_REGION",
                    "awslogs-stream-prefix": "ecs"
                }
            },
            "healthCheck": {
                "command": ["CMD-SHELL", "curl -f http://localhost:8080/health || exit 1"],
                "interval": 30,
                "timeout": 5,
                "retries": 3
            }
        }
    ]
}
EOF
    
    # Register task definition
    TASK_DEF_ARN=$(aws ecs register-task-definition \
        --cli-input-json file://task-definition.json \
        --region "$AWS_REGION" \
        --query 'taskDefinition.taskDefinitionArn' \
        --output text)
    
    log_info "Task definition registered: $TASK_DEF_ARN"
    
    # Create service
    aws ecs create-service \
        --cluster "talos-cluster" \
        --service-name "talos-service" \
        --task-definition "$TASK_DEF_ARN" \
        --desired-count 2 \
        --launch-type FARGATE \
        --network-configuration "awsvpcConfiguration={subnets=[subnet-$(aws ec2 describe-subnets --filters Name=availability-zone,Values=us-east-1a --query 'Subnets[0].SubnetId' --output text)],securityGroups=[sg-$(aws ec2 describe-security-groups --group-names default --query 'SecurityGroups[0].GroupId' --output text)],assignPublicIp=ENABLED}" \
        --region "$AWS_REGION" || log_warn "Service might already exist"
    
    log_info "ECS service created/updated"
}

# Set up Application Load Balancer
setup_load_balancer() {
    log_info "Setting up Application Load Balancer..."
    
    # Create target group
    TARGET_GROUP_ARN=$(aws elbv2 create-target-group \
        --name "talos-targets" \
        --protocol HTTP \
        --port 8080 \
        --target-type ip \
        --vpc-id "$(aws ec2 describe-vpcs --query 'Vpcs[0].VpcId' --output text)" \
        --health-check-path "/health" \
        --region "$AWS_REGION" \
        --query 'TargetGroups[0].TargetGroupArn' \
        --output text)
    
    # Create load balancer
    LB_ARN=$(aws elbv2 create-load-balancer \
        --name "talos-alb" \
        --subnets "$(aws ec2 describe-subnets --filters Name=availability-zone,Values=us-east-1a,us-east-1b --query 'Subnets[*].SubnetId' --output text | tr '\n' ' ')" \
        --security-groups "$(aws ec2 describe-security-groups --group-names default --query 'SecurityGroups[0].GroupId' --output text)" \
        --region "$AWS_REGION" \
        --query 'LoadBalancers[0].LoadBalancerArn' \
        --output text)
    
    # Create listener
    aws elbv2 create-listener \
        --load-balancer-arn "$LB_ARN" \
        --protocol HTTP \
        --port 80 \
        --default-actions Type=forward,TargetGroupArn="$TARGET_GROUP_ARN" \
        --region "$AWS_REGION"
    
    log_info "Load balancer set up successfully"
}

# Deploy monitoring
deploy_monitoring() {
    log_info "Deploying monitoring stack..."
    
    # Deploy CloudWatch dashboard
    aws cloudwatch put-dashboard \
        --dashboard-name "Talos-Metrics" \
        --dashboard-body file://monitoring/cloudwatch-dashboard.json \
        --region "$AWS_REGION" || log_warn "Dashboard creation failed"
    
    # Create CloudWatch alarms
    aws cloudwatch put-metric-alarm \
        --alarm-name "Talos-HighCPU" \
        --alarm-description "High CPU usage detected" \
        --metric-name CPUUtilization \
        --namespace "AWS/ECS" \
        --statistic Average \
        --period 300 \
        --threshold 80 \
        --comparison-operator GreaterThanThreshold \
        --evaluation-periods 2 \
        --alarm-actions arn:aws:sns:$(aws sts get-caller-identity --query Account --output text):talos-alerts \
        --region "$AWS_REGION"
    
    log_info "Monitoring deployed"
}

# Clean up temporary files
cleanup() {
    rm -f task-definition.json
    log_info "Cleaned up temporary files"
}

# Main deployment function
main() {
    log_info "Starting Talos Enterprise deployment to AWS..."
    
    check_prerequisites
    create_ecr_repository
    build_and_push_image
    deploy_to_ecs
    setup_load_balancer
    deploy_monitoring
    cleanup
    
    log_info "🎉 Talos Enterprise deployment completed successfully!"
    log_info "Access your application at: http://$(aws elbv2 describe-load-balancers --names talos-alb --query 'LoadBalancers[0].DNSName' --output text)"
}

# Handle script interruption
trap cleanup EXIT

# Run main function
main "$@"
