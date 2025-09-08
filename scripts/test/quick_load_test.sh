#!/bin/bash

# Quick S3 Metrics Adapter Load Test
# Simple script for rapid testing of upload/delete performance

set -e

# Configuration
BUCKET_NAME="s3event-test-s3"
REGION="us-east-1"
TEST_COUNT=100
EXTENSIONS=("txt" "json" "csv" "xml" "yaml" "log" "png" "pdf" "py" "js")

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
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

# Get metrics
get_metric() {
    local metric="$1"
    curl -s "http://localhost:8087/metrics" | grep "$metric" | awk '{print $2}' | head -1 || echo "0"
}

# Monitor metrics
show_metrics() {
    local phase="$1"
    log "Metrics during $phase:"
    echo "  Events: $(get_metric 's3_event_total')"
    echo "  Created: $(get_metric 's3_event_total.*Object Created')"
    echo "  Deleted: $(get_metric 's3_delete_total')"
    echo "  Messages/sec: $(get_metric 's3_poller_messages_per_second')"
    echo
}

# Create and upload test files
upload_test() {
    log "Starting upload test with $TEST_COUNT files..."
    
    local temp_dir="quick_test_$(date +%s)"
    mkdir -p "$temp_dir"
    
    show_metrics "before upload"
    
    local start_time=$(date +%s)
    
    for i in $(seq 1 $TEST_COUNT); do
        local ext="${EXTENSIONS[$((i % ${#EXTENSIONS[@]}))]}"
        local filename="quick_test_$i.$ext"
        local filepath="$temp_dir/$filename"
        
        # Create file
        echo "Quick test file $i with extension $ext" > "$filepath"
        
        # Upload to S3
        aws s3 cp "$filepath" "s3://$BUCKET_NAME/quick_test/$filename" --region "$REGION" --quiet
        
        # Show progress every 20 files
        if [ $((i % 20)) -eq 0 ]; then
            log "Uploaded $i/$TEST_COUNT files..."
        fi
    done
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    success "Uploaded $TEST_COUNT files in ${duration}s"
    show_metrics "after upload"
    
    # Wait for processing
    log "Waiting for events to be processed..."
    sleep 10
    
    # Clean up temp files
    rm -rf "$temp_dir"
}

# Delete test files
delete_test() {
    log "Starting deletion test..."
    
    show_metrics "before deletion"
    
    local start_time=$(date +%s)
    
    # Delete all files with quick_test/ prefix
    aws s3 rm "s3://$BUCKET_NAME/quick_test/" --recursive --region "$REGION" --quiet
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    success "Deleted all test files in ${duration}s"
    
    # Monitor deletion progress
    log "Monitoring deletion events..."
    for i in {1..10}; do
        sleep 5
        local deleted_count=$(get_metric 's3_delete_total')
        log "Delete events processed: $deleted_count (${i}0s elapsed)"
    done
    
    show_metrics "after deletion"
}

# Test different file extensions
extension_test() {
    log "Testing file extension metrics..."
    
    local extensions=("txt" "json" "csv" "xml" "yaml")
    local temp_dir="extension_test_$(date +%s)"
    mkdir -p "$temp_dir"
    
    for ext in "${extensions[@]}"; do
        local filename="test.$ext"
        local filepath="$temp_dir/$filename"
        
        echo "Test file with $ext extension" > "$filepath"
        aws s3 cp "$filepath" "s3://$BUCKET_NAME/extension_test/$filename" --region "$REGION" --quiet
        log "Uploaded $filename"
    done
    
    sleep 5
    
    # Show extension metrics
    log "File extension metrics:"
    curl -s "http://localhost:8087/metrics" | grep "s3_bucket_extension_files_total" | while read -r line; do
        echo "  $line"
    done
    
    # Clean up
    aws s3 rm "s3://$BUCKET_NAME/extension_test/" --recursive --region "$REGION" --quiet
    rm -rf "$temp_dir"
    
    success "Extension test completed"
}

# Main execution
main() {
    log "Starting Quick S3 Metrics Adapter Load Test"
    echo "==========================================="
    echo
    
    # Check if application is running
    if ! curl -s "http://localhost:8087/metrics" > /dev/null; then
        warning "Cannot access metrics endpoint. Is the application running?"
        exit 1
    fi
    
    # Run tests
    upload_test
    delete_test
    extension_test
    
    success "Quick load test completed!"
    echo
    log "Check your Grafana dashboard for detailed metrics visualization"
}

# Help
if [[ "$1" == "-h" || "$1" == "--help" ]]; then
    echo "Quick S3 Metrics Adapter Load Test"
    echo "Usage: $0 [options]"
    echo
    echo "Options:"
    echo "  -h, --help    Show this help"
    echo "  -c, --count N Number of files to test (default: 100)"
    echo
    echo "This script will:"
    echo "1. Upload test files with different extensions"
    echo "2. Monitor metrics during upload"
    echo "3. Delete all test files"
    echo "4. Monitor deletion metrics"
    echo "5. Test file extension metrics"
    exit 0
fi

# Parse count argument
if [[ "$1" == "-c" || "$1" == "--count" ]]; then
    TEST_COUNT="$2"
fi

main
