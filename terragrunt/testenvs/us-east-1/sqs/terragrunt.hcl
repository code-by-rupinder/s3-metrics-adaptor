include "env" {
  path = find_in_parent_folders("env.hcl")
  expose = true
}
include "provider" {
  path = find_in_parent_folders("provider.hcl")
}
terraform {
  source = "../../../modules/sqs_queue"
}

inputs = {
  queue_name                  = "${include.env.locals.prefix}-sqs"
  message_retention_seconds   = 345600
  visibility_timeout_seconds  = 30
  tags                       = include.env.locals.tags
}
