# AWS Credentials Configuration Guide

This guide covers different methods to configure AWS credentials for the S3 Event Exporter in various Kubernetes environments.

## Credential Methods Overview

| Method | Environment | Security | Complexity | Recommended |
|--------|-------------|----------|------------|-------------|
| **Cloud IAM** | EKS | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ✅ Managed K8s |
| **Instance Profile** | Self-managed on EC2 | ⭐⭐⭐⭐ | ⭐⭐ | ✅ EC2 clusters |
| **External Secret** | Any | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ✅ Production |
| **Kubernetes Secret** | Any | ⭐⭐⭐ | ⭐⭐ | ⚠️ Development |
| **Manual EnvVars** | Any | ⭐⭐ | ⭐ | ❌ Testing only |

## Method 1: Cloud Provider IAM (Managed Kubernetes)

### EKS with IAM Roles for Service Accounts (IRSA)

**Best for**: Amazon EKS clusters

```bash
# 1. Create IAM role and policy
aws iam create-policy \
  --policy-name S3EventExporterPolicy \
  --policy-document file://iam-policy.json

aws iam create-role \
  --role-name S3EventExporterRole \
  --assume-role-policy-document file://trust-policy.json

aws iam attach-role-policy \
  --role-name S3EventExporterRole \
  --policy-arn arn:aws:iam::ACCOUNT:policy/S3EventExporterPolicy

# 2. Associate role with service account
eksctl create iamserviceaccount \
  --cluster=my-cluster \
  --namespace=monitoring \
  --name=s3-metrics-adapter \
  --role-name=S3EventExporterRole \
  --approve

# 3. Install chart
helm install s3-metrics-adapter s3-metrics-adapter/s3-metrics-adapter \
  --set serviceAccount.annotations."eks\.amazonaws\.com/role-arn"="arn:aws:iam::ACCOUNT:role/S3EventExporterRole" \
  --namespace monitoring
```


## Method 2: EC2 Instance Profile (Self-managed on EC2)

**Best for**: Self-managed Kubernetes on EC2 instances

### Setup

```bash
# 1. Create IAM role for EC2 instances
aws iam create-role \
  --role-name EC2-S3EventExporter \
  --assume-role-policy-document '{
    "Version": "2012-10-17",
    "Statement": [{
      "Effect": "Allow",
      "Principal": {"Service": "ec2.amazonaws.com"},
      "Action": "sts:AssumeRole"
    }]
  }'

# 2. Attach policy
aws iam attach-role-policy \
  --role-name EC2-S3EventExporter \
  --policy-arn arn:aws:iam::ACCOUNT:policy/S3EventExporterPolicy

# 3. Create instance profile
aws iam create-instance-profile \
  --instance-profile-name EC2-S3EventExporter

aws iam add-role-to-instance-profile \
  --instance-profile-name EC2-S3EventExporter \
  --role-name EC2-S3EventExporter

# 4. Attach to EC2 instances
aws ec2 associate-iam-instance-profile \
  --instance-id i-1234567890abcdef0 \
  --iam-instance-profile Name=EC2-S3EventExporter
```

### Installation

```bash
helm install s3-metrics-adapter s3-metrics-adapter/s3-metrics-adapter \
  -f examples/values-ec2-instance-profile.yaml \
  --namespace monitoring
```

## Method 3: External Secret Management (Recommended for Production)

**Best for**: Production environments with external secret management

### Using External Secrets Operator

```bash
# 1. Install External Secrets Operator
helm repo add external-secrets https://charts.external-secrets.io
helm install external-secrets external-secrets/external-secrets \
  --namespace external-secrets-system \
  --create-namespace

# 2. Create SecretStore (AWS Secrets Manager example)
kubectl apply -f - <<EOF
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: aws-secrets-manager
  namespace: monitoring
spec:
  provider:
    aws:
      service: SecretsManager
      region: us-west-2
      auth:
        instanceProfile: {}
EOF

# 3. Create ExternalSecret
kubectl apply -f - <<EOF
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: s3-metrics-adapter-aws-credentials
  namespace: monitoring
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: aws-secrets-manager
    kind: SecretStore
  target:
    name: aws-credentials
    creationPolicy: Owner
  data:
  - secretKey: AWS_ACCESS_KEY_ID
    remoteRef:
      key: s3-metrics-adapter-credentials
      property: access_key_id
  - secretKey: AWS_SECRET_ACCESS_KEY
    remoteRef:
      key: s3-metrics-adapter-credentials
      property: secret_access_key
  - secretKey: AWS_REGION
    remoteRef:
      key: s3-metrics-adapter-credentials
      property: region
EOF

# 4. Install chart
helm install s3-metrics-adapter s3-metrics-adapter/s3-metrics-adapter \
  --set awsCredentials.existingSecret.name="aws-credentials" \
  --namespace monitoring
```

### Using HashiCorp Vault

```bash
# 1. Configure Vault auth
kubectl apply -f - <<EOF
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: vault-backend
  namespace: monitoring
spec:
  provider:
    vault:
      server: "https://vault.example.com"
      path: "secret"
      version: "v2"
      auth:
        kubernetes:
          mountPath: "kubernetes"
          role: "s3-metrics-adapter"
EOF

# 2. Create ExternalSecret for Vault
kubectl apply -f - <<EOF
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: s3-metrics-adapter-vault-credentials
  namespace: monitoring
spec:
  refreshInterval: 30m
  secretStoreRef:
    name: vault-backend
    kind: SecretStore
  target:
    name: aws-credentials
  data:
  - secretKey: AWS_ACCESS_KEY_ID
    remoteRef:
      key: aws/s3-metrics-adapter
      property: access_key_id
  - secretKey: AWS_SECRET_ACCESS_KEY
    remoteRef:
      key: aws/s3-metrics-adapter
      property: secret_access_key
EOF
```

## Method 4: Kubernetes Secret (Development/Testing)

**Best for**: Development environments and quick testing

### Manual Secret Creation

```bash
# 1. Create secret manually
kubectl create secret generic aws-credentials \
  --from-literal=AWS_ACCESS_KEY_ID=YOUR_ACCESS_KEY \
  --from-literal=AWS_SECRET_ACCESS_KEY=YOUR_SECRET_KEY \
  --from-literal=AWS_REGION=us-west-2 \
  --namespace monitoring

# 2. Install with existing secret
helm install s3-metrics-adapter s3-metrics-adapter/s3-metrics-adapter \
  --set awsCredentials.existingSecret.name="aws-credentials" \
  --namespace monitoring
```

### Using Helm Values (Not Recommended)

```bash
# ONLY for development - credentials will be visible in values
helm install s3-metrics-adapter s3-metrics-adapter/s3-metrics-adapter \
  --set awsCredentials.create=true \
  --set-string awsCredentials.accessKeyId=YOUR_ACCESS_KEY \
  --set-string awsCredentials.secretAccessKey=YOUR_SECRET_KEY \
  --set awsCredentials.region=us-west-2 \
  --namespace monitoring
```

## Method 5: Manual Environment Variables

**Best for**: Custom secret management or testing

```yaml
# values-manual-env.yaml
env:
  - name: AWS_ACCESS_KEY_ID
    valueFrom:
      secretKeyRef:
        name: my-custom-secret
        key: access-key
  - name: AWS_SECRET_ACCESS_KEY
    valueFrom:
      secretKeyRef:
        name: my-custom-secret
        key: secret-key
  - name: AWS_REGION
    value: "us-west-2"
  - name: AWS_SESSION_TOKEN
    valueFrom:
      secretKeyRef:
        name: my-custom-secret
        key: session-token
        optional: true
```

## Required IAM Permissions

All methods require the following IAM permissions:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "sqs:ReceiveMessage",
        "sqs:DeleteMessage",
        "sqs:GetQueueAttributes",
        "sqs:GetQueueUrl"
      ],
      "Resource": "arn:aws:sqs:*:*:*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:GetObjectVersion",
        "s3:ListBucket"
      ],
      "Resource": [
        "arn:aws:s3:::your-bucket-name/*",
        "arn:aws:s3:::your-bucket-name"
      ]
    }
  ]
}
```

## Deployment Examples

### Self-managed Kubernetes (Production)

```bash
# Using external secret management
helm install s3-metrics-adapter s3-metrics-adapter/s3-metrics-adapter \
  -f examples/values-selfmanaged.yaml \
  --set awsCredentials.existingSecret.name="aws-credentials" \
  --namespace monitoring \
  --create-namespace
```

### On-premises Kubernetes

```bash
# Simple deployment with manual secret
kubectl create secret generic aws-credentials \
  --from-literal=AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID \
  --from-literal=AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY \
  --from-literal=AWS_REGION=us-west-2 \
  --namespace monitoring

helm install s3-metrics-adapter s3-metrics-adapter/s3-metrics-adapter \
  -f examples/values-onpremises.yaml \
  --namespace monitoring
```

### EC2-based Kubernetes

```bash
# Using instance profile (no secrets needed)
helm install s3-metrics-adapter s3-metrics-adapter/s3-metrics-adapter \
  -f examples/values-ec2-instance-profile.yaml \
  --namespace monitoring
```

## Security Best Practices

### 1. Use Least Privilege
- Grant only necessary permissions
- Scope to specific resources when possible
- Use resource-based policies

### 2. Rotate Credentials Regularly
- Set up automatic rotation for access keys
- Use temporary credentials when possible
- Monitor credential usage

### 3. External Secret Management
- Use external secret stores (Vault, AWS Secrets Manager)
- Implement secret rotation
- Audit secret access

### 4. Network Security
- Enable network policies
- Restrict egress to AWS endpoints only
- Use VPC endpoints when available

### 5. Monitoring and Auditing
- Monitor credential usage with CloudTrail
- Set up alerts for unusual access patterns
- Regularly audit IAM permissions

## Troubleshooting

### Common Issues

1. **Credentials not found**
   ```bash
   # Check secret exists
   kubectl get secret aws-credentials -n monitoring
   
   # Check secret contents
   kubectl describe secret aws-credentials -n monitoring
   ```

2. **Permission denied**
   ```bash
   # Test credentials
   kubectl exec -it deployment/s3-metrics-adapter -n monitoring -- \
     aws sts get-caller-identity
   
   # Check IAM permissions
   aws iam simulate-principal-policy \
     --policy-source-arn arn:aws:iam::ACCOUNT:role/ROLE_NAME \
     --action-names sqs:ReceiveMessage \
     --resource-arns arn:aws:sqs:us-west-2:ACCOUNT:QUEUE_NAME
   ```

3. **Instance profile not working**
   ```bash
   # Check instance profile
   curl http://169.254.169.254/latest/meta-data/iam/security-credentials/
   
   # Check role assumption
   kubectl exec -it deployment/s3-metrics-adapter -n monitoring -- \
     curl http://169.254.169.254/latest/meta-data/iam/security-credentials/ROLE_NAME
   ```

4. **External secrets not syncing**
   ```bash
   # Check external secret status
   kubectl describe externalsecret s3-metrics-adapter-aws-credentials -n monitoring
   
   # Check secret store
   kubectl describe secretstore aws-secrets-manager -n monitoring
   ```

## Migration Between Methods

### From Access Keys to Instance Profile

```bash
# 1. Remove existing secret
kubectl delete secret aws-credentials -n monitoring

# 2. Update deployment to use instance profile
helm upgrade s3-metrics-adapter s3-metrics-adapter/s3-metrics-adapter \
  -f examples/values-ec2-instance-profile.yaml \
  --namespace monitoring
```

### From Manual to External Secret Management

```bash
# 1. Set up external secret (see examples above)
# 2. Update deployment
helm upgrade s3-metrics-adapter s3-metrics-adapter/s3-metrics-adapter \
  --set awsCredentials.existingSecret.name="aws-credentials" \
  --namespace monitoring
```

Choose the method that best fits your Kubernetes environment and security requirements!
