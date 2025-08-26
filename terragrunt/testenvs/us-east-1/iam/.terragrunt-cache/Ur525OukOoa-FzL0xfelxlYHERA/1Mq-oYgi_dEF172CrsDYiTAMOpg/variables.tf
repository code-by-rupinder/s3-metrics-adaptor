variable "role_name" { type = string }
variable "assume_role_policy" { type = string }
variable "description" { type = string }
variable "tags" { type = map(string) }

# Inline policies: list of objects { name = string, policy = string }
variable "inline_policies" {
	type    = list(object({ name = string, policy = string }))
	default = []
}

# Managed policy ARNs to attach
variable "managed_policy_arns" {
	type    = list(string)
	default = []
}
