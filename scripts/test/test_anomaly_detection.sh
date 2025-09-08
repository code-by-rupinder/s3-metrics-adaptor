#!/bin/bash

# Test Anomaly Detection for S3 Metrics Adapter
# This script tests the two types of anomalies that can be detected

set -e

# Configuration
BUCKET_NAME="s3event-test-s3"
REGION="us-east-1"
TEST_PREFIX="anomaly_test"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\0330;36m'
NC='\033[0m'

log() {
    echo -e "${BLUE}[$(date +'%H:%M:%S')]${NC} $1"
}

success() {
    echo -e "${GREEN}✅ $1${NC}"
}

warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

error() {
    echo -e "${RED}❌ $1${NC}"
}

test_info() {
    echo -e "${PURPLE}🧪 $1${NC}"
}

# Get metrics
get_metric() {
    local metric="$1"
    curl -s "http://localhost:8087/metrics" | grep "$metric" | awk '{print $2}' | head -1 || echo "0"
}

# Show current anomaly metrics
show_anomaly_metrics() {
    log "Current Anomaly Metrics:"
    echo "  🔍 System Deletes: $(get_metric 's3_event_anomaly_total.*system_delete')"
    echo "  🔍 Delete Markers: $(get_metric 's3_event_anomaly_total.*delete_marker_created')"
    echo "  🔍 Manual Deletes: $(get_metric 's3_event_anomaly_total.*manual_delete')"
    echo
}

# Test 1: System Delete Anomaly
test_system_delete() {
    test_info "Test 1: System Delete Anomaly"
    log "Creating files that will be deleted by S3 lifecycle..."
    
    # Create files with lifecycle policy that will trigger system deletion
    local temp_dir="anomaly_test_$(date +%s)"
    mkdir -p "$temp_dir"
    
    # Create test files
    for i in {1..5}; do
        local filename="system_delete_test_$i.txt"
        local filepath="$temp_dir/$filename"
        echo "System delete test file $i - $(date)" > "$filepath"
        
        # Upload with metadata that might trigger lifecycle
        aws s3 cp "$filepath" "s3://$BUCKET_NAME/$TEST_PREFIX/system/$filename" \
            --region "$REGION" \
            --metadata "lifecycle-test=true,expiry-date=$(date -d '+1 minute' -u +%Y-%m-%dT%H:%M:%SZ)" \
            --quiet
    done
    
    success "Created 5 files for system delete test"
    log "Note: System deletes require S3 lifecycle policies or automated processes"
    log "These files will only trigger system_delete if S3 automatically deletes them"
    
    # Clean up temp files
    rm -rf "$temp_dir"
}

# Test 2: Delete Marker Creation
test_delete_marker() {
    test_info "Test 2: Delete Marker Creation Anomaly"
    log "Creating delete markers by deleting versioned objects..."
    
    # Enable versioning on bucket (if not already enabled)
    log "Checking if versioning is enabled..."
    local versioning_status=$(aws s3api get-bucket-versioning --bucket "$BUCKET_NAME" --region "$REGION" --query 'Status' --output text 2>/dev/null || echo "NotEnabled")
    
    if [ "$versioning_status" != "Enabled" ]; then
        warning "Versioning is not enabled. Enabling versioning for delete marker test..."
        aws s3api put-bucket-versioning --bucket "$BUCKET_NAME" --versioning-configuration Status=Enabled --region "$REGION"
        log "Versioning enabled. Waiting 30 seconds for propagation..."
        sleep 30
    else
        success "Versioning is already enabled"
    fi
    
    # Create a test file
    local test_file="delete_marker_test.txt"
    local test_content="Delete marker test file - $(date)"
    
    # Upload file (creates version 1)
    echo "$test_content" | aws s3 cp - "s3://$BUCKET_NAME/$TEST_PREFIX/versioned/$test_file" --region "$REGION" --quiet
    success "Uploaded test file (version 1)"
    
    # Delete the file (creates delete marker)
    aws s3 rm "s3://$BUCKET_NAME/$TEST_PREFIX/versioned/$test_file" --region "$REGION" --quiet
    success "Deleted test file (should create delete marker)"
    
    # Upload again (creates version 2)
    echo "$test_content" | aws s3 cp - "s3://$BUCKET_NAME/$TEST_PREFIX/versioned/$test_file" --region "$REGION" --quiet
    success "Uploaded test file again (version 2)"
    
    # Delete again (creates another delete marker)
    aws s3 rm "s3://$BUCKET_NAME/$TEST_PREFIX/versioned/$test_file" --region "$REGION" --quiet
    success "Deleted test file again (should create another delete marker)"
    
    log "Delete markers should trigger 'delete_marker_created' anomalies"
}

# Test 3: Normal Operations (Should NOT trigger anomalies)
test_normal_operations() {
    test_info "Test 3: Normal Operations (Should NOT trigger anomalies)"
    log "Creating and deleting files normally..."
    
    local temp_dir="normal_test_$(date +%s)"
    mkdir -p "$temp_dir"
    
    # Create and delete files normally
    for i in {1..10}; do
        local filename="normal_test_$i.txt"
        local filepath="$temp_dir/$filename"
        echo "Normal test file $i - $(date)" > "$filepath"
        
        # Upload
        aws s3 cp "$filepath" "s3://$BUCKET_NAME/$TEST_PREFIX/normal/$filename" --region "$REGION" --quiet
        
        # Delete immediately
        aws s3 rm "s3://$BUCKET_NAME/$TEST_PREFIX/normal/$filename" --region "$REGION" --quiet
    done
    
    success "Created and deleted 10 files normally"
    log "These operations should NOT trigger any anomalies"
    
    # Clean up temp files
    rm -rf "$temp_dir"
}

# Test 4: Bulk Operations (Should NOT trigger anomalies with current logic)
test_bulk_operations() {
    test_info "Test 4: Bulk Operations (Should NOT trigger anomalies)"
    log "Creating and deleting many files in bulk..."
    
    local temp_dir="bulk_test_$(date +%s)"
    mkdir -p "$temp_dir"
    
    # Create many files
    for i in {1..50}; do
        local filename="bulk_test_$i.txt"
        local filepath="$temp_dir/$filename"
        echo "Bulk test file $i - $(date)" > "$filepath"
        aws s3 cp "$filepath" "s3://$BUCKET_NAME/$TEST_PREFIX/bulk/$filename" --region "$REGION" --quiet
    done
    
    success "Created 50 files for bulk test"
    
    # Wait a bit
    sleep 5
    
    # Delete all files
    aws s3 rm "s3://$BUCKET_NAME/$TEST_PREFIX/bulk/" --recursive --region "$REGION" --quiet
    success "Deleted all 50 files in bulk"
    
    log "Bulk operations should NOT trigger anomalies with current logic"
    
    # Clean up temp files
    rm -rf "$temp_dir"
}

# Monitor metrics during test
monitor_metrics() {
    local test_name="$1"
    log "Monitoring metrics during $test_name..."
    
    local initial_system=$(get_metric 's3_event_anomaly_total.*system_delete')
    local initial_marker=$(get_metric 's3_event_anomaly_total.*delete_marker_created')
    local initial_manual=$(get_metric 's3_event_anomaly_total.*manual_delete')
    
    echo "  📊 Initial: System=$initial_system, Marker=$initial_marker, Manual=$initial_manual"
    
    # Wait for events to be processed
    sleep 10
    
    local final_system=$(get_metric 's3_event_anomaly_total.*system_delete')
    local final_marker=$(get_metric 's3_event_anomaly_total.*delete_marker_created')
    local final_manual=$(get_metric 's3_event_anomaly_total.*manual_delete')
    
    echo "  📊 Final:   System=$final_system, Marker=$final_marker, Manual=$final_manual"
    
    # Calculate changes
    local system_change=$((final_system - initial_system))
    local marker_change=$((final_marker - initial_marker))
    local manual_change=$((final_manual - initial_manual))
    
    echo "  📈 Changes: System=+$system_change, Marker=+$marker_change, Manual=+$manual_change"
    echo
}

# Main execution
main() {
    log "Starting Anomaly Detection Tests"
    echo "================================="
    echo
    
    # Check if application is running
    if ! curl -s "http://localhost:8087/metrics" > /dev/null; then
        error "Cannot access metrics endpoint. Is the application running?"
        exit 1
    fi
    
    # Show initial metrics
    show_anomaly_metrics
    
    # Run tests
    test_normal_operations
    monitor_metrics "Normal Operations"
    
    test_bulk_operations
    monitor_metrics "Bulk Operations"
    
    test_delete_marker
    monitor_metrics "Delete Marker Test"
    
    test_system_delete
    monitor_metrics "System Delete Test"
    
    # Final metrics
    log "Final Anomaly Metrics:"
    show_anomaly_metrics
    
    success "Anomaly detection tests completed!"
    echo
    log "Expected Results:"
    echo "  ✅ Normal operations: No anomalies"
    echo "  ✅ Bulk operations: No anomalies"
    echo "  ✅ Delete markers: Should show 'delete_marker_created' anomalies"
    echo "  ✅ System deletes: May show 'system_delete' anomalies (depends on S3 lifecycle)"
    echo
    log "Check your Grafana dashboard for anomaly metrics visualization"
}

# Help function
show_help() {
    cat << EOF
Anomaly Detection Test for S3 Metrics Adapter

Usage: $0 [options]

This script tests the anomaly detection functionality by:
1. Normal operations (should NOT trigger anomalies)
2. Bulk operations (should NOT trigger anomalies)
3. Delete marker creation (should trigger 'delete_marker_created')
4. System deletions (may trigger 'system_delete' if S3 lifecycle is configured)

The script monitors metrics before and after each test to show changes.

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Run main function
main
