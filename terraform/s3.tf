# Create S3 bucket for testing
resource "aws_s3_bucket" "test_bucket" {
  bucket = "${var.project}-${var.environment}-test-bucket"

  tags = {
    Environment = var.environment
    Project     = var.project
  }
}

# Add bucket policy to allow EventBridge to receive events
resource "aws_s3_bucket_policy" "allow_eventbridge" {
  bucket = aws_s3_bucket.test_bucket.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AllowEventBridgeS3EventNotifications"
        Effect = "Allow"
        Principal = {
          Service = "events.amazonaws.com"
        }
        Action = [
          "s3:GetBucketNotification",
          "s3:PutBucketNotification"
        ]
        Resource = aws_s3_bucket.test_bucket.arn
      }
    ]
  })
}

# Enable EventBridge notifications for the bucket
resource "aws_s3_bucket_notification" "bucket_notification" {
  bucket      = aws_s3_bucket.test_bucket.id
  eventbridge = true
}

# Block public access to the bucket
resource "aws_s3_bucket_public_access_block" "test_bucket" {
  bucket = aws_s3_bucket.test_bucket.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# Enable versioning for the bucket
resource "aws_s3_bucket_versioning" "test_bucket" {
  bucket = aws_s3_bucket.test_bucket.id
  versioning_configuration {
    status = "Enabled"
  }
}
