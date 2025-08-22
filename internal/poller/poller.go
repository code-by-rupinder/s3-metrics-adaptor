package poller

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
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
	waitTimeSeconds = 20
	maxRetries      = 3
	backoffBase     = 2 * time.Second
)

// ErrInvalidQueueURL is returned when the queue URL is invalid
var ErrInvalidQueueURL = errors.New("invalid queue URL format")

type SQSPoller struct {
	queues        []string
	eventParser   parser.EventParser
	done          chan struct{}
	wg            sync.WaitGroup
	retryBackoff  time.Duration
	config        *config.Config
	clientFactory func(region string) (SQSClientInterface, error)
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

	for _, msg := range output.Messages {
		if err := p.processMessage(ctx, client, queueURL, msg); err != nil {
			// Log the message body for debugging
			logger.Debug(logger.LogContext{
				Component: "sqspoller",
				TraceID:   *msg.MessageId,
			}, fmt.Sprintf("Failed message content (Queue: %s): %s", queueURL, *msg.Body))

			logger.Error(logger.LogContext{
				Component: "sqspoller",
				TraceID:   *msg.MessageId,
			}, fmt.Errorf("failed to process message from %s: %w", queueURL, err))

			// Delete failed messages to avoid poison pill
			if _, delErr := client.DeleteMessage(ctx, &awsSQS.DeleteMessageInput{
				QueueUrl:      aws.String(queueURL),
				ReceiptHandle: msg.ReceiptHandle,
			}); delErr != nil {
				logger.Error(logger.LogContext{
					Component: "sqspoller",
					TraceID:   *msg.MessageId,
				}, fmt.Errorf("failed to delete failed message from %s: %w", queueURL, delErr))
			}
			continue // Continue processing other messages
		}
	}

	return nil
}

// processMessage processes a single SQS message
func (p *SQSPoller) processMessage(ctx context.Context, client SQSClientInterface, queueURL string, msg types.Message) error {
	if msg.Body == nil {
		return errors.New("received message with nil body")
	}

	// Parse the message
	parsedEvent, err := p.eventParser.Parse(*msg.Body)
	if err != nil {
		metrics.GetMetrics().IncreaseParserErrors()
		return fmt.Errorf("failed to parse message: %w", err)
	}

	// Check if bucket and prefix are allowed
	prefix := getPrefix(parsedEvent.ObjectKey)
	logger.Debug(logger.LogContext{
		Component: "sqspoller",
		TraceID:   parsedEvent.RequestID,
	}, fmt.Sprintf("Checking allowance for bucket=%s prefix=%s key=%s",
		parsedEvent.BucketName, prefix, parsedEvent.ObjectKey))

	if !p.config.IsAllowedBucketAndPrefix(parsedEvent.BucketName, prefix) {
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

	// Update metrics
	metrics.GetMetrics().UpdateMetrics(parsedEvent)

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
