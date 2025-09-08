package main

import (
	"flag"
	"os"
	"strings"
	"testing"

	"s3_metrics_adapter/internal/config"

	"github.com/stretchr/testify/assert"
)

// resetFlags resets all command line flags for testing
func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	// Redefine all flags
	configFile = flag.String("config", "config.yaml", "Path to configuration file")
	listenAddr = flag.String("listen-address", ":8087", "The address to listen on for HTTP requests")

	// Logging flags
	loggingDefault = flag.String("logging.default", "", "Default log level (debug, info, warn, error)")
	loggingComponentEvent = flag.String("logging.components.eventparser", "", "Log level for eventparser component")
	loggingComponentMetrics = flag.String("logging.components.metricsexporter", "", "Log level for metricsexporter component")
	loggingComponentSQS = flag.String("logging.components.sqspoller", "", "Log level for sqspoller component")
	loggingTimestampFormat = flag.String("logging.format.timestampFormat", "", "Timestamp format for logs")
	loggingPrettyPrint = flag.String("logging.format.prettyPrint", "", "Enable pretty-printed log output (true/false)")

	// SQS flags
	sqsQueue = flag.String("sqs-queue", "", "Single SQS queue URL (overrides config file)")
	sqsQueues = flag.String("sqs-queues", "", "Comma-separated list of SQS queue URLs (overrides config file)")
	sqsWorkerCount = flag.Int("sqs.workerCount", 0, "Number of SQS worker threads (0 uses config file)")
	sqsMaxMessages = flag.Int("sqs.maxMessages", 0, "Max messages per SQS request (0 uses config file)")
	sqsWaitTime = flag.Int("sqs.waitTime", 0, "SQS long polling wait time in seconds (0 uses config file)")
	sqsProcessUnlistedBuckets = flag.String("sqs.processUnlistedBuckets", "", "Process events from unlisted buckets (true/false)")
	sqsUseEventTransformer = flag.String("sqs.useEventTransformer", "", "Use EventBridge transformer (true/false)")

	// Metrics flags
	metricsEnabled = flag.String("metrics.enabled", "", "Enable metrics collection (true/false)")
	metricsPort = flag.Int("metrics.port", 0, "Metrics server port (0 uses config file)")
	metricsPrefixDepth = flag.Int("metrics.prefixDepth", 0, "Prefix depth for hierarchical tracking (0 uses config file)")
	metricsTypesEventTotal = flag.String("metrics.types.eventTotal", "", "Enable event total metrics (true/false)")
	metricsTypesObjectSize = flag.String("metrics.types.objectSize", "", "Enable object size metrics (true/false)")
	metricsTypesIPTotal = flag.String("metrics.types.ipTotal", "", "Enable IP total metrics (true/false)")
	metricsTypesPrefixTotal = flag.String("metrics.types.prefixTotal", "", "Enable prefix total metrics (true/false)")
	metricsTypesPrefixDepthTotal = flag.String("metrics.types.prefixDepthTotal", "", "Enable prefix depth metrics (true/false)")
	metricsTypesFileExtensionTotal = flag.String("metrics.types.fileExtensionTotal", "", "Enable file extension metrics (true/false)")
	metricsTypesLatency = flag.String("metrics.types.latency", "", "Enable latency metrics (true/false)")
	metricsTypesAnomalyDetection = flag.String("metrics.types.anomalyDetection", "", "Enable anomaly detection metrics (true/false)")
	metricsTypesLifecycleExpiration = flag.String("metrics.types.lifecycleExpiration", "", "Enable lifecycle expiration metrics (true/false)")
	metricsTypesDeleteTotal = flag.String("metrics.types.deleteTotal", "", "Enable delete total metrics (true/false)")
	metricsObjectSizeBuckets = flag.String("metrics.objectSizeBuckets", "", "Comma-separated list of object size bucket boundaries")

	// Bucket flags
	bucketName = flag.String("bucket", "", "Single S3 bucket name to monitor (overrides config file)")
	bucketNames = flag.String("buckets", "", "Comma-separated list of S3 bucket names to monitor (overrides config file)")
	processUnlistedBucketsFlag = flag.String("process-unlisted-buckets", "", "Process events from unlisted buckets (true/false)")
}

func TestParseBool(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
		isSet    bool
	}{
		{"", false, false},
		{"true", true, true},
		{"True", true, true},
		{"TRUE", true, true},
		{"1", true, true},
		{"false", false, true},
		{"False", false, true},
		{"FALSE", false, true},
		{"0", false, true},
		{"invalid", false, false},
		{"maybe", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			val, isSet := parseBool(tt.input)
			assert.Equal(t, tt.expected, val, "Expected value %v for input %s", tt.expected, tt.input)
			assert.Equal(t, tt.isSet, isSet, "Expected isSet %v for input %s", tt.isSet, tt.input)
		})
	}
}

func TestParseFloat64Slice(t *testing.T) {
	tests := []struct {
		input    string
		expected []float64
	}{
		{"", nil},
		{"1024", []float64{1024}},
		{"1024,2048", []float64{1024, 2048}},
		{"1024, 2048, 4096", []float64{1024, 2048, 4096}},
		{"1024.5,2048.7", []float64{1024.5, 2048.7}},
		{"invalid", nil},                             // parseFloat64Slice returns nil for invalid input
		{"1024,invalid,2048", []float64{1024, 2048}}, // skips invalid values
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseFloat64Slice(tt.input)
			assert.Equal(t, tt.expected, result, "Expected %v for input %s", tt.expected, tt.input)
		})
	}
}

func TestConfigurationOverride(t *testing.T) {
	// Create a base config
	cfg := &config.Config{
		Logging: struct {
			Default    string            `mapstructure:"default" yaml:"default"`
			Components map[string]string `mapstructure:"components" yaml:"components"`
			Format     struct {
				TimestampFormat string `mapstructure:"timestampFormat" yaml:"timestampFormat"`
				PrettyPrint     bool   `mapstructure:"prettyPrint" yaml:"prettyPrint"`
			} `mapstructure:"format" yaml:"format"`
		}{
			Default: "info",
			Components: map[string]string{
				"eventparser": "info",
			},
			Format: struct {
				TimestampFormat string `mapstructure:"timestampFormat" yaml:"timestampFormat"`
				PrettyPrint     bool   `mapstructure:"prettyPrint" yaml:"prettyPrint"`
			}{
				TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
				PrettyPrint:     false,
			},
		},
		SQS: struct {
			Queues  []string `mapstructure:"queues"`
			Buckets []struct {
				Name   string   `mapstructure:"name"`
				Prefix []string `mapstructure:"prefix"`
			} `mapstructure:"buckets"`
			ProcessUnlistedBuckets bool `mapstructure:"processUnlistedBuckets"`
			WorkerCount            int  `mapstructure:"workerCount"`
			MaxMessages            int  `mapstructure:"maxMessages"`
			WaitTime               int  `mapstructure:"waitTime"`
			UseEventTransformer    bool `mapstructure:"useEventTransformer"`
		}{
			Queues:      []string{"https://sqs.us-west-2.amazonaws.com/123456789012/original-queue"},
			WorkerCount: 5,
			MaxMessages: 10,
			WaitTime:    20,
		},
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
			Port:    8087,
			// PathLabeling - TODO: Implement in next release
		},
	}

	resetFlags()

	// Test logging configuration override
	t.Run("logging override", func(t *testing.T) {
		// Simulate command line: --logging.default debug --logging.format.prettyPrint true
		os.Args = []string{"cmd", "--logging.default", "debug", "--logging.format.prettyPrint", "true"}
		flag.Parse()

		// Apply overrides (simplified version of what main() does)
		if *loggingDefault != "" {
			cfg.Logging.Default = *loggingDefault
		}
		if prettyPrint, isSet := parseBool(*loggingPrettyPrint); isSet {
			cfg.Logging.Format.PrettyPrint = prettyPrint
		}

		assert.Equal(t, "debug", cfg.Logging.Default)
		assert.True(t, cfg.Logging.Format.PrettyPrint)
	})

	resetFlags()

	// Test SQS configuration override
	t.Run("sqs override", func(t *testing.T) {
		// Simulate command line: --sqs.workerCount 15 --sqs-queues "queue1,queue2"
		os.Args = []string{"cmd", "--sqs.workerCount", "15", "--sqs-queues", "queue1,queue2"}
		flag.Parse()

		// Apply overrides
		if *sqsWorkerCount > 0 {
			cfg.SQS.WorkerCount = *sqsWorkerCount
		}
		if *sqsQueues != "" {
			queueList := strings.Split(*sqsQueues, ",")
			for i := range queueList {
				queueList[i] = strings.TrimSpace(queueList[i])
			}
			cfg.SQS.Queues = queueList
		}

		assert.Equal(t, 15, cfg.SQS.WorkerCount)
		assert.Equal(t, []string{"queue1", "queue2"}, cfg.SQS.Queues)
	})

	resetFlags()

	// Test metrics configuration override
	t.Run("metrics override", func(t *testing.T) {
		// Simulate command line: --metrics.enabled false --metrics.port 9090
		os.Args = []string{"cmd", "--metrics.enabled", "false", "--metrics.port", "9090"}
		flag.Parse()

		// Apply overrides
		if enabled, isSet := parseBool(*metricsEnabled); isSet {
			cfg.Metrics.Enabled = enabled
		}
		if *metricsPort > 0 {
			cfg.Metrics.Port = *metricsPort
		}

		assert.False(t, cfg.Metrics.Enabled)
		assert.Equal(t, 9090, cfg.Metrics.Port)
	})
}

func TestBucketListParsing(t *testing.T) {
	resetFlags()

	t.Run("single bucket via buckets flag", func(t *testing.T) {
		os.Args = []string{"cmd", "--buckets", "bucket1"}
		flag.Parse()

		// Simulate bucket handling logic
		var bucketList []string
		if *bucketNames != "" {
			bucketList = strings.Split(*bucketNames, ",")
			for i := range bucketList {
				bucketList[i] = strings.TrimSpace(bucketList[i])
			}
		}

		expected := []string{"bucket1"}
		assert.Equal(t, expected, bucketList)
	})

	resetFlags()

	t.Run("multiple buckets via buckets flag", func(t *testing.T) {
		os.Args = []string{"cmd", "--buckets", "bucket1, bucket2, bucket3"}
		flag.Parse()

		// Simulate bucket handling logic
		var bucketList []string
		if *bucketNames != "" {
			bucketList = strings.Split(*bucketNames, ",")
			for i := range bucketList {
				bucketList[i] = strings.TrimSpace(bucketList[i])
			}
		}

		expected := []string{"bucket1", "bucket2", "bucket3"}
		assert.Equal(t, expected, bucketList)
	})

	resetFlags()

	t.Run("single bucket via bucket flag", func(t *testing.T) {
		os.Args = []string{"cmd", "--bucket", "single-bucket"}
		flag.Parse()

		// Simulate bucket handling logic
		var bucketList []string
		if *bucketName != "" {
			bucketList = []string{*bucketName}
		}

		expected := []string{"single-bucket"}
		assert.Equal(t, expected, bucketList)
	})
}

func TestObjectSizeBucketsParsing(t *testing.T) {
	resetFlags()

	t.Run("object size buckets parsing", func(t *testing.T) {
		os.Args = []string{"cmd", "--metrics.objectSizeBuckets", "1024, 2048, 4096.5"}
		flag.Parse()

		buckets := parseFloat64Slice(*metricsObjectSizeBuckets)
		expected := []float64{1024, 2048, 4096.5}
		assert.Equal(t, expected, buckets)
	})
}
