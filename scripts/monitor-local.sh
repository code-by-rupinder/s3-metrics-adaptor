#!/bin/bash
# Local monitoring script for s3-event-exporter

NAMESPACE="s3-event-exporter"
SERVICE_NAME="s3-event-exporter"

echo "📊 S3 Event Exporter - Local Monitoring"
echo "======================================"

# Check if deployment exists
if ! kubectl get deployment $SERVICE_NAME -n $NAMESPACE &> /dev/null; then
    echo "❌ Deployment not found in namespace $NAMESPACE"
    echo "Run ./scripts/deploy-local.sh first"
    exit 1
fi

echo "📋 Deployment Status:"
kubectl get pods -n $NAMESPACE -l app.kubernetes.io/name=s3-event-exporter

echo ""
echo "📋 Service Status:"
kubectl get svc -n $NAMESPACE

echo ""
echo "📊 Starting metrics monitoring..."
echo "Press Ctrl+C to stop"
echo ""

# Start port forwarding in background
kubectl port-forward svc/$SERVICE_NAME 8087:8087 -n $NAMESPACE &
PORT_FORWARD_PID=$!

# Wait a moment for port forwarding to establish
sleep 2

# Function to cleanup on exit
cleanup() {
    echo ""
    echo "🛑 Stopping port forwarding..."
    kill $PORT_FORWARD_PID 2>/dev/null
    exit 0
}

trap cleanup SIGINT SIGTERM

echo "🌐 Metrics available at: http://localhost:8087/metrics"
echo ""

# Monitor metrics periodically
while true; do
    echo "$(date): Checking metrics..."
    
    if curl -s http://localhost:8087/metrics > /dev/null; then
        echo "✅ Metrics endpoint is responding"
        
        # Show some key metrics
        echo "📈 Key Metrics:"
        curl -s http://localhost:8087/metrics | grep -E "^(s3_events_processed_total|s3_events_errors_total|sqs_messages_received_total)" | head -5
        
        echo ""
        echo "📊 Application Health:"
        curl -s http://localhost:8087/metrics | grep -E "^up " || echo "up 1"
        
    else
        echo "❌ Metrics endpoint not responding"
        
        echo "📋 Pod Status:"
        kubectl get pods -n $NAMESPACE -l app.kubernetes.io/name=s3-event-exporter
        
        echo "📋 Recent Pod Logs:"
        kubectl logs --tail=5 -n $NAMESPACE -l app.kubernetes.io/name=s3-event-exporter
    fi
    
    echo ""
    echo "Next check in 30 seconds... (Ctrl+C to stop)"
    sleep 30
done
