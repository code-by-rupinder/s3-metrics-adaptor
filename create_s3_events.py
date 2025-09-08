#!/usr/bin/env python3
"""
Script to create different types of S3 events for testing the metrics adapter.
This script uses boto3 to perform various S3 operations that will trigger different event types.
"""

import boto3
import os
import time
from botocore.exceptions import ClientError

def create_s3_events():
    # Initialize S3 client
    s3 = boto3.client('s3')
    bucket = 's3event-test-s3'
    prefix = 'finance/2025/q1/reports'
    
    print("Creating different types of S3 events...")
    
    try:
        # 1. Put Object (Object Created.Put)
        print("1. Creating Put event...")
        s3.put_object(
            Bucket=bucket,
            Key=f'{prefix}/test-put.txt',
            Body=b'Hello from Put operation',
            ContentType='text/plain'
        )
        print("   ✓ Put event created")
        
        # 2. Copy Object (Object Created.Copy)
        print("2. Creating Copy event...")
        s3.copy_object(
            Bucket=bucket,
            CopySource={'Bucket': bucket, 'Key': f'{prefix}/test-put.txt'},
            Key=f'{prefix}/test-copy.txt'
        )
        print("   ✓ Copy event created")
        
        # 3. Multipart Upload (Object Created.CompleteMultipartUpload)
        print("3. Creating Multipart Upload event...")
        # Create a large file for multipart upload
        large_file = 'large-test-file.txt'
        with open(large_file, 'wb') as f:
            f.write(b'x' * (10 * 1024 * 1024))  # 10MB file
        
        # Start multipart upload
        response = s3.create_multipart_upload(
            Bucket=bucket,
            Key=f'{prefix}/test-multipart.txt',
            ContentType='text/plain'
        )
        upload_id = response['UploadId']
        
        # Upload parts
        part_number = 1
        with open(large_file, 'rb') as f:
            while True:
                data = f.read(5 * 1024 * 1024)  # 5MB parts
                if not data:
                    break
                
                response = s3.upload_part(
                    Bucket=bucket,
                    Key=f'{prefix}/test-multipart.txt',
                    PartNumber=part_number,
                    UploadId=upload_id,
                    Body=data
                )
                part_number += 1
        
        # Complete multipart upload
        s3.complete_multipart_upload(
            Bucket=bucket,
            Key=f'{prefix}/test-multipart.txt',
            UploadId=upload_id
        )
        print("   ✓ Multipart upload event created")
        
        # 4. Delete Object (Object Deleted.Delete)
        print("4. Creating Delete event...")
        s3.delete_object(
            Bucket=bucket,
            Key=f'{prefix}/test-put.txt'
        )
        print("   ✓ Delete event created")
        
        # 5. Post Object (Object Created.Post) - This is tricky as it requires form-based upload
        # We'll simulate it by using put_object with specific metadata
        print("5. Creating Post-like event...")
        s3.put_object(
            Bucket=bucket,
            Key=f'{prefix}/test-post.txt',
            Body=b'Hello from Post operation',
            ContentType='application/x-www-form-urlencoded',
            Metadata={
                'upload-type': 'post',
                'source': 'web-form'
            }
        )
        print("   ✓ Post-like event created")
        
        # 6. Create another file for additional testing
        print("6. Creating additional Put event...")
        s3.put_object(
            Bucket=bucket,
            Key=f'{prefix}/test-additional.txt',
            Body=b'Additional test file',
            ContentType='text/plain'
        )
        print("   ✓ Additional Put event created")
        
        # Clean up local files
        if os.path.exists(large_file):
            os.remove(large_file)
        
        print("\n✅ All events created successfully!")
        print("Check your metrics at http://localhost:8087/metrics")
        print("Look for different subtypes in s3_event_total metric:")
        print("  - subtype=\"Put\"")
        print("  - subtype=\"Copy\"")
        print("  - subtype=\"CompleteMultipartUpload\"")
        print("  - subtype=\"Delete\"")
        print("  - subtype=\"Post\" (if supported by your S3 configuration)")
        
    except ClientError as e:
        print(f"❌ Error creating S3 events: {e}")
    except Exception as e:
        print(f"❌ Unexpected error: {e}")

if __name__ == "__main__":
    create_s3_events()
