# S3 Event Exporter - Docker Deployment Guide

This guide explains how to run the S3 Event Exporter as a Docker container, whether you want to use the pre-built image from Docker Hub or build it yourself.

## Prerequisites

- Docker installed and running
- AWS credentials configured (if monitoring real S3 buckets)
- An SQS queue already created to receive S3 event notifications
- An S3 bucket already created
- The S3 bucket must have EventBridge notifications enabled and configured to send events to the SQS queue
- Access to SQS queues that receive S3 event notifications (processed only if they originate from S3 EventBridge notifications; generic SQS events are not supported)

## Quick Start

### 1. Using Docker Hub Image

```bash
# Pull the latest image
docker pull codebyrupinder/s3-metrics-prom-adaptor:latest


# Run with your config and AWS credentials (for distroless nonroot image)
docker run -d \
  --name s3-metrics-adapter \
  -p 8087:8087 \
  -v /path/to/your/config.yaml:/app/config.yaml \
  -v ~/.aws:/home/nonroot/.aws:ro \
  codebyrupinder/s3-metrics-prom-adaptor:latest

# Alternative: Provide AWS credentials as environment variables
docker run -d \
  --name s3-metrics-adapter \
  -p 8087:8087 \
  -v /path/to/your/config.yaml:/app/config.yaml \
  -e AWS_ACCESS_KEY_ID=your-access-key \
  -e AWS_SECRET_ACCESS_KEY=your-secret-key \
  -e AWS_REGION=us-east-1 \
  codebyrupinder/s3-metrics-prom-adaptor:latest

# If running on an EC2 instance or ECS task with an attached IAM role,
# you do NOT need to provide credentials. The container will pick up credentials automatically.
```

### 2. Building from Source

```bash
# Clone the repository
git clone https://github.com/code-by-rupinder/s3-metrics-adaptor.git
cd s3-metrics-adapter

# Build the Docker image
docker build -t s3-metrics-prom-adaptor:local .

# Run the locally built image
docker run -d \
  --name s3-metrics-adapter \
  -p 8087:8087 \
  -v $(pwd)/config.yaml:/app/config.yaml \
  s3-metrics-prom-adaptor:local
```

## Configuration

### Environment Variables

You can override configuration values using environment variables:

```bash
docker run -d \
  --name s3-metrics-adapter \
  -p 8087:8087 \
  -e SQS_QUEUE_URL="https://sqs.us-west-2.amazonaws.com/123456789/your-queue" \
  -e S3_BUCKET_NAME="your-bucket-name" \
  -e METRICS_PORT="8087" \
  -e LOG_LEVEL="info" \
  codebyrupinder/s3-metrics-prom-adaptor:latest
```

### Custom Configuration File

Create a custom `config.yaml` file:

```yaml
logging:
  default: info
  components:
    eventparser: info
    metricsexporter: info
    sqspoller: info
  format:
    timestampFormat: "2006-01-02T15:04:05.000Z07:00"
    prettyPrint: false

sqs:
  queues:
    - https://sqs.us-west-2.amazonaws.com/123456789/your-s3-event-queue
  buckets:
    - name: your-s3-bucket-name
      prefixes:
        - logs/
        - data/
  processUnlistedBuckets: false
  workerCount: 5
  maxMessages: 10
  waitTime: 20

metrics:
  enabled: true
  types:
    eventTotal: true
    objectSize: true
    userTotal: true
    ipTotal: true
    prefixTotal: true
    prefixDepthTotal: true
    fileExtensionTotal: true
    latency: true
    anomalyDetection: true
    lifecycleExpiration: true
    deleteTotal: true
  prefixDepth: 4
  objectSizeBuckets:
    - 1024      # 1KB
    - 102400    # 100KB
    - 1048576   # 1MB
    - 10485760  # 10MB
  port: 8087
```

Mount this configuration:

```bash
docker run -d \
  --name s3-metrics-adapter \
  -p 8087:8087 \
  -v /path/to/your/config.yaml:/app/config.yaml \
  codebyrupinder/s3-metrics-prom-adaptor:latest -config /app/config.yaml
```

## AWS Credentials

### Option 1: AWS Credentials File

```bash
docker run -d \
  --name s3-metrics-adapter \
  -p 8087:8087 \
  -v ~/.aws:/home/app/.aws:ro \
  -v /path/to/config.yaml:/app/config.yaml \
  codebyrupinder/s3-metrics-prom-adaptor:latest
```

### Option 2: Environment Variables

```bash
docker run -d \
  --name s3-metrics-adapter \
  -p 8087:8087 \
  -e AWS_ACCESS_KEY_ID="your-access-key" \
  -e AWS_SECRET_ACCESS_KEY="your-secret-key" \
  -e AWS_REGION="us-west-2" \
  -v /path/to/config.yaml:/app/config.yaml \
  codebyrupinder/s3-metrics-prom-adaptor:latest
```

### Option 3: IAM Role (for EC2/ECS)

When running on AWS infrastructure, the container will automatically use the instance's IAM role:

```bash
docker run -d \
  --name s3-metrics-adapter \
  -p 8087:8087 \
  -v /path/to/config.yaml:/app/config.yaml \
  codebyrupinder/s3-metrics-prom-adaptor:latest
```

## Docker Compose

Create a `docker-compose.yml` file:

```yaml
version: '3.8'

services:
  s3-metrics-adapter:
    image: codebyrupinder/s3_metrics_adapter:latest
    container_name: s3-metrics-adapter
    ports:
      - "8087:8087"
    volumes:
      - ./config.yaml:/app/config.yaml:ro
      - ~/.aws:/home/app/.aws:ro
    environment:
      - AWS_REGION=us-west-2
      - LOG_LEVEL=info
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8087/metrics"]
      interval: 30s
      timeout: 3s
      retries: 3
      start_period: 10s

  prometheus:
    image: prom/prometheus:latest
    container_name: prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml:ro
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
    restart: unless-stopped

  grafana:
    image: grafana/grafana:latest
    container_name: grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana-storage:/var/lib/grafana
    restart: unless-stopped

volumes:
  grafana-storage:
```

Start the stack:

```bash
docker-compose up -d
```

## Monitoring and Health Checks

### Health Check Endpoint

The container includes a built-in health check that monitors the metrics endpoint:

```bash
# Check container health
docker ps

# Manual health check
curl http://localhost:8087/metrics
```

### Metrics Endpoint

Access Prometheus metrics at:
- `http://localhost:8087/metrics`

### Available Metrics

- `s3_events_total` - Total number of S3 events processed
- `s3_object_size_bytes` - Object size distribution
- `s3_events_by_user_total` - Events grouped by user
- `s3_events_by_ip_total` - Events grouped by source IP
- `s3_events_by_prefix_total` - Events grouped by object prefix
- `s3_processing_latency_seconds` - Event processing latency
- And many more...

## Scaling and Production Deployment

### Horizontal Scaling

Run multiple instances with different SQS queues:

```bash
# Instance 1
docker run -d \
  --name s3-exporter-1 \
  -p 8087:8087 \
  -e SQS_QUEUE_URL="https://sqs.us-west-2.amazonaws.com/123456789/queue-1" \
  codebyrupinder/s3-metrics-prom-adaptor:latest

# Instance 2  
docker run -d \
  --name s3-exporter-2 \
  -p 8088:8087 \
  -e SQS_QUEUE_URL="https://sqs.us-west-2.amazonaws.com/123456789/queue-2" \
  codebyrupinder/s3-metrics-prom-adaptor:latest
```

### Resource Limits

Set resource limits for production:

```bash
docker run -d \
  --name s3-metrics-adapter \
  -p 8087:8087 \
  --memory=512m \
  --memory-swap=1g \
  --cpus=1.0 \
  --restart=unless-stopped \
  codebyrupinder/s3-metrics-prom-adaptor:latest
```

### Kubernetes Deployment

Example Kubernetes deployment:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: s3-metrics-adapter
spec:
  replicas: 3
  selector:
    matchLabels:
      app: s3-metrics-adapter
  template:
    metadata:
      labels:
        app: s3-metrics-adapter
    spec:
      containers:
      - name: s3-metrics-adapter
        image: codebyrupinder/s3_metrics_adapter:latest
        ports:
        - containerPort: 8087
        env:
        - name: AWS_REGION
          value: "us-west-2"
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /metrics
            port: 8087
          initialDelaySeconds: 30
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /metrics
            port: 8087
          initialDelaySeconds: 5
          periodSeconds: 10
```

## Troubleshooting

### Container Won't Start

1. **Check logs:**
   ```bash
   docker logs s3-metrics-adapter
   ```

2. **Verify configuration:**
   ```bash
   docker run --rm -v /path/to/config.yaml:/app/config.yaml \
     codebyrupinder/s3_metrics_adapter:latest -config /app/config.yaml -help
   ```

3. **Test configuration syntax:**
   ```bash
   # Validate YAML syntax
   python -c "import yaml; yaml.safe_load(open('config.yaml'))"
   ```

### AWS Connection Issues

1. **Check AWS credentials:**
   ```bash
   docker run --rm -it \
     -v ~/.aws:/home/app/.aws:ro \
     codebyrupinder/s3_metrics_adapter:latest \
     sh -c "aws sts get-caller-identity"
   ```

2. **Verify SQS queue access:**
   ```bash
   docker run --rm -it \
     -v ~/.aws:/home/app/.aws:ro \
     codebyrupinder/s3_metrics_adapter:latest \
     sh -c "aws sqs get-queue-attributes --queue-url your-queue-url"
   ```

### Performance Issues

1. **Monitor resource usage:**
   ```bash
   docker stats s3-metrics-adapter
   ```

2. **Adjust worker count:**
   ```yaml
   sqs:
     workerCount: 10  # Increase for higher throughput
     maxMessages: 10  # Batch size per request
   ```

3. **Check metrics for bottlenecks:**
   ```bash
   curl -s http://localhost:8087/metrics | grep latency
   ```

## Security Best Practices

1. **Run as non-root user:** ✅ (Built into the image)
2. **Use read-only filesystem:**
   ```bash
   docker run -d --read-only --tmpfs /tmp \
     --name s3-metrics-adapter \
     codebyrupinder/s3_metrics_adapter:latest
   ```

3. **Limit network access:**
   ```bash
   docker network create --driver bridge restricted
   docker run -d --network restricted \
     --name s3-metrics-adapter \
     codebyrupinder/s3_metrics_adapter:latest
   ```

4. **Use secrets management:**
   ```bash
   # Using Docker secrets
   echo "your-secret-key" | docker secret create aws_secret_key -
   docker service create \
     --secret aws_secret_key \
     --name s3-metrics-adapter \
     codebyrupinder/s3_metrics_adapter:latest
   ```

## Integration Examples

### With Prometheus

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 's3-metrics-adapter'
    static_configs:
      - targets: ['s3-metrics-adapter:8087']
    scrape_interval: 30s
    metrics_path: /metrics
```

### With Grafana Dashboard

Import the provided Grafana dashboard JSON or create custom dashboards using the available metrics.

### With Alertmanager

Set up alerts for important metrics:

```yaml
# alerting rules
groups:
  - name: s3-events
    rules:
      - alert: HighS3EventVolume
        expr: rate(s3_events_total[5m]) > 100
        for: 2m
        annotations:
          description: "High volume of S3 events detected"
```

## Support

For issues, feature requests, or contributions:
- GitHub: https://github.com/code-by-rupinder/s3-metrics-adaptor
- Issues: https://github.com/code-by-rupinder/s3-metrics-adaptor/issues

