import boto3
import os
from datetime import datetime

def perform_multipart_upload():
    s3_client = boto3.client('s3')
    bucket_name = 's3-bucket'  # Replace with your bucket name
    file_name = f'multipart-test-{datetime.now().strftime("%Y%m%d-%H%M%S")}.dat'
    
    # Create a large file (100MB)
    file_size = 100 * 1024 * 1024  # 100MB
    with open('large_test_file.dat', 'wb') as f:
        f.write(os.urandom(file_size))
    
    # Initiate multipart upload
    mpu = s3_client.create_multipart_upload(Bucket=bucket_name, Key=file_name)
    
    # Upload parts
    chunk_size = 5 * 1024 * 1024  # 5MB chunks
    parts = []
    
    with open('large_test_file.dat', 'rb') as f:
        part_number = 1
        while True:
            data = f.read(chunk_size)
            if not data:
                break
                
            part = s3_client.upload_part(
                Bucket=bucket_name,
                Key=file_name,
                PartNumber=part_number,
                UploadId=mpu['UploadId'],
                Body=data
            )
            
            parts.append({
                'PartNumber': part_number,
                'ETag': part['ETag']
            })
            
            part_number += 1
    
    # Complete multipart upload
    s3_client.complete_multipart_upload(
        Bucket=bucket_name,
        Key=file_name,
        UploadId=mpu['UploadId'],
        MultipartUpload={'Parts': parts}
    )
    
    # Clean up test file
    os.remove('large_test_file.dat')
    print(f"Completed multipart upload: {file_name}")

if __name__ == '__main__':
    perform_multipart_upload()
