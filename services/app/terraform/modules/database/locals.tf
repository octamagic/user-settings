locals {
  aws_region  = data.aws_region.current.name
  environment = nonsensitive(data.aws_ssm_parameter.environment.value)

  user_settings_table = {
    name                   = "${var.system}-user-settings"
    billing_mode           = var.user_settings_table_read_capacity != null && var.user_settings_table_write_capacity != null ? "PROVISIONED" : "PAY_PER_REQUEST"
    hash_key               = "pk"
    range_key              = "sk"
    read_capacity          = var.user_settings_table_read_capacity
    write_capacity         = var.user_settings_table_write_capacity
    point_in_time_recovery = var.user_settings_table_pitr_enabled
    stream_enabled         = true
    stream_view_type       = "NEW_IMAGE"
  }

  tags = {
    Application = var.system
    GitRepo     = "github.com/soundcloud/user-settings"
    ManagedBy   = "Terraform"
    Environment = local.environment
  }
}
