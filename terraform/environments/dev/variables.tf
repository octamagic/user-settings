variable "aws_region" {
  description = "AWS region for this environment."
  type        = string
  default     = "us-east-1"
}

variable "vpc_id" {
  description = "VPC the service runs in."
  type        = string
}

variable "private_subnet_ids" {
  description = "Private subnets for the Fargate tasks."
  type        = list(string)
}

variable "alb_subnet_ids" {
  description = "Subnets for the internal load balancer."
  type        = list(string)
}

variable "certificate_arn" {
  description = "ACM certificate for HTTPS. null serves HTTP only."
  type        = string
  default     = null
}

variable "user_directory_base_url" {
  description = "Base URL of the upstream user directory."
  type        = string
}

variable "otel_exporter_otlp_endpoint" {
  description = "OTLP endpoint for traces. null disables tracing."
  type        = string
  default     = null
}

variable "image_tag" {
  description = "Image tag for the first apply. Later deploys are rolled by CI."
  type        = string
  default     = "bootstrap"
}

variable "github_repository" {
  description = "owner/repo allowed to assume the deploy role. null skips the role."
  type        = string
  default     = "patrickisgreat/user-settings"
}

variable "default_tags" {
  description = "Extra tags merged into every resource."
  type        = map(string)
  default     = {}
}
