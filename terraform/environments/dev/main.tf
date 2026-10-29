# dev environment of the user-settings service. Account-specific inputs
# (network, upstream URL, certificate) come from variables; sizing is fixed here
# so a diff between environments is a diff between these files.

module "service" {
  source = "../../modules/service"

  environment = "dev"

  vpc_id             = var.vpc_id
  private_subnet_ids = var.private_subnet_ids
  alb_subnet_ids     = var.alb_subnet_ids
  certificate_arn    = var.certificate_arn

  cpu           = "256"
  memory        = "512"
  desired_count = 1
  min_count     = 1
  max_count     = 2

  log_level          = "debug"
  log_retention_days = 14

  table_point_in_time_recovery = false
  table_deletion_protection    = false
  alb_deletion_protection      = false

  image_tag                   = var.image_tag
  user_directory_base_url     = var.user_directory_base_url
  otel_exporter_otlp_endpoint = var.otel_exporter_otlp_endpoint
  github_repository           = var.github_repository
}

output "service_url" {
  description = "Base URL of the service."
  value       = module.service.service_url
}

output "ecr_repository_url" {
  description = "Image repository CI pushes to."
  value       = module.service.ecr_repository_url
}

output "cluster_name" {
  description = "ECS cluster name for the deploy workflow."
  value       = module.service.cluster_name
}

output "service_name" {
  description = "ECS service name for the deploy workflow."
  value       = module.service.service_name
}

output "ci_role_arn" {
  description = "Set as the AWS_DEPLOY_ROLE_ARN repository variable."
  value       = module.service.ci_role_arn
}
