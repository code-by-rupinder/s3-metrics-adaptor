resource "aws_cloudwatch_event_rule" "this" {
  name        = var.rule_name
  description = "Capture all S3 events from test bucket"

  event_pattern = jsonencode({
    source = ["aws.s3"]
    detail-type = [
      "Object Access Tier Changed",
      "Object ACL Updated",
      "Object Created",
      "Object Deleted",
      "Object Restore Completed"
    ]
    detail = {
      bucket = {
        name = [var.bucket_name]
      }
    }
  })

  tags = var.tags
}

resource "aws_cloudwatch_event_target" "this" {
  for_each  = { for t in var.targets : t.target_id => t }
  rule      = aws_cloudwatch_event_rule.this.name
  target_id = each.value.target_id
  arn       = each.value.arn
  role_arn  = each.value.role_arn
}