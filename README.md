# S3 Metrics Adapter

[![Docker Hub](https://img.shields.io/docker/v/codebyrupinder/s3-metrics-adapter?label=Docker%20Hub)](https://hub.docker.com/r/codebyrupinder/s3-metrics-adapter)
[![Go Report Card](https://goreportcard.com/badge/github.com/codebyrupinder/s3-metrics-adapter)](https://goreportcard.com/report/github.com/codebyrupinder/s3-metrics-adapter)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/release/codebyrupinder/s3-metrics-adapter.svg)](https://github.com/codebyrupinder/s3-metrics-adapter/releases/latest)

A high-performance Prometheus metrics adapter that transforms Amazon S3 events from SQS queues into comprehensive monitoring metrics. Get deep insights into your S3 bucket activities, access patterns, and operational health.

## 🚀 Key Features

- **Real-time S3 Event Processing** - Convert S3 events from SQS into Prometheus metrics
- **Comprehensive Metrics** - 12+ metric types covering events, sizes, users, IPs, and anomalies
- **Advanced Analytics** - File extension tracking, hierarchical path analysis, and anomaly detection
- **Production Ready** - Built with Go, optimized Docker images (14.5MB), and comprehensive monitoring
- **Highly Configurable** - Enable/disable metrics individually, custom histogram buckets, flexible filtering
- **Security First** - Distroless container images, non-root execution, minimal attack surface
- **Multi-deployment Options** - Docker, Kubernetes, binary, or source deployment

## 📊 Available Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `s3_event_total` | Counter | Total S3 events by type and subtype |
| `s3_event_object_size_bytes` | Histogram | Object size distribution |
| `s3_event_user_total` | Counter | Events by user identity |
| `s3_event_ip_total` | Counter | Events by source IP |
| `s3_event_prefix_total` | Counter | Events by object prefix |
| `s3_events_hierarchical_path_total` | Counter | Events by hierarchical path depth |
| `s3_bucket_extension_files_total` | Gauge | File count by extension and type |
| `s3_event_latency_seconds` | Gauge | Event processing latency |
| `s3_event_anomaly_total` | Counter | Detected anomalies (system deletes, etc.) |
| `s3_lifecycle_expiration_total` | Counter | Lifecycle policy deletions |
| `s3_delete_total` | Counter | Delete events with detailed categorization |
| `s3_event_parser_errors_total` | Counter | Parsing error count |

[📖 Complete Metrics Reference](./docs/METRICS_REFERENCE.md)

## 🏃‍♂️ Quick Start

### Docker (Recommended)

```bash
# Pull and run the latest version
docker run -d \
  --name s3-metrics-adapter \
  -p 8087:8087 \
  -v $(pwd)/config.yaml:/app/config.yaml \
  -e AWS_REGION=us-west-2 \
  -e AWS_ACCESS_KEY_ID=your-key \
  -e AWS_SECRET_ACCESS_KEY=your-secret \
  codebyrupinder/s3-metrics-adapter:latest

# Check metrics
curl http://localhost:8087/metrics
```

### Binary Download

```bash
# Download for your platform
curl -LO https://github.com/codebyrupinder/s3-metrics-adapter/releases/latest/download/s3_metrics_adapter-linux-amd64
chmod +x s3_metrics_adapter-linux-amd64

# Run with config
./s3_metrics_adapter-linux-amd64 -config config.yaml
```

### Docker Compose (Full Stack)

```bash
git clone https://github.com/codebyrupinder/s3-metrics-adapter.git
cd s3-metrics-adapter
docker-compose up -d

# Access services
# Metrics: http://localhost:8087/metrics
# Prometheus: http://localhost:9090
# Grafana: http://localhost:3000 (admin/admin)
```

## ⚙️ Configuration

Create a `config.yaml` file:

```yaml
sqs:
  queues:
    - https://sqs.us-west-2.amazonaws.com/123456789/your-s3-event-queue
  buckets:
    - name: your-bucket-name
      prefix:
        - logs/
        - data/
  workerCount: 5

metrics:
  enabled: true
  port: 8087
  types:
    eventTotal: true
    objectSize: true
    userTotal: true
    ipTotal: true
    prefixTotal: true
    fileExtensionTotal: true
    anomalyDetection: true
    deleteTotal: true
  objectSizeBuckets:
    - 1024      # 1KB
    - 102400    # 100KB  
    - 1048576   # 1MB
    - 10485760  # 10MB

logging:
  default: info
```

## 📚 Documentation

| Guide | Description |
|-------|-------------|
| [Docker Deployment Guide](./docs/DOCKER_DEPLOYMENT_GUIDE.md) | Complete Docker, Compose & Kubernetes deployment |
| [Binary Deployment Guide](./docs/BINARY_DEPLOYMENT_GUIDE.md) | Download and run binary directly |
| [Configuration Reference](./docs/CONFIGURATION_REFERENCE.md) | All configuration options explained |
| [Metrics Reference](./docs/METRICS_REFERENCE.md) | Complete metrics documentation |
| [Sample Grafana Dashboard](./docs/grafana-dashboard-sample.json) | Ready-to-import dashboard |

## 🏗️ Prerequisites

### AWS Setup Required

1. **S3 Bucket** with event notifications enabled
2. **SQS Queue** to receive S3 events
3. **EventBridge** or legacy S3 notifications configured
4. **AWS Credentials** with permissions:
   ```json
   {
     "Version": "2012-10-17",
     "Statement": [
       {
         "Effect": "Allow",
         "Action": [
           "sqs:ReceiveMessage",
           "sqs:DeleteMessage",
           "sqs:GetQueueAttributes"
         ],
         "Resource": "arn:aws:sqs:*:*:your-queue-name"
       }
     ]
   }
   ```

### System Requirements

- **Memory**: 256MB minimum, 512MB recommended
- **CPU**: 0.25 cores minimum, 0.5 cores recommended  
- **Network**: Access to AWS SQS endpoints
- **Storage**: 100MB for logs and temporary files

## 🔧 Advanced Usage

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: s3-metrics-adapter
spec:
  replicas: 2
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
        image: codebyrupinder/s3-metrics-adapter:latest
        ports:
        - containerPort: 8087
        resources:
          requests:
            memory: 256Mi
            cpu: 250m
          limits:
            memory: 512Mi
            cpu: 500m
        livenessProbe:
          httpGet:
            path: /metrics
            port: 8087
          initialDelaySeconds: 30
          periodSeconds: 30
```

### Scaling Configuration

For high-volume S3 buckets:

```yaml
sqs:
  workerCount: 20        # Increase workers
  maxMessages: 10        # Batch size
  waitTime: 20          # Long polling

metrics:
  types:
    # Disable expensive metrics if needed
    fileExtensionTotal: false
    prefixDepthTotal: false
```

### Multi-Region Setup

```yaml
# config-us-east-1.yaml
sqs:
  queues:
    - https://sqs.us-east-1.amazonaws.com/123456789/s3-events-east
    - https://sqs.us-west-2.amazonaws.com/123456789/s3-events-west
    # Add more regions/queues as needed
```

## 📈 Monitoring Examples

### Prometheus Queries

```promql
# Event rate by bucket
sum(rate(s3_event_total[5m])) by (bucket)

# Large object uploads (>100MB)
histogram_quantile(0.95, s3_event_object_size_bytes_bucket{bucket="my-bucket"})

# Anomaly detection rate
sum(rate(s3_event_anomaly_total[5m])) by (type)

# Processing latency
avg(s3_event_latency_seconds) by (bucket)
```

### Alerting Rules

```yaml
groups:
  - name: s3-metrics-adapter
    rules:
      - alert: HighS3EventVolume
        expr: rate(s3_event_total[5m]) > 100
        for: 2m
        annotations:
          description: "High S3 event volume: {{ $value }} events/sec"
          
      - alert: S3ProcessingLatency
        expr: avg(s3_event_latency_seconds) > 60
        for: 1m
        annotations:
          description: "High processing latency: {{ $value }}s"
```

## 🛡️ Security

- **Distroless Images**: Minimal attack surface with no shell or package manager
- **Non-root Execution**: Container runs as non-root user
- **Secrets Management**: Support for AWS IAM roles, environment variables, or mounted credentials
- **Network Security**: Configurable resource limits and network policies
- **Read-only Filesystem**: Compatible with read-only root filesystems

## 🔍 Troubleshooting

### Common Issues

**Container won't start:**
```bash
# Check logs
docker logs s3-metrics-adapter

# Verify configuration
docker run --rm -v $(pwd)/config.yaml:/app/config.yaml \
  codebyrupinder/s3-metrics-adapter:latest -help
```

**AWS connection issues:**
```bash
# Test AWS credentials
docker run --rm -it \
  -e AWS_REGION=us-west-2 \
  amazon/aws-cli sts get-caller-identity

# Verify SQS access
aws sqs get-queue-attributes --queue-url your-queue-url
```

**High memory usage:**
```yaml
# Reduce enabled metrics
metrics:
  types:
    fileExtensionTotal: false  # Most memory intensive
    prefixDepthTotal: false
```

### Debug Mode

```yaml
logging:
  default: debug
  components:
    eventparser: debug
    sqspoller: debug
```

## 🚀 Performance

- **Throughput**: 1000+ events/second
- **Memory**: ~50MB runtime usage
- **Latency**: <100ms event processing
- **Startup**: <5 second container startup
- **Image Size**: 14.5MB (distroless)

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Setup

```bash
git clone https://github.com/codebyrupinder/s3-metrics-adapter.git
cd s3-metrics-adapter
go mod download
go build -o s3-metrics-adapter ./cmd
```

### Running Tests

```bash
go test ./...
```

### Building Docker Image

```bash
./docker-build.sh -v 1.0.0
```

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Issues**: [GitHub Issues](https://github.com/codebyrupinder/s3-metrics-adapter/issues)
- **Discussions**: [GitHub Discussions](https://github.com/codebyrupinder/s3-metrics-adapter/discussions)
- **Security**: Please report security issues privately via email

## 🏆 Acknowledgments

- AWS SDK for Go team
- Prometheus Go client library
- All contributors and users of this project

## 🔗 Related Projects

- [Prometheus](https://prometheus.io/) - Metrics collection
- [Grafana](https://grafana.com/) - Metrics visualization  
- [AWS CloudWatch](https://aws.amazon.com/cloudwatch/) - Native AWS monitoring
- [S3 Bucket Notifications](https://docs.aws.amazon.com/AmazonS3/latest/userguide/notification-content-structure.html) - Event structure reference

---

**Star this project ⭐ if it helps you monitor your S3 infrastructure!**