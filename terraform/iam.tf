# Create IAM role for EventBridge to send events to SQS
resource "aws_iam_role" "eventbridge_role" {
  name = "${var.project}-${var.environment}-eventbridge-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "events.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })

  tags = {
    Environment = var.environment
    Project     = var.project
  }
}

# Create IAM policy for EventBridge to send messages to SQS
resource "aws_iam_role_policy" "eventbridge_policy" {
  name = "${var.project}-${var.environment}-eventbridge-policy"
  role = aws_iam_role.eventbridge_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "sqs:SendMessage"
        ]
        Resource = aws_sqs_queue.s3_event_queue.arn
      }
    ]
  })
}
