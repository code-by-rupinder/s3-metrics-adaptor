#!/bin/bash

# Simple script to create different S3 event types using AWS CLI
# This script creates actual files and performs real S3 operations

BUCKET="s3event-test-s3"
PREFIX="finance/2025/q1/reports"

echo "Creating different types of S3 events..."

# 1. Put Object (Object Created.Put) - Text file
echo "1. Creating Put event (text file)..."
echo "Hello from Put operation" > test-put.txt
aws s3 cp test-put.txt s3://$BUCKET/$PREFIX/test-put.txt --content-type "text/plain"
echo "   ✓ Put event created (.txt)"

# Wait a moment for the event to be processed
sleep 2

# 2. Copy Object (Object Created.Copy) - JSON file
echo "2. Creating Copy event (JSON file)..."
echo '{"name": "test", "value": 123}' > test-copy.json
aws s3 cp test-copy.json s3://$BUCKET/$PREFIX/test-copy.json --content-type "application/json"
echo "   ✓ Copy event created (.json)"

# Wait a moment for the event to be processed
sleep 2

# 3. Create a large file for multipart upload (CSV file)
echo "3. Creating large file for multipart upload (CSV)..."
echo "id,name,email,date" > large-test-file.csv
for i in {1..10000}; do
    echo "$i,user$i,user$i@example.com,2025-01-01" >> large-test-file.csv
done
echo "   ✓ Large CSV file created"

# 4. Upload large file (triggers multipart upload)
echo "4. Uploading large file (multipart upload)..."
aws s3 cp large-test-file.csv s3://$BUCKET/$PREFIX/test-multipart.csv --content-type "text/csv"
echo "   ✓ Multipart upload event created (.csv)"

# Wait a moment for the event to be processed
sleep 2

# 5. Delete object (for delete events)
echo "5. Creating Delete event..."
aws s3 rm s3://$BUCKET/$PREFIX/test-put.txt
echo "   ✓ Delete event created (.txt)"

# Wait a moment for the event to be processed
sleep 2

# 6. Create additional files with different extensions
echo "6. Creating additional files with different extensions..."

# XML file
echo "6a. Creating XML file..."
echo '<?xml version="1.0"?><data><item>test</item></data>' > test-data.xml
aws s3 cp test-data.xml s3://$BUCKET/$PREFIX/test-data.xml --content-type "application/xml"
echo "   ✓ XML file created (.xml)"

sleep 1

# YAML file
echo "6b. Creating YAML file..."
echo "name: test
version: 1.0
description: Test YAML file" > test-config.yaml
aws s3 cp test-config.yaml s3://$BUCKET/$PREFIX/test-config.yaml --content-type "application/x-yaml"
echo "   ✓ YAML file created (.yaml)"

sleep 1

# Log file
echo "6c. Creating log file..."
echo "2025-01-01 10:00:00 INFO Application started
2025-01-01 10:00:01 DEBUG Loading configuration
2025-01-01 10:00:02 INFO Server listening on port 8080" > test-app.log
aws s3 cp test-app.log s3://$BUCKET/$PREFIX/test-app.log --content-type "text/plain"
echo "   ✓ Log file created (.log)"

sleep 1

# Image file (simulated)
echo "6d. Creating image file..."
echo "fake-image-data" > test-image.png
aws s3 cp test-image.png s3://$BUCKET/$PREFIX/test-image.png --content-type "image/png"
echo "   ✓ Image file created (.png)"

sleep 1

# Archive file
echo "6e. Creating archive file..."
echo "fake-archive-data" > test-backup.tar.gz
aws s3 cp test-backup.tar.gz s3://$BUCKET/$PREFIX/test-backup.tar.gz --content-type "application/gzip"
echo "   ✓ Archive file created (.tar.gz)"

sleep 1

# 7. Create a file that might trigger Post event (using specific headers)
echo "7. Creating Post-like event..."
echo "Post operation content" > test-post.txt
aws s3 cp test-post.txt s3://$BUCKET/$PREFIX/test-post.txt \
    --content-type "application/x-www-form-urlencoded" \
    --metadata "upload-type=post,source=web-form"
echo "   ✓ Post-like event created (.txt)"

# Clean up local files
rm -f test-put.txt test-copy.json test-data.xml test-config.yaml test-app.log test-image.png test-backup.tar.gz test-post.txt large-test-file.csv

echo ""
echo "✅ All events created successfully!"
echo "Check your metrics at http://localhost:8087/metrics"
echo ""
echo "Look for different subtypes in s3_event_total metric:"
echo "  - subtype=\"Put\""
echo "  - subtype=\"Copy\""
echo "  - subtype=\"CompleteMultipartUpload\""
echo "  - subtype=\"Delete\""
echo ""
echo "Look for different file extensions in s3_file_extension_total metric:"
echo "  - extension=\"txt\""
echo "  - extension=\"json\""
echo "  - extension=\"csv\""
echo "  - extension=\"xml\""
echo "  - extension=\"yaml\""
echo "  - extension=\"log\""
echo "  - extension=\"png\""
echo "  - extension=\"tar.gz\""
echo ""
echo "Note: Post events are rare and depend on how the upload was performed."
echo "Most uploads via AWS CLI will show as 'Put' events."
