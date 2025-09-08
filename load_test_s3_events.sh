#!/bin/bash

# S3 Metrics Adapter Load Testing Script
# This script performs comprehensive load testing by:
# 1. Uploading files with different extensions in batches
# 2. Performing bulk deletions to test delete event processing
# 3. Monitoring metrics during the test

set -e

# Configuration
BUCKET_NAME="s3event-test-s3"
REGION="us-east-1"
TEST_DIR="load_test_files"
BATCH_SIZE=50
TOTAL_FILES=500
EXTENSIONS=("txt" "json" "csv" "xml" "yaml" "log" "png" "jpg" "pdf" "zip" "tar.gz" "sql" "py" "js" "html" "css" "md" "conf" "ini" "properties")

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Metrics endpoints
METRICS_URL="http://localhost:8087/metrics"
PROMETHEUS_URL="http://localhost:9090/api/v1/query"

# Helper functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

log_test() {
    echo -e "${PURPLE}🧪 $1${NC}"
}

log_performance() {
    echo -e "${CYAN}📊 $1${NC}"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if AWS CLI is installed
    if ! command -v aws &> /dev/null; then
        log_error "AWS CLI is not installed. Please install it first."
        exit 1
    fi
    
    # Check if jq is installed
    if ! command -v jq &> /dev/null; then
        log_error "jq is not installed. Please install it first."
        exit 1
    fi
    
    # Check if curl is available
    if ! command -v curl &> /dev/null; then
        log_error "curl is not installed. Please install it first."
        exit 1
    fi
    
    # Check if metrics endpoint is accessible
    if ! curl -s "$METRICS_URL" > /dev/null; then
        log_error "Cannot access metrics endpoint at $METRICS_URL. Is the application running?"
        exit 1
    fi
    
    log_success "All prerequisites met"
}

# Create test files with different extensions
create_test_files() {
    log_info "Creating test files with different extensions..."
    
    # Create test directory
    mkdir -p "$TEST_DIR"
    
    local file_count=0
    local extension_index=0
    
    while [ $file_count -lt $TOTAL_FILES ]; do
        local extension="${EXTENSIONS[$extension_index]}"
        local filename="load_test_file_$(printf "%03d" $file_count).$extension"
        local filepath="$TEST_DIR/$filename"
        
        # Create file with different content based on extension
        case $extension in
            "txt"|"log"|"md")
                echo "This is a test file for load testing the S3 metrics adapter. File number: $file_count" > "$filepath"
                ;;
            "json")
                echo "{\"test_file\": true, \"file_number\": $file_count, \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\", \"data\": \"load test data\"}" > "$filepath"
                ;;
            "csv")
                echo "id,name,value,timestamp" > "$filepath"
                echo "$file_count,test_file_$file_count,$(($RANDOM % 1000)),$(date -u +%Y-%m-%dT%H:%M:%SZ)" >> "$filepath"
                ;;
            "xml")
                echo "<?xml version=\"1.0\" encoding=\"UTF-8\"?>" > "$filepath"
                echo "<test_file id=\"$file_count\">" >> "$filepath"
                echo "  <name>load_test_file_$file_count</name>" >> "$filepath"
                echo "  <timestamp>$(date -u +%Y-%m-%dT%H:%M:%SZ)</timestamp>" >> "$filepath"
                echo "  <data>load test data</data>" >> "$filepath"
                echo "</test_file>" >> "$filepath"
                ;;
            "yaml"|"yml")
                echo "test_file: true" > "$filepath"
                echo "file_number: $file_count" >> "$filepath"
                echo "timestamp: $(date -u +%Y-%m-%dT%H:%M:%SZ)" >> "$filepath"
                echo "data: load test data" >> "$filepath"
                ;;
            "py"|"js"|"html"|"css")
                echo "// Test file for load testing" > "$filepath"
                echo "// File number: $file_count" >> "$filepath"
                echo "// Generated at: $(date -u +%Y-%m-%dT%H:%M:%SZ)" >> "$filepath"
                ;;
            "conf"|"ini"|"properties")
                echo "# Test configuration file" > "$filepath"
                echo "file.number=$file_count" >> "$filepath"
                echo "file.timestamp=$(date -u +%Y-%m-%dT%H:%M:%SZ)" >> "$filepath"
                echo "file.data=load test data" >> "$filepath"
                ;;
            *)
                echo "Test file content for load testing. File number: $file_count" > "$filepath"
                ;;
        esac
        
        ((file_count++))
        ((extension_index++))
        if [ $extension_index -ge ${#EXTENSIONS[@]} ]; then
            extension_index=0
        fi
        
        # Show progress every 50 files
        if [ $((file_count % 50)) -eq 0 ]; then
            log_info "Created $file_count test files..."
        fi
    done
    
    log_success "Created $TOTAL_FILES test files with different extensions"
}

# Get current metrics
get_metrics() {
    local metric_name="$1"
    local query="$2"
    
    if [ -n "$query" ]; then
        # Use Prometheus API for complex queries
        curl -s "$PROMETHEUS_URL?query=$query" | jq -r '.data.result[0].value[1] // "0"'
    else
        # Use direct metrics endpoint
        curl -s "$METRICS_URL" | grep "$metric_name" | awk '{print $2}' | head -1 || echo "0"
    fi
}

# Monitor metrics during test
monitor_metrics() {
    local phase="$1"
    
    log_performance "Metrics during $phase:"
    
    # Get basic metrics
    local event_total=$(get_metrics "s3_event_total")
    local delete_total=$(get_metrics "s3_delete_total")
    local created_total=$(get_metrics "s3_event_total" "sum(s3_event_total{event=\"Object Created\"})")
    local deleted_total=$(get_metrics "s3_delete_total" "sum(s3_delete_total)")
    
    # Get performance metrics
    local messages_per_sec=$(get_metrics "s3_poller_messages_per_second" "avg(s3_poller_messages_per_second)")
    local parse_time=$(get_metrics "s3_poller_parse_time_seconds" "avg(s3_poller_parse_time_seconds)")
    local batch_size=$(get_metrics "s3_poller_batch_size" "avg(s3_poller_batch_size)")
    
    # Get cardinality metrics
    local cardinality_total=$(get_metrics "s3_metrics_cardinality_total" "sum(s3_metrics_cardinality_total)")
    
    echo "  📈 Total Events: $event_total"
    echo "  📈 Created Events: $created_total"
    echo "  📈 Deleted Events: $deleted_total"
    echo "  ⚡ Messages/sec: $messages_per_sec"
    echo "  ⏱️  Parse Time: $parse_time seconds"
    echo "  📦 Batch Size: $batch_size"
    echo "  🔢 Cardinality: $cardinality_total"
    echo
}

# Upload files in batches
upload_files() {
    log_test "Starting file upload test..."
    
    local uploaded_count=0
    local batch_count=0
    
    # Get initial metrics
    monitor_metrics "before upload"
    
    for file in "$TEST_DIR"/*; do
        if [ -f "$file" ]; then
            local filename=$(basename "$file")
            local s3_key="load_test/$filename"
            
            # Upload file to S3
            aws s3 cp "$file" "s3://$BUCKET_NAME/$s3_key" --region "$REGION" --quiet
            
            ((uploaded_count++))
            
            # Process in batches
            if [ $((uploaded_count % BATCH_SIZE)) -eq 0 ]; then
                ((batch_count++))
                log_info "Uploaded batch $batch_count ($uploaded_count files total)"
                
                # Wait a bit for events to be processed
                sleep 2
                
                # Monitor metrics after each batch
                monitor_metrics "after batch $batch_count"
            fi
        fi
    done
    
    log_success "Uploaded $uploaded_count files to S3"
    monitor_metrics "after upload"
}

# Wait for events to be processed
wait_for_processing() {
    log_info "Waiting for events to be processed..."
    
    local initial_events=$(get_metrics "s3_event_total" "sum(s3_event_total)")
    local current_events=$initial_events
    local wait_time=0
    local max_wait=300  # 5 minutes max wait
    
    while [ $wait_time -lt $max_wait ]; do
        sleep 10
        current_events=$(get_metrics "s3_event_total" "sum(s3_event_total)")
        
        if [ "$current_events" != "$initial_events" ]; then
            log_info "Events are being processed... (wait time: ${wait_time}s)"
            initial_events=$current_events
        else
            log_info "Waiting for more events... (wait time: ${wait_time}s)"
        fi
        
        ((wait_time += 10))
    done
    
    log_success "Event processing monitoring complete"
}

# Perform bulk deletion test
bulk_delete_test() {
    log_test "Starting bulk deletion test..."
    
    # Get initial delete count
    local initial_deletes=$(get_metrics "s3_delete_total" "sum(s3_delete_total)")
    
    # Delete all files in the load_test/ prefix
    log_info "Deleting all files with prefix 'load_test/'..."
    aws s3 rm "s3://$BUCKET_NAME/load_test/" --recursive --region "$REGION" --quiet
    
    log_success "Bulk deletion initiated"
    
    # Monitor deletion metrics
    local delete_count=0
    local max_wait=180  # 3 minutes max wait for deletions
    
    for i in $(seq 1 18); do  # Check every 10 seconds for 3 minutes
        sleep 10
        local current_deletes=$(get_metrics "s3_delete_total" "sum(s3_delete_total)")
        delete_count=$((current_deletes - initial_deletes))
        
        log_info "Deletion progress: $delete_count delete events processed (${i}0s elapsed)"
        
        if [ $delete_count -ge $TOTAL_FILES ]; then
            log_success "All deletion events processed!"
            break
        fi
    done
    
    monitor_metrics "after bulk deletion"
}

# Test different file extensions separately
test_extension_metrics() {
    log_test "Testing file extension metrics..."
    
    # Get file extension metrics
    local extension_metrics=$(curl -s "$METRICS_URL" | grep "s3_bucket_extension_files_total" || echo "")
    
    if [ -n "$extension_metrics" ]; then
        log_info "File extension metrics found:"
        echo "$extension_metrics" | while read -r line; do
            echo "  $line"
        done
    else
        log_warning "No file extension metrics found"
    fi
}

# Test performance under load
performance_test() {
    log_test "Running performance test..."
    
    # Test with rapid uploads
    log_info "Testing rapid upload performance..."
    local rapid_start=$(date +%s)
    
    for i in {1..20}; do
        local temp_file="$TEST_DIR/rapid_test_$i.txt"
        echo "Rapid test file $i" > "$temp_file"
        aws s3 cp "$temp_file" "s3://$BUCKET_NAME/rapid_test/rapid_$i.txt" --region "$REGION" --quiet
        rm "$temp_file"
    done
    
    local rapid_end=$(date +%s)
    local rapid_duration=$((rapid_end - rapid_start))
    
    log_info "Rapid upload test completed in ${rapid_duration}s"
    
    # Wait for processing
    sleep 5
    
    # Test rapid deletions
    log_info "Testing rapid deletion performance..."
    local delete_start=$(date +%s)
    
    aws s3 rm "s3://$BUCKET_NAME/rapid_test/" --recursive --region "$REGION" --quiet
    
    local delete_end=$(date +%s)
    local delete_duration=$((delete_end - delete_start))
    
    log_info "Rapid deletion test completed in ${delete_duration}s"
    
    monitor_metrics "after performance test"
}

# Generate test report
generate_report() {
    log_test "Generating load test report..."
    
    local report_file="load_test_report_$(date +%Y%m%d_%H%M%S).txt"
    
    {
        echo "S3 Metrics Adapter Load Test Report"
        echo "===================================="
        echo "Test Date: $(date)"
        echo "Total Files Tested: $TOTAL_FILES"
        echo "File Extensions: ${EXTENSIONS[*]}"
        echo "Batch Size: $BATCH_SIZE"
        echo ""
        echo "Final Metrics:"
        echo "=============="
        curl -s "$METRICS_URL" | grep -E "(s3_event_total|s3_delete_total|s3_poller_|s3_metrics_cardinality_total)" | sort
    } > "$report_file"
    
    log_success "Test report saved to: $report_file"
}

# Cleanup function
cleanup() {
    log_info "Cleaning up test files..."
    rm -rf "$TEST_DIR"
    log_success "Cleanup complete"
}

# Main test execution
main() {
    log_info "Starting S3 Metrics Adapter Load Test"
    echo "=========================================="
    echo
    
    # Set up trap for cleanup on exit
    trap cleanup EXIT
    
    # Run tests
    check_prerequisites
    create_test_files
    upload_files
    wait_for_processing
    test_extension_metrics
    performance_test
    bulk_delete_test
    generate_report
    
    log_success "Load test completed successfully!"
    echo
    log_info "Check the generated report for detailed metrics"
    log_info "Monitor your Grafana dashboard for real-time metrics visualization"
}

# Help function
show_help() {
    cat << EOF
S3 Metrics Adapter Load Testing Script

Usage: $0 [options]

Options:
  -h, --help          Show this help message
  -f, --files N       Number of files to test (default: 500)
  -b, --batch N       Batch size for uploads (default: 50)
  -r, --region REGION AWS region (default: us-east-1)
  -B, --bucket NAME   S3 bucket name (default: s3event-test-s3)

Examples:
  $0                           # Run with default settings
  $0 -f 1000 -b 100           # Test 1000 files in batches of 100
  $0 -r us-west-2 -B my-bucket # Use different region and bucket

This script will:
1. Create test files with different extensions
2. Upload them to S3 in batches
3. Monitor metrics during upload
4. Test bulk deletion performance
5. Generate a detailed report

Prerequisites:
- AWS CLI configured with appropriate permissions
- S3 metrics adapter running and accessible
- jq installed for JSON processing
- curl available for HTTP requests

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
        -r|--region)
            REGION="$2"
            shift 2
            ;;
        -B|--bucket)
            BUCKET_NAME="$2"
            shift 2
            ;;
        *)
            log_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Run main function
main
