data "aws_region" "current" {}

data "aws_ssm_parameter" "environment" {
  name = "/aft/account-request/custom-fields/environment"
}
