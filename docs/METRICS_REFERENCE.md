# S3 Metrics Adapter - Metrics Documentation

## Overview
This S3 metrics adapter collects and exposes various Prometheus metrics for S3 event monitoring. All metrics are configurable and can be enabled/disabled individually through the configuration.

## Available Metrics

### 1. Event Total (`s3_event_total`)
- **Type**: Counter
- **Description**: Total number of S3 events received, including specific event subtypes
- **Labels**:
  - `event`: Main event type (e.g., "Object Created", "Object Deleted")
  - `bucket`: S3 bucket name
  - `subtype`: Event subtype (e.g., "Put", "Post", "DeleteMarkerCreated")
- **Example**: `s3_event_total{event="Object Created",bucket="my-bucket",subtype="Put"}`

### 2. Object Size (`s3_event_object_size_bytes`)
- **Type**: Histogram
- **Description**: Distribution of S3 object sizes in bytes
- **Labels**:
  - `bucket`: S3 bucket name
  - `prefix`: Object prefix (top-level directory)
- **Buckets**: Configurable, defaults to exponential buckets from 1KB to 16TB
- **Note**: Only tracks size for "Object Created" events; other events record 0


### 4. IP Total (`s3_event_ip_total`)
- **Type**: Counter
- **Description**: Total number of S3 events by source IP address
- **Labels**:
  - `ip`: Source IP address or "unknown" if not available

### 5. Prefix Total (`s3_event_prefix_total`)
- **Type**: Counter
- **Description**: Total number of S3 events by object prefix (top-level directory)
- **Labels**:
  - `prefix`: Top-level prefix or "/" for root-level objects

### 6. Hierarchical Path Total (`s3_events_hierarchical_path_total`)
- **Type**: Counter
- **Description**: Total number of S3 events grouped by hierarchical path at configured directory depth
- **Labels**:
  - `path`: Path at specified depth (e.g., "folder1/subfolder" for depth=2)
  - `bucket`: S3 bucket name
- **Configuration**: Depth is configurable via `PrefixDepth` setting

### 7. File Extension Total (`s3_bucket_extension_files_total`)
- **Type**: Gauge
- **Description**: Total number of files by extension in bucket
- **Labels**:
  - `bucket`: S3 bucket name
  - `extension`: File extension (lowercase with dot) or "none" for files without extension
  - `prefix`: Object prefix
  - `filetype`: Either "file" or "directory"
- **Behavior**:
  - Increments on "Object Created" events
  - Sets to -1 on "Object Deleted" events
- **Special Cases**:
  - Hidden files (starting with .) are marked as extension "none"
  - Directories (ending with /) are marked as filetype "directory"

### 8. Latency (`s3_event_latency_seconds`)
- **Type**: Gauge
- **Description**: Time in seconds between event creation and processing
- **Labels**:
  - `bucket`: S3 bucket name
  - `event`: Full event type

### 9. Anomaly Total (`s3_event_anomaly_total`)
- **Type**: Counter
- **Description**: Total number of detected anomalies in S3 operations
- **Labels**:
  - `type`: Type of anomaly detected
- **Anomaly Types**:
  - `system_delete`: Deletions initiated by S3 system (`s3.amazonaws.com`)
  - `delete_marker_created`: Delete marker creation events
  - `manual_delete`: Manual deletions via DeleteObject API

### 10. Lifecycle Expiration Total (`s3_lifecycle_expiration_total`)
- **Type**: Counter
- **Description**: Total number of objects deleted via lifecycle expiration policies
- **Labels**:
  - `bucket`: S3 bucket name
  - `prefix`: Object prefix
- **Trigger**: Only increments when deletion reason is "Lifecycle Expiration"

### 11. Delete Total (`s3_delete_total`)
- **Type**: Counter
- **Description**: Total number of delete events with detailed categorization
- **Labels**:
  - `bucket`: S3 bucket name
  - `deletion_type`: Type of deletion (e.g., "Delete", "DeleteMarkerCreated")
  - `reason`: Reason for deletion (e.g., "DeleteObject", "Lifecycle Expiration")

### 12. Parser Errors Total (`s3_event_parser_errors_total`)
- **Type**: Counter
- **Description**: Total number of S3 event parsing errors
- **Note**: Always enabled regardless of configuration
- **Usage**: Incremented via `IncreaseParserErrors()` method

## Configuration

### Enabling/Disabling Metrics
All metrics (except parser errors) can be individually enabled or disabled through the configuration:

```yaml
metrics:
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
```

### Object Size Buckets
The histogram buckets for object size can be customized via `config.Metrics.ObjectSizeBuckets`. If not specified, defaults to exponential buckets from 1KB to 16TB.

### Prefix Depth
The hierarchical path depth can be configured via `config.Metrics.PrefixDepth` to control how deep the path tracking goes.

## Usage Pattern

1. **Initialize**: Call `Initialize(cfg)` with your configuration
2. **Update**: Call `UpdateMetrics(event)` for each S3 event
3. **Error Tracking**: Call `IncreaseParserErrors()` when parsing fails
4. **Access**: Use `GetMetrics()` to get the global metrics instance

## Metric Categories

### Core Event Metrics
- Event Total
- Object Size
- IP Total

### Organization Metrics
- Prefix Total
- Hierarchical Path Total
- File Extension Total

### Performance Metrics
- Latency

### Operational Metrics
- Anomaly Total
- Lifecycle Expiration Total
- Delete Total
- Parser Errors Total
