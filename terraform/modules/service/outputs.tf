output "service_url" {
  description = "Base URL of the service behind the load balancer."
  value       = "${var.certificate_arn != null ? "https" : "http"}://${module.alb.dns_name}"
}

output "ecr_repository_url" {
  description = "Where CI pushes images."
  value       = module.ecr.repository_url
}

output "table_name" {
  description = "Settings table name (USER_SETTINGS_TABLE_NAME)."
  value       = module.table.name
}

output "cluster_name" {
  description = "ECS cluster name, for the deploy workflow."
  value       = module.cluster.name
}

output "service_name" {
  description = "ECS service name, for the deploy workflow."
  value       = module.service.service_name
}

output "log_group_name" {
  description = "CloudWatch log group for the service."
  value       = aws_cloudwatch_log_group.this.name
}

output "ci_role_arn" {
  description = "Role GitHub Actions assumes to deploy; set it as the AWS_DEPLOY_ROLE_ARN repository variable. null when github_repository is unset."
  value       = one(module.ci_role[*].arn)
}
