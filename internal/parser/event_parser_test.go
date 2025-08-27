package parser

import (
	"testing"
)

const (
	// Test data
	expectedParseErrorMsg = "Expected successful parse, got error: %v"
	testBucket            = "test-bucket"
	testObject            = "test.txt"
	testUser              = "user123"
	testSourceIP          = "1.2.3.4"
	testRequestID         = "req123"

	// Event types and reasons
	eventTypeObjectDeleted    = "Object Deleted"
	eventTypeObjectCreated    = "Object Created"
	eventTypeObjectCreatedPut = "ObjectCreated:Put"
	reasonLifecycleExpiration = "Lifecycle Expiration"
	reasonDeleteObject        = "DeleteObject"

	// Extensions and types
	extensionTxt = "txt"

	// Error messages
	errBucketName = "Expected bucket name '%s', got %s"
	errObjectKey  = "Expected object key '%s', got %s"
	errEventType  = "Expected event type '%s', got %s"
	errFileType   = "Expected file type '%s', got %s"
	errReason     = "Expected reason '%s', got %s"

	// Test case names
	testCheckEventType  = "check event type"
	testCheckBucketName = "check bucket name"
	testCheckObjectKey  = "check object key"
	testCheckFileType   = "check file extension"
	testCheckReason     = "check reason"
)

// TestSQSMessage tests parsing of an SQS message containing an S3 event
func TestSQSMessage(t *testing.T) {
	sqsMessage := `{
		"MessageId": "12345",
		"Body": "{\"version\":\"1\",\"id\":\"event-id\",\"detail-type\":\"Object Created\",\"source\":\"aws.s3\",\"account\":\"123456789012\",\"time\":\"2023-01-01T00:00:00Z\",\"region\":\"us-west-2\",\"resources\":[],\"detail\":{\"version\":\"1\",\"bucket\":{\"name\":\"test-bucket\"},\"object\":{\"key\":\"test.txt\",\"size\":100,\"etag\":\"abc123\",\"version-id\":\"1\"},\"request-id\":\"req123\",\"requester\":\"user123\",\"source-ip-address\":\"1.2.3.4\",\"reason\":\"PutObject\"}}"
	}`

	parser := S3EventParser{}
	event, err := parser.Parse(sqsMessage)
	if err != nil {
		t.Errorf(expectedParseErrorMsg, err)
	}

	if event.BucketName != testBucket {
		t.Errorf("Expected bucket name '%s', got %s", testBucket, event.BucketName)
	}
}

// TestDirectS3Event tests parsing of a direct S3 notification event
func TestDirectS3Event(t *testing.T) {
	s3Event := `{
		"Records": [{
			"eventVersion": "1",
			"eventTime": "2023-01-01T00:00:00Z",
			"eventName": "ObjectCreated:Put",
			"userIdentity": {
				"principalId": "user123"
			},
			"requestParameters": {
				"sourceIPAddress": "1.2.3.4"
			},
			"responseElements": {
				"x-amz-request-id": "req123"
			},
			"s3": {
				"bucket": {
					"name": "test-bucket"
				},
				"object": {
					"key": "test.txt",
					"size": 100,
					"eTag": "abc123"
				}
			}
		}]
	}`

	parser := S3EventParser{}
	event, err := parser.Parse(s3Event)
	if err != nil {
		t.Errorf(expectedParseErrorMsg, err)
	}

	t.Run(testCheckEventType, func(t *testing.T) {
		if event.EventType != eventTypeObjectCreatedPut {
			t.Errorf(errEventType, eventTypeObjectCreatedPut, event.EventType)
		}
	})

	t.Run(testCheckBucketName, func(t *testing.T) {
		if event.BucketName != testBucket {
			t.Errorf(errBucketName, testBucket, event.BucketName)
		}
	})

	t.Run(testCheckObjectKey, func(t *testing.T) {
		if event.ObjectKey != testObject {
			t.Errorf(errObjectKey, testObject, event.ObjectKey)
		}
	})

	t.Run(testCheckFileType, func(t *testing.T) {
		if event.FileType != extensionTxt {
			t.Errorf(errFileType, extensionTxt, event.FileType)
		}
	})
}

// TestDeleteEvent tests parsing of a deletion event
func TestDeleteEvent(t *testing.T) {
	deleteEvent := `{
		"version": "1",
		"id": "event-id",
		"detail-type": "Object Deleted",
		"source": "aws.s3",
		"account": "123456789012",
		"time": "2023-01-01T00:00:00Z",
		"region": "us-west-2",
		"resources": [],
		"detail": {
			"version": "1",
			"bucket": {
				"name": "test-bucket"
			},
			"object": {
				"key": "test.txt",
				"version-id": "v1"
			},
			"request-id": "req123",
			"requester": "user123",
			"source-ip-address": "1.2.3.4",
			"reason": "DeleteObject"
		}
	}`
	parser := S3EventParser{}
	event, err := parser.Parse(deleteEvent)
	if err != nil {
		t.Errorf(expectedParseErrorMsg, err)
	}

	t.Run(testCheckEventType, func(t *testing.T) {
		if event.EventType != eventTypeObjectDeleted {
			t.Errorf(errEventType, eventTypeObjectDeleted, event.EventType)
		}
	})

	t.Run("check bucket name", func(t *testing.T) {
		if event.BucketName != "test-bucket" {
			t.Errorf("Expected bucket name 'test-bucket', got %s", event.BucketName)
		}
	})

	t.Run("check object key", func(t *testing.T) {
		if event.ObjectKey != "test.txt" {
			t.Errorf("Expected object key 'test.txt', got %s", event.ObjectKey)
		}
	})
}

// TestInvalidJSON tests handling of invalid JSON input
func TestInvalidJSON(t *testing.T) {
	invalidJSON := "{"

	parser := S3EventParser{}
	_, err := parser.Parse(invalidJSON)
	if err == nil {
		t.Error("Expected error parsing invalid JSON, got nil")
	}
}

// TestLifecycleExpirationEvent tests parsing of lifecycle expiration events
func TestLifecycleExpirationEvent(t *testing.T) {
	expirationEvent := `{
		"version": "1",
		"id": "event-id",
		"detail-type": "Object Deleted",
		"source": "aws.s3",
		"account": "123456789012",
		"time": "2023-01-01T00:00:00Z",
		"region": "us-west-2",
		"resources": [],
		"detail": {
			"version": "1",
			"bucket": {
				"name": "test-bucket"
			},
			"object": {
				"key": "test.txt"
			},
			"request-id": "req123",
			"requester": "user123",
			"source-ip-address": "1.2.3.4",
			"reason": "Lifecycle Expiration",
			"deletion-type": "Lifecycle"
		}
	}`

	parser := S3EventParser{}
	event, err := parser.Parse(expirationEvent)
	if err != nil {
		t.Errorf(expectedParseErrorMsg, err)
	}

	t.Run(testCheckEventType, func(t *testing.T) {
		if event.EventType != eventTypeObjectDeleted {
			t.Errorf(errEventType, eventTypeObjectDeleted, event.EventType)
		}
	})

	t.Run(testCheckReason, func(t *testing.T) {
		if event.Reason != reasonLifecycleExpiration {
			t.Errorf(errReason, reasonLifecycleExpiration, event.Reason)
		}
	})

	t.Run(testCheckBucketName, func(t *testing.T) {
		if event.BucketName != testBucket {
			t.Errorf(errBucketName, testBucket, event.BucketName)
		}
	})

	t.Run(testCheckObjectKey, func(t *testing.T) {
		if event.ObjectKey != testObject {
			t.Errorf(errObjectKey, testObject, event.ObjectKey)
		}
	})
}

// TestInvalidEventSource tests handling of non-S3 event source
func TestInvalidEventSource(t *testing.T) {
	nonS3Event := `{
		"version": "1",
		"id": "event-id",
		"detail-type": "EC2 Instance State-change Notification",
		"source": "aws.ec2",
		"account": "123456789012",
		"time": "2023-01-01T00:00:00Z",
		"region": "us-west-2",
		"resources": [],
		"detail": {}
	}`

	parser := S3EventParser{}
	_, err := parser.Parse(nonS3Event)
	if err == nil {
		t.Error("Expected error parsing non-S3 event, got nil")
	}
}
