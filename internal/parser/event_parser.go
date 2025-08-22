package parser

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"s3_metrics_adapter/internal/logger"
	"strings"
	"time"
)

type ParsedEvent struct {
	EventType   string    // The type of event (e.g., "Object Created", "Object Deleted")
	BucketName  string    // The name of the S3 bucket
	ObjectKey   string    // The key/path of the object in the bucket
	Size        int64     // Size of the object in bytes
	FileType    string    // Extracted file extension
	Time        time.Time // When the event occurred
	RequesterID string    // Who requested the operation
	SourceIP    string    // Source IP of the request
	Reason      string    // Reason for the operation (e.g., "PutObject")
	RequestID   string    // AWS request ID for tracking
}

type EventParser interface {
	Parse(message string) (*ParsedEvent, error)
}

type S3EventParser struct{}

type SQSMessage struct {
	MessageId  string `json:"MessageId"`
	Body       string `json:"Body"`
	Attributes struct {
		SenderId                         string `json:"SenderId"`
		ApproximateReceiveCount          string `json:"ApproximateReceiveCount"`
		SentTimestamp                    string `json:"SentTimestamp"`
		ApproximateFirstReceiveTimestamp string `json:"ApproximateFirstReceiveTimestamp"`
	} `json:"Attributes"`
}

type S3Object struct {
	Key       string `json:"key"`
	ETag      string `json:"etag"`
	VersionId string `json:"version-id"`
	Sequencer string `json:"sequencer"`
	Size      int64  `json:"size,omitempty"` // Size might be present in some events
}

type S3Event struct {
	Version    string    `json:"version"`
	ID         string    `json:"id"`
	DetailType string    `json:"detail-type"`
	Source     string    `json:"source"`
	Account    string    `json:"account"`
	Time       time.Time `json:"time"`
	Region     string    `json:"region"`
	Resources  []string  `json:"resources"`
	Detail     struct {
		Version string `json:"version"`
		Bucket  struct {
			Name string `json:"name"`
		} `json:"bucket"`
		Object          S3Object `json:"object"`
		RequestID       string   `json:"request-id"`
		Requester       string   `json:"requester"`
		SourceIPAddress string   `json:"source-ip-address"`
		Reason          string   `json:"reason"`
		DeletionType    string   `json:"deletion-type,omitempty"`
	} `json:"detail"`
}

type LegacyS3Event struct {
	Records []struct {
		EventVersion string    `json:"eventVersion"`
		EventTime    time.Time `json:"eventTime"`
		EventName    string    `json:"eventName"`
		UserIdentity struct {
			PrincipalId string `json:"principalId"`
		} `json:"userIdentity"`
		RequestParameters struct {
			SourceIPAddress string `json:"sourceIPAddress"`
		} `json:"requestParameters"`
		ResponseElements struct {
			RequestID string `json:"x-amz-request-id"`
		} `json:"responseElements"`
		S3 struct {
			Bucket struct {
				Name string `json:"name"`
			} `json:"bucket"`
			Object struct {
				Key  string `json:"key"`
				Size int64  `json:"size"`
				ETag string `json:"eTag"`
			} `json:"object"`
		} `json:"s3"`
	} `json:"Records"`
}

func (p *S3EventParser) Parse(message string) (*ParsedEvent, error) {
	// First try to parse as SQS message to extract the actual event
	var sqsMessage SQSMessage
	if err := json.Unmarshal([]byte(message), &sqsMessage); err == nil && sqsMessage.Body != "" {
		logger.Debug(logger.LogContext{
			Component: "event_parser",
		}, fmt.Sprintf("Processing SQS message %s", sqsMessage.MessageId))
		message = sqsMessage.Body
	}

	// Try parsing as EventBridge format first
	parsedEvent, err := parseEventBridgeFormat(message)
	if err == nil {
		logger.Debug(logger.LogContext{
			Component: "event_parser",
			TraceID:   parsedEvent.RequestID,
		}, fmt.Sprintf("Successfully parsed EventBridge format event: %s on %s/%s",
			parsedEvent.EventType, parsedEvent.BucketName, parsedEvent.ObjectKey))
		return parsedEvent, nil
	}

	// Try parsing as legacy S3 notification format
	parsedEvent, err = parseLegacyFormat(message)
	if err == nil {
		logger.Debug(logger.LogContext{
			Component: "event_parser",
			TraceID:   parsedEvent.RequestID,
		}, fmt.Sprintf("Successfully parsed legacy format event: %s on %s/%s",
			parsedEvent.EventType, parsedEvent.BucketName, parsedEvent.ObjectKey))
		return parsedEvent, nil
	}

	return nil, fmt.Errorf("failed to parse event: unrecognized format")
}

func parseEventBridgeFormat(message string) (*ParsedEvent, error) {
	var event S3Event
	if err := json.Unmarshal([]byte(message), &event); err != nil {
		return nil, fmt.Errorf("failed to unmarshal EventBridge format: %v", err)
	}
	if event.Source != "aws.s3" {
		return nil, fmt.Errorf("not an EventBridge S3 event")
	}

	parsedEvent := &ParsedEvent{
		EventType:   event.DetailType,
		BucketName:  event.Detail.Bucket.Name,
		ObjectKey:   event.Detail.Object.Key,
		Size:        event.Detail.Object.Size,
		FileType:    getFileExtension(event.Detail.Object.Key),
		Time:        event.Time,
		RequesterID: event.Detail.Requester,
		SourceIP:    event.Detail.SourceIPAddress,
		Reason:      event.Detail.Reason,
		RequestID:   event.Detail.RequestID,
	}

	// Validate required fields
	if err := parsedEvent.validate(); err != nil {
		return nil, fmt.Errorf("invalid EventBridge event: %v", err)
	}

	return parsedEvent, nil
}

func parseLegacyFormat(message string) (*ParsedEvent, error) {
	var legacyEvent LegacyS3Event
	if err := json.Unmarshal([]byte(message), &legacyEvent); err != nil {
		return nil, fmt.Errorf("failed to unmarshal legacy format: %v", err)
	}

	if len(legacyEvent.Records) == 0 {
		return nil, fmt.Errorf("no records found in legacy event")
	}

	record := legacyEvent.Records[0]

	// Create event
	event := &ParsedEvent{
		EventType:   record.EventName,
		BucketName:  record.S3.Bucket.Name,
		ObjectKey:   record.S3.Object.Key,
		Size:        record.S3.Object.Size,
		FileType:    getFileExtension(record.S3.Object.Key),
		Time:        record.EventTime,
		RequesterID: record.UserIdentity.PrincipalId,
		SourceIP:    record.RequestParameters.SourceIPAddress,
		Reason:      record.EventName,
		RequestID:   record.ResponseElements.RequestID,
	}

	// Validate required fields
	if err := event.validate(); err != nil {
		return nil, fmt.Errorf("invalid legacy event: %v", err)
	}

	return event, nil
}

// validate checks if the required fields are present and valid
func (p *ParsedEvent) validate() error {
	if p.EventType == "" {
		return fmt.Errorf("event type is required")
	}
	if p.BucketName == "" {
		return fmt.Errorf("bucket name is required")
	}
	if p.ObjectKey == "" {
		return fmt.Errorf("object key is required")
	}
	if p.Time.IsZero() {
		return fmt.Errorf("event time is required")
	}
	if p.RequestID == "" {
		return fmt.Errorf("request ID is required")
	}
	return nil
}

func getFileExtension(objectKey string) string {
	ext := filepath.Ext(objectKey)
	if ext == "" {
		return "no_extension"
	}
	// Remove the leading dot and convert to lowercase
	return strings.ToLower(ext[1:])
}

func (p *ParsedEvent) LogEventDetails() {
	// Log minimal info at INFO level
	logger.Info(logger.LogContext{
		Component: "event_parser",
		TraceID:   p.RequestID,
	}, fmt.Sprintf("S3 event processed: %s on %s/%s", p.EventType, p.BucketName, p.ObjectKey))

	// Log detailed info at DEBUG level
	logger.Debug(logger.LogContext{
		Component: "event_parser",
		TraceID:   p.RequestID,
	}, fmt.Sprintf("S3 event details: size=%d bytes, type=%s, requester=%s, source_ip=%s, time=%s",
		p.Size, p.FileType, p.RequesterID, p.SourceIP, p.Time.Format(time.RFC3339)))
}
