locals {
  aws_profile = data.aws_caller_identity.current.account_id
  aws_region  = data.aws_region.current.name
  environment = nonsensitive(data.aws_ssm_parameter.environment.value)

  base_account_name = nonsensitive(data.aws_ssm_parameter.base_account_name.value)
  container_image   = "145582369311.dkr.ecr.eu-central-1.amazonaws.com/user-settings:${var.commit_sha}"
  ecs_cluster       = "${local.base_account_name}-${local.aws_region}-${local.environment}"

  log_levels = {
    production  = "error"
    staging     = "info"
    development = "debug"
  }
  log_level = local.log_levels[local.environment]

  spot_capacity      = local.environment == "production" ? 0 : 90
  on_demand_capacity = 100 - local.spot_capacity

  short_env_names = {
    production  = "prod"
    staging     = "stage"
    development = "dev"
  }
  short_env_name = local.short_env_names[local.environment]

  tags = {
    Application = var.system
    GitRepo     = "github.com/soundcloud/user-settings"
    ManagedBy   = "Terraform"
    Environment = local.environment
  }

  telemetry_config = {
    OTEL_EXPORTER_OTLP_PROTOCOL          = "grpc"
    OTEL_GO_X_DEPRECATED_RUNTIME_METRICS = "true"
    OTEL_GO_X_RESOURCE                   = "true"
    OTEL_LOGS_EXPORTER                   = "otlp"
    OTEL_METRICS_EXPORTER                = "otlp"
    OTEL_SC_COMPONENT                    = "api"
    OTEL_SC_ENV                          = local.environment
    OTEL_SC_OWNER                        = "principal-engineers-team"
    OTEL_SC_REPO                         = "soundcloud/user-settings"
    OTEL_SC_SYSTEM                       = var.system
    OTEL_SC_VERSION                      = var.commit_sha
    OTEL_SERVICE_NAME                    = var.system
    OTEL_TRACES_EXPORTER                 = "otlp"
    OTEL_TRACES_SAMPLER                  = "parentbased_traceidratio"
    OTEL_TRACES_SAMPLER_ARG              = local.trace_ratio
  }

  trace_ratios = {
    production  = "0.001"
    staging     = "0.1"
    development = "1"
  }
  trace_ratio = local.trace_ratios[local.environment]
}
