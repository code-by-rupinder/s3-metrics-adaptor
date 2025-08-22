# S3 Metrics Adapter - Configuration Reference

This document explains each configuration parameter available in `config.yaml`, including valid values, defaults, and usage notes.

---

## logging

### `default`
- **Type:** string
- **Default:** `info`
- **Description:** Sets the default log level. Options: `debug`, `info`, `warn`, `error`.
- **When to use:** Set to `debug` for troubleshooting, `info` for normal operation.

### `components`
- **Type:** object
- **Description:** Per-component log levels. Keys: `eventparser`, `metricsexporter`, `sqspoller`.
- **When to use:** Override log level for specific components.

### `format.timestampFormat`
- **Type:** string
- **Default:** `2006-01-02T15:04:05.000Z07:00`
- **Description:** Timestamp format for logs (Go time format).

### `format.prettyPrint`
- **Type:** bool
- **Default:** `false`
- **Description:** Pretty-print logs for easier reading.

---

## sqs

### `queues`
- **Type:** list of strings
- **Description:** SQS queue URLs to poll for S3 events.
- **When to use:** Required. Add at least one queue URL.

### `buckets`
- **Type:** list of objects
- **Description:** Restrict processing to specific S3 buckets and prefixes.
- **Fields:**
  - `name` (string): S3 bucket name (required)
  - `prefix` (list of strings): Only process objects with these prefixes (optional)
- **When to use:** Use to limit processing to certain buckets or folders.

### `processUnlistedBuckets`
- **Type:** bool
- **Default:** `false`
- **Description:** If true, process events from buckets not listed in `buckets`.

### `workerCount`
- **Type:** int
- **Default:** `5`
- **Description:** Number of concurrent SQS poller workers.

### `maxMessages`
- **Type:** int
- **Default:** `10`
- **Description:** Max SQS messages to fetch per poll.

### `waitTime`
- **Type:** int
- **Default:** `20`
- **Description:** SQS long polling wait time (seconds).

---

## metrics

### `enabled`
- **Type:** bool
- **Default:** `true`
- **Description:** Enable Prometheus metrics endpoint.

### `types`
- **Type:** object
- **Description:** Enable/disable specific metric types. All are boolean.
  - `eventTotal`, `objectSize`, `userTotal`, `ipTotal`, `prefixTotal`, `prefixDepthTotal`, `fileExtensionTotal`, `latency`, `anomalyDetection`, `lifecycleExpiration`, `deleteTotal`
- **When to use:** Disable metrics you don't need to reduce resource usage.

### `prefixDepth`
- **Type:** int
- **Default:** `4`
- **Description:** How many path segments to use for prefix-based metrics.

### `objectSizeBuckets`
- **Type:** list of ints
- **Description:** Defines the bucket boundaries (in bytes) for object size histograms in Prometheus metrics. Each value represents an upper limit for a bucket. The adapter will count how many S3 objects fall into each size range when processing events.

**How it works:**
If you set:
```yaml
objectSizeBuckets:
  - 1024      # 1KB
  - 102400    # 100KB
  - 1048576   # 1MB
  - 10485760  # 10MB
```
Then the following buckets are created:
- Objects ≤ 1KB (1024 bytes)
- Objects > 1KB and ≤ 100KB (102400 bytes)
- Objects > 100KB and ≤ 1MB (1048576 bytes)
- Objects > 1MB and ≤ 10MB (10485760 bytes)
- Objects > 10MB (everything larger)

**Example:**
If your S3 events include objects of 500 bytes, 10KB, 500KB, and 20MB:
- 500 bytes → counted in the 1KB bucket
- 10KB (10240 bytes) → counted in the 100KB bucket
- 500KB (512000 bytes) → counted in the 1MB bucket
- 20MB (20971520 bytes) → counted in the overflow bucket (>10MB)

**Tip:**
Adjust these values to match the typical file sizes in your S3 buckets for more meaningful metrics.

### `port`
- **Type:** int
- **Default:** `8087`
- **Description:** Port for the Prometheus metrics HTTP server.

---

## Example

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
