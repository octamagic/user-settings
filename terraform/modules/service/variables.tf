variable "name" {
  description = "Service name. Resource names are <name>-<environment>."
  type        = string
  default     = "user-settings"
}

variable "environment" {
  description = "Deployment environment: dev, staging or production."
  type        = string

  validation {
    condition     = contains(["dev", "staging", "production"], var.environment)
    error_message = "environment must be dev, staging or production."
  }
}

# ── network ────────────────────────────────────────────────────────────────

variable "vpc_id" {
  description = "VPC the service and its load balancer live in."
  type        = string
}

variable "private_subnet_ids" {
  description = "Subnets for the Fargate tasks. Private; tasks never get public IPs."
  type        = list(string)
}

variable "alb_subnet_ids" {
  description = "Subnets for the load balancer. Private for an internal ALB, public for an internet-facing one."
  type        = list(string)
}

variable "alb_internal" {
  description = "Create an internal load balancer. The service is a backend API, so internal is the default."
  type        = bool
  default     = true
}

variable "alb_ingress_cidr_blocks" {
  description = "CIDRs allowed to reach the load balancer."
  type        = list(string)
  default     = ["10.0.0.0/8"]
}

variable "certificate_arn" {
  description = "ACM certificate for HTTPS on the load balancer. null serves plain HTTP only, which is acceptable for an internal ALB inside the VPC."
  type        = string
  default     = null
}

variable "alb_deletion_protection" {
  description = "Protect the load balancer from deletion."
  type        = bool
  default     = false
}

# ── compute ────────────────────────────────────────────────────────────────

variable "image_tag" {
  description = "Image tag to run. Deploys roll new tags by registering task definition revisions; this value only matters on first apply and when Terraform re-registers the task definition."
  type        = string
  default     = "bootstrap"
}

variable "cpu_architecture" {
  description = "X86_64 or ARM64. Must match how the image was built."
  type        = string
  default     = "X86_64"
}

variable "cpu" {
  description = "Task CPU units."
  type        = string
  default     = "256"
}

variable "memory" {
  description = "Task memory in MiB."
  type        = string
  default     = "512"
}

variable "container_port" {
  description = "Port the service listens on."
  type        = number
  default     = 8000
}

variable "desired_count" {
  description = "Initial task count. Autoscaling adjusts it between min_count and max_count afterwards."
  type        = number
  default     = 1
}

variable "min_count" {
  description = "Autoscaling lower bound."
  type        = number
  default     = 1
}

variable "max_count" {
  description = "Autoscaling upper bound."
  type        = number
  default     = 2
}

variable "cpu_target_percent" {
  description = "Average CPU utilisation the autoscaler holds the service at."
  type        = number
  default     = 60
}

variable "log_retention_days" {
  description = "CloudWatch log retention."
  type        = number
  default     = 30
}

# ── application configuration ──────────────────────────────────────────────

variable "user_directory_base_url" {
  description = "Base URL of the upstream user directory (USER_DIRECTORY_BASE_URL)."
  type        = string
}

variable "log_level" {
  description = "LOG_LEVEL for the service."
  type        = string
  default     = "info"
}

variable "otel_exporter_otlp_endpoint" {
  description = "OTLP endpoint for traces. null disables the OpenTelemetry SDK."
  type        = string
  default     = null
}

variable "extra_environment" {
  description = "Additional environment variables for the container. Never put secrets here; use SSM or Secrets Manager references."
  type        = map(string)
  default     = {}
}

# ── data ───────────────────────────────────────────────────────────────────

variable "table_point_in_time_recovery" {
  description = "Enable PITR on the settings table."
  type        = bool
  default     = true
}

variable "table_deletion_protection" {
  description = "Protect the settings table from deletion."
  type        = bool
  default     = true
}

# ── CI identity ────────────────────────────────────────────────────────────

variable "github_repository" {
  description = "owner/repo allowed to assume the deploy role via GitHub OIDC. null skips creating the role."
  type        = string
  default     = null
}

variable "github_oidc_subjects" {
  description = "Token subjects allowed to assume the deploy role. Defaults to pushes to main and any GitHub Environment of github_repository."
  type        = list(string)
  default     = null
}

variable "tags" {
  description = "Tags applied to every resource."
  type        = map(string)
  default     = {}
}
