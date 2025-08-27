
variable "queue_name" {
	description = "The name of the SQS queue."
	type        = string
}

variable "message_retention_seconds" {
	description = "The number of seconds Amazon SQS retains a message."
	type        = number
}

variable "visibility_timeout_seconds" {
	description = "The visibility timeout for the queue, in seconds."
	type        = number
}


variable "tags" {
	description = "A map of tags to assign to the resource."
	type        = map(string)
}
// ...existing code from terraform/modules/sqs_queue/variables.tf...
