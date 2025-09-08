package config

import (
	"os"
	"testing"
)

const (
	testBucketName = "test-bucket"
	queueURL       = "https://sqs.us-west-2.amazonaws.com/123456789012/test-queue"

	// Constants for bucket prefix tests
	s3BucketName  = "s3-bucket"
	s3Bucket2Name = "s3-bucket-2"
	folder1       = "folder-1"
	folder2       = "folder-2"
)

func verifySQSConfig(t *testing.T, cfg *Config) {
	t.Helper()
	if len(cfg.SQS.Queues) != 1 {
		t.Errorf("Expected 1 queue, got %d", len(cfg.SQS.Queues))
	}

	if len(cfg.SQS.Buckets) != 2 {
		t.Errorf("Expected 2 buckets, got %d", len(cfg.SQS.Buckets))
	}

	if cfg.SQS.WorkerCount != 3 {
		t.Errorf("Expected worker count 3, got %d", cfg.SQS.WorkerCount)
	}
}

func verifyLoggingConfig(t *testing.T, cfg *Config) {
	t.Helper()
	if cfg.Logging.Default != "info" {
		t.Errorf("Expected default log level 'info', got %s", cfg.Logging.Default)
	}

	if cfg.Logging.Components["sqspoller"] != "debug" {
		t.Errorf("Expected sqspoller log level 'debug', got %s", cfg.Logging.Components["sqspoller"])
	}

	if cfg.Logging.Format.TimestampFormat != "2006-01-02T15:04:05.000Z07:00" {
		t.Errorf("Expected RFC3339Nano timestamp format, got %s", cfg.Logging.Format.TimestampFormat)
	}
}

func verifyMetricsConfig(t *testing.T, cfg *Config) {
	t.Helper()
	if !cfg.Metrics.Enabled {
		t.Error("Expected metrics to be enabled")
	}

	if cfg.Metrics.Port != 8087 {
		t.Errorf("Expected metrics port 8087, got %d", cfg.Metrics.Port)
	}

	if !cfg.Metrics.Types.EventTotal {
		t.Error("Expected eventTotal metric to be enabled")
	}

	if !cfg.Metrics.Types.PrefixDepthTotal {
		t.Error("Expected prefixDepthTotal metric to be enabled")
	}

	if !cfg.Metrics.Types.FileExtensionTotal {
		t.Error("Expected fileExtensionTotal metric to be enabled")
	}

	if !cfg.Metrics.Types.TimestampMetrics {
		t.Error("Expected timestampMetrics to be enabled")
	}

	if cfg.Metrics.PrefixDepth != 2 {
		t.Errorf("Expected prefix depth 2, got %d", cfg.Metrics.PrefixDepth)
	}

	expectedBuckets := []float64{1024, 102400, 1048576}
	if len(cfg.Metrics.ObjectSizeBuckets) != len(expectedBuckets) {
		t.Errorf("Expected %d object size buckets, got %d", len(expectedBuckets), len(cfg.Metrics.ObjectSizeBuckets))
	}
	for i, bucket := range expectedBuckets {
		if cfg.Metrics.ObjectSizeBuckets[i] != bucket {
			t.Errorf("Expected bucket %f at position %d, got %f", bucket, i, cfg.Metrics.ObjectSizeBuckets[i])
		}
	}
}

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tmpConfig := `logging:
  default: info
  components:
    sqspoller: debug
    eventparser: info
    metricsexporter: info
  format:
    timestampFormat: "2006-01-02T15:04:05.000Z07:00"
    prettyPrint: false

metrics:
  enabled: true
  port: 8087
  types:
    eventTotal: true
    objectSize: true
    ipTotal: true
    prefixTotal: true
    prefixDepthTotal: true
    fileExtensionTotal: true
    latency: true
    anomalyDetection: false
    lifecycleExpiration: true
    deleteTotal: true
    timestampMetrics: true
  prefixDepth: 2
  objectSizeBuckets:
    - 1024
    - 102400
    - 1048576

sqs:
  queues:
    - https://sqs.us-west-2.amazonaws.com/123456789012/test-queue
  buckets:
    - name: test-bucket
      prefix:
        - folder1/
        - folder2/
    - name: test-bucket-2
  processUnlistedBuckets: true
  useEventTransformer: true
  workerCount: 3
  maxMessages: 5
  waitTime: 10
`
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	tmpfileName := tmpfile.Name()
	defer func() {
		if err := os.Remove(tmpfileName); err != nil {
			t.Logf("failed to remove temp file: %v", err)
		}
	}()

	if _, err := tmpfile.Write([]byte(tmpConfig)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Test loading config
	cfg, err := LoadConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Print the loaded config for debugging
	t.Logf("Loaded config: %+v", cfg)
	t.Logf("Logging components: %+v", cfg.Logging.Components)

	// Verify configuration using helper functions
	verifySQSConfig(t, cfg)
	verifyLoggingConfig(t, cfg)
	verifyMetricsConfig(t, cfg)
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &Config{
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
						"SQSPoller": "debug",
					},
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
					DeleteEventFiltering: struct {
						Enabled               bool `mapstructure:"enabled"`
						IncludeActualDeletes  bool `mapstructure:"includeActualDeletes"`
						IncludeVersionDeletes bool `mapstructure:"includeVersionDeletes"`
						IncludeDeleteMarkers  bool `mapstructure:"includeDeleteMarkers"`
					}{
						Enabled:               true,
						IncludeActualDeletes:  true,
						IncludeVersionDeletes: false,
						IncludeDeleteMarkers:  false,
					},
					// PathLabeling - TODO: Implement in next release
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
					Queues: []string{queueURL},
					Buckets: []struct {
						Name   string   `mapstructure:"name"`
						Prefix []string `mapstructure:"prefix"`
					}{{Name: testBucketName}},
					UseEventTransformer: true,
				},
			},
			wantErr: false,
		},
		{
			name: "no queues",
			cfg: &Config{
				Logging: struct {
					Default    string            `mapstructure:"default" yaml:"default"`
					Components map[string]string `mapstructure:"components" yaml:"components"`
					Format     struct {
						TimestampFormat string `mapstructure:"timestampFormat" yaml:"timestampFormat"`
						PrettyPrint     bool   `mapstructure:"prettyPrint" yaml:"prettyPrint"`
					} `mapstructure:"format" yaml:"format"`
				}{
					Default: "info",
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
					Queues: []string{},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Removed unused constants

func TestIsAllowedBucketAndPrefix(t *testing.T) {
	cfg := &Config{
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
			Buckets: []struct {
				Name   string   `mapstructure:"name"`
				Prefix []string `mapstructure:"prefix"`
			}{
				{
					Name:   s3BucketName,
					Prefix: []string{folder1, folder2},
				},
				{
					Name:   s3Bucket2Name,
					Prefix: []string{},
				},
			},
			ProcessUnlistedBuckets: true,
			UseEventTransformer:    true,
		},
	}

	tests := []struct {
		name     string
		bucket   string
		prefix   string
		expected bool
	}{
		// Test s3-bucket with folder-1
		{"s3-bucket with folder-1", s3BucketName, folder1 + "/file.txt", true},
		{"s3-bucket with folder-1 root", s3BucketName, folder1, true},
		{"s3-bucket with folder-1 subfolder", s3BucketName, folder1 + "/subfolder/file.txt", true},

		// Test s3-bucket with folder-2
		{"s3-bucket with folder-2", s3BucketName, folder2 + "/file.txt", true},
		{"s3-bucket with folder-2 root", s3BucketName, folder2, true},
		{"s3-bucket with folder-2 subfolder", s3BucketName, folder2 + "/subfolder/file.txt", true},

		// Test s3-bucket with non-allowed prefix
		{"s3-bucket with wrong prefix", s3BucketName, "folder-3/file.txt", false},
		{"s3-bucket with wrong prefix root", s3BucketName, "folder-3", false},

		// Test s3-bucket-2 (no prefix restrictions)
		{"s3-bucket-2 with any prefix", s3Bucket2Name, "any/path/file.txt", true},
		{"s3-bucket-2 root file", s3Bucket2Name, "file.txt", true},
		{"s3-bucket-2 with folder", s3Bucket2Name, "folder/file.txt", true},

		// Test unlisted bucket (should be allowed due to processUnlistedBuckets: true)
		{"unlisted bucket", "other-bucket", "some/path/file.txt", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cfg.IsAllowedBucketAndPrefix(tt.bucket, tt.prefix)
			if result != tt.expected {
				t.Errorf("IsAllowedBucketAndPrefix() = %v, want %v", result, tt.expected)
			}
		})
	}
}
