package test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"s3_metrics_adapter/internal/config"
	"s3_metrics_adapter/internal/poller"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	skipIntegrationMsg = "Skipping integration test in short mode"
	configPath         = "../config.yaml"
)

// extractRegionFromURL is a helper to test region extraction logic
func extractRegionFromURL(queueURL string) (string, error) {
	// Example: https://sqs.us-west-2.amazonaws.com/123456789012/queue-name
	parts := strings.Split(queueURL, ".")
	if len(parts) < 4 {
		return "", assert.AnError
	}
	return parts[1], nil
}

// TestRealAWSIntegration tests the application with real AWS resources
// This test requires:
// 1. AWS credentials configured (via ~/.aws/credentials, environment variables, or IAM role)
// 2. Real SQS queue and S3 bucket as configured in config.yaml
func TestRealAWSIntegration(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip(skipIntegrationMsg)
	}

	// Check if AWS credentials are available
	if os.Getenv("AWS_PROFILE") == "" && os.Getenv("AWS_ACCESS_KEY_ID") == "" {
		// Try to check if default credentials exist
		homeDir, err := os.UserHomeDir()
		if err != nil {
			t.Skip("Cannot determine home directory for AWS credentials")
		}

		if _, err := os.Stat(homeDir + "/.aws/credentials"); os.IsNotExist(err) {
			t.Skip("No AWS credentials found - skipping integration test")
		}
	}

	// Load actual configuration
	cfg, err := config.LoadConfig(configPath)
	require.NoError(t, err, "Failed to load config.yaml")

	// Verify we have real AWS resources configured
	require.NotEmpty(t, cfg.SQS.Queues, "No SQS queues configured")

	t.Logf("Testing with SQS queues: %v", cfg.SQS.Queues)
	t.Logf("Testing with S3 buckets: %v", cfg.SQS.Buckets)

	// Test poller creation with real config
	sqsPoller := poller.NewPoller(cfg)
	require.NotNil(t, sqsPoller)

	// Test that we can extract regions from real queue URLs (using our helper)
	for _, queueURL := range cfg.SQS.Queues {
		region, err := extractRegionFromURL(queueURL)
		if err == nil && region != "" {
			t.Logf("Queue %s appears to be in region: %s", queueURL, region)
		}
	}

	// Test brief polling (this will attempt to connect to real AWS resources)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start polling
	err = sqsPoller.StartPolling(ctx)
	assert.NoError(t, err, "Failed to start polling")

	// Let it run for a short time to test real connectivity
	time.Sleep(2 * time.Second)

	// Graceful shutdown
	sqsPoller.Shutdown()

	t.Log("Successfully completed integration test with real AWS resources")
}

// TestConfigurationWithRealResources validates that the config works with real AWS
func TestConfigurationWithRealResources(t *testing.T) {
	if testing.Short() {
		t.Skip(skipIntegrationMsg)
	}

	cfg, err := config.LoadConfig(configPath)
	require.NoError(t, err)

	// Test bucket/prefix validation with real configuration
	if len(cfg.SQS.Buckets) > 0 {
		bucket := cfg.SQS.Buckets[0]
		t.Logf("Testing bucket: %s with prefixes: %v", bucket.Name, bucket.Prefix)

		// Test allowed paths
		if len(bucket.Prefix) > 0 {
			testPath := bucket.Prefix[0] + "test-file.txt"
			allowed := cfg.IsAllowedBucketAndPrefix(bucket.Name, testPath)
			assert.True(t, allowed, "Expected path %s to be allowed for bucket %s", testPath, bucket.Name)
		}

		// Test unlisted bucket behavior
		allowed := cfg.IsAllowedBucketAndPrefix("unlisted-bucket", "any/path.txt")
		assert.Equal(t, cfg.SQS.ProcessUnlistedBuckets, allowed)
	}
}

// TestPollerWithRealQueue tests the poller against a real SQS queue
func TestPollerWithRealQueue(t *testing.T) {
	if testing.Short() {
		t.Skip(skipIntegrationMsg)
	}

	cfg, err := config.LoadConfig(configPath)
	require.NoError(t, err)

	if len(cfg.SQS.Queues) == 0 {
		t.Skip("No queues configured for testing")
	}

	sqsPoller := poller.NewPoller(cfg)
	require.NotNil(t, sqsPoller)

	// Test with a short-lived context to avoid hanging
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Start polling (this will test actual AWS connectivity)
	err = sqsPoller.StartPolling(ctx)
	assert.NoError(t, err)

	// Let it attempt to poll
	time.Sleep(1 * time.Second)

	// Shutdown should complete successfully
	done := make(chan struct{})
	go func() {
		sqsPoller.Shutdown()
		close(done)
	}()

	select {
	case <-done:
		t.Log("Poller shutdown completed successfully")
	case <-time.After(5 * time.Second):
		t.Error("Poller shutdown timed out")
	}
}

// TestMetricsServerWithRealConfig tests metrics server startup with real config
func TestMetricsServerWithRealConfig(t *testing.T) {
	if testing.Short() {
		t.Skip(skipIntegrationMsg)
	}

	cfg, err := config.LoadConfig(configPath)
	require.NoError(t, err)

	if !cfg.Metrics.Enabled {
		t.Skip("Metrics not enabled in configuration")
	}

	// Test that metrics configuration is valid
	assert.True(t, cfg.Metrics.Enabled)
	assert.Greater(t, cfg.Metrics.Port, 0)

	t.Logf("Metrics configured on port: %d", cfg.Metrics.Port)
	t.Logf("Metrics types enabled: %+v", cfg.Metrics.Types)

	// Verify metrics types configuration
	enabled := 0
	if cfg.Metrics.Types.EventTotal {
		enabled++
	}
	if cfg.Metrics.Types.ObjectSize {
		enabled++
	}
	if cfg.Metrics.Types.UserTotal {
		enabled++
	}
	if cfg.Metrics.Types.IPTotal {
		enabled++
	}
	if cfg.Metrics.Types.PrefixTotal {
		enabled++
	}
	if cfg.Metrics.Types.PrefixDepthTotal {
		enabled++
	}
	if cfg.Metrics.Types.FileExtensionTotal {
		enabled++
	}
	if cfg.Metrics.Types.Latency {
		enabled++
	}
	if cfg.Metrics.Types.AnomalyDetection {
		enabled++
	}
	if cfg.Metrics.Types.LifecycleExpiration {
		enabled++
	}
	if cfg.Metrics.Types.DeleteTotal {
		enabled++
	}

	assert.Greater(t, enabled, 0, "At least one metrics type should be enabled")
	t.Logf("Total metrics types enabled: %d", enabled)
}

// TestEnvironmentCompatibility tests that the app works in different environments
func TestEnvironmentCompatibility(t *testing.T) {
	if testing.Short() {
		t.Skip(skipIntegrationMsg)
	}

	// Test loading config with environment variables

	originalWorkerCount := os.Getenv("SQS_WORKER_COUNT")
	defer func() {
		if originalWorkerCount == "" {
			if err := os.Unsetenv("SQS_WORKER_COUNT"); err != nil {
				t.Logf("failed to unset SQS_WORKER_COUNT: %v", err)
			}
		} else {
			if err := os.Setenv("SQS_WORKER_COUNT", originalWorkerCount); err != nil {
				t.Logf("failed to restore SQS_WORKER_COUNT: %v", err)
			}
		}
	}()

	// Set environment variable
	if err := os.Setenv("SQS_WORKER_COUNT", "3"); err != nil {
		t.Fatalf("failed to set SQS_WORKER_COUNT: %v", err)
	}

	cfg, err := config.LoadConfig(configPath)
	require.NoError(t, err)

	// Environment variables would override config file if implemented
	// For now, just verify config loads successfully
	assert.NotNil(t, cfg)
	assert.Greater(t, cfg.SQS.WorkerCount, 0)
}

// TestPollerShutdownGracefully tests that shutdown works correctly with real resources
func TestPollerShutdownGracefully(t *testing.T) {
	if testing.Short() {
		t.Skip(skipIntegrationMsg)
	}

	cfg, err := config.LoadConfig(configPath)
	require.NoError(t, err)

	sqsPoller := poller.NewPoller(cfg)
	require.NotNil(t, sqsPoller)

	// Start polling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = sqsPoller.StartPolling(ctx)
	require.NoError(t, err)

	// Let it run briefly
	time.Sleep(500 * time.Millisecond)

	// Test shutdown timing
	start := time.Now()
	sqsPoller.Shutdown()
	shutdownTime := time.Since(start)

	// Shutdown should complete within reasonable time
	assert.Less(t, shutdownTime, 10*time.Second, "Shutdown took too long: %v", shutdownTime)
	t.Logf("Shutdown completed in: %v", shutdownTime)
}

// TestConfigReload tests configuration reloading (if implemented)
func TestConfigReload(t *testing.T) {
	if testing.Short() {
		t.Skip(skipIntegrationMsg)
	}

	// Load config twice to ensure it's reloadable
	cfg1, err := config.LoadConfig(configPath)
	require.NoError(t, err)

	cfg2, err := config.LoadConfig(configPath)
	require.NoError(t, err)

	// Configs should be equivalent
	assert.Equal(t, cfg1.SQS.Queues, cfg2.SQS.Queues)
	assert.Equal(t, cfg1.SQS.WorkerCount, cfg2.SQS.WorkerCount)
	assert.Equal(t, cfg1.Metrics.Enabled, cfg2.Metrics.Enabled)
}
