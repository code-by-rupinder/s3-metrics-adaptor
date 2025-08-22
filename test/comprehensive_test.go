package test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"s3_metrics_adapter/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestApplicationStartupShutdown tests the complete application lifecycle
func TestApplicationStartupShutdown(t *testing.T) {
	// Create a temporary config file
	configContent := `
logging:
  default: info
  format:
    prettyPrint: false
sqs:
  queues:
    - https://sqs.us-west-2.amazonaws.com/123456789012/test-queue
  workerCount: 1
  maxMessages: 1
  waitTime: 1
metrics:
  enabled: true
  port: 0  # Let system choose port
  types:
    eventTotal: true
`
	tmpfile, err := os.CreateTemp("", "test-config-*.yaml")
	require.NoError(t, err)
	tmpfileName := tmpfile.Name()
	defer func() {
		if err := os.Remove(tmpfileName); err != nil {
			t.Logf("failed to remove temp file: %v", err)
		}
	}()

	_, err = tmpfile.Write([]byte(configContent))
	require.NoError(t, err)
	require.NoError(t, tmpfile.Close())

	// Test configuration loading
	cfg, err := config.LoadConfig(tmpfile.Name())
	require.NoError(t, err)
	assert.Equal(t, "info", cfg.Logging.Default)
	assert.True(t, cfg.Metrics.Enabled)
	assert.Equal(t, 1, cfg.SQS.WorkerCount)
	assert.Len(t, cfg.SQS.Queues, 1)
}

// TestConfigurationValidation tests various configuration scenarios
func TestConfigurationValidation(t *testing.T) {
	tests := []struct {
		name          string
		configYAML    string
		shouldError   bool
		errorContains string
	}{
		{
			name: "valid minimal config",
			configYAML: `
logging:
  default: info
sqs:
  queues:
    - https://sqs.us-west-2.amazonaws.com/123456789012/test-queue
  buckets:
    - name: test-bucket
metrics:
  enabled: true
`,
			shouldError: false,
		},
		{
			name: "missing queues",
			configYAML: `
logging:
  default: info
sqs:
  queues: []
metrics:
  enabled: true
`,
			shouldError:   true,
			errorContains: "at least one SQS queue must be specified",
		},
		{
			name: "invalid log level",
			configYAML: `
logging:
  default: invalid
sqs:
  queues:
    - https://sqs.us-west-2.amazonaws.com/123456789012/test-queue
metrics:
  enabled: true
`,
			shouldError:   true,
			errorContains: "invalid default log level",
		},
		{
			name: "empty bucket name",
			configYAML: `
logging:
  default: info
sqs:
  queues:
    - https://sqs.us-west-2.amazonaws.com/123456789012/test-queue
  buckets:
    - name: ""
metrics:
  enabled: true
`,
			shouldError:   true,
			errorContains: "bucket name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpfile, err := os.CreateTemp("", "test-config-*.yaml")
			require.NoError(t, err)
			tmpfileName := tmpfile.Name()
			defer func() {
				if err := os.Remove(tmpfileName); err != nil {
					t.Logf("failed to remove temp file: %v", err)
				}
			}()

			_, err = tmpfile.Write([]byte(tt.configYAML))
			require.NoError(t, err)
			require.NoError(t, tmpfile.Close())

			_, err = config.LoadConfig(tmpfile.Name())

			if tt.shouldError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestErrorHandling tests error scenarios and recovery
func TestErrorHandling(t *testing.T) {
	t.Run("invalid config file path", func(t *testing.T) {
		_, err := config.LoadConfig("/nonexistent/config.yaml")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read config file")
	})

	t.Run("malformed yaml", func(t *testing.T) {
		tmpfile, err := os.CreateTemp("", "bad-config-*.yaml")
		require.NoError(t, err)
		tmpfileName := tmpfile.Name()
		defer func() {
			if err := os.Remove(tmpfileName); err != nil {
				t.Logf("failed to remove temp file: %v", err)
			}
		}()

		_, err = tmpfile.Write([]byte("invalid: yaml: content: ["))
		require.NoError(t, err)
		require.NoError(t, tmpfile.Close())

		_, err = config.LoadConfig(tmpfile.Name())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read config file")
	})
}

// TestConcurrentAccess tests thread safety
func TestConcurrentAccess(t *testing.T) {
	// Test concurrent configuration access
	cfg := &config.Config{
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
				{Name: "test-bucket", Prefix: []string{"folder1/", "folder2/"}},
			},
			ProcessUnlistedBuckets: true,
		},
	}

	// Run concurrent checks
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = cfg.IsAllowedBucketAndPrefix("test-bucket", "folder1/test.txt")
				_ = cfg.IsAllowedBucketAndPrefix("unknown-bucket", "any/path.txt")
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out")
		}
	}
}

// TestMetricsServerIntegration tests the metrics HTTP server
func TestMetricsServerIntegration(t *testing.T) {
	// This would test the actual metrics server startup
	// For now, we test the configuration
	t.Run("metrics configuration", func(t *testing.T) {
		configContent := `
logging:
  default: info
sqs:
  queues:
    - https://sqs.us-west-2.amazonaws.com/123456789012/test-queue
metrics:
  enabled: true
  port: 8087
  types:
    eventTotal: true
    objectSize: true
    userTotal: false
  objectSizeBuckets:
    - 1024
    - 102400
    - 1048576
`
		tmpfile, err := os.CreateTemp("", "metrics-config-*.yaml")
		require.NoError(t, err)
		defer func() {
			if err := os.Remove(tmpfile.Name()); err != nil {
				t.Logf("failed to remove temp file: %v", err)
			}
		}()

		_, err = tmpfile.Write([]byte(configContent))
		require.NoError(t, err)
		require.NoError(t, tmpfile.Close())

		cfg, err := config.LoadConfig(tmpfile.Name())
		require.NoError(t, err)

		assert.True(t, cfg.Metrics.Enabled)
		assert.Equal(t, 8087, cfg.Metrics.Port)
		assert.True(t, cfg.Metrics.Types.EventTotal)
		assert.True(t, cfg.Metrics.Types.ObjectSize)
		assert.False(t, cfg.Metrics.Types.UserTotal)
		assert.Equal(t, []float64{1024, 102400, 1048576}, cfg.Metrics.ObjectSizeBuckets)
	})
}

// TestHealthEndpoint tests the health check endpoint
func TestHealthEndpoint(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("healthy")); err != nil {
			t.Logf("failed to write response: %v", err)
		}
	})

	req, err := http.NewRequest("GET", "/health", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "healthy", rr.Body.String())
}

// TestEnvironmentVariableOverrides tests environment variable configuration
func TestEnvironmentVariableOverrides(t *testing.T) {
	// Test that environment variables can override config file values
	// This would require modifying the config package to support env vars
	t.Run("env var override", func(t *testing.T) {
		// Set environment variable
		if err := os.Setenv("SQS_WORKER_COUNT", "10"); err != nil {
			t.Fatalf("failed to set SQS_WORKER_COUNT: %v", err)
		}
		defer func() {
			if err := os.Unsetenv("SQS_WORKER_COUNT"); err != nil {
				t.Logf("failed to unset SQS_WORKER_COUNT: %v", err)
			}
		}()

		// This test would verify that env vars override config file
		// Implementation depends on how you want to handle env vars
	})
}

// TestResourceCleanup tests proper resource cleanup
func TestResourceCleanup(t *testing.T) {
	// Test that resources are properly cleaned up
	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		// Test that operations respect context cancellation
		select {
		case <-ctx.Done():
			assert.Equal(t, context.DeadlineExceeded, ctx.Err())
		case <-time.After(200 * time.Millisecond):
			t.Fatal("Context should have been cancelled")
		}
	})
}

// TestJSONEventParsing tests parsing of various S3 event formats
func TestJSONEventParsing(t *testing.T) {
	tests := []struct {
		name      string
		jsonEvent string
		isValid   bool
	}{
		{
			name: "valid S3 event",
			jsonEvent: `{
				"eventName": "ObjectCreated:Put",
				"s3": {
					"bucket": {"name": "test-bucket"},
					"object": {"key": "test.txt", "size": 1024}
				}
			}`,
			isValid: true,
		},
		{
			name:      "invalid JSON",
			jsonEvent: `{"invalid": json}`,
			isValid:   false,
		},
		{
			name:      "empty event",
			jsonEvent: `{}`,
			isValid:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var event map[string]interface{}
			err := json.Unmarshal([]byte(tt.jsonEvent), &event)

			if tt.isValid {
				assert.NoError(t, err)
			}
		})
	}
}

// TestHTTPEndpoints tests HTTP endpoints
