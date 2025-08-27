#!/bin/bash
BUCKET_NAME="s3-event-exporter-dev-test-bucket"  # Replace with your bucket name
# Create test files of different sizes
dd if=/dev/zero of=small.txt bs=512 count=1         # 512B (below 1KB)
dd if=/dev/zero of=medium.txt bs=50K count=1        # 50KB (between 1KB and 100KB)
dd if=/dev/zero of=large.txt bs=500K count=1        # 500KB (between 100KB and 1MB)
dd if=/dev/zero of=xlarge.txt bs=2M count=1         # 2MB (above 1MB)

# Upload files to different paths
aws s3 cp small.txt s3://${BUCKET_NAME}/finance/2025/q1/reports/small.txt
sleep 3
aws s3 cp medium.txt s3://${BUCKET_NAME}/finance/2025/q1/reports/medium.txt
sleep 5
aws s3 cp large.txt s3://${BUCKET_NAME}/finance/2025/q2/reports/large.txt
sleep 5
aws s3 cp xlarge.txt s3://${BUCKET_NAME}/finance/2025/q2/reports/xlarge.txt
sleep 5


# Create some files in different paths to demonstrate prefix tracking
aws s3 cp small.txt s3://${BUCKET_NAME}/hr/2025/employees/list.txt
aws s3 cp medium.txt s3://${BUCKET_NAME}/it/2025/servers/config.yaml
aws s3 cp large.txt s3://${BUCKET_NAME}/marketing/2025/campaigns/data.csv

# Clean up local files
rm small.txt medium.txt large.txt xlarge.txt
