#!/bin/bash

# Configuration
BUCKET_NAME="s3-event-exporter-dev-test-bucket"  # Replace with your bucket name
FILE_SIZE_MB=100
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
TEST_FILE="test_file_${TIMESTAMP}.dat"
S3_KEY="multipart-test/${TEST_FILE}"

echo "Starting multipart upload test..."

# Create a test file of specified size
echo "Creating test file of ${FILE_SIZE_MB}MB..."
dd if=/dev/urandom of="${TEST_FILE}" bs=1M count=${FILE_SIZE_MB} 2>/dev/null

# Calculate MD5 of the local file
echo "Calculating local file MD5..."
LOCAL_MD5=$(md5sum "${TEST_FILE}" | cut -d ' ' -f 1)
echo "Local file MD5: ${LOCAL_MD5}"

# Upload file using multipart upload
echo "Uploading file to S3 bucket: ${BUCKET_NAME}"
aws s3 cp "${TEST_FILE}" "s3://${BUCKET_NAME}/${S3_KEY}" \
    --expected-size $((FILE_SIZE_MB * 1024 * 1024)) \
    --storage-class STANDARD

if [ $? -eq 0 ]; then
    echo "Upload successful!"
    
    # Verify the upload by downloading and comparing MD5
    echo "Downloading file to verify..."
    aws s3 cp "s3://${BUCKET_NAME}/${S3_KEY}" "${TEST_FILE}.downloaded"
    
    REMOTE_MD5=$(md5sum "${TEST_FILE}.downloaded" | cut -d ' ' -f 1)
    echo "Remote file MD5: ${REMOTE_MD5}"
    
    if [ "${LOCAL_MD5}" = "${REMOTE_MD5}" ]; then
        echo "MD5 verification successful! Files match."
    else
        echo "Warning: MD5 mismatch between local and remote files!"
    fi
    
    # Clean up
    echo "Cleaning up test files..."
    rm -f "${TEST_FILE}" "${TEST_FILE}.downloaded"
    
    # Display object info
    echo "Object details from S3:"
    aws s3api head-object --bucket "${BUCKET_NAME}" --key "${S3_KEY}"
else
    echo "Upload failed!"
    rm -f "${TEST_FILE}"
    exit 1
fi

echo "Test complete!"
