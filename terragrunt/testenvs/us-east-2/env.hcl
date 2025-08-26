locals {
  environment = "ohio"
  project     = "s3-event-prometheus"
  region      = "us-east-2"
  prefix      = "s3event-test-ohio"
  tags = {
    Environment = local.environment
    Project     = local.project
  }
}
