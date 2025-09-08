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

	// Convert S3 event names to our format
	// For EventBridge, we need to construct the full event name from detail-type and reason
	eventType := convertEventBridgeEventName(event.DetailType, event.Detail.Reason)

	parsedEvent := &ParsedEvent{
		EventType:   eventType,
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

	// Convert S3 event names to our format
	eventType := convertS3EventName(record.EventName)

	// Create event
	event := &ParsedEvent{
		EventType:   eventType,
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

// convertS3EventName converts S3 event names to our internal format
// Examples:
//
//	"s3:ObjectCreated:Put" -> "Object Created.Put"
//	"s3:ObjectCreated:Post" -> "Object Created.Post"
//	"s3:ObjectCreated:Copy" -> "Object Created.Copy"
//	"s3:ObjectCreated:CompleteMultipartUpload" -> "Object Created.CompleteMultipartUpload"
//	"s3:ObjectRemoved:Delete" -> "Object Deleted.Delete"
//	"s3:ObjectRemoved:DeleteMarkerCreated" -> "Object Deleted.DeleteMarkerCreated"
func convertS3EventName(s3EventName string) string {
	// Handle S3 event names like "s3:ObjectCreated:Put"
	if strings.HasPrefix(s3EventName, "s3:") {
		parts := strings.Split(s3EventName, ":")
		if len(parts) >= 3 {
			eventCategory := parts[1] // "ObjectCreated" or "ObjectRemoved"
			eventSubtype := parts[2]  // "Put", "Post", "Copy", etc.

			// Convert to our format
			switch eventCategory {
			case "ObjectCreated":
				return fmt.Sprintf("Object Created.%s", eventSubtype)
			case "ObjectRemoved":
				return fmt.Sprintf("Object Deleted.%s", eventSubtype)
			default:
				// For other event types, just return as-is
				return s3EventName
			}
		}
	}

	// If it doesn't match the expected format, return as-is
	return s3EventName
}

// convertEventBridgeEventName converts EventBridge S3 events to our internal format
// EventBridge sends simplified detail-type like "Object Created" and puts the specific operation in reason
// Examples:
//
//	detailType: "Object Created", reason: "PutObject" -> "Object Created.Put"
//	detailType: "Object Created", reason: "PostObject" -> "Object Created.Post"
//	detailType: "Object Created", reason: "CopyObject" -> "Object Created.Copy"
//	detailType: "Object Created", reason: "CompleteMultipartUpload" -> "Object Created.CompleteMultipartUpload"
//	detailType: "Object Deleted", reason: "DeleteObject" -> "Object Deleted.Delete"
//	detailType: "Object Deleted", reason: "DeleteMarkerCreated" -> "Object Deleted.DeleteMarkerCreated"
func convertEventBridgeEventName(detailType, reason string) string {
	// Map reason to subtype
	subtype := ""
	switch reason {
	case "PutObject":
		subtype = "Put"
	case "PostObject":
		subtype = "Post"
	case "CopyObject":
		subtype = "Copy"
	case "CompleteMultipartUpload":
		subtype = "CompleteMultipartUpload"
	case "DeleteObject":
		subtype = "Delete"
	case "DeleteMarkerCreated":
		subtype = "DeleteMarkerCreated"
	default:
		// For unknown reasons, try to use the reason as-is
		subtype = reason
	}

	// Combine detail-type with subtype
	if subtype != "" {
		return fmt.Sprintf("%s.%s", detailType, subtype)
	}

	// If no subtype could be determined, return just the detail-type
	return detailType
}
