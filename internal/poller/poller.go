package poller

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"s3_metrics_adapter/internal/config"
	"s3_metrics_adapter/internal/logger"
	"s3_metrics_adapter/internal/metrics"
	"s3_metrics_adapter/internal/parser"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsCfg "github.com/aws/aws-sdk-go-v2/config"
	awsSQS "github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/sirupsen/logrus"
)

// SQSClientInterface defines the interface for SQS operations
type SQSClientInterface interface {
	ReceiveMessage(ctx context.Context, params *awsSQS.ReceiveMessageInput, optFns ...func(*awsSQS.Options)) (*awsSQS.ReceiveMessageOutput, error)
	DeleteMessage(ctx context.Context, params *awsSQS.DeleteMessageInput, optFns ...func(*awsSQS.Options)) (*awsSQS.DeleteMessageOutput, error)
}

// Constants for SQS polling
const (
	maxMessages     = 10
	waitTimeSeconds = 2
	maxRetries      = 3
	backoffBase     = 2 * time.Second
	// Circuit breaker constants
	circuitBreakerFailureThreshold = 5
	circuitBreakerTimeout          = 30 * time.Second
	// Batch processing constants
	batchSize    = 10
	batchTimeout = 100 * time.Millisecond
	// Cache constants - TODO: Implement in next release
)

// ErrInvalidQueueURL is returned when the queue URL is invalid
var ErrInvalidQueueURL = errors.New("invalid queue URL format")

// CircuitBreakerState represents the state of the circuit breaker
type CircuitBreakerState int

const (
	CircuitBreakerClosed CircuitBreakerState = iota
	CircuitBreakerOpen
	CircuitBreakerHalfOpen
)

// CircuitBreaker implements the circuit breaker pattern for parsing failures
type CircuitBreaker struct {
	state        CircuitBreakerState
	failureCount int32
	lastFailTime time.Time
	mutex        sync.RWMutex
}

// PathLabelCacheEntry - TODO: Implement in next release

// BatchProcessor handles batch processing of messages
type BatchProcessor struct {
	batch     []types.Message
	mutex     sync.Mutex
	timer     *time.Timer
	processor func([]types.Message) error
}

// PerformanceMetrics tracks processing performance
type PerformanceMetrics struct {
	MessagesProcessed int64
	MessagesPerSecond float64
	ParseTimeTotal    int64
	ParseTimeCount    int64
	LastUpdateTime    time.Time
	mutex             sync.RWMutex
}

type SQSPoller struct {
	queues        []string
	eventParser   parser.EventParser
	done          chan struct{}
	wg            sync.WaitGroup
	retryBackoff  time.Duration
	config        *config.Config
	clientFactory func(region string) (SQSClientInterface, error)
	// New optimization components
	circuitBreaker *CircuitBreaker
	// Path label cache - TODO: Implement in next release
	batchProcessor     *BatchProcessor
	performanceMetrics *PerformanceMetrics
}

// NewPoller creates a new SQS poller instance
func NewPoller(cfg *config.Config) *SQSPoller {
	return &SQSPoller{
		queues:        cfg.SQS.Queues,
		eventParser:   &parser.S3EventParser{},
		done:          make(chan struct{}),
		retryBackoff:  backoffBase,
		config:        cfg,
		clientFactory: defaultClientFactory,
		// Initialize optimization components
		circuitBreaker: &CircuitBreaker{
			state:        CircuitBreakerClosed,
			failureCount: 0,
		},
		// Path label cache - TODO: Implement in next release
		performanceMetrics: &PerformanceMetrics{
			LastUpdateTime: time.Now(),
		},
	}
}

// Circuit breaker methods
func (cb *CircuitBreaker) CanExecute() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	if cb.state == CircuitBreakerClosed {
		return true
	}

	if cb.state == CircuitBreakerOpen {
		return time.Since(cb.lastFailTime) > circuitBreakerTimeout
	}

	return true // Half-open state
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failureCount = 0
	cb.state = CircuitBreakerClosed
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failureCount++
	cb.lastFailTime = time.Now()

	if cb.failureCount >= circuitBreakerFailureThreshold {
		cb.state = CircuitBreakerOpen
	}
}

// Path label cache methods - TODO: Implement in next release

// Performance metrics methods
func (pm *PerformanceMetrics) RecordMessageProcessed() {
	atomic.AddInt64(&pm.MessagesProcessed, 1)
}

func (pm *PerformanceMetrics) RecordParseTime(duration time.Duration) {
	atomic.AddInt64(&pm.ParseTimeTotal, int64(duration.Nanoseconds()))
	atomic.AddInt64(&pm.ParseTimeCount, 1)
}

func (pm *PerformanceMetrics) GetMessagesPerSecond() float64 {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	now := time.Now()
	elapsed := now.Sub(pm.LastUpdateTime).Seconds()
	if elapsed == 0 {
		return 0
	}

	messages := atomic.LoadInt64(&pm.MessagesProcessed)
	pm.LastUpdateTime = now

	return float64(messages) / elapsed
}

func (pm *PerformanceMetrics) GetAverageParseTime() time.Duration {
	parseTimeTotal := atomic.LoadInt64(&pm.ParseTimeTotal)
	parseTimeCount := atomic.LoadInt64(&pm.ParseTimeCount)

	if parseTimeCount == 0 {
		return 0
	}

	return time.Duration(parseTimeTotal / parseTimeCount)
}

// Message filtering - check if message should be processed before parsing
func (p *SQSPoller) shouldProcessMessage(msgBody string) bool {
	// Temporarily disable filtering to debug metrics issue
	// TODO: Re-enable with proper S3 event detection
	return true

	// Quick JSON parsing to extract bucket and key without full parsing
	// This is a simple heuristic - in production you might want more sophisticated filtering
	if !strings.Contains(msgBody, "Records") {
		return false
	}

	// Check for S3 event patterns
	if !strings.Contains(msgBody, "s3") && !strings.Contains(msgBody, "S3") {
		return false
	}

	return true
}

// extractPathLabels - TODO: Implement in next release
// This function will extract structured labels from S3 object paths

// updatePerformanceMetrics updates performance metrics for the poller
func (p *SQSPoller) updatePerformanceMetrics(queueURL string, batchSize int) {
	m := metrics.GetMetrics()
	if m == nil {
		return
	}

	// Get performance metrics
	messagesPerSecond := p.performanceMetrics.GetMessagesPerSecond()
	avgParseTime := p.performanceMetrics.GetAverageParseTime()

	// Update metrics
	m.UpdatePerformanceMetrics(queueURL, messagesPerSecond, avgParseTime, batchSize, "success")
}

// Batch processing methods
func (bp *BatchProcessor) AddMessage(msg types.Message) {
	bp.mutex.Lock()
	defer bp.mutex.Unlock()

	bp.batch = append(bp.batch, msg)

	// Start timer if this is the first message in the batch
	if len(bp.batch) == 1 {
		bp.timer = time.AfterFunc(batchTimeout, func() {
			bp.processBatch()
		})
	}

	// Process immediately if batch is full
	if len(bp.batch) >= batchSize {
		bp.timer.Stop()
		bp.processBatch()
	}
}

func (bp *BatchProcessor) processBatch() {
	bp.mutex.Lock()
	defer bp.mutex.Unlock()

	if len(bp.batch) == 0 {
		return
	}

	// Create a copy of the batch and reset
	batch := make([]types.Message, len(bp.batch))
	copy(batch, bp.batch)
	bp.batch = bp.batch[:0]

	// Process the batch
	if bp.processor != nil {
		bp.processor(batch)
	}
}

// NewPollerWithClientFactory creates a new SQS poller instance with a custom client factory (for testing)
func NewPollerWithClientFactory(cfg *config.Config, clientFactory func(region string) (SQSClientInterface, error)) *SQSPoller {
	return &SQSPoller{
		queues:        cfg.SQS.Queues,
		eventParser:   &parser.S3EventParser{},
		done:          make(chan struct{}),
		retryBackoff:  backoffBase,
		config:        cfg,
		clientFactory: clientFactory,
		// Initialize optimization components
		circuitBreaker: &CircuitBreaker{
			state:        CircuitBreakerClosed,
			failureCount: 0,
		},
		// Path label cache - TODO: Implement in next release
		performanceMetrics: &PerformanceMetrics{
			LastUpdateTime: time.Now(),
		},
	}
}

// defaultClientFactory creates a real AWS SQS client
func defaultClientFactory(region string) (SQSClientInterface, error) {
	cfg, err := awsCfg.LoadDefaultConfig(context.Background(), awsCfg.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	return awsSQS.NewFromConfig(cfg), nil
}

// extractRegionFromQueueURL extracts the AWS region from an SQS queue URL
func (p *SQSPoller) extractRegionFromQueueURL(queueURL string) (string, error) {
	// URL format: https://sqs.{region}.amazonaws.com/{account}/{queue}
	parts := strings.Split(queueURL, ".")
	if len(parts) < 4 {
		return "", fmt.Errorf("%w: %s", ErrInvalidQueueURL, queueURL)
	}
	return parts[1], nil
}

// StartPolling starts polling all configured SQS queues
func (p *SQSPoller) StartPolling(ctx context.Context) error {
	// Log startup at INFO level
	logger.Info(logger.LogContext{
		Component: "sqspoller",
	}, fmt.Sprintf("Starting SQS poller with %d queues", len(p.queues)))

	// Log detailed configuration at DEBUG level
	logger.Debug(logger.LogContext{
		Component: "sqspoller",
	}, fmt.Sprintf("Queue URLs: %s", strings.Join(p.queues, ", ")))

	// Start polling each queue in a separate goroutine
	for _, queueURL := range p.queues {
		p.wg.Add(1)
		go p.pollQueue(ctx, queueURL)
	}

	return nil
}

// Shutdown gracefully shuts down the poller
func (p *SQSPoller) Shutdown() {
	logger.Info(logger.LogContext{
		Component: "sqspoller",
	}, "Initiating graceful shutdown of SQS poller")
	close(p.done)
	p.wg.Wait()
	logger.Info(logger.LogContext{
		Component: "sqspoller",
	}, "SQS poller shutdown complete")
}

// pollQueue continuously polls a single SQS queue
func (p *SQSPoller) pollQueue(ctx context.Context, queueURL string) {
	defer p.wg.Done()

	region, err := p.extractRegionFromQueueURL(queueURL)
	if err != nil {
		logger.Error(logger.LogContext{
			Component: "sqspoller",
		}, fmt.Errorf("invalid queue URL %s: %w", queueURL, err))
		return
	}

	client, err := p.clientFactory(region)
	if err != nil {
		logger.Error(logger.LogContext{
			Component: "sqspoller",
		}, fmt.Errorf("failed to create SQS client for region %s: %w", region, err))
		return
	}

	logger.Info(logger.LogContext{
		Component: "sqspoller",
	}, fmt.Sprintf("Started polling queue %s in region %s", queueURL, region))

	retryCount := 0
	for {
		select {
		case <-ctx.Done():
			logger.GetLogger().WithFields(logrus.Fields{
				"queue_url": queueURL,
			}).Info("Context cancelled for queue")
			return
		case <-p.done:
			logger.GetLogger().WithFields(logrus.Fields{
				"queue_url": queueURL,
			}).Info("Stopping poll for queue")
			return
		default:
			if err := p.receiveAndProcessMessages(ctx, client, queueURL); err != nil {
				retryCount++
				backoff := p.calculateBackoff(retryCount)
				logger.GetLogger().WithFields(logrus.Fields{
					"queue_url":   queueURL,
					"retry_count": retryCount,
					"error":       err.Error(),
					"retry_after": backoff.String(),
				}).Error("Error processing queue")

				select {
				case <-time.After(backoff):
					continue
				case <-ctx.Done():
					return
				case <-p.done:
					return
				}
			} else {
				retryCount = 0 // Reset retry count on successful processing
			}
		}
	}
}

// receiveAndProcessMessages receives and processes messages from SQS
func (p *SQSPoller) receiveAndProcessMessages(ctx context.Context, client SQSClientInterface, queueURL string) error {
	output, err := client.ReceiveMessage(ctx, &awsSQS.ReceiveMessageInput{
		QueueUrl:            aws.String(queueURL),
		MaxNumberOfMessages: int32(maxMessages),
		WaitTimeSeconds:     int32(waitTimeSeconds),
	})

	if err != nil {
		return fmt.Errorf("failed to receive messages: %w", err)
	}

	// Filter messages before processing
	var validMessages []types.Message
	for _, msg := range output.Messages {
		if msg.Body != nil {
			logger.Info(logger.LogContext{
				Component: "sqspoller",
				TraceID:   *msg.MessageId,
			}, fmt.Sprintf("Received message (Queue: %s): %s", queueURL, *msg.Body))

			if p.shouldProcessMessage(*msg.Body) {
				validMessages = append(validMessages, msg)
				logger.Info(logger.LogContext{
					Component: "sqspoller",
					TraceID:   *msg.MessageId,
				}, "Message passed filtering - will be processed")
			} else {
				// Delete filtered messages immediately
				logger.Info(logger.LogContext{
					Component: "sqspoller",
					TraceID:   *msg.MessageId,
				}, fmt.Sprintf("Filtering out message (Queue: %s): %s", queueURL, *msg.Body))

				if _, delErr := client.DeleteMessage(ctx, &awsSQS.DeleteMessageInput{
					QueueUrl:      aws.String(queueURL),
					ReceiptHandle: msg.ReceiptHandle,
				}); delErr != nil {
					logger.Error(logger.LogContext{
						Component: "sqspoller",
						TraceID:   *msg.MessageId,
					}, fmt.Errorf("failed to delete filtered message from %s: %w", queueURL, delErr))
				}
			}
		}
	}

	// Process valid messages in batch
	if len(validMessages) > 0 {
		return p.processBatchMessages(ctx, client, queueURL, validMessages)
	}

	return nil
}

// processBatchMessages processes a batch of messages with optimizations
func (p *SQSPoller) processBatchMessages(ctx context.Context, client SQSClientInterface, queueURL string, messages []types.Message) error {
	// Check circuit breaker
	if !p.circuitBreaker.CanExecute() {
		logger.Warn(logger.LogContext{
			Component: "sqspoller",
		}, "Circuit breaker is open, skipping message processing")
		return nil
	}

	// Process messages in parallel (with controlled concurrency)
	semaphore := make(chan struct{}, 5) // Limit concurrent processing
	var wg sync.WaitGroup
	var errors []error
	var errorMutex sync.Mutex

	for _, msg := range messages {
		wg.Add(1)
		go func(message types.Message) {
			defer wg.Done()

			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			if err := p.processMessage(ctx, client, queueURL, message); err != nil {
				errorMutex.Lock()
				errors = append(errors, err)
				errorMutex.Unlock()

				// Record circuit breaker failure
				p.circuitBreaker.RecordFailure()

				// Delete failed message
				if _, delErr := client.DeleteMessage(ctx, &awsSQS.DeleteMessageInput{
					QueueUrl:      aws.String(queueURL),
					ReceiptHandle: message.ReceiptHandle,
				}); delErr != nil {
					logger.Error(logger.LogContext{
						Component: "sqspoller",
						TraceID:   *message.MessageId,
					}, fmt.Errorf("failed to delete failed message from %s: %w", queueURL, delErr))
				}
			} else {
				// Record circuit breaker success
				p.circuitBreaker.RecordSuccess()
			}
		}(msg)
	}

	wg.Wait()

	// Log any errors but don't fail the entire batch
	if len(errors) > 0 {
		logger.Error(logger.LogContext{
			Component: "sqspoller",
		}, fmt.Errorf("processed batch with %d errors out of %d messages", len(errors), len(messages)))
	}

	// Update performance metrics
	p.updatePerformanceMetrics(queueURL, len(messages))

	return nil
}

// processMessage processes a single SQS message
func (p *SQSPoller) processMessage(ctx context.Context, client SQSClientInterface, queueURL string, msg types.Message) error {
	if msg.Body == nil {
		return errors.New("received message with nil body")
	}

	// Record performance metrics
	startTime := time.Now()
	defer func() {
		p.performanceMetrics.RecordMessageProcessed()
		p.performanceMetrics.RecordParseTime(time.Since(startTime))
	}()

	// Parse the message
	parsedEvent, err := p.eventParser.Parse(*msg.Body)
	if err != nil {
		metrics.GetMetrics().IncreaseParserErrors()
		return fmt.Errorf("failed to parse message: %w", err)
	}

	// Check if bucket and prefix are allowed using the full object key
	logger.Debug(logger.LogContext{
		Component: "sqspoller",
		TraceID:   parsedEvent.RequestID,
	}, fmt.Sprintf("Checking allowance for bucket=%s key=%s",
		parsedEvent.BucketName, parsedEvent.ObjectKey))

	if !p.config.IsAllowedBucketAndPrefix(parsedEvent.BucketName, parsedEvent.ObjectKey) {
		logger.Debug(logger.LogContext{
			Component: "sqspoller",
			TraceID:   parsedEvent.RequestID,
		}, fmt.Sprintf("Skipping event - bucket/prefix not allowed: %s/%s",
			parsedEvent.BucketName, parsedEvent.ObjectKey))

		// Delete the filtered message from the queue
		_, err = client.DeleteMessage(ctx, &awsSQS.DeleteMessageInput{
			QueueUrl:      aws.String(queueURL),
			ReceiptHandle: msg.ReceiptHandle,
		})
		if err != nil {
			return fmt.Errorf("failed to delete filtered message: %w", err)
		}
		return nil
	}

	logger.Debug(logger.LogContext{
		Component: "sqspoller",
		TraceID:   parsedEvent.RequestID,
	}, fmt.Sprintf("Processing allowed event: %s/%s", parsedEvent.BucketName, parsedEvent.ObjectKey))

	// Path labeling - TODO: Implement in next release

	// Update metrics
	m := metrics.GetMetrics()
	if m != nil {
		m.UpdateMetrics(parsedEvent)
	}

	// Log event details
	parsedEvent.LogEventDetails()

	// Delete the message from the queue
	_, err = client.DeleteMessage(ctx, &awsSQS.DeleteMessageInput{
		QueueUrl:      aws.String(queueURL),
		ReceiptHandle: msg.ReceiptHandle,
	})

	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	return nil
}

// calculateBackoff calculates exponential backoff with jitter
func (p *SQSPoller) calculateBackoff(retryCount int) time.Duration {
	maxShift := 20
	shift := retryCount
	if shift > maxShift {
		shift = maxShift
	}
	backoff := p.retryBackoff * time.Duration(1<<uint(shift)) // #nosec G115
	if backoff > 1*time.Minute {
		backoff = 1 * time.Minute // Cap at 1 minute
	}
	// Add jitter (±20%)
	jitter := time.Duration(float64(backoff) * (0.8 + 0.4*float64(time.Now().UnixNano()%100)/100.0))
	return jitter
}

// getPrefix extracts the top-level prefix from an object key
func getPrefix(key string) string {
	if key == "" {
		return ""
	}
	parts := strings.Split(key, "/")
	if len(parts) > 1 {
		return parts[0]
	}
	return ""
}
