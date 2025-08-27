package test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"s3_metrics_adapter/internal/config"
)

// BenchmarkEventParsing tests the performance of event parsing
func BenchmarkEventParsing(b *testing.B) {
	eventJSON := `{
		"eventName": "ObjectCreated:Put",
		"eventSource": "aws:s3",
		"eventTime": "2023-08-15T12:00:00.000Z",
		"userIdentity": {
			"type": "IAMUser",
			"principalId": "AIDAIOSFODNN7EXAMPLE"
		},
		"s3": {
			"bucket": {
				"name": "test-bucket"
			},
			"object": {
				"key": "test/file.txt",
				"size": 1024,
				"eTag": "d41d8cd98f00b204e9800998ecf8427e"
			}
		}
	}`

	// For benchmarking, we'll test JSON parsing performance
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var event map[string]interface{}
		err := json.Unmarshal([]byte(eventJSON), &event)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkConfigurationLoading tests config loading performance
func BenchmarkConfigurationLoading(b *testing.B) {
	// Test configuration validation performance
	testConfig := &config.Config{
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
			Queues:      []string{"https://sqs.us-west-2.amazonaws.com/123456789012/queue1"},
			WorkerCount: 10,
			MaxMessages: 10,
			WaitTime:    20,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := config.ValidateConfig(testConfig)
		_ = err // Ignore errors for benchmark
	}
}

// BenchmarkBucketPrefixChecking tests bucket/prefix validation performance
func BenchmarkBucketPrefixChecking(b *testing.B) {
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
				{
					Name:   "bucket1",
					Prefix: []string{"folder1/", "folder2/", "folder3/"},
				},
				{
					Name:   "bucket2",
					Prefix: []string{"data/", "logs/"},
				},
				{
					Name:   "bucket3",
					Prefix: []string{}, // No prefixes
				},
			},
			ProcessUnlistedBuckets: true,
		},
	}

	testCases := []struct {
		bucket string
		prefix string
	}{
		{"bucket1", "folder1/file.txt"},
		{"bucket2", "data/file.csv"},
		{"bucket3", "any/path/file.json"},
		{"unknown-bucket", "some/file.txt"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testCase := testCases[i%len(testCases)]
		cfg.IsAllowedBucketAndPrefix(testCase.bucket, testCase.prefix)
	}
}

// BenchmarkConcurrentBucketChecking tests concurrent access performance
func BenchmarkConcurrentBucketChecking(b *testing.B) {
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

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cfg.IsAllowedBucketAndPrefix("test-bucket", "folder1/test.txt")
		}
	})
}

// BenchmarkMemoryUsage tests memory allocation patterns
func BenchmarkMemoryUsage(b *testing.B) {
	eventJSON := `{
		"eventName": "ObjectCreated:Put",
		"s3": {
			"bucket": {"name": "test-bucket"},
			"object": {"key": "test/large/path/file.txt", "size": 1048576}
		}
	}`

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var event map[string]interface{}
		err := json.Unmarshal([]byte(eventJSON), &event)
		if err != nil {
			b.Fatal(err)
		}
		_ = event // Use the result to prevent optimization
	}
}

// BenchmarkContextCancellation tests context handling performance
func BenchmarkContextCancellation(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)

		select {
		case <-ctx.Done():
			// Expected case
		case <-time.After(10 * time.Millisecond):
			b.Fatal("Context should have been cancelled")
		}

		cancel()
	}
}

// BenchmarkStringOperations tests common string operations used in the app
func BenchmarkStringOperations(b *testing.B) {
	paths := []string{
		"folder1/subfolder/file.txt",
		"data/2023/08/15/events.json",
		"logs/application/error.log",
		"archive/2023/Q3/backup.tar.gz",
	}

	b.Run("HasPrefix", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			path := paths[i%len(paths)]
			_ = hasPrefix(path, "folder1/")
		}
	})

	b.Run("SplitPath", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			path := paths[i%len(paths)]
			_ = splitPath(path)
		}
	})
}

// Helper functions for benchmarks
func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func splitPath(path string) []string {
	// Simple path splitting simulation
	result := make([]string, 0, 4)
	current := ""
	for _, r := range path {
		if r == '/' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(r)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
