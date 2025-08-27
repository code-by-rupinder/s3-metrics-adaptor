package metrics

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestEventMetrics(t *testing.T) {
	cfg := createTestConfig()
	m := Initialize(cfg)
	defer resetMetrics(m)

	tests := []struct {
		name     string
		event    *ParsedEvent
		mainType string
		subType  string
		check    func(t *testing.T, m *Metrics)
	}{
		{
			name: "object creation event",
			event: &ParsedEvent{
				EventType:   eventMainObjectCreated + "." + eventSubTypePut,
				BucketName:  testBucket,
				ObjectKey:   testObjectKey,
				Size:        1024,
				Time:        time.Now(),
				RequesterID: testUser,
				SourceIP:    testSourceIP,
			},
			mainType: eventMainObjectCreated,
			subType:  eventSubTypePut,
			check: func(t *testing.T, m *Metrics) {
				checkEventMetric(t, m, eventMainObjectCreated, testBucket, eventSubTypePut, 1)
				if m.objectSize == nil {
					t.Errorf(errMetricNotInit, "objectSize")
				}
				expectedPath := getPrefixAtDepth(testObjectKey, m.config.Metrics.PrefixDepth)
				checkPrefixDepthMetric(t, m, expectedPath, testBucket, 1)
			},
		},
		{
			name: "delete marker creation",
			event: &ParsedEvent{
				EventType:   eventMainObjectDeleted + "." + eventSubTypeDeleteMarkerCreated,
				BucketName:  testBucket,
				ObjectKey:   testFolder + "/suspicious.txt",
				RequesterID: testSystemUser,
				Reason:      "DeleteObject",
			},
			mainType: eventMainObjectDeleted,
			subType:  eventSubTypeDeleteMarkerCreated,
			check: func(t *testing.T, m *Metrics) {
				if anomalyCount := testutil.ToFloat64(m.anomalyTotal.WithLabelValues("delete_marker_created")); anomalyCount != 1 {
					t.Errorf(errMetricValue, "anomalyTotal", anomalyCount, 1)
				}
				checkEventMetric(t, m, eventMainObjectDeleted, testBucket, eventSubTypeDeleteMarkerCreated, 1)
			},
		},
		{
			name: "multipart upload completion",
			event: &ParsedEvent{
				EventType:   eventMainObjectCreated + "." + eventSubTypeMultipartComplete,
				BucketName:  testBucket,
				ObjectKey:   testObjectKey,
				Size:        1048576,
				Time:        time.Now(),
				RequesterID: testUser,
				SourceIP:    testSourceIP,
			},
			mainType: eventMainObjectCreated,
			subType:  eventSubTypeMultipartComplete,
			check: func(t *testing.T, m *Metrics) {
				checkEventMetric(t, m, eventMainObjectCreated, testBucket, eventSubTypeMultipartComplete, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetMetrics(m)
			m = Initialize(cfg)
			m.UpdateMetrics(tt.event)
			tt.check(t, m)
		})
	}
}
