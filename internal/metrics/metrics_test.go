package metrics

import (
	"testing"
)

func TestInitialize(t *testing.T) {
	cfg := createTestConfig()
	metrics := Initialize(cfg)

	if metrics == nil {
		t.Fatal("Expected non-nil metrics instance")
	}

	// Test that metrics are properly initialized
	if metrics.eventTotal == nil {
		t.Errorf(errMetricNotInit, "eventTotal")
	}
	if metrics.objectSize == nil {
		t.Errorf(errMetricNotInit, "objectSize")
	}
	if metrics.ipTotal == nil {
		t.Errorf(errMetricNotInit, "ipTotal")
	}
	if metrics.prefixDepthTotal == nil {
		t.Errorf(errMetricNotInit, "prefixDepthTotal")
	}
}
