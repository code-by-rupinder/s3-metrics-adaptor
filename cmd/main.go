package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"s3_metrics_adapter/internal/config"
	"s3_metrics_adapter/internal/logger"
	"s3_metrics_adapter/internal/metrics"
	"s3_metrics_adapter/internal/poller"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Version information (set during build)
	Version   string = "dev"
	BuildTime string = "unknown"
	GitCommit string = "unknown"
	GitBranch string = "unknown"

	// Command line flags
	configFile = flag.String("config", "config.yaml", "Path to configuration file")
	listenAddr = flag.String("listen-address", ":8087", "The address to listen on for HTTP requests")
	version    = flag.Bool("version", false, "Show version information and exit")

	// Logging flags with hierarchical naming
	loggingDefault          = flag.String("logging.default", "", "Default log level (debug, info, warn, error)")
	loggingComponentEvent   = flag.String("logging.components.eventparser", "", "Log level for eventparser component")
	loggingComponentMetrics = flag.String("logging.components.metricsexporter", "", "Log level for metricsexporter component")
	loggingComponentSQS     = flag.String("logging.components.sqspoller", "", "Log level for sqspoller component")
	loggingTimestampFormat  = flag.String("logging.format.timestampFormat", "", "Timestamp format for logs")
	loggingPrettyPrint      = flag.String("logging.format.prettyPrint", "", "Enable pretty-printed log output (true/false)")

	// SQS flags
	sqsQueue                  = flag.String("sqs-queue", "", "Single SQS queue URL (overrides config file)")
	sqsQueues                 = flag.String("sqs-queues", "", "Comma-separated list of SQS queue URLs (overrides config file)")
	sqsWorkerCount            = flag.Int("sqs.workerCount", 0, "Number of SQS worker threads (0 uses config file)")
	sqsMaxMessages            = flag.Int("sqs.maxMessages", 0, "Max messages per SQS request (0 uses config file)")
	sqsWaitTime               = flag.Int("sqs.waitTime", 0, "SQS long polling wait time in seconds (0 uses config file)")
	sqsProcessUnlistedBuckets = flag.String("sqs.processUnlistedBuckets", "", "Process events from unlisted buckets (true/false)")
	sqsUseEventTransformer    = flag.String("sqs.useEventTransformer", "", "Use EventBridge transformer (true/false)")

	// Metrics flags
	metricsEnabled                  = flag.String("metrics.enabled", "", "Enable metrics collection (true/false)")
	metricsPort                     = flag.Int("metrics.port", 0, "Metrics server port (0 uses config file)")
	metricsPrefixDepth              = flag.Int("metrics.prefixDepth", 0, "Prefix depth for hierarchical tracking (0 uses config file)")
	metricsTypesEventTotal          = flag.String("metrics.types.eventTotal", "", "Enable event total metrics (true/false)")
	metricsTypesObjectSize          = flag.String("metrics.types.objectSize", "", "Enable object size metrics (true/false)")
	metricsTypesIPTotal             = flag.String("metrics.types.ipTotal", "", "Enable IP total metrics (true/false)")
	metricsTypesPrefixTotal         = flag.String("metrics.types.prefixTotal", "", "Enable prefix total metrics (true/false)")
	metricsTypesPrefixDepthTotal    = flag.String("metrics.types.prefixDepthTotal", "", "Enable prefix depth metrics (true/false)")
	metricsTypesFileExtensionTotal  = flag.String("metrics.types.fileExtensionTotal", "", "Enable file extension metrics (true/false)")
	metricsTypesLatency             = flag.String("metrics.types.latency", "", "Enable latency metrics (true/false)")
	metricsTypesAnomalyDetection    = flag.String("metrics.types.anomalyDetection", "", "Enable anomaly detection metrics (true/false)")
	metricsTypesLifecycleExpiration = flag.String("metrics.types.lifecycleExpiration", "", "Enable lifecycle expiration metrics (true/false)")
	metricsTypesDeleteTotal         = flag.String("metrics.types.deleteTotal", "", "Enable delete total metrics (true/false)")
	metricsObjectSizeBuckets        = flag.String("metrics.objectSizeBuckets", "", "Comma-separated list of object size bucket boundaries")

	// Bucket flags (backward compatibility)
	bucketName                 = flag.String("bucket", "", "Single S3 bucket name to monitor (overrides config file)")
	bucketNames                = flag.String("buckets", "", "Comma-separated list of S3 bucket names to monitor (overrides config file)")
	processUnlistedBucketsFlag = flag.String("process-unlisted-buckets", "", "Process events from unlisted buckets (true/false)")
)

// Helper function to parse boolean strings
func parseBool(s string) (bool, bool) {
	if s == "" {
		return false, false // not set
	}
	val, err := strconv.ParseBool(s)
	if err != nil {
		return false, false // invalid value, treat as not set
	}
	return val, true // valid value
}

// Helper function to parse float64 slice from comma-separated string
func parseFloat64Slice(s string) []float64 {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var result []float64
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if val, err := strconv.ParseFloat(part, 64); err == nil {
			result = append(result, val)
		}
	}
	return result
}

// showVersion displays version information
func showVersion() {
	fmt.Printf("s3_metrics_adapter version %s\n", Version)
	fmt.Printf("Build time: %s\n", BuildTime)
	fmt.Printf("Git commit: %s\n", GitCommit)
	fmt.Printf("Git branch: %s\n", GitBranch)
	fmt.Println()
	fmt.Println("s3_metrics_adapter - S3 Metrics Adapter")
	fmt.Println("Adapts S3 events from SQS queues into Prometheus metrics")
	fmt.Println("Documentation: https://github.com/codebyrupinder/s3_metrics_adapter")
}

func main() {
	// Parse command line flags
	flag.Parse()

	// Show version and exit if requested
	if *version {
		showVersion()
		return
	}

	mainCtx := logger.LogContext{
		Operation: "Startup",
		Component: "Main",
		ExtraFields: map[string]interface{}{
			"version": Version,
		},
	}

	// Load configuration
	logger.Info(mainCtx, "Loading configuration file: "+*configFile)
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		logger.Fatal(mainCtx, err, "Failed to load configuration")
	}

	// Override config with command line flags if provided
	// Logging configuration
	if *loggingDefault != "" {
		cfg.Logging.Default = *loggingDefault
	}
	if *loggingComponentEvent != "" {
		if cfg.Logging.Components == nil {
			cfg.Logging.Components = make(map[string]string)
		}
		cfg.Logging.Components["eventparser"] = *loggingComponentEvent
	}
	if *loggingComponentMetrics != "" {
		if cfg.Logging.Components == nil {
			cfg.Logging.Components = make(map[string]string)
		}
		cfg.Logging.Components["metricsexporter"] = *loggingComponentMetrics
	}
	if *loggingComponentSQS != "" {
		if cfg.Logging.Components == nil {
			cfg.Logging.Components = make(map[string]string)
		}
		cfg.Logging.Components["sqspoller"] = *loggingComponentSQS
	}
	if *loggingTimestampFormat != "" {
		cfg.Logging.Format.TimestampFormat = *loggingTimestampFormat
	}
	if prettyPrint, isSet := parseBool(*loggingPrettyPrint); isSet {
		cfg.Logging.Format.PrettyPrint = prettyPrint
	}

	// Handle SQS queues (both single and list)
	var queueList []string
	if *sqsQueue != "" {
		queueList = []string{*sqsQueue}
	}
	if *sqsQueues != "" {
		// Split comma-separated queue URLs
		queueList = strings.Split(*sqsQueues, ",")
		for i := range queueList {
			queueList[i] = strings.TrimSpace(queueList[i])
		}
	}
	if len(queueList) > 0 {
		cfg.SQS.Queues = queueList
	}
	if *sqsWorkerCount > 0 {
		cfg.SQS.WorkerCount = *sqsWorkerCount
	}
	if *sqsMaxMessages > 0 {
		cfg.SQS.MaxMessages = *sqsMaxMessages
	}
	if *sqsWaitTime > 0 {
		cfg.SQS.WaitTime = *sqsWaitTime
	}
	if processUnlisted, isSet := parseBool(*sqsProcessUnlistedBuckets); isSet {
		cfg.SQS.ProcessUnlistedBuckets = processUnlisted
	}
	if useTransformer, isSet := parseBool(*sqsUseEventTransformer); isSet {
		cfg.SQS.UseEventTransformer = useTransformer
	}

	// Metrics configuration
	if enabled, isSet := parseBool(*metricsEnabled); isSet {
		cfg.Metrics.Enabled = enabled
	}
	if *metricsPort > 0 {
		cfg.Metrics.Port = *metricsPort
	}
	if *metricsPrefixDepth > 0 {
		cfg.Metrics.PrefixDepth = *metricsPrefixDepth
	}

	// Metrics types configuration
	if eventTotal, isSet := parseBool(*metricsTypesEventTotal); isSet {
		cfg.Metrics.Types.EventTotal = eventTotal
	}
	if objectSize, isSet := parseBool(*metricsTypesObjectSize); isSet {
		cfg.Metrics.Types.ObjectSize = objectSize
	}
	if ipTotal, isSet := parseBool(*metricsTypesIPTotal); isSet {
		cfg.Metrics.Types.IPTotal = ipTotal
	}
	if prefixTotal, isSet := parseBool(*metricsTypesPrefixTotal); isSet {
		cfg.Metrics.Types.PrefixTotal = prefixTotal
	}
	if prefixDepthTotal, isSet := parseBool(*metricsTypesPrefixDepthTotal); isSet {
		cfg.Metrics.Types.PrefixDepthTotal = prefixDepthTotal
	}
	if fileExtTotal, isSet := parseBool(*metricsTypesFileExtensionTotal); isSet {
		cfg.Metrics.Types.FileExtensionTotal = fileExtTotal
	}
	if latency, isSet := parseBool(*metricsTypesLatency); isSet {
		cfg.Metrics.Types.Latency = latency
	}
	if anomaly, isSet := parseBool(*metricsTypesAnomalyDetection); isSet {
		cfg.Metrics.Types.AnomalyDetection = anomaly
	}
	if lifecycle, isSet := parseBool(*metricsTypesLifecycleExpiration); isSet {
		cfg.Metrics.Types.LifecycleExpiration = lifecycle
	}
	if deleteTotal, isSet := parseBool(*metricsTypesDeleteTotal); isSet {
		cfg.Metrics.Types.DeleteTotal = deleteTotal
	}
	if *metricsObjectSizeBuckets != "" {
		if buckets := parseFloat64Slice(*metricsObjectSizeBuckets); len(buckets) > 0 {
			cfg.Metrics.ObjectSizeBuckets = buckets
		}
	}

	// Handle buckets (both single and list)
	var bucketList []string
	if *bucketName != "" {
		bucketList = []string{*bucketName}
	}
	if *bucketNames != "" {
		// Split comma-separated bucket names
		bucketList = strings.Split(*bucketNames, ",")
		for i := range bucketList {
			bucketList[i] = strings.TrimSpace(bucketList[i])
		}
	}
	if len(bucketList) > 0 {
		var buckets []struct {
			Name   string   `mapstructure:"name"`
			Prefix []string `mapstructure:"prefix"`
		}
		for _, bucket := range bucketList {
			if bucket != "" {
				buckets = append(buckets, struct {
					Name   string   `mapstructure:"name"`
					Prefix []string `mapstructure:"prefix"`
				}{
					Name:   bucket,
					Prefix: []string{},
				})
			}
		}
		cfg.SQS.Buckets = buckets
	}

	// Handle legacy processUnlistedBucketsFlag (for backward compatibility)
	if *processUnlistedBucketsFlag != "" {
		if processUnlisted, isSet := parseBool(*processUnlistedBucketsFlag); isSet {
			cfg.SQS.ProcessUnlistedBuckets = processUnlisted
		}
	} // Log the loaded configuration
	logger.Info(logger.LogContext{
		ExtraFields: map[string]interface{}{
			"config_file":      *configFile,
			"buckets":          cfg.SQS.Buckets,
			"process_unlisted": cfg.SQS.ProcessUnlistedBuckets,
		},
	}, "Configuration loaded")

	// Update context with configuration details
	mainCtx.ExtraFields = map[string]interface{}{
		"config_file": *configFile,
		"queue_count": len(cfg.SQS.Queues),
		"port":        *listenAddr,
	}

	// Validate configuration
	if err := config.ValidateConfig(cfg); err != nil {
		logger.Fatal(mainCtx, err, "Invalid configuration")
	}

	// Initialize metrics
	metrics.Initialize(cfg)

	// Create context that listens for cancellation signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create and start the SQS poller
	sqsPoller := poller.NewPoller(cfg)
	if err := sqsPoller.StartPolling(ctx); err != nil {
		logger.Fatal(mainCtx, err, "Failed to start poller")
	}

	// Log successful startup
	logger.Info(mainCtx, "Application started successfully")

	// Set up HTTP server for Prometheus metrics
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Get cardinality health information
		cardinalityHealth := metrics.GetMetrics().GetCardinalityHealth()

		// Determine overall health status
		status := "healthy"
		if cardinalityHealth["status"] == "critical" {
			status = "critical"
			w.WriteHeader(http.StatusInternalServerError)
		} else if cardinalityHealth["status"] == "warning" {
			status = "warning"
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		// Create health response
		healthResponse := map[string]interface{}{
			"status":                 status,
			"timestamp":              time.Now().Format(time.RFC3339),
			"cardinality_monitoring": cardinalityHealth,
		}

		// Convert to JSON
		jsonResponse, err := json.MarshalIndent(healthResponse, "", "  ")
		if err != nil {
			logger.Warn(logger.LogContext{
				Operation: "HealthHandler",
				Component: "Main",
			}, err, "Failed to marshal health response")
			w.WriteHeader(http.StatusInternalServerError)
			if _, writeErr := w.Write([]byte(`{"status":"error","message":"Failed to generate health response"}`)); writeErr != nil {
				logger.Error(logger.LogContext{
					Operation: "HealthHandler",
					Component: "Main",
				}, writeErr, "Failed to write error response")
			}
			return
		}

		if _, err := w.Write(jsonResponse); err != nil {
			logger.Warn(logger.LogContext{
				Operation: "HealthHandler",
				Component: "Main",
			}, err, "Failed to write health response")
		}
	})

	// Start HTTP server in a goroutine
	server := &http.Server{
		Addr:              *listenAddr,
		Handler:           http.DefaultServeMux,
		ReadHeaderTimeout: 5 * time.Second, // Prevent Slowloris attacks

	}

	serverCtx := logger.LogContext{
		Operation: "MetricsServer",
		Component: "Main",
		ExtraFields: map[string]interface{}{
			"address": *listenAddr,
		},
	}

	go func() {
		logger.Info(serverCtx, "Starting metrics server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal(serverCtx, err, "Failed to start metrics server")
		}
	}()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for interrupt signal
	<-sigChan

	shutdownCtx := logger.LogContext{
		Operation: "Shutdown",
		Component: "Main",
	}

	logger.Info(shutdownCtx, "Received shutdown signal")

	// Initiate graceful shutdown
	cancel() // Cancel the context to stop the poller
	sqsPoller.Shutdown()

	// Shutdown the HTTP server
	if err := server.Shutdown(context.Background()); err != nil {
		logger.Error(shutdownCtx, err, "Error shutting down metrics server")
	}

	logger.Info(shutdownCtx, "Shutdown complete")
}
