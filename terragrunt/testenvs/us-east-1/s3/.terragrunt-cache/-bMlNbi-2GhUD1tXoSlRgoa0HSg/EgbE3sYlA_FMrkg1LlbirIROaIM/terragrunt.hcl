include "env" {
  path = find_in_parent_folders("env.hcl")
  expose = true
}
include "provider" {
  path = find_in_parent_folders("provider.hcl")
}
terraform {
  source = "../../../modules/s3_bucket"
}

inputs = {
  bucket_name                  = "${include.env.locals.prefix}-s3"
  tags                         = include.env.locals.tags
  enable_versioning            = true
  enable_public_access_block   = true
  enable_eventbridge_notification = true
  policy_json                  = file("../../../policies/s3-bucket-policy.json")
}
