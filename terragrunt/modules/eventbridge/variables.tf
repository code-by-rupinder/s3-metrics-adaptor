variable "bucket_name" {
	description = "The S3 bucket name to match in EventBridge rule."
	type        = string
}

variable "tags" {
	description = "A map of tags to assign to the resource."
	type        = map(string)
}

variable "rule_name" {
	description = "Name of the EventBridge rule."
	type        = string
}

variable "description" {
	description = "Description of the EventBridge rule."
	type        = string
	default     = ""
}


variable "s3_bucket_arn" {
	description = "ARN of the S3 bucket to monitor."
	type        = string
}

variable "environment" {
	description = "Environment name (e.g., dev, test, prod)."
	type        = string
}



variable "targets" {
	description = "List of targets for the rule. Each target is an object with target_id, arn, and role_arn."
	type = list(object({
		target_id = string
		arn       = string
		role_arn  = string
	}))
	default = []
}
