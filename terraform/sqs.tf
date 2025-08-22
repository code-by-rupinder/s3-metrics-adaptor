# Create SQS Queue
resource "aws_sqs_queue" "s3_event_queue" {
  name                       = "${var.project}-${var.environment}-${var.queue_name}"
  message_retention_seconds  = var.message_retention_seconds
  visibility_timeout_seconds = var.visibility_timeout_seconds

  # Enable server-side encryption
  sqs_managed_sse_enabled = true

  # Allow all principals to send and receive messages (for testing purposes)
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "events.amazonaws.com"
        }
        Action = "sqs:SendMessage"
        Resource = "arn:aws:sqs:${var.aws_region}:*:${var.project}-${var.environment}-${var.queue_name}"
        Condition = {
          ArnEquals = {
            "aws:SourceArn": aws_cloudwatch_event_rule.s3_events.arn
          }
        }
      },
      {
        Effect = "Allow"
        Principal = {
          AWS = "*"
        }
        Action = [
          "sqs:ReceiveMessage",
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes"
        ]
        Resource = "arn:aws:sqs:${var.aws_region}:*:${var.project}-${var.environment}-${var.queue_name}"
      }
    ]
  })

  tags = {
    Environment = var.environment
    Project     = var.project
  }
}
