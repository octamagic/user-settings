variable "error_duration" {
  description = "Set the duration in minutes for the high errors alert"
  type        = number
  default     = 2
}

variable "error_threshold" {
  description = "Set the threshold percentage for the high errors alert"
  type        = string
  default     = "5"
}

variable "latency_duration" {
  description = "Set the duration in minutes for the high response time alert"
  type        = number
  default     = 2
}

variable "latency_threshold" {
  description = "Set the threshold for the high response time alert"
  type        = string
  default     = "3"
}

variable "runbook_url" {
  description = "The user-settings runbook URL"
  type        = string
  default     = "https://eng-doc.soundcloud.org/runbooks/"
}

variable "system" {
  type        = string
  description = "The component that the table belongs to"
  default     = "user-settings"
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
