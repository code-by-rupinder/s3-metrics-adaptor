include "env" {
  path = find_in_parent_folders("env.hcl")
  expose = true
}
include "provider" {
  path = find_in_parent_folders("provider.hcl")
}
terraform {
  source = "../../../modules/iam_role"
}

inputs = {
  role_name           = "${include.env.locals.prefix}-eventbridge-role"
  description         = "IAM role for EventBridge in ${include.env.locals.region}"
  assume_role_policy  = file("../../../policies/assume-role-policy.json")
  managed_policy_arns = ["arn:aws:iam::aws:policy/AmazonSQSFullAccess"]
  inline_policies     = []
  tags                = include.env.locals.tags
}
