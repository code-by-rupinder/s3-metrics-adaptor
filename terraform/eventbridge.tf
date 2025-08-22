# Create EventBridge rule to capture all S3 events
resource "aws_cloudwatch_event_rule" "s3_events" {
  name        = "${var.project}-${var.environment}-s3-events"
  description = "Capture all S3 events from test bucket"

  event_pattern = jsonencode({
    source = ["aws.s3"]
    detail-type = [
      "Object Created",
      "Object Created.Put",
      "Object Created.Post",
      "Object Created.Copy",
      "Object Created.CompleteMultipartUpload",
      "Object Deleted",
      "Object Deleted.Delete",
      "Object Deleted.DeleteMarkerCreated"
    ]
    resources = [aws_s3_bucket.test_bucket.arn]
  })

  tags = {
    Environment = var.environment
    Project     = var.project
  }
}

# Add SQS queue as target for the EventBridge rule
resource "aws_cloudwatch_event_target" "sqs" {
  rule      = aws_cloudwatch_event_rule.s3_events.name
  target_id = "SendToSQS"
  arn       = aws_sqs_queue.s3_event_queue.arn
  role_arn  = aws_iam_role.eventbridge_role.arn
}
