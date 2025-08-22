package metrics

import (
	"testing"
	"time"
)

func TestFileExtensionMetrics(t *testing.T) {
	cfg := createTestConfig()
	m := Initialize(cfg)
	defer resetMetrics(m)

	tests := []struct {
		name  string
		event *ParsedEvent
		check func(t *testing.T, m *Metrics)
	}{
		{
			name: "file with extension",
			event: &ParsedEvent{
				EventType:   eventMainObjectCreated + "." + eventSubTypePut,
				BucketName:  testBucket,
				ObjectKey:   testObjectKey,
				Size:        1024,
				Time:        time.Now(),
				RequesterID: testUser,
			},
			check: func(t *testing.T, m *Metrics) {
				checkFileExtensionMetric(t, m, testBucket, ".txt", testFolder, "file", 1)
			},
		},
		{
			name: "file without extension",
			event: &ParsedEvent{
				EventType:   eventMainObjectCreated + "." + eventSubTypePut,
				BucketName:  testBucket,
				ObjectKey:   testFileNoExt,
				Size:        1024,
				Time:        time.Now(),
				RequesterID: testUser,
			},
			check: func(t *testing.T, m *Metrics) {
				checkFileExtensionMetric(t, m, testBucket, "none", testFolder, "file", 1)
			},
		},
		{
			name: "directory",
			event: &ParsedEvent{
				EventType:   eventMainObjectCreated + "." + eventSubTypePut,
				BucketName:  testBucket,
				ObjectKey:   testDirectory,
				Size:        0,
				Time:        time.Now(),
				RequesterID: testUser,
			},
			check: func(t *testing.T, m *Metrics) {
				checkFileExtensionMetric(t, m, testBucket, "none", testFolder, "directory", 1)
			},
		},
		{
			name: "file deletion",
			event: &ParsedEvent{
				EventType:   eventMainObjectDeleted + "." + eventSubTypeDelete,
				BucketName:  testBucket,
				ObjectKey:   testObjectKey,
				RequesterID: testUser,
			},
			check: func(t *testing.T, m *Metrics) {
				checkFileExtensionMetric(t, m, testBucket, ".txt", testFolder, "file", -1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetMetrics(m)  // Unregister existing metrics
			m = Initialize(cfg)  // Create new metrics instance
			
			// For deletion tests, set initial value after initialization but before processing event
			if tt.event.EventType == eventMainObjectDeleted+"."+eventSubTypeDelete && m.fileExtensionTotal != nil {
				info := getFileInfo(tt.event.ObjectKey)
				m.fileExtensionTotal.WithLabelValues(
					tt.event.BucketName,
					info.extension,
					testFolder,
					info.fileType,
				).Set(1)
			}

			// Process the event
			m.UpdateMetrics(tt.event)
			tt.check(t, m)
		})
	}
}
