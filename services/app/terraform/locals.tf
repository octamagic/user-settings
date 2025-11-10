locals {
  environment = nonsensitive(data.aws_ssm_parameter.environment.value)
  aws_region  = data.aws_region.current.name

  tags = {
    Application = var.system
    GitRepo     = "github.com/soundcloud/user-settings"
    ManagedBy   = "Terraform"
    Owner       = "api-team"
    Environment = var.environment
  }
}
