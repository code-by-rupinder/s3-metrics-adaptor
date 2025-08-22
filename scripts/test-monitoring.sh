#!/bin/bash
# Test script to validate monitoring configurations

set -e

echo "🧪 Testing Helm Chart Monitoring Configurations"
echo "=============================================="

# Test 1: Default configuration (no monitoring)
echo "📋 Test 1: Default configuration (ServiceMonitor disabled)"
helm template test helm/s3-event-exporter/ --dry-run > /tmp/default-test.yaml
if grep -q "ServiceMonitor" /tmp/default-test.yaml; then
    echo "❌ FAIL: ServiceMonitor should not be created by default"
    exit 1
else
    echo "✅ PASS: ServiceMonitor not created by default"
fi

# Test 2: Prometheus Operator configuration
echo "📋 Test 2: Prometheus Operator configuration"
helm template test helm/s3-event-exporter/ \
    --set serviceMonitor.enabled=true \
    --set serviceMonitor.additionalLabels.release=prometheus \
    --dry-run > /tmp/operator-test.yaml

if grep -q "kind: ServiceMonitor" /tmp/operator-test.yaml; then
    echo "✅ PASS: ServiceMonitor created when enabled"
else
    echo "❌ FAIL: ServiceMonitor not created when enabled"
    exit 1
fi

if grep -q "release: prometheus" /tmp/operator-test.yaml; then
    echo "✅ PASS: Additional labels applied to ServiceMonitor"
else
    echo "❌ FAIL: Additional labels not applied"
    exit 1
fi

# Test 3: External Prometheus configuration
echo "📋 Test 3: External Prometheus configuration"
helm template test helm/s3-event-exporter/ \
    -f helm/examples/values-external-prometheus.yaml \
    --dry-run > /tmp/external-test.yaml

if grep -q "prometheus.io/scrape.*true" /tmp/external-test.yaml; then
    echo "✅ PASS: Service annotations for external Prometheus"
else
    echo "❌ FAIL: Service annotations missing"
    exit 1
fi

if grep -q "ServiceMonitor" /tmp/external-test.yaml; then
    echo "❌ FAIL: ServiceMonitor should not be created for external Prometheus"
    exit 1
else
    echo "✅ PASS: No ServiceMonitor created for external Prometheus"
fi

# Test 4: Chart validation
echo "📋 Test 4: Chart validation"
if helm lint helm/s3-event-exporter/ | grep -q "0 chart(s) failed"; then
    echo "✅ PASS: Chart linting successful"
else
    echo "❌ FAIL: Chart linting failed"
    exit 1
fi

# Test 5: Template validation with various configurations
echo "📋 Test 5: Template validation with edge cases"

# Test with both ServiceMonitor and external annotations (should work)
helm template test helm/s3-event-exporter/ \
    --set serviceMonitor.enabled=true \
    --set service.annotations.'prometheus\.io/scrape'=true \
    --dry-run > /tmp/both-test.yaml

if grep -q "ServiceMonitor" /tmp/both-test.yaml && grep -q "prometheus.io/scrape" /tmp/both-test.yaml; then
    echo "✅ PASS: Both ServiceMonitor and annotations can coexist"
else
    echo "❌ FAIL: Configuration conflict detected"
    exit 1
fi

echo ""
echo "🎉 All monitoring configuration tests passed!"
echo ""
echo "Summary of available configurations:"
echo "1. Default: No monitoring (users choose their method)"
echo "2. Prometheus Operator: ServiceMonitor enabled with labels"
echo "3. External Prometheus: Service annotations for scraping"
echo "4. Hybrid: Both methods can coexist if needed"
echo ""
echo "For deployment guidance, see:"
echo "- helm/MONITORING.md - Complete monitoring setup guide"
echo "- helm/examples/values-external-prometheus.yaml"
echo "- helm/examples/values-prometheus-operator.yaml"
