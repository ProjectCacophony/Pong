terraform {
  backend "s3" {
    bucket = "cacophony-terraform"
    key    = "lambda-pong"
    region = "us-east-1"
  }
}
provider "aws" {
  region = "us-east-1"
}