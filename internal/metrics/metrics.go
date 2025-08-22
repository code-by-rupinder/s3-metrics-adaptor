package metrics

import (
	"fmt"
	"strings"
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
	userTotal          *prometheus.CounterVec
	ipTotal            *prometheus.CounterVec
	prefixTotal        *prometheus.CounterVec
	prefixDepthTotal   *prometheus.CounterVec
	fileExtensionTotal *prometheus.GaugeVec // Track total files by extension
	latency            *prometheus.GaugeVec
	anomalyTotal       *prometheus.CounterVec
	lifecycleTotal     *prometheus.CounterVec
	deleteTotal        *prometheus.CounterVec
	parserErrorTotal   prometheus.Counter
}

// initializeMetric initializes a single metric
func (m *Metrics) initializeMetric(metricType string) {
	switch metricType {
	case "eventTotal":
		m.eventTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "s3_event_total",
				Help: "Total number of S3 events received, including specific event subtypes (e.g., Object Created.Put)",
			},
			[]string{"event", "bucket", "subtype"},
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
			[]string{"bucket", "prefix"},
		)
		prometheus.MustRegister(m.objectSize)

	case "userTotal":
		m.userTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "s3_event_user_total",
				Help: "Total number of S3 events by user",
			},
			[]string{"user"},
		)
		prometheus.MustRegister(m.userTotal)

	case "ipTotal":
		m.ipTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "s3_event_ip_total",
				Help: "Total number of S3 events by source IP",
			},
			[]string{"ip"},
		)
		prometheus.MustRegister(m.ipTotal)

	case "prefixTotal":
		m.prefixTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "s3_event_prefix_total",
				Help: "Total number of S3 events by object prefix",
			},
			[]string{"prefix"},
		)
		prometheus.MustRegister(m.prefixTotal)

	case "prefixDepthTotal":
		m.prefixDepthTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "s3_events_hierarchical_path_total",
				Help: "Total number of S3 events grouped by hierarchical path at configured directory depth. Example: depth=2 for 'folder1/subfolder/file.txt' tracks as 'folder1/subfolder'",
			},
			[]string{"path", "bucket"},
		)
		prometheus.MustRegister(m.prefixDepthTotal)

	case "latency":
		m.latency = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "s3_event_latency_seconds",
				Help: "Time between event creation and processing",
			},
			[]string{"bucket", "event"},
		)
		prometheus.MustRegister(m.latency)

	case "anomalyTotal":
		m.anomalyTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "s3_event_anomaly_total",
				Help: "Total number of detected anomalies (system_delete, delete_marker_created, manual_delete)",
			},
			[]string{"type"},
		)
		prometheus.MustRegister(m.anomalyTotal)

	case "lifecycleTotal":
		m.lifecycleTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "s3_lifecycle_expiration_total",
				Help: "Total number of objects deleted via lifecycle expiration",
			},
			[]string{"bucket", "prefix"},
		)
		prometheus.MustRegister(m.lifecycleTotal)

	case "deleteTotal":
		m.deleteTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "s3_delete_total",
				Help: "Total number of delete events with type (e.g., Delete Marker Created) and reason (e.g., DeleteObject, Lifecycle Expiration)",
			},
			[]string{"bucket", "deletion_type", "reason"},
		)
		prometheus.MustRegister(m.deleteTotal)

	case "fileExtensionTotal":
		m.fileExtensionTotal = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "s3_bucket_extension_files_total",
				Help: "Total number of files by extension in bucket. Extension is 'none' for files without extension. Filetype is either 'file' (for all files) or 'directory' (for folders).",
			},
			[]string{"bucket", "extension", "prefix", "filetype"},
		)
		prometheus.MustRegister(m.fileExtensionTotal)
	}
}

// Initialize sets up all the metrics based on configuration
func Initialize(cfg *config.Config) *Metrics {
	metrics = &Metrics{
		config: cfg,
	}

	if !cfg.Metrics.Enabled {
		return metrics
	}

	initializeEnabledMetrics(metrics, cfg)
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
		"userTotal":          metricsConfig.UserTotal,
		"ipTotal":            metricsConfig.IPTotal,
		"prefixTotal":        metricsConfig.PrefixTotal,
		"prefixDepthTotal":   metricsConfig.PrefixDepthTotal,
		"fileExtensionTotal": metricsConfig.FileExtensionTotal,
		"latency":            metricsConfig.Latency,
		"anomalyTotal":       metricsConfig.AnomalyDetection,
		"lifecycleTotal":     metricsConfig.LifecycleExpiration,
		"deleteTotal":        metricsConfig.DeleteTotal,
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
func (m *Metrics) updateBaseMetrics(event *ParsedEvent, prefix string) {
	if m.eventTotal != nil {
		// Split event type into main type and subtype (e.g., "Object Created.Put" -> "Object Created" and "Put")
		eventType := event.EventType
		subType := ""
		if parts := strings.Split(event.EventType, "."); len(parts) > 1 {
			eventType = parts[0]
			subType = parts[1]
		}

		// Track only the main event type with subtype
		m.eventTotal.WithLabelValues(eventType, event.BucketName, subType).Inc()
	}

	// Track object size for all events, with 0 for non-creation events
	if m.objectSize != nil {
		size := float64(0)
		if event.Size > 0 && strings.HasPrefix(event.EventType, "Object Created") {
			size = float64(event.Size)
		}
		m.objectSize.WithLabelValues(event.BucketName, prefix).Observe(size)
	}

	// Track user identity, handling both IAM users and system operations
	if m.userTotal != nil {
		var userType string
		switch event.RequesterID {
		case "s3.amazonaws.com":
			userType = "system"
		case "":
			userType = "unknown"
		default:
			userType = "iam_user"
		}
		m.userTotal.WithLabelValues(event.RequesterID).Inc()
		m.userTotal.WithLabelValues(userType).Inc()
	}

	// Track source IP, with special handling for system operations
	if m.ipTotal != nil {
		sourceIP := event.SourceIP
		if sourceIP == "" {
			sourceIP = "unknown"
		}
		m.ipTotal.WithLabelValues(sourceIP).Inc()
	}
}

// updateAdvancedMetrics updates metrics related to prefixes, latency, and lifecycle
func (m *Metrics) updateAdvancedMetrics(event *ParsedEvent, prefix string) {
	if m.prefixTotal != nil {
		m.prefixTotal.WithLabelValues(prefix).Inc()
	}

	if m.prefixDepthTotal != nil {
		prefixPath := getPrefixAtDepth(event.ObjectKey, m.config.Metrics.PrefixDepth)
		if prefixPath != "" {
			m.prefixDepthTotal.WithLabelValues(prefixPath, event.BucketName).Inc()
		}
	}

	if m.latency != nil {
		latency := time.Since(event.Time).Seconds()
		m.latency.WithLabelValues(event.BucketName, event.EventType).Set(latency)
	}

	// Lifecycle expiration is handled in updateDeleteMetrics
}

// updateDeleteMetrics updates metrics related to delete operations
func (m *Metrics) updateDeleteMetrics(event *ParsedEvent) {
	isDeleteEvent := strings.HasPrefix(event.EventType, "Object Deleted")

	if m.deleteTotal != nil && isDeleteEvent {
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
		// Detect system-initiated deletions
		if event.RequesterID == "s3.amazonaws.com" {
			m.anomalyTotal.WithLabelValues("system_delete").Inc()
		}

		// Detect delete marker creation
		if strings.HasSuffix(event.EventType, "DeleteMarkerCreated") {
			m.anomalyTotal.WithLabelValues("delete_marker_created").Inc()
		}

		// Detect bulk deletions
		if event.Reason == "DeleteObject" {
			m.anomalyTotal.WithLabelValues("manual_delete").Inc()
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

	// Update file extension metrics
	if m.fileExtensionTotal != nil {
		fileInfo := getFileInfo(event.ObjectKey)
		labels := []string{
			event.BucketName,
			fileInfo.extension,
			prefix,
			fileInfo.fileType,
		}

		switch {
		case strings.HasPrefix(event.EventType, "Object Created"):
			m.fileExtensionTotal.WithLabelValues(labels...).Inc()

		case strings.HasPrefix(event.EventType, "Object Deleted"):
			// For deletion events, always set to -1
			m.fileExtensionTotal.WithLabelValues(labels...).Set(-1)
		}
	}
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

// Import the ParsedEvent from parser package
type ParsedEvent = parser.ParsedEvent
