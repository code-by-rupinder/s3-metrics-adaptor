# S3 Event Exporter Helm Chart

This directory contains the official Helm chart for deploying the S3 Event Exporter on Kubernetes clusters.

[![Helm Chart](https://img.shields.io/badge/helm-chart-blue.svg)](https://helm.sh)
[![Kubernetes](https://img.shields.io/badge/kubernetes-1.19%2B-green.svg)](https://kubernetes.io)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## Overview

The S3 Event Exporter is a Kubernetes-native application that processes S3 events from SQS queues and exports metrics for monitoring. This Helm chart provides a production-ready deployment with comprehensive configuration options for various Kubernetes environments.

## Features

- 🚀 **Production Ready**: Comprehensive configuration for enterprise deployments
- 🔐 **Multi-Auth Support**: AWS IRSA, instance profiles, secrets, and external secret management
- 📊 **Monitoring Ready**: Built-in Prometheus metrics and ServiceMonitor support
- 🔄 **Auto-scaling**: Horizontal Pod Autoscaler configuration
- 🛡️ **Security First**: Security contexts, network policies, and non-root containers
- 🌐 **Multi-Environment**: Support for EKS, GKE, self-managed, and on-premises Kubernetes

## Quick Start

### Install from GitHub Pages (Recommended)

```bash
# Add the Helm repository
helm repo add helm-charts https://code-by-rupinder.github.io/helm-charts/
helm repo update

# Install the chart
helm install my-s3-exporter helm-charts/s3-metrics-adapter \
  --version 1.0.0 \
  --namespace monitoring \
  --create-namespace \
  --set 'config.sqs.queues[0]=https://sqs.us-west-2.amazonaws.com/123456789/your-queue'
```

## Chart Information

| Field | Value |
|-------|-------|
| **Chart Version** | 1.0.0 |
| **App Version** | 1.0.1 |
| **Kubernetes Version** | 1.19+ |
| **Helm Version** | 3.0+ |

## Directory Structure

```
helm/
├── s3-metrics-adapter/
│   ├── Chart.yaml                 # Chart metadata
│   ├── values.yaml               # Default configuration values
│   ├── .helmignore              # Files to ignore during packaging
│   ├── templates/               # Kubernetes manifest templates
│   │   ├── deployment.yaml      # Main application deployment
│   │   ├── service.yaml         # Service for metrics exposure
│   │   ├── serviceaccount.yaml  # Service account with annotations
│   │   ├── secret.yaml          # AWS credentials secret (conditional)
│   │   ├── configmap.yaml       # Application configuration
│   │   ├── hpa.yaml             # Horizontal Pod Autoscaler
│   │   ├── ingress.yaml         # Ingress configuration
│   │   ├── networkpolicy.yaml   # Network policies
│   │   ├── poddisruptionbudget.yaml # Pod disruption budget
│   │   ├── servicemonitor.yaml  # Prometheus ServiceMonitor
│   │   ├── podmonitor.yaml      # Prometheus PodMonitor
│   │   └── NOTES.txt           # Post-installation notes
│   └── examples/               # Example configuration files
│       ├── values-production.yaml
│       ├── values-development.yaml
│       ├── values-staging.yaml
│       └── values-ec2-instance-profile.yaml
├── DEPLOYMENT.md               # Comprehensive deployment guide
├── GITHUB-PAGES-SETUP.md      # GitHub Pages setup documentation
└── README.md                  # This file
```

## Configuration

### Essential Configuration

```yaml
# Minimum required configuration
config:
  sqs:
    queues:
      - "https://sqs.us-west-2.amazonaws.com/123456789/your-queue"

# AWS credentials (choose one method)
# Method 1: EKS IRSA (Recommended for EKS)
serviceAccount:
  annotations:
    eks.amazonaws.com/role-arn: "arn:aws:iam::ACCOUNT:role/S3EventExporterRole"

# Method 2: Existing Secret
awsCredentials:
  existingSecret:
    name: "aws-credentials"

# Method 3: Create Secret via Helm (Development only)
awsCredentials:
  create: true
  accessKeyId: "AKIA..."
  secretAccessKey: "secret..."
  region: "us-west-2"
```

### Production Configuration Example

```yaml
replicaCount: 3

image:
  repository: codebyrupinder/s3_metrics_adapter
  tag: "1.0.1"
  pullPolicy: IfNotPresent

serviceAccount:
  create: true
  annotations:
    eks.amazonaws.com/role-arn: "arn:aws:iam::ACCOUNT:role/S3EventExporterRole"

resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 100m
    memory: 128Mi

config:
  sqs:
    queues:
      - "https://sqs.us-west-2.amazonaws.com/ACCOUNT/prod-s3-events"
    buckets:
      - name: "prod-data-bucket"
        prefixes: ["logs/", "analytics/"]
    workerCount: 10
  logging:
    default: info

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70

serviceMonitor:
  enabled: true
  interval: 30s
  additionalLabels:
    release: prometheus

podDisruptionBudget:
  enabled: true
  minAvailable: 2

networkPolicy:
  enabled: true
```

## Deployment Scenarios

### 1. Amazon EKS (with IRSA)

```bash
# Create IAM role and associate with service account
eksctl create iamserviceaccount \
  --cluster=my-cluster \
  --namespace=monitoring \
  --name=s3-metrics-adapter \
  --role-name=S3EventExporterRole \
  --attach-policy-arn=arn:aws:iam::ACCOUNT:policy/S3EventExporterPolicy \
  --approve

# Install chart
helm install s3-metrics-adapter helm-charts/s3-metrics-adapter \
  --version 1.0.0 \
  --set serviceAccount.annotations."eks\.amazonaws\.com/role-arn"="arn:aws:iam::ACCOUNT:role/S3EventExporterRole" \
  --namespace monitoring \
  --create-namespace
```

### 2. Self-managed Kubernetes (EC2 Instance Profile)

```bash
helm install s3-metrics-adapter helm-charts/s3-metrics-adapter \
  --version 1.0.0 \
  -f helm/examples/values-ec2-instance-profile.yaml \
  --namespace monitoring \
  --create-namespace
```

### 3. On-premises / Any Kubernetes (with Secrets)

```bash
# Create AWS credentials secret
kubectl create secret generic aws-credentials \
  --from-literal=AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID \
  --from-literal=AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY \
  --from-literal=AWS_REGION=us-west-2 \
  --namespace monitoring

# Install chart
helm install s3-metrics-adapter helm-charts/s3-metrics-adapter \
  --version 1.0.0 \
  --set awsCredentials.existingSecret.name="aws-credentials" \
  --namespace monitoring \
  --create-namespace
```

## Values Reference

### Core Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of replicas | `1` |
| `image.repository` | Container image repository | `codebyrupinder/s3_metrics_adapter` |
| `image.tag` | Container image tag | `1.0.1` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |

### AWS Credentials

| Parameter | Description | Default |
|-----------|-------------|---------|
| `awsCredentials.create` | Create AWS credentials secret | `false` |
| `awsCredentials.existingSecret.name` | Name of existing secret | `""` |
| `awsCredentials.accessKeyId` | AWS Access Key ID | `""` |
| `awsCredentials.secretAccessKey` | AWS Secret Access Key | `""` |
| `awsCredentials.region` | AWS Region | `us-west-2` |

### Application Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `config.sqs.queues` | List of SQS queue URLs | `[]` |
| `config.sqs.workerCount` | Number of SQS workers | `5` |
| `config.sqs.maxMessages` | Max messages per request | `10` |
| `config.sqs.waitTime` | Long polling wait time | `20` |
| `config.logging.default` | Log level | `info` |

### Resources and Scaling

| Parameter | Description | Default |
|-----------|-------------|---------|
| `resources.limits.cpu` | CPU limit | `200m` |
| `resources.limits.memory` | Memory limit | `256Mi` |
| `resources.requests.cpu` | CPU request | `100m` |
| `resources.requests.memory` | Memory request | `128Mi` |
| `autoscaling.enabled` | Enable HPA | `false` |
| `autoscaling.minReplicas` | Minimum replicas | `1` |
| `autoscaling.maxReplicas` | Maximum replicas | `5` |

### Monitoring

| Parameter | Description | Default |
|-----------|-------------|---------|
| `serviceMonitor.enabled` | Enable Prometheus ServiceMonitor | `false` |
| `serviceMonitor.interval` | Scrape interval | `30s` |
| `podMonitor.enabled` | Enable Prometheus PodMonitor | `false` |
| `service.port` | Service port | `8087` |

### Security

| Parameter | Description | Default |
|-----------|-------------|---------|
| `securityContext.runAsNonRoot` | Run as non-root user | `true` |
| `securityContext.runAsUser` | User ID | `65532` |
| `securityContext.readOnlyRootFilesystem` | Read-only filesystem | `true` |
| `networkPolicy.enabled` | Enable network policies | `false` |

## Monitoring Integration

The S3 Event Exporter supports multiple monitoring approaches:

### 1. Prometheus Operator (Recommended for Managed Clusters)

If you have Prometheus Operator installed:

```bash
helm install s3-metrics-adapter helm-charts/s3-metrics-adapter \
  --version 1.0.0 \
  --set serviceMonitor.enabled=true \
  --set serviceMonitor.additionalLabels.release=prometheus \
  --namespace monitoring
```

### 2. External Prometheus (Any Prometheus Setup)

For standard Prometheus installations:

```bash
helm install s3-metrics-adapter helm-charts/s3-metrics-adapter \
  --version 1.0.0 \
  --set service.annotations.'prometheus\.io/scrape'=true \
  --set service.annotations.'prometheus\.io/port'=8087 \
  --set service.annotations.'prometheus\.io/path'=/metrics \
  --namespace monitoring
```

### 3. Check Your Setup

Not sure which method to use? Check if Prometheus Operator is installed:

```bash
# Check if ServiceMonitor CRD exists
kubectl get crd servicemonitors.monitoring.coreos.com

# If this command succeeds, use Method 1 (Prometheus Operator)
# If it fails, use Method 2 (External Prometheus)
```

**📚 Complete monitoring setup guide**: See [MONITORING.md](MONITORING.md) for detailed configuration options and troubleshooting.

### Prometheus Metrics

The application exposes metrics on port `8087` at `/metrics` endpoint:

- `s3_events_processed_total` - Total number of S3 events processed
- `s3_events_errors_total` - Total number of processing errors
- `sqs_messages_received_total` - Total SQS messages received
- `processing_duration_seconds` - Event processing duration

## Troubleshooting

### Common Issues

1. **Permission Denied Errors**
   ```bash
   # Check service account
   kubectl describe serviceaccount s3-metrics-adapter -n monitoring
   
   # Verify AWS permissions
   kubectl exec -it deployment/s3-metrics-adapter -n monitoring -- aws sts get-caller-identity
   ```

2. **SQS Connection Issues**
   ```bash
   # Check network policies
   kubectl get networkpolicy -n monitoring
   
   # Test SQS connectivity
   kubectl exec -it deployment/s3-metrics-adapter -n monitoring -- \
     aws sqs get-queue-attributes --queue-url YOUR_QUEUE_URL
   ```

3. **Pod Startup Issues**
   ```bash
   # Check pod logs
   kubectl logs -l app.kubernetes.io/name=s3-metrics-adapter -n monitoring
   
   # Describe pod
   kubectl describe pod -l app.kubernetes.io/name=s3-metrics-adapter -n monitoring
   ```

### Debug Commands

```bash
# Validate chart
helm repo add helm-charts https://code-by-rupinder.github.io/helm-charts/
helm repo update
helm lint helm-charts/s3-metrics-adapter

# Dry run install
helm install s3-metrics-adapter helm/s3_metrics_adapter/ --dry-run --debug

# Check template output
helm template s3-metrics-adapter helm/s3_metrics_adapter/ --debug

# Test chart
helm test s3-metrics-adapter -n monitoring
```

## Upgrading

### Chart Upgrades

```bash
# Update repository
helm repo update

# Upgrade to latest version
helm upgrade s3-metrics-adapter helm-charts/s3-metrics-adapter \
  --version 1.0.0 \
  --namespace monitoring

# Upgrade with new values
helm upgrade s3-metrics-adapter helm-charts/s3-metrics-adapter \
  --version 1.0.0 \
  -f new-values.yaml \
  --namespace monitoring
```

### Rolling Back

```bash
# View release history
helm history s3-metrics-adapter -n monitoring

# Rollback to previous version
helm rollback s3-metrics-adapter 1 -n monitoring
```

## Development

### Testing Locally

```bash
# Lint chart
helm lint helm/s3_metrics_adapter/

# Package chart
helm package helm/s3_metrics_adapter/

# Install locally
helm install s3-metrics-adapter ./s3-metrics-adapter-1.0.0.tgz
```

### Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes to the chart
4. Test thoroughly
5. Submit a pull request

### Release Process

1. Update version in `Chart.yaml`
2. Update `CHANGELOG.md`
3. Create a Git tag
4. GitHub Actions will automatically publish to GitHub Pages

## Support

- **Documentation**: [DEPLOYMENT.md](DEPLOYMENT.md)
- **GitHub Issues**: [Report bugs or request features](https://github.com/codebyrupinder/s3_metrics_adapter/issues)
- **GitHub Discussions**: [Community support](https://github.com/codebyrupinder/s3_metrics_adapter/discussions)

## License

This Helm chart is licensed under the MIT License. See [LICENSE](../LICENSE) file for details.

## Changelog

### v1.0.0 (2024-01-XX)
- Initial release
- Support for multiple AWS credential methods
- Production-ready templates
- Comprehensive monitoring integration
- Multi-environment deployment support
