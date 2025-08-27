# S3 Metrics Adapter - Binary Deployment Guide

This guide explains how to download and run the S3 Metrics Adapter binary directly on your system.

## Download the Latest Release

Visit the [latest release page](https://github.com/code-by-rupinder/s3-metrics-adaptor/releases/latest) to download the binary for your operating system and architecture:

| OS      | Architecture | Download Filename                       |
|---------|--------------|-----------------------------------------|
| macOS   | amd64        | s3_metrics_adapter-darwin-amd64         |
| macOS   | arm64        | s3_metrics_adapter-darwin-arm64         |
| Linux   | amd64        | s3_metrics_adapter-linux-amd64          |
| Linux   | arm64        | s3_metrics_adapter-linux-arm64          |
| Windows | amd64        | s3_metrics_adapter-windows-amd64.exe    |
| Windows | arm64        | s3_metrics_adapter-windows-arm64.exe    |

Example direct download link for Linux amd64:
```
curl -LO https://github.com/codebyrupinder/s3-metrics-adapter/releases/latest/download/s3_metrics_adapter-linux-amd64
chmod +x s3_metrics_adapter-linux-amd64
```

## Verify the Download (Optional)

Each release provides a SHA256 checksum. You can verify the binary:
```
sha256sum s3_metrics_adapter-linux-amd64
# Compare the output to the checksum listed on the release page
```

## Create a Configuration File


Create a `config.yaml` file with your desired settings. Minimal example:
```yaml
logging:
  default: info
sqs:
  queues:
    - https://sqs.us-west-2.amazonaws.com/123456789/your-s3-event-queue
  buckets:
    - name: your-s3-bucket-name
      prefix:
        - logs/
        - data/
metrics:
  enabled: true
  port: 8087
```

### Detailed Configuration Example

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
      prefix:
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

See the [full documentation](./DOCKER_DEPLOYMENT_GUIDE.md) for more configuration options and explanations.

## Run the Binary

```bash
./s3_metrics_adapter-linux-amd64 -config /path/to/config.yaml
```

## Provide AWS Credentials

You can provide AWS credentials in one of the following ways:
- Mount or place your AWS credentials/config files in `~/.aws/` (default location)
- Set environment variables:
  ```bash
  export AWS_ACCESS_KEY_ID=your-access-key
  export AWS_SECRET_ACCESS_KEY=your-secret-key
  export AWS_REGION=us-west-2
  ```
- If running on an EC2 instance or ECS task with an attached IAM role, credentials will be picked up automatically.

## Check Version

```bash
./s3_metrics_adapter-linux-amd64 --version
```

For more details, see the [full documentation](./DOCKER_DEPLOYMENT_GUIDE.md) or the [GitHub repository](https://github.com/codebyrupinder/s3-metrics-adapter).
