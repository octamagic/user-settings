locals {
  aws_region  = data.aws_region.current.name
  environment = nonsensitive(data.aws_ssm_parameter.environment.value)
}
