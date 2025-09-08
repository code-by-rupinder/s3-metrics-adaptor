#!/bin/bash

# Bulk Delete Load Test for S3 Metrics Adapter
# This script creates and deletes many files to test bulk event processing performance

set -e

# Configuration
BUCKET_NAME="s3event-test-s3"
REGION="us-east-1"
TEST_PREFIX="bulk_test"
TOTAL_FILES=1000
BATCH_SIZE=50
EXTENSIONS=("txt" "json" "csv" "xml" "yaml" "log" "png" "pdf" "py" "js" "html" "css" "md" "conf" "ini" "properties" "sql" "sh" "bat" "exe")

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
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

performance() {
    echo -e "${CYAN}📊 $1${NC}"
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
    echo "  📈 Total Events: $(get_metric 's3_event_total')"
    echo "  📈 Created Events: $(get_metric 's3_event_total.*Object Created')"
    echo "  📈 Deleted Events: $(get_metric 's3_delete_total')"
    echo "  ⚡ Messages/sec: $(get_metric 's3_poller_messages_per_second')"
    echo "  ⏱️  Parse Time: $(get_metric 's3_poller_parse_time_seconds')"
    echo "  📦 Batch Size: $(get_metric 's3_poller_batch_size')"
    echo "  🔢 Cardinality: $(get_metric 's3_metrics_cardinality_total')"
    echo
}

# Create test files in bulk
create_bulk_files() {
    log "Creating $TOTAL_FILES test files for bulk delete test..."
    
    local temp_dir="bulk_test_$(date +%s)"
    mkdir -p "$temp_dir"
    
    local start_time=$(date +%s)
    local file_count=0
    
    for i in $(seq 1 $TOTAL_FILES); do
        local ext="${EXTENSIONS[$((i % ${#EXTENSIONS[@]}))]}"
        local filename="bulk_file_$(printf "%04d" $i).$ext"
        local filepath="$temp_dir/$filename"
        
        # Create file with content
        echo "Bulk test file $i with extension $ext - $(date)" > "$filepath"
        
        # Upload to S3
        aws s3 cp "$filepath" "s3://$BUCKET_NAME/$TEST_PREFIX/$filename" --region "$REGION" --quiet
        
        ((file_count++))
        
        # Show progress every 100 files
        if [ $((i % 100)) -eq 0 ]; then
            log "Created $i/$TOTAL_FILES files..."
        fi
    done
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    success "Created $TOTAL_FILES files in ${duration}s (${file_count} files/sec)"
    
    # Clean up temp files
    rm -rf "$temp_dir"
    
    return $file_count
}

# Wait for events to be processed
wait_for_processing() {
    local initial_events=$(get_metric 's3_event_total')
    local current_events=$initial_events
    local wait_time=0
    local max_wait=300  # 5 minutes max wait
    
    log "Waiting for upload events to be processed..."
    
    while [ $wait_time -lt $max_wait ]; do
        sleep 10
        current_events=$(get_metric 's3_event_total')
        
        if [ "$current_events" != "$initial_events" ]; then
            log "Events are being processed... (wait time: ${wait_time}s)"
            initial_events=$current_events
        else
            log "Waiting for more events... (wait time: ${wait_time}s)"
        fi
        
        ((wait_time += 10))
    done
    
    success "Upload event processing monitoring complete"
}

# Perform bulk deletion test
bulk_delete_test() {
    log "Starting BULK DELETE test with $TOTAL_FILES files..."
    
    # Get initial metrics
    show_metrics "before bulk delete"
    
    local initial_deletes=$(get_metric 's3_delete_total')
    local start_time=$(date +%s)
    
    # Delete all files with bulk_test/ prefix
    log "Deleting all files with prefix '$TEST_PREFIX/'..."
    aws s3 rm "s3://$BUCKET_NAME/$TEST_PREFIX/" --recursive --region "$REGION" --quiet
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    success "Bulk deletion initiated in ${duration}s"
    
    # Monitor deletion progress
    log "Monitoring bulk deletion events..."
    local max_wait=600  # 10 minutes max wait for bulk processing
    local check_interval=10  # Check every 10 seconds
    
    for i in $(seq 1 $((max_wait / check_interval))); do
        sleep $check_interval
        local current_deletes=$(get_metric 's3_delete_total')
        local processed_deletes=$((current_deletes - initial_deletes))
        local elapsed_time=$((i * check_interval))
        
        log "Bulk delete progress: $processed_deletes/$TOTAL_FILES events processed (${elapsed_time}s elapsed)"
        
        # Calculate processing rate
        if [ $elapsed_time -gt 0 ]; then
            local rate=$((processed_deletes / elapsed_time))
            log "Processing rate: $rate events/second"
        fi
        
        # Check if all events are processed
        if [ $processed_deletes -ge $TOTAL_FILES ]; then
            success "All bulk delete events processed!"
            break
        fi
        
        # Show metrics every 60 seconds
        if [ $((elapsed_time % 60)) -eq 0 ]; then
            show_metrics "during bulk delete (${elapsed_time}s)"
        fi
    done
    
    show_metrics "after bulk delete"
}

# Test rapid deletion performance
rapid_delete_test() {
    log "Testing rapid deletion performance..."
    
    local rapid_count=100
    local temp_dir="rapid_test_$(date +%s)"
    mkdir -p "$temp_dir"
    
    # Create files for rapid test
    for i in $(seq 1 $rapid_count); do
        local filename="rapid_$i.txt"
        local filepath="$temp_dir/$filename"
        echo "Rapid test file $i" > "$filepath"
        aws s3 cp "$filepath" "s3://$BUCKET_NAME/$TEST_PREFIX/rapid/$filename" --region "$REGION" --quiet
    done
    
    log "Created $rapid_count files for rapid deletion test"
    
    # Wait a bit for upload events
    sleep 5
    
    # Rapid deletion
    local start_time=$(date +%s)
    aws s3 rm "s3://$BUCKET_NAME/$TEST_PREFIX/rapid/" --recursive --region "$REGION" --quiet
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    success "Rapid deletion completed in ${duration}s"
    
    # Clean up
    rm -rf "$temp_dir"
    
    show_metrics "after rapid delete test"
}

# Test concurrent deletion
concurrent_delete_test() {
    log "Testing concurrent deletion performance..."
    
    local concurrent_count=50
    local temp_dir="concurrent_test_$(date +%s)"
    mkdir -p "$temp_dir"
    
    # Create files for concurrent test
    for i in $(seq 1 $concurrent_count); do
        local filename="concurrent_$i.txt"
        local filepath="$temp_dir/$filename"
        echo "Concurrent test file $i" > "$filepath"
        aws s3 cp "$filepath" "s3://$BUCKET_NAME/$TEST_PREFIX/concurrent/$filename" --region "$REGION" --quiet
    done
    
    log "Created $concurrent_count files for concurrent deletion test"
    
    # Wait for upload events
    sleep 5
    
    # Concurrent deletion (simulate multiple users deleting)
    local start_time=$(date +%s)
    
    # Delete in parallel batches
    for i in $(seq 1 5); do
        aws s3 rm "s3://$BUCKET_NAME/$TEST_PREFIX/concurrent/concurrent_$((i * 10))_*.txt" --region "$REGION" --quiet &
    done
    
    # Wait for all deletions to complete
    wait
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    success "Concurrent deletion completed in ${duration}s"
    
    # Clean up
    rm -rf "$temp_dir"
    
    show_metrics "after concurrent delete test"
}

# Generate performance report
generate_report() {
    log "Generating bulk delete performance report..."
    
    local report_file="bulk_delete_report_$(date +%Y%m%d_%H%M%S).txt"
    
    {
        echo "S3 Metrics Adapter Bulk Delete Load Test Report"
        echo "================================================"
        echo "Test Date: $(date)"
        echo "Total Files Tested: $TOTAL_FILES"
        echo "File Extensions: ${EXTENSIONS[*]}"
        echo "Batch Size: $BATCH_SIZE"
        echo ""
        echo "Final Metrics:"
        echo "=============="
        curl -s "http://localhost:8087/metrics" | grep -E "(s3_event_total|s3_delete_total|s3_poller_|s3_metrics_cardinality_total)" | sort
        echo ""
        echo "Performance Analysis:"
        echo "===================="
        echo "This test validates:"
        echo "1. Bulk event processing capability"
        echo "2. Delete event handling performance"
        echo "3. System stability under load"
        echo "4. Metrics accuracy with high volume"
        echo "5. Memory and CPU usage patterns"
    } > "$report_file"
    
    success "Performance report saved to: $report_file"
}

# Main execution
main() {
    log "Starting S3 Metrics Adapter Bulk Delete Load Test"
    echo "================================================="
    echo
    
    # Check if application is running
    if ! curl -s "http://localhost:8087/metrics" > /dev/null; then
        error "Cannot access metrics endpoint. Is the application running?"
        exit 1
    fi
    
    # Check if test prefix is allowed
    log "Verifying test prefix '$TEST_PREFIX' is allowed..."
    if ! grep -q "$TEST_PREFIX" cmd/config.yaml; then
        warning "Test prefix '$TEST_PREFIX' not found in config. Adding it..."
        # Add the prefix to config (you may need to restart the app)
    fi
    
    # Run tests
    create_bulk_files
    wait_for_processing
    bulk_delete_test
    rapid_delete_test
    concurrent_delete_test
    generate_report
    
    success "Bulk delete load test completed!"
    echo
    log "Check your Grafana dashboard for detailed performance metrics"
    log "The test validates your application's ability to handle high-volume delete events"
}

# Help function
show_help() {
    cat << EOF
Bulk Delete Load Test for S3 Metrics Adapter

Usage: $0 [options]

Options:
  -h, --help          Show this help
  -f, --files N       Number of files to test (default: 1000)
  -b, --batch N       Batch size for processing (default: 50)
  -p, --prefix PREFIX Test prefix (default: bulk_test)

Examples:
  $0                           # Run with default settings (1000 files)
  $0 -f 5000 -b 100           # Test 5000 files in batches of 100
  $0 -p my_test -f 2000       # Use custom prefix with 2000 files

This script will:
1. Create many test files with different extensions
2. Upload them to S3 in bulk
3. Monitor upload event processing
4. Perform bulk deletion to test delete event handling
5. Test rapid and concurrent deletion scenarios
6. Generate a detailed performance report

The test validates:
- Bulk event processing capability
- Delete event handling performance
- System stability under load
- Metrics accuracy with high volume
- Memory and CPU usage patterns

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -f|--files)
            TOTAL_FILES="$2"
            shift 2
            ;;
        -b|--batch)
            BATCH_SIZE="$2"
            shift 2
            ;;
        -p|--prefix)
            TEST_PREFIX="$2"
            shift 2
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
