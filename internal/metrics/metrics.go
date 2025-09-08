package metrics

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"s3_metrics_adapter/internal/config"
	"s3_metrics_adapter/internal/logger"
	"s3_metrics_adapter/internal/parser"
)

var metrics *Metrics

// Metrics holds all the Prometheus metrics
type Metrics struct {
	config             *config.Config
	eventTotal         *prometheus.CounterVec
	objectSize         *prometheus.HistogramVec
	ipTotal            *prometheus.CounterVec
	prefixTotal        *prometheus.CounterVec
	prefixDepthTotal   *prometheus.CounterVec
	fileExtensionTotal *prometheus.GaugeVec // Track total files by extension
	latency            *prometheus.GaugeVec
	anomalyTotal       *prometheus.CounterVec
	lifecycleTotal     *prometheus.CounterVec
	deleteTotal        *prometheus.CounterVec
	parserErrorTotal   prometheus.Counter
	// Timestamp metrics
	eventTimestamp      *prometheus.GaugeVec // Unix timestamp when S3 event occurred
	eventProcessingTime *prometheus.GaugeVec // Unix timestamp when event was processed
	eventAge            *prometheus.GaugeVec // Age of event when processed (seconds)
	// Cardinality monitoring
	cardinalityTotal *prometheus.GaugeVec // Track cardinality of each metric
	cardinalityStats map[string]int       // Track current cardinality per metric
	lastLogTime      time.Time            // Last time cardinality was logged
	cardinalityMutex sync.RWMutex         // Protect cardinality stats from concurrent access
	// Performance metrics
	messagesPerSecond *prometheus.GaugeVec     // Messages processed per second
	parseTime         *prometheus.HistogramVec // Parse time histogram
	batchSize         *prometheus.HistogramVec // Batch size histogram
}

// buildLabelNames creates label names based on metric type and configuration
func (m *Metrics) buildLabelNames(metricType string) []string {
	switch metricType {
	case "eventTotal":
		return []string{"event", "bucket", "subtype"}
	case "objectSize":
		return []string{"bucket", "prefix"}
	case "ipTotal":
		return []string{"ip"}
	case "prefixTotal":
		return []string{"prefix", "bucket"}
	case "prefixDepthTotal":
		return []string{"path", "bucket", "event", "subtype"}
	case "latency":
		return []string{"bucket", "event"}
	case "anomalyTotal":
		return []string{"type"}
	case "lifecycleTotal":
		return []string{"bucket", "prefix"}
	case "deleteTotal":
		return []string{"bucket", "deletion_type", "reason"}
	case "fileExtensionTotal":
		return []string{"bucket", "extension", "prefix", "filetype"}
	case "timestampMetrics":
		return []string{"event_type", "bucket"}
	case "cardinalityTotal":
		return []string{"metric_name", "status"}
	case "messagesPerSecond":
		return []string{"queue"}
	case "parseTime":
		return []string{"queue", "status"}
	case "batchSize":
		return []string{"queue"}
	default:
		return []string{}
	}
}

// buildLabelValues creates label values for a metric based on event data
func (m *Metrics) buildLabelValues(metricType string, event *ParsedEvent, prefix string) []string {
	switch metricType {
	case "eventTotal":
		// Split event type into main type and subtype (e.g., "Object Created.Put" -> "Object Created" and "Put")
		eventType := event.EventType
		subType := ""
		if parts := strings.Split(event.EventType, "."); len(parts) > 1 {
			eventType = parts[0]
			subType = parts[1]
		}
		return []string{eventType, event.BucketName, subType}
	case "objectSize":
		return []string{event.BucketName, prefix}
	case "ipTotal":
		ip := event.SourceIP
		if ip == "" {
			ip = "unknown"
		}
		return []string{ip}
	case "prefixTotal":
		return []string{prefix, event.BucketName}
	case "prefixDepthTotal":
		prefixPath := getPrefixAtDepth(event.ObjectKey, m.config.Metrics.PrefixDepth)
		// Split event type into main type and subtype (e.g., "Object Created.Put" -> "Object Created" and "Put")
		eventType := event.EventType
		subType := ""
		if parts := strings.Split(event.EventType, "."); len(parts) > 1 {
			eventType = parts[0]
			subType = parts[1]
		}
		return []string{prefixPath, event.BucketName, eventType, subType}
	case "latency":
		return []string{event.BucketName, event.EventType}
	case "anomalyTotal":
		anomalyType := "unknown"
		if strings.Contains(event.EventType, "Delete") {
			if event.SourceIP == "s3.amazonaws.com" {
				anomalyType = "system_delete"
			} else if strings.Contains(event.EventType, "DeleteMarkerCreated") {
				anomalyType = "delete_marker_created"
			} else {
				anomalyType = "manual_delete"
			}
		}
		return []string{anomalyType}
	case "lifecycleTotal":
		return []string{event.BucketName, prefix}
	case "deleteTotal":
		deleteType := "unknown"
		reason := "unknown"
		if strings.Contains(event.EventType, "Delete") {
			if event.SourceIP == "s3.amazonaws.com" {
				deleteType = "system"
				reason = "lifecycle_expiration"
			} else if strings.Contains(event.EventType, "DeleteMarkerCreated") {
				deleteType = "delete_marker"
				reason = "delete_object"
			} else {
				deleteType = "manual"
				reason = "delete_object"
			}
		}
		return []string{event.BucketName, deleteType, reason}
	case "fileExtensionTotal":
		fileInfo := getFileInfo(event.ObjectKey)
		return []string{event.BucketName, fileInfo.extension, prefix, fileInfo.fileType}
	case "timestampMetrics":
		return []string{event.EventType, event.BucketName}
	case "cardinalityTotal":
		// This will be handled separately in updateCardinalityMetrics
		return []string{"", ""}
	case "messagesPerSecond":
		// This will be handled separately in updatePerformanceMetrics
		return []string{""}
	case "parseTime":
		// This will be handled separately in updatePerformanceMetrics
		return []string{"", ""}
	case "batchSize":
		// This will be handled separately in updatePerformanceMetrics
		return []string{""}
	default:
		return []string{}
	}
}

// initializeMetric initializes a single metric using dynamic label configuration
func (m *Metrics) initializeMetric(metricType string) {
	labelNames := m.buildLabelNames(metricType)
	switch metricType {
	case "eventTotal":
		m.eventTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "s3_event_total",
				Help: "Total number of S3 events received, including specific event subtypes (e.g., Object Created.Put)",
			},
			labelNames,
		)
		prometheus.MustRegister(m.eventTotal)

	case "objectSize":
		buckets := m.config.Metrics.ObjectSizeBuckets
		if len(buckets) == 0 {
			buckets = prometheus.ExponentialBuckets(1024, 2, 15) // 1KB to 16TB
		}
		m.objectSize = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "s3_event_object_size_bytes",
				Help:    "Distribution of S3 object sizes",
				Buckets: buckets,
			},
			labelNames,
		)
		prometheus.MustRegister(m.objectSize)

	case "ipTotal":
		m.ipTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "s3_event_ip_total",
				Help: "Total number of S3 events by source IP",
			},
			labelNames,
		)
		prometheus.MustRegister(m.ipTotal)

	case "prefixTotal":
		m.prefixTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "s3_event_prefix_total",
				Help: "Total number of S3 events by object prefix",
			},
			labelNames,
		)
		prometheus.MustRegister(m.prefixTotal)

	case "prefixDepthTotal":
		m.prefixDepthTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "s3_events_hierarchical_path_total",
				Help: "Total number of S3 events grouped by hierarchical path at configured directory depth. Example: depth=2 for 'folder1/subfolder/file.txt' tracks as 'folder1/subfolder'",
			},
			labelNames,
		)
		prometheus.MustRegister(m.prefixDepthTotal)

	case "latency":
		m.latency = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "s3_event_latency_seconds",
				Help: "Time between event creation and processing",
			},
			labelNames,
		)
		prometheus.MustRegister(m.latency)

	case "anomalyTotal":
		m.anomalyTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "s3_event_anomaly_total",
				Help: "Total number of detected anomalies (system_delete, delete_marker_created, manual_delete)",
			},
			labelNames,
		)
		prometheus.MustRegister(m.anomalyTotal)

	case "lifecycleTotal":
		m.lifecycleTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "s3_lifecycle_expiration_total",
				Help: "Total number of objects deleted via lifecycle expiration",
			},
			labelNames,
		)
		prometheus.MustRegister(m.lifecycleTotal)

	case "deleteTotal":
		m.deleteTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "s3_delete_total",
				Help: "Total number of delete events with type (e.g., Delete Marker Created) and reason (e.g., DeleteObject, Lifecycle Expiration)",
			},
			labelNames,
		)
		prometheus.MustRegister(m.deleteTotal)

	case "fileExtensionTotal":
		m.fileExtensionTotal = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "s3_bucket_extension_files_total",
				Help: "Total number of files by extension in bucket. Extension is 'none' for files without extension. Filetype is either 'file' (for all files) or 'directory' (for folders).",
			},
			labelNames,
		)
		prometheus.MustRegister(m.fileExtensionTotal)

	case "timestampMetrics":
		// Event timestamp - when the S3 event occurred (aggregated by event type)
		m.eventTimestamp = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "s3_event_timestamp_seconds",
				Help: "Unix timestamp when the S3 event occurred (aggregated by event type)",
			},
			labelNames,
		)
		prometheus.MustRegister(m.eventTimestamp)

		// Processing timestamp - when the event was processed by the adapter (aggregated by event type)
		m.eventProcessingTime = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "s3_event_processing_timestamp_seconds",
				Help: "Unix timestamp when the S3 event was processed by the adapter (aggregated by event type)",
			},
			labelNames,
		)
		prometheus.MustRegister(m.eventProcessingTime)

		// Event age - how old the event was when processed (aggregated by event type)
		m.eventAge = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "s3_event_age_seconds",
				Help: "Age of the S3 event when processed (event_timestamp - processing_timestamp) in seconds (aggregated by event type)",
			},
			labelNames,
		)
		prometheus.MustRegister(m.eventAge)
	}
}

// Initialize sets up all the metrics based on configuration
func Initialize(cfg *config.Config) *Metrics {
	metrics = &Metrics{
		config:           cfg,
		cardinalityStats: make(map[string]int),
		lastLogTime:      time.Now(),
	}

	if !cfg.Metrics.Enabled {
		return metrics
	}

	initializeEnabledMetrics(metrics, cfg)

	// Initialize cardinality monitoring if enabled
	if cfg.Metrics.CardinalityMonitoring.Enabled {
		initializeCardinalityMonitoring(metrics, cfg)
	}

	// Initialize performance metrics (always enabled)
	initializePerformanceMetrics(metrics, cfg)

	return metrics
}

// initializeEnabledMetrics initializes all enabled metrics based on configuration
func initializeEnabledMetrics(metrics *Metrics, cfg *config.Config) {
	// Initialize parser error counter - this is always enabled
	metrics.parserErrorTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "s3_event_parser_errors_total",
		Help: "Total number of S3 event parsing errors",
	})
	prometheus.MustRegister(metrics.parserErrorTotal)

	metricsConfig := cfg.Metrics.Types
	metricTypes := map[string]bool{
		"eventTotal":         metricsConfig.EventTotal,
		"objectSize":         metricsConfig.ObjectSize,
		"ipTotal":            metricsConfig.IPTotal,
		"prefixTotal":        metricsConfig.PrefixTotal,
		"prefixDepthTotal":   metricsConfig.PrefixDepthTotal,
		"fileExtensionTotal": metricsConfig.FileExtensionTotal,
		"latency":            metricsConfig.Latency,
		"anomalyTotal":       metricsConfig.AnomalyDetection,
		"lifecycleTotal":     metricsConfig.LifecycleExpiration,
		"deleteTotal":        metricsConfig.DeleteTotal,
		"timestampMetrics":   metricsConfig.TimestampMetrics,
	}

	// Collect enabled metrics for logging
	var enabledMetrics []string
	var disabledMetrics []string

	for metricType, enabled := range metricTypes {
		if enabled {
			metrics.initializeMetric(metricType)
			enabledMetrics = append(enabledMetrics, metricType)
		} else {
			disabledMetrics = append(disabledMetrics, metricType)
		}
	}

	// Log metrics configuration
	logger.Info(logger.LogContext{
		Component: "metrics",
	}, fmt.Sprintf("Metrics configuration initialized with %d enabled and %d disabled metrics",
		len(enabledMetrics), len(disabledMetrics)))
}

// updateBaseMetrics updates the basic metrics (event total, object size, user total, IP total)
// updateBaseMetrics updates the basic metrics (event total, object size, IP total)
func (m *Metrics) updateBaseMetrics(event *ParsedEvent, prefix string) {
	if m.eventTotal != nil {
		labels := m.buildLabelValues("eventTotal", event, prefix)
		m.eventTotal.WithLabelValues(labels...).Inc()
	}

	// Track object size for all events, with 0 for non-creation events
	if m.objectSize != nil {
		size := float64(0)
		if event.Size > 0 && strings.HasPrefix(event.EventType, "Object Created") {
			size = float64(event.Size)
		}
		labels := m.buildLabelValues("objectSize", event, prefix)
		m.objectSize.WithLabelValues(labels...).Observe(size)
	}

	// Track source IP, with special handling for system operations
	if m.ipTotal != nil {
		labels := m.buildLabelValues("ipTotal", event, prefix)
		m.ipTotal.WithLabelValues(labels...).Inc()
	}
}

// updateAdvancedMetrics updates metrics related to prefixes, latency, and lifecycle
func (m *Metrics) updateAdvancedMetrics(event *ParsedEvent, prefix string) {
	if m.prefixTotal != nil {
		labels := m.buildLabelValues("prefixTotal", event, prefix)
		m.prefixTotal.WithLabelValues(labels...).Inc()
	}

	if m.prefixDepthTotal != nil {
		labels := m.buildLabelValues("prefixDepthTotal", event, prefix)
		if labels[0] != "" { // Check if path is not empty
			m.prefixDepthTotal.WithLabelValues(labels...).Inc()
		}
	}

	if m.latency != nil {
		latency := time.Since(event.Time).Seconds()
		labels := m.buildLabelValues("latency", event, prefix)
		m.latency.WithLabelValues(labels...).Set(latency)
	}

	// Lifecycle expiration is handled in updateDeleteMetrics
}

// updateTimestampMetrics updates timestamp-related metrics
func (m *Metrics) updateTimestampMetrics(event *ParsedEvent, prefix string) {
	// Only update if timestamp metrics are enabled
	if m.eventTimestamp == nil && m.eventProcessingTime == nil && m.eventAge == nil {
		return
	}

	// Get current processing time
	processingTime := time.Now()
	eventTime := event.Time

	// Use helper method to get label values
	labels := m.buildLabelValues("timestampMetrics", event, prefix)

	// Set event timestamp (when the S3 event occurred)
	if m.eventTimestamp != nil {
		m.eventTimestamp.WithLabelValues(labels...).Set(float64(eventTime.Unix()))
	}

	// Set processing timestamp (when the event was processed by the adapter)
	if m.eventProcessingTime != nil {
		m.eventProcessingTime.WithLabelValues(labels...).Set(float64(processingTime.Unix()))
	}

	// Set event age (how old the event was when processed)
	if m.eventAge != nil {
		ageSeconds := processingTime.Sub(eventTime).Seconds()
		m.eventAge.WithLabelValues(labels...).Set(ageSeconds)
	}
}

// updateDeleteMetrics updates metrics related to delete operations
func (m *Metrics) updateDeleteMetrics(event *ParsedEvent) {
	isDeleteEvent := strings.HasPrefix(event.EventType, "Object Deleted")

	if m.deleteTotal != nil && isDeleteEvent {
		// Check if this delete event should be counted based on filtering configuration
		if !m.shouldCountDeleteEvent(event) {
			return
		}

		// Extract deletion type from EventType (e.g., "Object Deleted.DeleteMarkerCreated" -> "DeleteMarkerCreated")
		deletionType := "Delete"
		if parts := strings.Split(event.EventType, "."); len(parts) > 1 {
			deletionType = parts[1]
		}

		labels := []string{
			event.BucketName,
			deletionType,
			event.Reason,
		}
		m.deleteTotal.WithLabelValues(labels...).Inc()

		// Special handling for lifecycle expirations
		if event.Reason == "Lifecycle Expiration" && m.lifecycleTotal != nil {
			prefix := getPrefix(event.ObjectKey)
			m.lifecycleTotal.WithLabelValues(event.BucketName, prefix).Inc()
		}
	}

	// Detect anomalies
	if m.anomalyTotal != nil && isDeleteEvent {
		// Only detect anomalies for delete events that are being counted
		if m.shouldCountDeleteEvent(event) {
			// Detect system-initiated deletions (lifecycle, automated)
			if event.RequesterID == "s3.amazonaws.com" {
				m.anomalyTotal.WithLabelValues("system_delete").Inc()
			}

			// Detect delete marker creation (versioning-related)
			if strings.HasSuffix(event.EventType, "DeleteMarkerCreated") {
				m.anomalyTotal.WithLabelValues("delete_marker_created").Inc()
			}

			// TODO: Implement proper anomaly detection
			// - Bulk deletion spikes (more than X deletions in Y minutes)
			// - Unusual timing (deletions outside business hours)
			// - Suspicious IPs (deletions from unexpected sources)
			// - Rapid succession (many deletions in short time)
			// - Move operations (copy + delete) should not trigger anomalies
		}
	}
}

// fileInfo holds information about a file's type and extension
type fileInfo struct {
	extension string
	fileType  string
}

// getFileInfo extracts file information from an object key
func getFileInfo(key string) fileInfo {
	// Check if it's a directory (ends with slash)
	if strings.HasSuffix(key, "/") {
		return fileInfo{
			extension: "none",
			fileType:  "directory",
		}
	}

	// Get the filename
	parts := strings.Split(key, "/")
	filename := parts[len(parts)-1]

	// Extract extension
	if idx := strings.LastIndex(filename, "."); idx >= 0 {
		// Handle hidden files (files starting with .)
		if idx == 0 {
			return fileInfo{
				extension: "none",
				fileType:  "file",
			}
		}
		return fileInfo{
			extension: strings.ToLower(filename[idx:]), // Return lowercase extension with dot
			fileType:  "file",
		}
	}

	// No extension
	return fileInfo{
		extension: "none",
		fileType:  "file",
	}
}

// UpdateMetrics updates all enabled metrics based on the event
func (m *Metrics) UpdateMetrics(event *ParsedEvent) {
	if !m.config.Metrics.Enabled {
		return
	}

	prefix := getPrefix(event.ObjectKey)

	m.updateBaseMetrics(event, prefix)
	m.updateAdvancedMetrics(event, prefix)
	m.updateDeleteMetrics(event)

	// Update cardinality monitoring
	m.logCardinalityStats()

	// Update file extension metrics
	if m.fileExtensionTotal != nil {
		labels := m.buildLabelValues("fileExtensionTotal", event, prefix)

		switch {
		case strings.HasPrefix(event.EventType, "Object Created"):
			m.fileExtensionTotal.WithLabelValues(labels...).Inc()

		case strings.HasPrefix(event.EventType, "Object Deleted"):
			// File extension metrics are counters - they only increment
			// Deletions don't decrement the count, they represent total files created
			// No action needed for delete events
		}
	}

	// Update timestamp metrics
	m.updateTimestampMetrics(event, prefix)
}

// getPrefix extracts the top-level prefix from an object key
func getPrefix(key string) string {
	parts := strings.Split(key, "/")
	if len(parts) > 1 {
		return parts[0]
	}
	return "/" // Return "/" for root-level objects
}

// getPrefixAtDepth extracts the prefix at the specified depth
func getPrefixAtDepth(key string, depth int) string {
	if depth <= 0 {
		return ""
	}

	parts := strings.Split(key, "/")
	if len(parts) < depth {
		return strings.Join(parts, "/")
	}

	return strings.Join(parts[:depth], "/")
}

// shouldCountDeleteEvent determines if a delete event should be counted based on filtering configuration
func (m *Metrics) shouldCountDeleteEvent(event *ParsedEvent) bool {
	// If delete event filtering is disabled, count all delete events
	if !m.config.Metrics.DeleteEventFiltering.Enabled {
		return true
	}

	// Extract deletion type from EventType
	deletionType := "Delete"
	if parts := strings.Split(event.EventType, "."); len(parts) > 1 {
		deletionType = parts[1]
	}

	// Check if this type of deletion should be counted
	switch deletionType {
	case "Delete":
		return m.config.Metrics.DeleteEventFiltering.IncludeActualDeletes
	case "DeleteMarkerCreated":
		return m.config.Metrics.DeleteEventFiltering.IncludeDeleteMarkers
	default:
		// For version deletions and other types, check if version deletes are included
		return m.config.Metrics.DeleteEventFiltering.IncludeVersionDeletes
	}
}

// IncreaseParserErrors increments the parser error counter
func (m *Metrics) IncreaseParserErrors() {
	if m != nil && m.parserErrorTotal != nil {
		m.parserErrorTotal.Inc()
	}
}

// GetMetrics returns the global metrics instance
func GetMetrics() *Metrics {
	return metrics
}

// initializeCardinalityMonitoring initializes cardinality monitoring metrics
func initializeCardinalityMonitoring(metrics *Metrics, cfg *config.Config) {
	metrics.cardinalityTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "s3_metrics_cardinality_total",
		Help: "Number of unique label combinations for each metric",
	}, []string{"metric_name", "status"})

	prometheus.MustRegister(metrics.cardinalityTotal)

	logger.Info(logger.LogContext{
		Component: "metrics",
		Operation: "initializeCardinalityMonitoring",
	}, "Cardinality monitoring initialized",
		"log_interval", cfg.Metrics.CardinalityMonitoring.LogInterval,
		"alert_threshold", cfg.Metrics.CardinalityMonitoring.AlertThreshold,
		"critical_threshold", cfg.Metrics.CardinalityMonitoring.CriticalThreshold)
}

// updateCardinalityMetrics updates cardinality statistics for all metrics
func (m *Metrics) updateCardinalityMetrics() {
	if m.cardinalityTotal == nil {
		return
	}

	// Get all registered metrics
	registry := prometheus.DefaultRegisterer.(*prometheus.Registry)
	metricFamilies, err := registry.Gather()
	if err != nil {
		logger.Error(logger.LogContext{
			Component: "metrics",
			Operation: "updateCardinalityMetrics",
		}, err, "Failed to gather metrics for cardinality monitoring")
		return
	}

	// Protect cardinality stats with mutex
	m.cardinalityMutex.Lock()
	defer m.cardinalityMutex.Unlock()

	// Reset cardinality stats
	m.cardinalityStats = make(map[string]int)

	// Count cardinality for each metric family
	for _, family := range metricFamilies {
		metricName := family.GetName()
		cardinality := 0

		for range family.GetMetric() {
			cardinality++
		}

		m.cardinalityStats[metricName] = cardinality

		// Update cardinality metric with status
		status := m.getCardinalityStatus(metricName, cardinality)
		m.cardinalityTotal.WithLabelValues(metricName, status).Set(float64(cardinality))
	}
}

// getCardinalityStatus determines the status of a metric based on its cardinality
func (m *Metrics) getCardinalityStatus(metricName string, cardinality int) string {
	cfg := m.config.Metrics.CardinalityMonitoring

	if cardinality >= cfg.CriticalThreshold {
		return "critical"
	} else if cardinality >= cfg.AlertThreshold {
		return "warning"
	} else if cardinality >= cfg.MaxCardinality {
		return "high"
	}
	return "normal"
}

// logCardinalityStats logs cardinality statistics if interval has passed
func (m *Metrics) logCardinalityStats() {
	if m.cardinalityTotal == nil {
		return
	}

	cfg := m.config.Metrics.CardinalityMonitoring
	now := time.Now()

	// Check if enough time has passed since last log (with mutex protection)
	m.cardinalityMutex.RLock()
	timeSinceLastLog := now.Sub(m.lastLogTime).Seconds()
	m.cardinalityMutex.RUnlock()

	if timeSinceLastLog < float64(cfg.LogInterval) {
		return
	}

	// Update last log time and update metrics (with write lock)
	m.cardinalityMutex.Lock()
	m.lastLogTime = now
	m.cardinalityMutex.Unlock()

	m.updateCardinalityMetrics()

	// Get cardinality stats for logging (with read lock)
	m.cardinalityMutex.RLock()
	cardinalityStats := make(map[string]int)
	for k, v := range m.cardinalityStats {
		cardinalityStats[k] = v
	}
	m.cardinalityMutex.RUnlock()

	// Log cardinality statistics
	logger.Info(logger.LogContext{
		Component: "metrics",
		Operation: "logCardinalityStats",
	}, "Cardinality statistics",
		"stats", cardinalityStats)

	// Check for alerts
	m.checkCardinalityAlerts(cardinalityStats)
}

// checkCardinalityAlerts checks for cardinality threshold violations
func (m *Metrics) checkCardinalityAlerts(cardinalityStats map[string]int) {
	cfg := m.config.Metrics.CardinalityMonitoring

	for metricName, cardinality := range cardinalityStats {
		if cardinality >= cfg.CriticalThreshold {
			logger.Error(logger.LogContext{
				Component: "metrics",
				Operation: "checkCardinalityAlerts",
			}, fmt.Errorf("metric %s cardinality %d exceeded critical threshold %d", metricName, cardinality, cfg.CriticalThreshold),
				"CRITICAL: Metric cardinality exceeded critical threshold")
		} else if cardinality >= cfg.AlertThreshold {
			logger.Warn(logger.LogContext{
				Component: "metrics",
				Operation: "checkCardinalityAlerts",
			}, fmt.Errorf("metric %s cardinality %d exceeded alert threshold %d", metricName, cardinality, cfg.AlertThreshold),
				"WARNING: Metric cardinality exceeded alert threshold")
		}
	}
}

// estimateCardinality estimates cardinality for a new metric before creation
func (m *Metrics) estimateCardinality(metricType string, sampleEvents []*ParsedEvent) int {
	if len(sampleEvents) == 0 {
		return 0
	}

	labelCombinations := make(map[string]bool)

	for _, event := range sampleEvents {
		prefix := getPrefix(event.ObjectKey)
		labelValues := m.buildLabelValues(metricType, event, prefix)
		labelKey := strings.Join(labelValues, "|")
		labelCombinations[labelKey] = true
	}

	return len(labelCombinations)
}

// GetCardinalityStats returns current cardinality statistics
func (m *Metrics) GetCardinalityStats() map[string]int {
	m.cardinalityMutex.RLock()
	defer m.cardinalityMutex.RUnlock()

	if m.cardinalityStats == nil {
		return make(map[string]int)
	}

	// Return a copy to avoid race conditions
	stats := make(map[string]int)
	for k, v := range m.cardinalityStats {
		stats[k] = v
	}
	return stats
}

// GetCardinalityHealth returns health status based on cardinality
func (m *Metrics) GetCardinalityHealth() map[string]interface{} {
	if m.cardinalityTotal == nil {
		return map[string]interface{}{
			"status":  "disabled",
			"message": "Cardinality monitoring is disabled",
		}
	}

	m.updateCardinalityMetrics()

	cfg := m.config.Metrics.CardinalityMonitoring
	status := "healthy"
	alerts := []string{}

	// Get cardinality stats with read lock
	m.cardinalityMutex.RLock()
	cardinalityStats := make(map[string]int)
	for k, v := range m.cardinalityStats {
		cardinalityStats[k] = v
	}
	m.cardinalityMutex.RUnlock()

	for metricName, cardinality := range cardinalityStats {
		if cardinality >= cfg.CriticalThreshold {
			status = "critical"
			alerts = append(alerts, fmt.Sprintf("%s: %d (critical threshold: %d)",
				metricName, cardinality, cfg.CriticalThreshold))
		} else if cardinality >= cfg.AlertThreshold {
			if status == "healthy" {
				status = "warning"
			}
			alerts = append(alerts, fmt.Sprintf("%s: %d (alert threshold: %d)",
				metricName, cardinality, cfg.AlertThreshold))
		}
	}

	return map[string]interface{}{
		"status":            status,
		"cardinality_stats": cardinalityStats,
		"alerts":            alerts,
		"thresholds": map[string]int{
			"alert":    cfg.AlertThreshold,
			"critical": cfg.CriticalThreshold,
			"max":      cfg.MaxCardinality,
		},
	}
}

// initializePerformanceMetrics initializes performance monitoring metrics
func initializePerformanceMetrics(metrics *Metrics, cfg *config.Config) {
	metrics.messagesPerSecond = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "s3_poller_messages_per_second",
		Help: "Number of messages processed per second",
	}, []string{"queue"})

	metrics.parseTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "s3_poller_parse_time_seconds",
		Help:    "Time taken to parse S3 events",
		Buckets: prometheus.DefBuckets,
	}, []string{"queue", "status"})

	metrics.batchSize = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "s3_poller_batch_size",
		Help:    "Number of messages processed in each batch",
		Buckets: []float64{1, 2, 5, 10, 20, 50, 100},
	}, []string{"queue"})

	prometheus.MustRegister(metrics.messagesPerSecond)
	prometheus.MustRegister(metrics.parseTime)
	prometheus.MustRegister(metrics.batchSize)

	logger.Info(logger.LogContext{
		Component: "metrics",
		Operation: "initializePerformanceMetrics",
	}, "Performance metrics initialized")
}

// UpdatePerformanceMetrics updates performance metrics
func (m *Metrics) UpdatePerformanceMetrics(queue string, messagesPerSecond float64, parseTime time.Duration, batchSize int, status string) {
	if m.messagesPerSecond != nil {
		m.messagesPerSecond.WithLabelValues(queue).Set(messagesPerSecond)
	}

	if m.parseTime != nil {
		m.parseTime.WithLabelValues(queue, status).Observe(parseTime.Seconds())
	}

	if m.batchSize != nil {
		m.batchSize.WithLabelValues(queue).Observe(float64(batchSize))
	}
}

// Import the ParsedEvent from parser package
type ParsedEvent = parser.ParsedEvent
