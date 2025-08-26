

# S3 Metrics Adapter Helm Chart

Helm chart for deploying S3 Metrics Adapter on Kubernetes clusters.

---

## Prerequisites

- Kubernetes 1.19+
- Helm 3.x
- Access to an SQS queue with S3 event notifications
- AWS credentials (via IRSA, EC2 instance profile, or secret)

---

## Usage Overview

1. **Copy and edit the sample values.yaml**
   - `cp helm/s3_metrics_adapter/values.yaml my-values.yaml`
   - Edit `my-values.yaml` to set your SQS queue(s), AWS auth, and any custom config.
2. **Install the chart**
   - `helm repo add helm-charts https://code-by-rupinder.github.io/helm-charts/`
   - `helm repo update`
   - `helm install s3-metrics-adapter helm-charts/s3-metrics-adapter -n monitoring --create-namespace -f my-values.yaml`

3. **Upgrade**
  - `helm upgrade s3-metrics-adapter helm-charts/s3-metrics-adapter -n monitoring -f my-values.yaml`

---

## AWS Authentication Options

**EKS (IRSA):**
- Set `serviceAccount.annotations.eks.amazonaws.com/role-arn` in your values file.
- Example:
  ```yaml
  serviceAccount:
    annotations:
      eks.amazonaws.com/role-arn: "arn:aws:iam::ACCOUNT:role/S3EventExporterRole"
  ```

**Self-managed K8s (EC2 Instance Profile):**
- No extra config needed if nodes have proper IAM role.

**Self-managed K8s (AWS Secret):**
- Create a secret:
  ```bash
  kubectl create secret generic aws-credentials \
    --from-literal=AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID \
    --from-literal=AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY \
    --from-literal=AWS_REGION=us-west-2 \
    --namespace monitoring
  ```
- Reference it in your values file:
  ```yaml
  awsCredentials:
    existingSecret:
      name: aws-credentials
  ```

---

## Minimal Example values.yaml

```yaml
config:
  sqs:
    queues:
      - "https://sqs.us-west-2.amazonaws.com/123456789/your-queue"
  logging:
    default: info
serviceAccount:
  create: true
  annotations:
    eks.amazonaws.com/role-arn: "arn:aws:iam::ACCOUNT:role/S3EventExporterRole" # For EKS IRSA, else leave blank
```

---

## Common Commands

**Install:**
```bash
helm install s3-metrics-adapter helm-charts/s3-metrics-adapter -n monitoring --create-namespace -f my-values.yaml
```

**Upgrade:**
```bash
helm upgrade s3-metrics-adapter helm-charts/s3-metrics-adapter -n monitoring -f my-values.yaml
```



---

## Troubleshooting

- Check pod logs:
  ```bash
  kubectl logs -l app.kubernetes.io/name=s3-metrics-adapter -n monitoring
  ```
- Describe pod for events:
  ```bash
  kubectl describe pod -l app.kubernetes.io/name=s3-metrics-adapter -n monitoring
  ```
- Validate Helm values:
  ```bash
  helm lint helm-charts/s3-metrics-adapter
  helm template s3-metrics-adapter helm-charts/s3-metrics-adapter -f my-values.yaml
  ```

---

## More Information

- See `values.yaml` for all configuration options.
- For application usage, see the main project README.
- For advanced deployment, see `DEPLOYMENT.md` in this directory.
For more configuration options, see the chart's `values.yaml` or the full documentation.

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



## Support

- **Documentation**: [DEPLOYMENT.md](DEPLOYMENT.md)
- **GitHub Issues**: [Report bugs or request features](https://github.com/codebyrupinder/s3_metrics_adapter/issues)

## License

This Helm chart is licensed under the MIT License. See [LICENSE](../LICENSE) file for details.

## Changelog

### v1.0.0 (2024-01-XX)
- Initial release
- Support for multiple AWS credential methods
- Production-ready templates
- Comprehensive monitoring integration
- Multi-environment deployment support
