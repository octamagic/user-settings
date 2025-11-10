module "api_service" {
  source       = "github.com/soundcloud/tf-common//modules/aws-ecs-service?ref=aws-ecs-service@3.38.0"
  service_name = var.system
  aws_region   = local.aws_region
  aws_profile  = local.aws_profile
  cluster_name = local.ecs_cluster
  environment  = local.environment
  vpc_id       = data.aws_vpc.sc.id
  cpu          = var.cpu
  memory       = var.memory
  subnets      = data.aws_subnets.private.ids
  environment_variables = merge(local.telemetry_config, {
    APP_NAME                 = var.system
    LOG_LEVEL                = local.log_level
    ENVIRONMENT              = local.environment
    PORT                     = var.container_port
    USER_SETTINGS_TABLE_NAME = var.user_settings_table_name
    OKIDOKI_BASE_URL         = "http://http.okidoki.prod.moshimoshi.srv.db.s-cloud.net"
  })
  docker_image_url   = local.container_image
  ecs_min_tasks      = var.ecs_min_tasks
  ecs_desired_tasks  = var.ecs_desired_tasks
  ecs_max_tasks      = var.ecs_max_tasks
  scaleup_cooldown   = var.scaleup_cooldown
  scaledown_cooldown = var.scaledown_cooldown
  capacity_provider_strategy = {
    FARGATE      = local.on_demand_capacity
    FARGATE_SPOT = local.spot_capacity
  }
  minimum_non_spot_tasks = 1
  target_tracking_policy = {
    target_value           = 40
    scaleup_cooldown       = 60
    scaledown_cooldown     = 60
    predefined_metric_type = "ECSServiceAverageCPUUtilization"
    resource_label         = null
  }
  target_tracking_policy_enabled = true
  use_docker_healthcheck         = false
  virtual_node_healthcheck_config = {
    healthy_threshold   = 3
    unhealthy_threshold = 2
    interval            = 5
    timeout             = 2
    path                = var.api_healthcheck_path
  }
  enable_execute_command = true
  enable_app_mesh        = true
  mesh_name              = local.short_env_name
  debug_app_mesh         = local.environment == "development"
  ingress_cidr_ranges    = ["10.0.0.0/8"]
  mesh_backends = [
    "http.okidoki.prod.moshimoshi.srv.db.s-cloud.net"
  ]
  policies                          = []
  container_port                    = var.container_port
  envoy_listener_max_retries        = 5
  envoy_image_version               = "v1.29.6.1-prod"
  container_secrets_manager_secrets = []
  enable_otel_collector_sidecar     = true
  enable_ecs_container_metrics      = false
  tags = merge(local.tags, {
    Component = "api"
  })
}
