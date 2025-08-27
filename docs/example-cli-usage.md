# S3 Event Exporter - Complete CLI Configuration

This document shows how to use all configuration parameters as command-line arguments using hierarchical dot notation.

## Hierarchical Configuration Parameters

You can now pass all configuration parameters using dot notation that mirrors your YAML structure:

### Logging Configuration

```bash
# Basic logging
--logging.default debug
--logging.format.timestampFormat "2006-01-02T15:04:05.000Z07:00"
--logging.format.prettyPrint true

# Component-specific logging
--logging.components.eventparser debug
--logging.components.metricsexporter info
--logging.components.sqspoller warn
```

### SQS Configuration

```bash
# Basic SQS settings
--sqs.workerCount 10
--sqs.maxMessages 5
--sqs.waitTime 20
--sqs.processUnlistedBuckets true
--sqs.useEventTransformer false

# SQS queues (multiple ways)
--sqs-queue "https://sqs.us-west-2.amazonaws.com/305018987196/s3-metrics-adapter-dev-s3-event-queue"
--sqs-queues "queue1,queue2,queue3"
```

### Metrics Configuration

```bash
# Basic metrics
--metrics.enabled true
--metrics.port 8087
--metrics.prefixDepth 4

# Metrics types (enable/disable specific metrics)
--metrics.types.eventTotal true
--metrics.types.objectSize true
--metrics.types.ipTotal true
--metrics.types.prefixTotal true
--metrics.types.prefixDepthTotal true
--metrics.types.fileExtensionTotal true
--metrics.types.latency true
--metrics.types.anomalyDetection true
--metrics.types.lifecycleExpiration true
--metrics.types.deleteTotal true

# Object size buckets
--metrics.objectSizeBuckets "1024,102400,1048576"
```

### Bucket Configuration

```bash
# Single bucket
--bucket "my-bucket"

# Multiple buckets
--buckets "bucket1,bucket2,bucket3"

# Process unlisted buckets
--process-unlisted-buckets true
```

## Complete Examples

### Example 1: Override specific settings
```bash
./main \
  --config config.yaml \
  --logging.default debug \
  --metrics.port 9090 \
  --sqs.workerCount 15
```

### Example 2: Complete CLI configuration (minimal config file needed)
```bash
./main \
  --config minimal.yaml \
  --logging.default info \
  --logging.format.prettyPrint true \
  --logging.format.timestampFormat "2006-01-02 15:04:05" \
  --sqs-queues "https://sqs.us-west-2.amazonaws.com/305018987196/queue1,https://sqs.us-west-2.amazonaws.com/305018987196/queue2" \
  --buckets "prod-bucket,staging-bucket" \
  --sqs.workerCount 10 \
  --sqs.maxMessages 5 \
  --sqs.processUnlistedBuckets false \
  --metrics.enabled true \
  --metrics.port 8087 \
  --metrics.types.eventTotal true \
  --metrics.types.objectSize true \
  --metrics.objectSizeBuckets "1024,102400,1048576,10485760"
```

### Example 3: Development override
```bash
./main \
  --config config.yaml \
  --logging.default debug \
  --logging.components.eventparser trace \
  --metrics.port 8088 \
  --sqs.processUnlistedBuckets true
```

## Configuration Priority

1. **Command-line arguments** (highest priority) - override everything
2. **Environment variables** (middle priority) - as supported by Viper
3. **Configuration file** (lowest priority) - default values

## Boolean Parameters

All boolean parameters accept:
- `true`, `True`, `TRUE`, `1`
- `false`, `False`, `FALSE`, `0`

## List Parameters

Comma-separated values are automatically split and trimmed:
- `--buckets "bucket1, bucket2, bucket3"` 
- `--metrics.objectSizeBuckets "1024, 2048, 4096"`

## Backward Compatibility

Legacy parameters are still supported:
- `--sqs-queue` (single queue)
- `--bucket` (single bucket) 
- `--process-unlisted-buckets`

But new hierarchical parameters are recommended for consistency.
