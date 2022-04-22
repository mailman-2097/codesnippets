terraform {
  backend "s3" {
    bucket = "dc-tf-state-bucket"
    key    = "org-import"
    region = "ap-southeast-2"
  }
}