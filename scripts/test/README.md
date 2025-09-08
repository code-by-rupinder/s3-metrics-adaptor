# Test Scripts

This directory contains various test scripts for the S3 Metrics Adapter.

## Load Testing Scripts

### `bulk_delete_load_test.sh`

**Purpose:** Comprehensive load testing for bulk deletion scenarios

- **Usage:** `./bulk_delete_load_test.sh [options]`
- **Options:**
  - `-f, --files N`: Number of files to test (default: 1000)
  - `-b, --batch N`: Batch size for processing (default: 50)
  - `-p, --prefix PREFIX`: Test prefix (default: bulk_test)
- **Features:**
  - Creates files with 20 different extensions
  - Monitors real-time metrics during testing
  - Tests bulk deletion performance
  - Generates performance reports
- **Example:** `./bulk_delete_load_test.sh -f 5000 -b 100`

### `quick_load_test.sh`

**Purpose:** Quick validation testing for basic functionality

- **Usage:** `./quick_load_test.sh [options]`
- **Options:**
  - `-f, --files N`: Number of files to test (default: 100)
  - `-p, --prefix PREFIX`: Test prefix (default: quick_test)
- **Features:**
  - Rapid testing and validation
  - Checks metrics after uploads and deletions
  - Validates application functionality
- **Example:** `./quick_load_test.sh -f 200`

### `test_anomaly_detection.sh`

**Purpose:** Test anomaly detection functionality

- **Usage:** `./test_anomaly_detection.sh`
- **Features:**
  - Tests normal operations (should not trigger anomalies)
  - Tests bulk operations (should not trigger anomalies)
  - Tests delete marker creation (should trigger anomalies)
  - Tests system deletions (may trigger anomalies)
  - Monitors metrics before and after each test
- **Expected Results:**
  - Normal operations: No anomalies
  - Delete markers: `delete_marker_created` anomalies
  - System deletes: `system_delete` anomalies (if S3 lifecycle configured)

## Event Simulation Scripts

### `simple_s3_events.sh`

**Purpose:** Simple S3 event simulation for basic testing

- **Usage:** `./simple_s3_events.sh`
- **Features:**
  - Creates various file types with different extensions
  - Simulates Put, Copy, Multipart Upload events
  - Basic functionality validation

### `test_different_events.sh`

**Purpose:** Test different S3 event types

- **Usage:** `./test_different_events.sh`
- **Features:**
  - Tests various S3 event scenarios
  - Validates event parsing and processing

## Path Labeling Test Scripts

### `create_test_folder_structure.sh`

**Purpose:** Create folder structure for path labeling testing

- **Usage:** `./create_test_folder_structure.sh`
- **Features:**
  - Creates department-based folder structure
  - Sets up test environment for path labeling

### `upload_files_to_structure.sh`

**Purpose:** Upload files to existing folder structure

- **Usage:** `./upload_files_to_structure.sh`
- **Features:**
  - Uploads files to department folders
  - Tests path labeling functionality

## Prerequisites

Before running any test scripts:

1. **AWS Credentials:** Ensure AWS credentials are configured
2. **S3 Bucket:** Have an S3 bucket with event notifications enabled
3. **SQS Queue:** Configure SQS queue to receive S3 events
4. **Application Running:** Start the S3 Metrics Adapter application
5. **Configuration:** Ensure test prefixes are allowed in your config

## Configuration

Add these prefixes to your `config.yaml`:

```yaml
sqs:
  buckets:
    - name: your-bucket-name
      prefix:
        - "bulk_test"
        - "quick_test"
        - "anomaly_test"
        - "extension_test"
        - "logs/test"
        - "prod"
        - "finance"
        - "hr"
        - "it"
        - "marketing"
        - "operations"
        - "legal"
        - "sales"
        - "research"
```

## Running Tests

1. **Start the application:**

   ```bash
   go run cmd/main.go --config config.yaml
   ```

2. **Run load tests:**

   ```bash
   cd scripts/test
   ./quick_load_test.sh
   ./bulk_delete_load_test.sh
   ./test_anomaly_detection.sh
   ```

3. **Monitor metrics:**

   ```bash
   curl http://localhost:8087/metrics
   ```

4. **Check Grafana dashboard:**
   - Import the sample dashboard
   - Monitor real-time metrics
   - Validate test results

## Troubleshooting

- **Events not processed:** Check if prefixes are allowed in config
- **No metrics:** Ensure application is running and accessible
- **Permission errors:** Verify AWS credentials and S3/SQS permissions
- **High cardinality:** Monitor cardinality metrics and adjust configuration

## Performance Expectations

- **Quick Load Test:** 100 files in ~30 seconds
- **Bulk Load Test:** 1000 files in ~5-10 minutes
- **Processing Rate:** 10-50 events/second (depending on system)
- **Memory Usage:** < 100MB for normal operations
- **CPU Usage:** < 50% during bulk operations
