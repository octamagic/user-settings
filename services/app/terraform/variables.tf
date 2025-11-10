variable "alerts_enabled" {
  description = "Whether or not to enable alerts for the environment"
  type        = bool
  default     = false
}

variable "commit_sha" {
  description = "The commit being deployed, used for the docker image tags"
  type        = string
}

variable "environment" {
  description = "The environment to deploy to"
  type        = string
}

variable "system" {
  description = "The name of the system"
  type        = string
  default     = "user-settings"
}

variable "tags" {
  description = "Tags to attach to all resources"
  type        = map(string)
  default     = {}
}

variable "user_settings_table_read_capacity" {
  description = "Dynamo read capacity for user_settings table when using provisioned billing mode"
  type        = number
  default     = null
}

variable "user_settings_table_write_capacity" {
  description = "Dynamo write capacity for versions table when using provisioned billing mode"
  type        = number
  default     = null
}

variable "user_settings_table_pitr_enabled" {
  description = "Whether or not to enable point in time recovery on the user_settings table"
  type        = bool
  default     = false
}

variable "default_rule_severity" {
  description = "The severity level to use for the default rules (not the severity to use as a default for custom rules). Should be 'info', 'warning', or 'critical'."
  type        = string
  default     = "warning"
  validation {
    condition     = contains(["info", "warning", "critical"], var.default_rule_severity)
    error_message = "Invalid severity level. Must be one of: info, warning, critical."
  }
}

variable "enable_pagerduty" {
  description = "Whether to enable PagerDuty notifications for the alerts."
  type        = bool
  default     = false
}

variable "alert_on_no_data" {
  description = "Whether to alert on missing data. Defaults to true."
  type        = bool
  default     = false
}
