# S3 Metrics Adapter - Binary Deployment Guide

This guide explains how to download and run the S3 Metrics Adapter binary directly on your system.

## Download the Latest Release

Visit the [latest release page](https://github.com/code-by-rupinder/s3-metrics-adaptor/releases/latest) to download the binary for your operating system and architecture:

| OS      | Architecture | Download Filename                    |
| ------- | ------------ | ------------------------------------ |
| macOS   | amd64        | s3_metrics_adapter-darwin-amd64      |
| macOS   | arm64        | s3_metrics_adapter-darwin-arm64      |
| Linux   | amd64        | s3_metrics_adapter-linux-amd64       |
| Linux   | arm64        | s3_metrics_adapter-linux-arm64       |
| Windows | amd64        | s3_metrics_adapter-windows-amd64.exe |
| Windows | arm64        | s3_metrics_adapter-windows-arm64.exe |

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
  useEventTransformer: false # Whether EventBridge transformer is enabled

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
    timestampMetrics: true # Enable Unix timestamp metrics for events
  prefixDepth: 3 # Configures how many path segments to include in hierarchical path tracking
  objectSizeBuckets:
    - 1024.0 # 1KB
    - 102400.0 # 100KB
    - 1048576.0 # 1MB
    - 10485760.0 # 10MB
  port: 8087

  # Cardinality monitoring configuration
  cardinalityMonitoring:
    enabled: true
    logInterval: 300
    alertThreshold: 1000
    criticalThreshold: 5000
    maxCardinality: 10000

  # Delete event filtering configuration
  deleteEventFiltering:
    enabled: true
    includeActualDeletes: true
    includeVersionDeletes: false
    includeDeleteMarkers: false
```

### Configuration Options Explained

#### SQS Configuration

- `processUnlistedBuckets`: Whether to process events from buckets not explicitly listed (default: false)
- `workerCount`: Number of concurrent workers processing messages (default: 5)
- `maxMessages`: Maximum number of messages to fetch per batch (default: 10)
- `waitTime`: Long polling wait time in seconds (default: 20)
- `useEventTransformer`: Whether EventBridge transformer is enabled (default: false)

#### Metrics Configuration

- `timestampMetrics`: Enable Unix timestamp metrics for events (default: true)
- `prefixDepth`: Number of path segments to include in hierarchical path tracking (default: 3)
- `objectSizeBuckets`: Histogram buckets for object size distribution (in bytes)

#### Cardinality Monitoring

- `enabled`: Enable cardinality monitoring to prevent metric explosion (default: true)
- `logInterval`: Interval in seconds for cardinality logging (default: 300)
- `alertThreshold`: Alert when cardinality exceeds this value (default: 1000)
- `criticalThreshold`: Critical alert threshold (default: 5000)
- `maxCardinality`: Maximum allowed cardinality per metric (default: 10000)

#### Delete Event Filtering

- `enabled`: Enable delete event filtering to control what deletions are counted (default: true)
- `includeActualDeletes`: Include actual file deletions (recommended: true)
- `includeVersionDeletes`: Include version deletions (recommended: false)
- `includeDeleteMarkers`: Include delete markers (recommended: false)

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

## Load Testing

The S3 Metrics Adapter includes several load testing scripts to validate performance and functionality:

### Quick Load Test

```bash
# Test with 100 files (default)
./scripts/test/quick_load_test.sh

# Test with custom number of files
./scripts/test/quick_load_test.sh -f 500
```

### Bulk Load Test

```bash
# Full-scale load test with 1000 files
./scripts/test/bulk_delete_load_test.sh

# Custom load test
./scripts/test/bulk_delete_load_test.sh -f 5000 -b 100
```

### Anomaly Detection Test

```bash
# Test anomaly detection functionality
./scripts/test/test_anomaly_detection.sh
```

These scripts will:

- Create test files with various extensions
- Monitor metrics in real-time
- Test bulk deletion performance
- Validate anomaly detection
- Generate performance reports

## Check Version

```bash
./s3_metrics_adapter-linux-amd64 --version
```

For more details, see the [full documentation](./DOCKER_DEPLOYMENT_GUIDE.md) or the [GitHub repository](https://github.com/codebyrupinder/s3-metrics-adapter).
