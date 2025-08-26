locals {
  environment = "test"
  project     = "s3-event-exporter"
  region      = "us-east-1"
  prefix      = "s3event-test"
  tags = {
    Environment = local.environment
    Project     = local.project
  }
}
