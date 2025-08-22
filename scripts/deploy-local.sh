#!/bin/bash
# Local Kubernetes deployment script for s3-event-exporter

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 S3 Event Exporter - Local Kubernetes Deployment${NC}"
echo "=================================================="

# Check if kubectl is available and connected
echo -e "${YELLOW}📋 Checking Kubernetes cluster...${NC}"
if ! kubectl cluster-info &> /dev/null; then
    echo -e "${RED}❌ kubectl is not connected to a cluster${NC}"
    echo "Please start your local Kubernetes cluster (Docker Desktop, minikube, kind, etc.)"
    exit 1
fi

echo -e "${GREEN}✅ Kubernetes cluster is accessible${NC}"
kubectl get nodes

# Check if namespace exists, create if not
NAMESPACE="s3-event-exporter"
echo -e "${YELLOW}📋 Setting up namespace: ${NAMESPACE}${NC}"
if ! kubectl get namespace $NAMESPACE &> /dev/null; then
    kubectl create namespace $NAMESPACE
    echo -e "${GREEN}✅ Namespace ${NAMESPACE} created${NC}"
else
    echo -e "${GREEN}✅ Namespace ${NAMESPACE} already exists${NC}"
fi

# Check for AWS credentials
echo -e "${YELLOW}📋 Checking AWS credentials...${NC}"
if ! kubectl get secret aws-credentials -n $NAMESPACE &> /dev/null; then
    echo -e "${YELLOW}⚠️  AWS credentials secret not found${NC}"
    echo -e "${BLUE}Creating AWS credentials secret...${NC}"
    
    # Check for AWS credentials in environment
    if [[ -z "$AWS_ACCESS_KEY_ID" || -z "$AWS_SECRET_ACCESS_KEY" ]]; then
        echo -e "${RED}❌ AWS credentials not found in environment${NC}"
        echo "Please set your AWS credentials:"
        echo "export AWS_ACCESS_KEY_ID=your_access_key"
        echo "export AWS_SECRET_ACCESS_KEY=your_secret_key"
        echo "export AWS_REGION=us-west-2"
        echo ""
        echo "Or create the secret manually:"
        echo "kubectl create secret generic aws-credentials \\"
        echo "  --from-literal=AWS_ACCESS_KEY_ID=your_access_key \\"
        echo "  --from-literal=AWS_SECRET_ACCESS_KEY=your_secret_key \\"
        echo "  --from-literal=AWS_REGION=us-west-2 \\"
        echo "  --namespace $NAMESPACE"
        exit 1
    fi
    
    # Create the secret
    kubectl create secret generic aws-credentials \
        --from-literal=AWS_ACCESS_KEY_ID="$AWS_ACCESS_KEY_ID" \
        --from-literal=AWS_SECRET_ACCESS_KEY="$AWS_SECRET_ACCESS_KEY" \
        --from-literal=AWS_REGION="${AWS_REGION:-us-west-2}" \
        --namespace $NAMESPACE
    
    echo -e "${GREEN}✅ AWS credentials secret created${NC}"
else
    echo -e "${GREEN}✅ AWS credentials secret already exists${NC}"
fi

# Deploy the application
echo -e "${YELLOW}📋 Deploying S3 Event Exporter...${NC}"
helm upgrade --install s3-event-exporter ./helm/s3-event-exporter \
    --namespace $NAMESPACE \
    --values helm/examples/values-local-testing.yaml \
    --wait \
    --timeout 300s

echo -e "${GREEN}✅ Deployment completed${NC}"

# Wait for pods to be ready
echo -e "${YELLOW}📋 Waiting for pods to be ready...${NC}"
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=s3-event-exporter -n $NAMESPACE --timeout=120s

# Show deployment status
echo -e "${YELLOW}📋 Deployment Status:${NC}"
kubectl get pods -n $NAMESPACE -l app.kubernetes.io/name=s3-event-exporter

# Show service information
echo -e "${YELLOW}📋 Service Information:${NC}"
kubectl get svc -n $NAMESPACE

echo -e "${BLUE}🎉 Deployment Complete!${NC}"
echo ""
echo -e "${YELLOW}📊 To access metrics:${NC}"
echo "kubectl port-forward svc/s3-event-exporter 8087:8087 -n $NAMESPACE"
echo "curl http://localhost:8087/metrics"
echo ""
echo -e "${YELLOW}📋 To view logs:${NC}"
echo "kubectl logs -f deployment/s3-event-exporter -n $NAMESPACE"
echo ""
echo -e "${YELLOW}🔧 To update configuration:${NC}"
echo "Edit helm/examples/values-local-testing.yaml and run:"
echo "helm upgrade s3-event-exporter ./helm/s3-event-exporter -n $NAMESPACE -f helm/examples/values-local-testing.yaml"
echo ""
echo -e "${YELLOW}🗑️  To uninstall:${NC}"
echo "helm uninstall s3-event-exporter -n $NAMESPACE"
echo "kubectl delete namespace $NAMESPACE"
