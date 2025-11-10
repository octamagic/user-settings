variable "commit_sha" {
  description = "The commit being deployed, used for the docker image tags"
  type        = string
}

variable "component" {
  type        = string
  description = "The service component name"
  default     = "api"
}

variable "cpu" {
  description = "Task CPU limit in units"
  type        = number
  default     = 1024
}

variable "memory" {
  description = "Task memory limit in MB"
  type        = number
  default     = 2048
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

variable "api_healthcheck_path" {
  description = "The route that the lambda should healthcheck on"
  type        = string
  default     = "/-/health"
}

variable "ecs_min_tasks" {
  description = "Minimum number of api containers to run per region"
  type        = number
  default     = 1
}

variable "ecs_desired_tasks" {
  description = "Desired number of api containers to run per region"
  type        = number
  default     = 1
}

variable "ecs_max_tasks" {
  description = "Maximum number of api containers to run per region"
  type        = number
  default     = 3
}

variable "scaleup_cooldown" {
  description = "Time to wait after scale up before scaling up again"
  type        = number
  default     = 60
}

variable "scaledown_cooldown" {
  description = "Time to wait after scale down before scaling down again"
  type        = number
  default     = 300
}

variable "container_port" {
  description = "Container port"
  type        = number
  default     = 8000
}

variable "user_settings_table_name" {
  description = "Table name for user settings"
  type        = string
}
