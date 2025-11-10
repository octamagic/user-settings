data "aws_ssm_parameter" "environment" {
  name = "/aft/account-request/custom-fields/environment"
}

data "aws_ssm_parameter" "base_account_name" {
  name = "/aft/account-request/custom-fields/base_account_name"
}

data "aws_region" "current" {}

data "aws_vpc" "sc" {
  filter {
    name   = "tag:Name"
    values = ["sc-default-vpc"]
  }
}

data "aws_subnets" "private" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.sc.id]
  }
  tags = {
    Tier = "Private"
  }
}
