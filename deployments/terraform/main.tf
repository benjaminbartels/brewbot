provider "aws" {}

terraform {
  backend "s3" {
    key = "brewbot"
  }
}

data "terraform_remote_state" "infrastucture" {
  backend = "s3"
  config = {
    bucket = var.terraformStateS3BucketName
    key    = "infrastucture"
  }
}
