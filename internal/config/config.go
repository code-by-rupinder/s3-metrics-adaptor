package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"s3_metrics_adapter/internal/logger"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	Logging struct {
		Default    string            `mapstructure:"default" yaml:"default"`       // Default log level
		Components map[string]string `mapstructure:"components" yaml:"components"` // Component-specific log levels
		Format     struct {
			TimestampFormat string `mapstructure:"timestampFormat" yaml:"timestampFormat"`
			PrettyPrint     bool   `mapstructure:"prettyPrint" yaml:"prettyPrint"`
		} `mapstructure:"format" yaml:"format"`
	} `mapstructure:"logging" yaml:"logging"`

	SQS struct {
		Queues  []string `mapstructure:"queues"`
		Buckets []struct {
			Name   string   `mapstructure:"name"`
			Prefix []string `mapstructure:"prefix"`
		} `mapstructure:"buckets"`
		ProcessUnlistedBuckets bool `mapstructure:"processUnlistedBuckets"`
		WorkerCount            int  `mapstructure:"workerCount"`
		MaxMessages            int  `mapstructure:"maxMessages"`
		WaitTime               int  `mapstructure:"waitTime"`
		UseEventTransformer    bool `mapstructure:"useEventTransformer"` // Whether EventBridge transformer is enabled
	} `mapstructure:"sqs"`
	Metrics struct {
		Enabled bool `mapstructure:"enabled"`
		Types   struct {
			EventTotal          bool `mapstructure:"eventTotal"`
			ObjectSize          bool `mapstructure:"objectSize"`
			UserTotal           bool `mapstructure:"userTotal"`
			IPTotal             bool `mapstructure:"ipTotal"`
			PrefixTotal         bool `mapstructure:"prefixTotal"`
			PrefixDepthTotal    bool `mapstructure:"prefixDepthTotal"`
			FileExtensionTotal  bool `mapstructure:"fileExtensionTotal"` // Track total files by extension in bucket/prefix
			Latency             bool `mapstructure:"latency"`
			AnomalyDetection    bool `mapstructure:"anomalyDetection"`
			LifecycleExpiration bool `mapstructure:"lifecycleExpiration"`
			DeleteTotal         bool `mapstructure:"deleteTotal"`
		} `mapstructure:"types"`
		ObjectSizeBuckets []float64 `mapstructure:"objectSizeBuckets"`
		PrefixDepth       int       `mapstructure:"prefixDepth"`
		Port              int       `mapstructure:"port"`
	} `mapstructure:"metrics"`
}

// LoadConfig loads the config from the given file path using Viper
// validateLogLevels checks if the provided log levels are valid
func validateLogLevels(defaultLevel string, componentLevels map[string]string) error {
	// Validate default level
	if _, err := logrus.ParseLevel(defaultLevel); err != nil {
		return fmt.Errorf("invalid default log level '%s': %w", defaultLevel, err)
	}

	// Validate component levels
	for component, level := range componentLevels {
		if _, err := logrus.ParseLevel(level); err != nil {
			return fmt.Errorf("invalid log level '%s' for component '%s': %w", level, component, err)
		}
	}

	return nil
}

func LoadConfig(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml") // Explicitly tell viper this is YAML
	v.AutomaticEnv()        // Enable environment variable support

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := v.UnmarshalExact(&cfg); err != nil { // Use UnmarshalExact to be strict about config structure
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	// Configure logging
	if cfg.Logging.Default == "" {
		cfg.Logging.Default = "info"
	}
	if cfg.Logging.Format.TimestampFormat == "" {
		cfg.Logging.Format.TimestampFormat = time.RFC3339Nano
	}
	if cfg.Logging.Components == nil {
		cfg.Logging.Components = make(map[string]string)
	}

	// Validate logging levels
	if err := validateLogLevels(cfg.Logging.Default, cfg.Logging.Components); err != nil {
		return nil, fmt.Errorf("invalid logging configuration: %w", err)
	}

	// Initialize logger with component-specific levels
	if err := logger.InitLogger(cfg.Logging.Default, os.Stdout, cfg.Logging.Format.PrettyPrint, cfg.Logging.Components); err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Validate the rest of the configuration
	if err := ValidateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// ValidateConfig validates the loaded config
func ValidateConfig(cfg *Config) error {
	if len(cfg.SQS.Queues) == 0 {
		return errors.New("at least one SQS queue must be specified")
	}
	for _, b := range cfg.SQS.Buckets {
		if b.Name == "" {
			return errors.New("bucket name cannot be empty")
		}
	}

	// Set defaults for SQS polling if not specified
	if cfg.SQS.WorkerCount <= 0 {
		cfg.SQS.WorkerCount = 5
	}
	if cfg.SQS.MaxMessages <= 0 {
		cfg.SQS.MaxMessages = 10
	}
	if cfg.SQS.WaitTime <= 0 {
		cfg.SQS.WaitTime = 20
	}
	return nil
}

// IsAllowedBucketAndPrefix checks if the bucket and prefix are allowed
func (cfg *Config) IsAllowedBucketAndPrefix(bucket, prefix string) bool {
	logger.Debug(logger.LogContext{
		ExtraFields: map[string]interface{}{
			"bucket":           bucket,
			"object_key":       prefix,
			"check_type":       "start_check",
			"config_buckets":   cfg.SQS.Buckets,
			"process_unlisted": cfg.SQS.ProcessUnlistedBuckets,
		},
	}, "Starting bucket/prefix check")

	// First check if the bucket is listed in configuration
	for _, b := range cfg.SQS.Buckets {
		logger.Debug(logger.LogContext{
			ExtraFields: map[string]interface{}{
				"checking_bucket": b.Name,
				"current_bucket":  bucket,
				"prefix_rules":    b.Prefix,
			},
		}, "Checking against configured bucket")

		if b.Name == bucket {
			// This is a listed bucket, check prefix rules
			logger.Debug(logger.LogContext{
				ExtraFields: map[string]interface{}{
					"bucket":       bucket,
					"object_key":   prefix,
					"prefix_rules": b.Prefix,
					"check_type":   "listed_bucket",
				},
			}, "Found matching bucket in config")

			// If no prefixes specified for this bucket, allow all events from it
			if len(b.Prefix) == 0 {
				logger.Debug(logger.LogContext{
					ExtraFields: map[string]interface{}{
						"bucket": bucket,
						"reason": "no_prefixes_specified",
					},
				}, "All prefixes allowed for bucket")
				return true
			}

			// Check if the object matches any of the allowed prefixes
			for _, p := range b.Prefix {
				logger.Debug(logger.LogContext{
					ExtraFields: map[string]interface{}{
						"bucket":          bucket,
						"object_key":      prefix,
						"checking_prefix": p,
					},
				}, "Checking prefix")

				if strings.HasPrefix(prefix, p) {
					logger.Info(logger.LogContext{
						ExtraFields: map[string]interface{}{
							"bucket":          bucket,
							"object_key":      prefix,
							"matching_prefix": p,
						},
					}, "Prefix matched")
					return true
				}
			}

			// Bucket is listed but prefix doesn't match
			logger.Info(logger.LogContext{
				ExtraFields: map[string]interface{}{
					"bucket":           bucket,
					"object_key":       prefix,
					"allowed_prefixes": b.Prefix,
				},
			}, "Event blocked: prefix not allowed for listed bucket")
			return false
		}
	}

	// Bucket is not listed, check processUnlistedBuckets
	allowed := cfg.SQS.ProcessUnlistedBuckets
	logger.Info(logger.LogContext{
		ExtraFields: map[string]interface{}{
			"bucket":           bucket,
			"object_key":       prefix,
			"check_type":       "unlisted_bucket",
			"process_unlisted": cfg.SQS.ProcessUnlistedBuckets,
			"allowed":          allowed,
		},
	}, "Checking unlisted bucket")
	return allowed
}
