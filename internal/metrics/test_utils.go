package metrics

import (
	"s3_metrics_adapter/internal/config"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

// Constants for event types
const (
	// Event main types
	eventMainObjectCreated = "Object Created"
	eventMainObjectDeleted = "Object Deleted"

	// Event sub types
	eventSubTypePut                 = "Put"
	eventSubTypeDeleteMarkerCreated = "DeleteMarkerCreated"
	eventSubTypeDelete              = "Delete"
	eventSubTypeMultipartComplete   = "CompleteMultipartUpload"

	// Test data
	testBucket     = "test-bucket"
	testFolder     = "folder"
	testUser       = "test-user"
	testSourceIP   = "192.168.1.1"
	testObjectKey  = "folder/test.txt"
	testSystemUser = "s3.amazonaws.com"
	testNestedPath = "folder1/subfolder/file.txt"
	testFileNoExt  = "folder/testfile"
	testDirectory  = "folder/subfolder/"

	// Error messages
	errMetricNotInit  = "%s metric not initialized"
	errMetricValue    = "%s = %v, want %v"
	wantEventCountMsg = "eventTotal = %v, want 1"

	// Common strings
)

func createTestConfig() *config.Config {
	return &config.Config{
		Metrics: struct {
			Enabled bool `mapstructure:"enabled"`
			Types   struct {
				EventTotal          bool `mapstructure:"eventTotal"`
				ObjectSize          bool `mapstructure:"objectSize"`
				IPTotal             bool `mapstructure:"ipTotal"`
				PrefixTotal         bool `mapstructure:"prefixTotal"`
				PrefixDepthTotal    bool `mapstructure:"prefixDepthTotal"`
				FileExtensionTotal  bool `mapstructure:"fileExtensionTotal"`
				Latency             bool `mapstructure:"latency"`
				AnomalyDetection    bool `mapstructure:"anomalyDetection"`
				LifecycleExpiration bool `mapstructure:"lifecycleExpiration"`
				DeleteTotal         bool `mapstructure:"deleteTotal"`
				TimestampMetrics    bool `mapstructure:"timestampMetrics"`
			} `mapstructure:"types"`
			ObjectSizeBuckets []float64 `mapstructure:"objectSizeBuckets"`
			PrefixDepth       int       `mapstructure:"prefixDepth"`
			Port              int       `mapstructure:"port"`
			// PathLabeling - TODO: Implement in next release
			CardinalityMonitoring struct {
				Enabled           bool `mapstructure:"enabled"`           // Enable cardinality monitoring
				LogInterval       int  `mapstructure:"logInterval"`       // Interval in seconds for cardinality logging
				AlertThreshold    int  `mapstructure:"alertThreshold"`    // Alert when cardinality exceeds this value
				CriticalThreshold int  `mapstructure:"criticalThreshold"` // Critical alert threshold
				MaxCardinality    int  `mapstructure:"maxCardinality"`    // Maximum allowed cardinality per metric
			} `mapstructure:"cardinalityMonitoring"`
			DeleteEventFiltering struct {
				Enabled               bool `mapstructure:"enabled"`
				IncludeActualDeletes  bool `mapstructure:"includeActualDeletes"`
				IncludeVersionDeletes bool `mapstructure:"includeVersionDeletes"`
				IncludeDeleteMarkers  bool `mapstructure:"includeDeleteMarkers"`
			} `mapstructure:"deleteEventFiltering"`
		}{
			Enabled: true,
			Types: struct {
				EventTotal          bool `mapstructure:"eventTotal"`
				ObjectSize          bool `mapstructure:"objectSize"`
				IPTotal             bool `mapstructure:"ipTotal"`
				PrefixTotal         bool `mapstructure:"prefixTotal"`
				PrefixDepthTotal    bool `mapstructure:"prefixDepthTotal"`
				FileExtensionTotal  bool `mapstructure:"fileExtensionTotal"`
				Latency             bool `mapstructure:"latency"`
				AnomalyDetection    bool `mapstructure:"anomalyDetection"`
				LifecycleExpiration bool `mapstructure:"lifecycleExpiration"`
				DeleteTotal         bool `mapstructure:"deleteTotal"`
				TimestampMetrics    bool `mapstructure:"timestampMetrics"`
			}{
				EventTotal:          true,
				ObjectSize:          true,
				IPTotal:             true,
				PrefixTotal:         true,
				PrefixDepthTotal:    true,
				FileExtensionTotal:  true,
				Latency:             true,
				AnomalyDetection:    true,
				LifecycleExpiration: true,
				DeleteTotal:         true,
				TimestampMetrics:    true,
			},
			ObjectSizeBuckets: []float64{1024, 102400, 1048576},
			PrefixDepth:       2,
			// PathLabeling - TODO: Implement in next release
			DeleteEventFiltering: struct {
				Enabled               bool `mapstructure:"enabled"`
				IncludeActualDeletes  bool `mapstructure:"includeActualDeletes"`
				IncludeVersionDeletes bool `mapstructure:"includeVersionDeletes"`
				IncludeDeleteMarkers  bool `mapstructure:"includeDeleteMarkers"`
			}{
				Enabled:               true,
				IncludeActualDeletes:  true,
				IncludeVersionDeletes: false,
				IncludeDeleteMarkers:  true, // Enable for tests that expect delete marker anomalies
			},
		},
	}
}

func checkFileExtensionMetric(t *testing.T, m *Metrics, bucket, ext, prefix, fileType string, expected float64) {
	if m.fileExtensionTotal == nil {
		t.Errorf(errMetricNotInit, "fileExtensionTotal")
		return
	}
	count := testutil.ToFloat64(m.fileExtensionTotal.WithLabelValues(bucket, ext, prefix, fileType))
	if count != expected {
		t.Errorf(errMetricValue, "fileExtensionTotal", count, expected)
	}
}

func checkEventMetric(t *testing.T, m *Metrics, mainType, bucket, subType string, expected float64) {
	if m.eventTotal == nil {
		t.Errorf(errMetricNotInit, "eventTotal")
		return
	}
	count := testutil.ToFloat64(m.eventTotal.WithLabelValues(mainType, bucket, subType))
	if count != expected {
		t.Errorf(wantEventCountMsg, count)
	}
}

func checkPrefixDepthMetric(t *testing.T, m *Metrics, path, bucket string, expected float64) {
	if m.prefixDepthTotal == nil {
		t.Errorf(errMetricNotInit, "prefixDepthTotal")
		return
	}
	count := testutil.ToFloat64(m.prefixDepthTotal.WithLabelValues(path, bucket, "Object Created", "Put"))
	if count != expected {
		t.Errorf(errMetricValue, "prefixDepthTotal", count, expected)
	}
}

func resetMetrics(m *Metrics) {
	// Unregister all metric fields from the Metrics struct
	metricsToReset := []prometheus.Collector{
		m.eventTotal,
		m.objectSize,
		m.ipTotal,
		m.prefixTotal,
		m.prefixDepthTotal,
		m.fileExtensionTotal,
		m.latency,
		m.anomalyTotal,
		m.lifecycleTotal,
		m.deleteTotal,
		m.parserErrorTotal,
		m.eventTimestamp,
		m.eventProcessingTime,
		m.eventAge,
	}

	// Only add cardinalityTotal if it's not nil
	if m.cardinalityTotal != nil {
		metricsToReset = append(metricsToReset, m.cardinalityTotal)
	}

	// Only add performance metrics if they're not nil
	if m.messagesPerSecond != nil {
		metricsToReset = append(metricsToReset, m.messagesPerSecond)
	}
	if m.parseTime != nil {
		metricsToReset = append(metricsToReset, m.parseTime)
	}
	if m.batchSize != nil {
		metricsToReset = append(metricsToReset, m.batchSize)
	}

	// Unregister each metric if it exists
	for _, metric := range metricsToReset {
		if metric != nil {
			prometheus.DefaultRegisterer.Unregister(metric)
		}
	}

	// Reset cardinality stats
	if m.cardinalityStats != nil {
		m.cardinalityStats = make(map[string]int)
	}
}
