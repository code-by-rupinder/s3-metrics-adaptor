include "env" {
  path = find_in_parent_folders("env.hcl")
  expose = true
}
include "provider" {
  path = find_in_parent_folders("provider.hcl")
}
terraform {
  source = "../../../modules/eventbridge"
}

dependency "iam" {
  config_path = "../iam"
  mock_outputs = {
    role_arn = "arn:aws:iam::123456789012:role/mock-eventbridge-role"
  }
}

dependency "s3" {
  config_path = "../s3"
  mock_outputs = {
    bucket_arn = "arn:aws:s3:::mock-bucket"
    bucket_id = "${include.env.locals.prefix}-s3"
  }
}


dependency "sqs" {
  config_path = "../sqs"
  mock_outputs = {
    queue_arn = "arn:aws:sqs:us-east-2:123456789012:mock-queue"
  }
}

inputs = {
  rule_name    = "${include.env.locals.prefix}-eventbridge-rule"
  description  = "Capture S3 events in ${include.env.locals.region}"
  bucket_name = dependency.s3.outputs.bucket_id
  s3_bucket_arn = dependency.s3.outputs.bucket_arn
  environment   = include.env.locals.environment
  project       = include.env.locals.project
  targets = [{
    target_id = "SendToSQS"
    arn       = dependency.sqs.outputs.queue_arn
    role_arn  = dependency.iam.outputs.role_arn
  }]
  tags = include.env.locals.tags
}
