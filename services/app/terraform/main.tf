module "alerts" {
  count  = var.alerts_enabled ? 1 : 0
  source = "./modules/alerts"

  system                = var.system
  alert_on_no_data      = var.alert_on_no_data
  enable_pagerduty      = var.enable_pagerduty
  default_rule_severity = var.default_rule_severity
}

module "database" {
  source = "./modules/database"

  system = var.system
  tags   = merge(var.tags, local.tags)

  user_settings_table_read_capacity  = var.user_settings_table_read_capacity
  user_settings_table_write_capacity = var.user_settings_table_write_capacity
  user_settings_table_pitr_enabled   = var.user_settings_table_pitr_enabled
}

module "api_service" {
  source = "./modules/service"

  commit_sha               = var.commit_sha
  tags                     = merge(var.tags, local.tags)
  user_settings_table_name = module.database.user_settings_table_name
}
