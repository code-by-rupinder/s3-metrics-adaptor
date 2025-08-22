# Integration Tests

This directory contains integration tests for the S3 Event Exporter application. These tests verify that all components work together correctly and test the complete data flow from event parsing to metrics export.

## Test Coverage

### 1. HTTP Endpoints Integration (`TestHTTPEndpoints`)
- Tests the `/health` endpoint returns proper health status
- Tests the `/metrics` endpoint serves Prometheus metrics correctly
- Verifies HTTP response codes and content types

### 2. Event Processing Integration (`TestEventProcessingIntegration`)
- Tests the complete pipeline: event parsing → metrics updating → HTTP export
- Verifies that parsed events correctly update metrics
- Tests anomaly detection integration (system deletions)
- Validates metrics are accessible via HTTP endpoints

### 3. Configuration Integration (`TestConfigurationIntegration`)
- Tests loading valid YAML configuration files
- Tests configuration validation (rejects invalid configs)
- Verifies all configuration options are properly parsed
- Tests component-specific logging configuration

### 4. Parser Integration (`TestParserIntegration`)
- Tests parsing of different S3 event formats:
  - EventBridge format
  - Legacy S3 notification format
  - SQS-wrapped messages
- Tests error handling for invalid JSON
- Verifies parsed event fields are correct

### 5. End-to-End Metrics (`TestMetricsEndToEnd`)
- Tests processing multiple events of different types
- Verifies all metric types are generated correctly:
  - Event counters
  - Prefix tracking
  - Anomaly detection
  - File extension tracking
  - Hierarchical path tracking
- Tests root-level file handling (prefix="/")
- Validates metric labels and values via HTTP export

## Running the Tests

### Run all integration tests:
```bash
go test ./test/... -v
```

### Run specific test:
```bash
go test ./test/ -run TestHTTPEndpoints -v
```

### Run with coverage:
```bash
go test ./test/... -cover -v
```

## Test Dependencies

The integration tests use:
- `testify/assert` and `testify/require` for assertions
- `httptest` for HTTP server testing
- Temporary files for configuration testing
- In-memory buffers for logging capture

## Test Data

Tests use predefined constants for consistency:
- `testBucketName`: "test-bucket"
- `testObjectKey`: "test.txt"
- `eventTypeCreatedPut`: "Object Created.Put"

## Anomaly Detection Testing

The tests verify that the following anomalies are properly detected:
- **System deletions**: Events with `RequesterID="s3.amazonaws.com"`
- **Delete marker creation**: Events with `EventName` ending in "DeleteMarkerCreated"
- **Manual deletions**: Events with `Reason="DeleteObject"`

## Benefits of Integration Tests

1. **End-to-End Validation**: Ensures all components work together
2. **Regression Prevention**: Catches issues that unit tests might miss
3. **Configuration Testing**: Validates real-world configuration scenarios
4. **HTTP API Testing**: Ensures metrics are properly exposed
5. **Multiple Format Support**: Tests different S3 event formats
6. **Anomaly Detection**: Validates security-related features

## Complementing Unit Tests

These integration tests work alongside the existing unit tests:
- **Unit tests**: Test individual components in isolation
- **Integration tests**: Test component interactions and data flow

Both types of tests are important for comprehensive coverage and confidence in the application's reliability.
