
variable "bucket_name" {
	description = "Name of the S3 bucket."
	type        = string
}

variable "acl" {
	description = "Canned ACL to apply. Defaults to private."
	type        = string
	default     = "private"
}

variable "tags" {
	description = "Tags to apply to the bucket."
	type        = map(string)
	default     = {}
}

variable "enable_versioning" {
	description = "Enable versioning for the bucket."
	type        = bool
	default     = true
}

variable "enable_public_access_block" {
	description = "Block all public access to the bucket."
	type        = bool
	default     = true
}

variable "enable_eventbridge_notification" {
	description = "Enable EventBridge notifications for the bucket."
	type        = bool
	default     = false
}

variable "policy_json" {
	description = "JSON IAM policy to attach to the bucket."
	type        = string
	default     = ""
}
