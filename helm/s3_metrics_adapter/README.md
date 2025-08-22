# S3 Event Exporter Helm Chart

A Helm chart for deploying the S3 Event Exporter on Kubernetes. This chart provides a comprehensive monitoring solution for AWS S3 events with Prometheus metrics integration.

## Overview

The S3 Event Exporter is a Prometheus exporter that monitors AWS S3 events through SQS queues and provides detailed metrics about object operations, user activities, and storage patterns.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+
- AWS credentials configured (IAM roles, access keys, or IRSA)
- SQS q## Support

- **GitHub Issues**: [Create an issue](https://github.com/code-by-rupinder/s3-metrics-adaptor/issues)

- **AWS Credentials Guide**: [Detailed credential configuration](../../docs/aws-credentials-guide.md)
- **Docker Hub**: [codebyrupinder/s3_metrics_adapter](https://hub.docker.com/r/codebyrupinder/s3_metrics_adapter)

## Quick Reference by Environment

| Environment | Credential Method | Example Values File |
|-------------|------------------|-------------------|
| **Amazon EKS** | IRSA | `values-production.yaml` |
| **Self-managed on EC2** | Instance Profile | `examples/values-ec2-instance-profile.yaml` |
| **On-premises** | Kubernetes Secret | `examples/values-onpremises.yaml` |
| **Bare Metal** | External Secret | `examples/values-selfmanaged.yaml` |
| **Development** | Manual Secret | `examples/values-dev.yaml` |

For detailed setup instructions for each environment, see the [AWS Credentials Guide](../../docs/aws-credentials-guide.md). configured to receive S3 event notifications

## Installation

### Add the Helm Repository

```bash
# Add the repository (when published to GitHub Pages)
helm repo add s3-metrics-adapter https://codebyrupinder.github.io/s3-metrics-adapter/
helm repo update
```

### Install from Source

```bash
# Clone the repository
git clone https://github.com/code-by-rupinder/s3-metrics-adapter.git
cd s3-metrics-adapter

# Install the chart
helm install s3-metrics-adapter ./charts/s3-metrics-adapter \
  --namespace monitoring \
  --create-namespace \
  --set config.sqs.queues[0]="https://sqs.us-west-2.amazonaws.com/123456789/my-s3-events-queue"
```

### Quick Installation

```bash
helm install s3-metrics-adapter s3-metrics-adapter/s3-metrics-adapter \
  --namespace monitoring \
  --create-namespace \
  --set config.sqs.queues[0]="your-sqs-queue-url"
```

## Configuration

### Basic Configuration

The following table lists the configurable parameters and their default values:

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of replicas | `1` |
| `image.repository` | Container image repository | `codebyrupinder/s3_metrics_adapter` |
| `image.tag` | Container image tag | `1.0.1` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |

### AWS Configuration

The chart supports multiple AWS credential methods depending on your Kubernetes environment:

| Method | Environment | Security | Recommended |
|--------|-------------|----------|-------------|
| **Cloud IAM** | EKS, GKE | ⭐⭐⭐⭐⭐ | ✅ Managed K8s |
| **Instance Profile** | Self-managed on EC2 | ⭐⭐⭐⭐ | ✅ EC2 clusters |
| **External Secret** | Any | ⭐⭐⭐⭐ | ✅ Production |
| **Kubernetes Secret** | Any | ⭐⭐⭐ | ⚠️ Development |

#### Method 1: Cloud Provider IAM (Managed Kubernetes)

**For Amazon EKS with IRSA:**
```bash
helm install s3-metrics-adapter s3-metrics-adapter/s3-metrics-adapter \
  --set serviceAccount.annotations."eks\.amazonaws\.com/role-arn"="arn:aws:iam::123456789:role/S3EventExporterRole" \
  --set config.sqs.queues[0]="https://sqs.us-west-2.amazonaws.com/123456789/my-queue"
```


#### Method 2: EC2 Instance Profile (Self-managed on EC2)

**Best for**: Self-managed Kubernetes on EC2 instances

```bash
# No additional configuration needed - uses EC2 instance profile
helm install s3-metrics-adapter s3-metrics-adapter/s3-metrics-adapter \
  -f examples/values-ec2-instance-profile.yaml \
  --set config.sqs.queues[0]="https://sqs.us-west-2.amazonaws.com/123456789/my-queue"
```

#### Method 3: External Secret Management (Production)

**Best for**: Production environments with external secret stores

```bash
# Using existing secret from external secret management
helm install s3-metrics-adapter s3-metrics-adapter/s3-metrics-adapter \
  --set awsCredentials.existingSecret.name="aws-credentials" \
  --set config.sqs.queues[0]="https://sqs.us-west-2.amazonaws.com/123456789/my-queue"
```

#### Method 4: Kubernetes Secret (Development)

**For self-managed Kubernetes, on-premises, bare metal:**

```bash
# Create AWS credentials secret
kubectl create secret generic aws-credentials \
  --from-literal=AWS_ACCESS_KEY_ID=YOUR_ACCESS_KEY \
  --from-literal=AWS_SECRET_ACCESS_KEY=YOUR_SECRET_KEY \
  --from-literal=AWS_REGION=us-west-2 \
  --namespace monitoring

# Install with automatic credential configuration
helm install s3-metrics-adapter s3-metrics-adapter/s3-metrics-adapter \
  --set awsCredentials.create=true \
  --set awsCredentials.region="us-west-2" \
  --set config.sqs.queues[0]="https://sqs.us-west-2.amazonaws.com/123456789/my-queue"
```

**Alternative with existing secret:**
```bash
helm install s3-metrics-adapter s3-metrics-adapter/s3-metrics-adapter \
  --set awsCredentials.existingSecret.name="aws-credentials" \
  --set config.sqs.queues[0]="https://sqs.us-west-2.amazonaws.com/123456789/my-queue"
```

### SQS and S3 Configuration

```yaml
config:
  sqs:
**GitHub Issues**: [Create an issue](https://github.com/code-by-rupinder/s3-metrics-adaptor/issues)
**Docker Hub**: [codebyrupinder/s3-metrics-adaptor](https://hub.docker.com/r/codebyrupinder/s3-metrics-adaptor)
        prefixes:
          - "logs/"
          - "data/"
    processUnlistedBuckets: true
    workerCount: 5
helm repo add s3-metrics-adaptor https://code-by-rupinder.github.io/s3-metrics-adaptor/
    waitTime: 20
```

### Monitoring Configuration

#### Enable ServiceMonitor for Prometheus Operator

git clone https://github.com/code-by-rupinder/s3-metrics-adaptor.git
cd s3-metrics-adaptor
  enabled: true
  interval: 30s
  scrapeTimeout: 10s
  additionalLabels:
    release: prometheus
```
helm install s3-metrics-adaptor ./charts/s3-metrics-adaptor \

#### Enable PodMonitor

```yaml
podMonitor:
  enabled: true
  interval: 30s
  scrapeTimeout: 10s
  repository: ghcr.io/code-by-rupinder/s3-metrics-adaptor
```

### Autoscaling

 GitHub Issues: [Create an issue](https://github.com/code-by-rupinder/s3-metrics-adaptor/issues)
 Docker Hub: [codebyrupinder/s3-metrics-adaptor](https://hub.docker.com/r/codebyrupinder/s3-metrics-adaptor)
  minReplicas: 1
  maxReplicas: 10
  targetCPUUtilizationPercentage: 80
  targetMemoryUtilizationPercentage: 80
```

### Security

#### Pod Security Context

```yaml
podSecurityContext:
  runAsNonRoot: true
  runAsUser: 65532
  fsGroup: 65532

securityContext:
  allowPrivilegeEscalation: false
  capabilities:
    drop:
      - ALL
  readOnlyRootFilesystem: true
  runAsNonRoot: true
  runAsUser: 65532
```

#### Network Policy

```yaml
networkPolicy:
  enabled: true
  policyTypes:
    - Ingress
    - Egress
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              name: monitoring
      ports:
        - protocol: TCP
          port: 8087
  egress:
    - to: []
      ports:
        - protocol: TCP
          port: 443  # HTTPS for AWS APIs
```

## Examples

### Complete Production Configuration

```yaml
# values-production.yaml
replicaCount: 2

image:
  repository: codebyrupinder/s3_metrics_adapter
  tag: "1.0.1"
  pullPolicy: IfNotPresent

serviceAccount:
  create: true
  annotations:
    eks.amazonaws.com/role-arn: "arn:aws:iam::123456789:role/S3EventExporterRole"

resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 100m
    memory: 128Mi

config:
  logging:
    default: info
  sqs:
    queues:
      - "https://sqs.us-west-2.amazonaws.com/123456789/s3-events-prod"
    buckets:
      - name: "prod-data-bucket"
        prefixes:
          - "logs/"
          - "analytics/"
      - name: "prod-backup-bucket"
    processUnlistedBuckets: false
    workerCount: 10
    maxMessages: 10

autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70

serviceMonitor:
  enabled: true
  interval: 30s
  additionalLabels:
    release: prometheus

podDisruptionBudget:
  enabled: true
  minAvailable: 1

networkPolicy:
  enabled: true

nodeSelector:
  node-type: monitoring

tolerations:
  - key: "monitoring"
    operator: "Equal"
    value: "true"
    effect: "NoSchedule"
```

### Development Configuration

```yaml
# values-dev.yaml
replicaCount: 1

resources:
  limits:
    cpu: 200m
    memory: 256Mi
  requests:
    cpu: 50m
    memory: 64Mi

config:
  logging:
    default: debug
  sqs:
    queues:
      - "https://sqs.us-west-2.amazonaws.com/123456789/s3-events-dev"
    processUnlistedBuckets: true
    workerCount: 2

envFrom:
  - secretRef:
      name: aws-dev-credentials

env:
  - name: AWS_REGION
    value: "us-west-2"
```

## Deployment Commands

### Managed Kubernetes (EKS, GKE, AKS)

```bash
# Amazon EKS with IRSA
helm install s3-metrics-adapter s3-metrics-adapter/s3-metrics-adapter \
  -f values-production.yaml \
  --set serviceAccount.annotations."eks\.amazonaws\.com/role-arn"="arn:aws:iam::123456789:role/S3EventExporterRole" \
  --namespace monitoring \
  --create-namespace

# Google GKE with Workload Identity
helm install s3-metrics-adapter s3-metrics-adapter/s3-metrics-adapter \
  -f values-production.yaml \
  --set serviceAccount.annotations."iam\.gke\.io/gcp-service-account"="s3-exporter@project.iam.gserviceaccount.com" \
  --namespace monitoring \
  --create-namespace
```

### Self-managed Kubernetes on EC2

```bash
# Using EC2 instance profile (recommended)
helm install s3-metrics-adapter s3-metrics-adapter/s3-metrics-adapter \
  -f examples/values-ec2-instance-profile.yaml \
  --namespace monitoring \
  --create-namespace

# With external secret management
helm install s3-metrics-adapter s3-metrics-adapter/s3-metrics-adapter \
  -f examples/values-selfmanaged.yaml \
  --set awsCredentials.existingSecret.name="aws-credentials" \
  --namespace monitoring \
  --create-namespace
```

### On-premises / Bare Metal Kubernetes

```bash
# Create AWS credentials secret first
kubectl create secret generic aws-credentials \
  --from-literal=AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID \
  --from-literal=AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY \
  --from-literal=AWS_REGION=us-west-2 \
  --namespace monitoring

# Simple deployment
helm install s3-metrics-adapter s3-metrics-adapter/s3-metrics-adapter \
  -f examples/values-onpremises.yaml \
  --namespace monitoring \
  --create-namespace

# Production deployment with external secrets
helm install s3-metrics-adapter s3-metrics-adapter/s3-metrics-adapter \
  -f examples/values-selfmanaged.yaml \
  --set awsCredentials.existingSecret.name="aws-credentials" \
  --namespace monitoring \
  --create-namespace
```

### Development / Testing

```bash
# Quick test deployment (not recommended for production)
helm install s3-metrics-adapter-dev s3-metrics-adapter/s3-metrics-adapter \
  --set awsCredentials.create=true \
  --set-string awsCredentials.accessKeyId=$AWS_ACCESS_KEY_ID \
  --set-string awsCredentials.secretAccessKey=$AWS_SECRET_ACCESS_KEY \
  --set awsCredentials.region=us-west-2 \
  --set 'config.sqs.queues[0]=https://sqs.us-west-2.amazonaws.com/123456789/test-queue' \
  --namespace dev \
  --create-namespace
```

### Upgrade

```bash
helm upgrade s3-metrics-adapter s3-metrics-adapter/s3-metrics-adapter \
  -f values-production.yaml \
  --namespace monitoring
```

### Uninstall

```bash
helm uninstall s3-metrics-adapter --namespace monitoring
```

## Monitoring and Observability

### Metrics Endpoints

- **Metrics**: `http://service:8087/metrics`
- **Health**: `http://service:8087/metrics` (also serves as health check)

### Available Metrics

- `s3_events_total` - Total number of S3 events processed
- `s3_object_size_bytes` - Distribution of object sizes
- `s3_events_by_user_total` - Events grouped by user
- `s3_events_by_ip_total` - Events grouped by source IP
- `s3_processing_latency_seconds` - Processing latency metrics

### Grafana Dashboard

A Grafana dashboard is available in the `grafana/` directory of the repository.

## Troubleshooting

### Common Issues

1. **AWS Permissions**
   ```bash
   # Check service account annotations
   kubectl describe serviceaccount s3-metrics-adapter -n monitoring
   
   # Check pod logs
   kubectl logs -l app.kubernetes.io/name=s3-metrics-adapter -n monitoring
   ```

2. **SQS Connectivity**
   ```bash
   # Test SQS connectivity from pod
   kubectl exec -it deployment/s3-metrics-adapter -n monitoring -- /usr/local/bin/s3-metrics-adapter -help
   ```

3. **Metrics Not Appearing**
   ```bash
   # Check metrics endpoint
   kubectl port-forward svc/s3-metrics-adapter 8087:8087 -n monitoring
   curl http://localhost:8087/metrics
   ```

### Debugging

```bash
# Enable debug logging
helm upgrade s3-metrics-adapter s3-metrics-adapter/s3-metrics-adapter \
  --set config.logging.default=debug \
  --namespace monitoring

# Check events
kubectl get events -n monitoring

# Check pod status
kubectl describe pod -l app.kubernetes.io/name=s3-metrics-adapter -n monitoring
```

## AWS IAM Configuration

The exporter needs IAM permissions to access SQS and S3. Here's an example IAM policy:

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
            "Resource": "arn:aws:sqs:*:*:*"
        },
        {
            "Effect": "Allow",
            "Action": [
                "s3:GetObject",
                "s3:ListBucket"
            ],
            "Resource": [
                "arn:aws:s3:::*/*",
                "arn:aws:s3:::*"
            ]
        }
    ]
}
```

## Values Reference

See [values.yaml](./values.yaml) for the complete list of configurable parameters.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](../../LICENSE) file for details.

## Support

- GitHub Issues: [Create an issue](https://github.com/codebyrupinder/s3_metrics_adapter/issues)
- Documentation: [Project README](../../README.md)
- Docker Hub: [codebyrupinder/s3_metrics_adapter](https://hub.docker.com/r/codebyrupinder/s3_metrics_adapter)
