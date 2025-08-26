
output "rule_arn" {
	value       = aws_cloudwatch_event_rule.this.arn
	description = "The ARN of the EventBridge rule."
}

output "rule_id" {
	value       = aws_cloudwatch_event_rule.this.id
	description = "The ID of the EventBridge rule."
}

output "target_arns" {
	value       = [for t in aws_cloudwatch_event_target.this : t.arn]
	description = "The ARNs of the EventBridge targets."
}
