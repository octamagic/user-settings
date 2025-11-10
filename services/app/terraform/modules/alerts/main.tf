module "alerts" {
  source                       = "github.com/soundcloud/tf-common//modules/grafana-alerts?ref=grafana-alerts@1.5.0"
  system                       = var.system
  repo                         = "soundcloud/user-settings"
  owner                        = "api-team"
  component                    = "api"
  service_name                 = var.system
  environment                  = local.environment
  runbook_url                  = var.runbook_url
  pagerduty_service_name       = "API Team (Non GQL)"
  slack_channel                = "api-team-alerts"
  alert_on_no_data             = var.alert_on_no_data
  enable_pagerduty             = var.enable_pagerduty
  create_default_rules         = true
  default_rule_severity        = var.default_rule_severity # This is the severity level to use for default rules. You might want to change this to info or warning for dev or stage environments.
  error_threshold              = var.error_threshold
  error_aggregation_duration   = var.error_duration
  latency_threshold            = var.latency_threshold
  latency_aggregation_duration = var.latency_duration
}
