variable "aws_region" {
  description = "AWS Region"
  type        = string
  default     = "us-west-2"  # Change this to your preferred region
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "dev"
}

variable "project" {
  description = "Project name"
  type        = string
  default     = "s3-event-exporter"
}

variable "queue_name" {
  description = "Name of the SQS queue"
  type        = string
  default     = "s3-event-queue"
}

variable "message_retention_seconds" {
  description = "Number of seconds to retain messages in the queue"
  type        = number
  default     = 345600  # 4 days
}

variable "visibility_timeout_seconds" {
  description = "The visibility timeout for the queue in seconds"
  type        = number
  default     = 30
}
