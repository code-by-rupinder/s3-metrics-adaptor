# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.2.0] - 2025-01-08

### Added

- **Bulk Load Testing Scripts**: Comprehensive load testing capabilities
  - `bulk_delete_load_test.sh`: Full-scale bulk deletion testing (1000+ files)
  - `quick_load_test.sh`: Quick validation testing (100 files)
  - `test_anomaly_detection.sh`: Anomaly detection testing suite
- **Enhanced Grafana Dashboard**: Improved visualization and metrics display
  - Fixed counter metrics to use `increase()` function for time-based queries
  - Updated file extension metrics to show positive values
  - Improved anomaly detection visualization
  - Enhanced delete event tracking with proper labels
- **Performance Monitoring**: Real-time performance metrics during load tests
  - Processing rate calculation (events per second)
  - Memory and CPU usage patterns
  - System stability validation under load

### Changed

- **File Extension Metrics**: Fixed negative values on delete events
  - File extension metrics now only increment (counter behavior)
  - Delete events no longer decrement file extension counts
  - Metrics represent total files created by extension, not current files
- **Anomaly Detection**: Improved anomaly detection logic
  - Removed flawed "manual_delete" anomaly that flagged every delete operation
  - Only real anomalies are now detected (system deletes, delete markers)
  - Normal operations and bulk deletions no longer trigger false anomalies
- **Grafana Dashboard Queries**: Updated all counter metrics to use proper time-based functions
  - `s3_delete_total` now uses `increase()` for accurate time-based counting
  - `s3_event_total` queries updated for proper time range display
  - `s3_event_anomaly_total` queries fixed for meaningful anomaly visualization
  - `s3_lifecycle_expiration_total` queries updated for lifecycle tracking

### Fixed

- **Concurrent Map Writes**: Fixed race condition in cardinality monitoring
  - Added `sync.RWMutex` protection for `cardinalityStats` map
  - Prevents application crashes during high-volume operations
  - Ensures thread-safe access to shared data structures
- **Configuration Management**: Fixed configuration file handling
  - Application now properly uses updated configuration files
  - Added support for new test prefixes (`bulk_test`, `anomaly_test`)
  - Improved configuration validation and error handling
- **Event Processing**: Enhanced event processing pipeline
  - Better handling of bulk operations
  - Improved error recovery and logging
  - More accurate metric generation

### Security

- **Docker Security**: Enhanced Docker Compose security practices
  - Non-root user execution
  - Security options and resource limits
  - Pinned image versions for reproducibility
  - Health checks for all services

### Documentation

- **Updated Configuration Reference**: Comprehensive configuration documentation
- **Enhanced Binary Deployment Guide**: Updated deployment instructions
- **Load Testing Guide**: Complete guide for performance testing
- **Anomaly Detection Guide**: Testing and troubleshooting anomaly detection

## [1.1.0] - 2024-12-15

### Added

- **Cardinality Monitoring**: Track unique label combinations to prevent metric explosion
- **Delete Event Filtering**: Control which delete events are counted
- **Performance Metrics**: Messages per second, parse time, batch size tracking
- **Poller Optimizations**: Batch processing, message filtering, circuit breaker pattern
- **Timestamp Metrics**: Unix timestamp metrics for events and processing
- **Enhanced Event Parsing**: Better S3 event subtype extraction

### Changed

- **Metrics Structure**: Improved metric organization and labeling
- **Configuration Format**: Enhanced configuration options and validation
- **Helm Chart**: Updated Helm chart with new features and configurations

### Fixed

- **Event Parsing**: Fixed event type and subtype extraction
- **Metric Cardinality**: Improved metric cardinality management
- **Configuration Loading**: Better configuration file handling

## [1.0.0] - 2024-12-01

### Added

- **Initial Release**: Core S3 metrics adapter functionality
- **SQS Integration**: Process S3 events from SQS queues
- **Prometheus Metrics**: Export S3 events as Prometheus metrics
- **Grafana Dashboard**: Basic monitoring dashboard
- **Docker Support**: Containerized deployment
- **Helm Chart**: Kubernetes deployment support
- **Configuration Management**: Flexible configuration system
- **Event Types**: Support for Object Created, Object Deleted events
- **File Extension Tracking**: Track files by extension type
- **IP Address Tracking**: Monitor source IP addresses
- **Prefix Tracking**: Track S3 object prefixes
- **Object Size Metrics**: Track object size distributions
