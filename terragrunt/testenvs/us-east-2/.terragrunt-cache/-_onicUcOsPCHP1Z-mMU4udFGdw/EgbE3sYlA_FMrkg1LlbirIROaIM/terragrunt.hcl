include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../modules/s3_bucket"
}

inputs = {
  bucket_name = "test-bucket-use2"
  tags = {
    Environment = "test"
    Project     = "s3-event-exporter"
  }
}
