#!/bin/bash

# Test script to create different types of S3 events
# Make sure your AWS CLI is configured and the application is running

BUCKET="s3event-test-s3"
PREFIX="finance/2025/q1/reports"

echo "Creating different types of S3 events..."

# 1. Put Object (already working)
echo "1. Creating Put event..."
echo "Hello from Put operation" > test-put.txt
aws s3 cp test-put.txt s3://$BUCKET/$PREFIX/test-put.txt --content-type "text/plain"

# 2. Copy Object
echo "2. Creating Copy event..."
aws s3 cp s3://$BUCKET/$PREFIX/test-put.txt s3://$BUCKET/$PREFIX/test-copy.txt

# 3. Create a larger file for multipart upload
echo "3. Creating large file for multipart upload..."
dd if=/dev/zero of=large-test-file.txt bs=1M count=10 2>/dev/null

# 4. Upload large file (triggers multipart upload)
echo "4. Uploading large file (multipart upload)..."
aws s3 cp large-test-file.txt s3://$BUCKET/$PREFIX/test-multipart.txt

# 5. Delete object (for delete events)
echo "5. Creating Delete event..."
aws s3 rm s3://$BUCKET/$PREFIX/test-put.txt

# 6. Create a file with POST-like metadata
echo "6. Creating file with metadata (Post-like)..."
echo "Hello from Post operation" > test-post.txt
aws s3 cp test-post.txt s3://$BUCKET/$PREFIX/test-post.txt \
    --metadata "upload-type=post,source=web-form" \
    --content-type "application/x-www-form-urlencoded"

# Clean up
rm -f large-test-file.txt test-put.txt test-post.txt

echo "Done! Check your metrics at http://localhost:8087/metrics"
echo "Look for different subtypes in s3_event_total metric:"
echo "  - subtype=\"Put\""
echo "  - subtype=\"Copy\""
echo "  - subtype=\"CompleteMultipartUpload\""
echo "  - subtype=\"Delete\""
