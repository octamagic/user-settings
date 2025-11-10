variable "component" {
  type        = string
  description = "The service component name"
  default     = "api"
}

variable "system" {
  type        = string
  description = "The component that the table belongs to"
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

variable "gcp_data_accounts" {
  description = "List of accounts for WIF"
  type        = list(string)
  default = [
    "115045750519978460092",
    "111649564186482509886"
  ]
}
