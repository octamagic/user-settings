# Everything the user-settings service needs in one AWS account and region,
# composed from github.com/patrickisgreat/tf-tools modules. Each environment root
# under ../../environments instantiates this once.
#
# Module sources are pinned to a tf-tools release tag. Bump every ref together.

locals {
  prefix         = "${var.name}-${var.environment}"
  container_name = "app"
  otel_enabled   = var.otel_exporter_otlp_endpoint != null

  tags = merge(var.tags, {
    Service     = var.name
    Environment = var.environment
  })

  # The container's environment mirrors config/.env.* and internal/config.
  environment = merge(
    {
      APP_NAME                 = var.name
      ENVIRONMENT              = var.environment
      HOST                     = "0.0.0.0"
      PORT                     = tostring(var.container_port)
      LOG_LEVEL                = var.log_level
      LOG_PRETTY               = "false"
      AWS_REGION               = data.aws_region.current.region
      USER_SETTINGS_TABLE_NAME = module.table.name
      USER_DIRECTORY_BASE_URL  = var.user_directory_base_url
      OTEL_SDK_DISABLED        = local.otel_enabled ? "false" : "true"
    },
    local.otel_enabled ? {
      OTEL_EXPORTER_OTLP_ENDPOINT = var.otel_exporter_otlp_endpoint
      OTEL_SERVICE_NAME           = local.prefix
    } : {},
    var.extra_environment,
  )

  create_ci_role = var.github_repository != null
  github_oidc_subjects = coalesce(
    var.github_oidc_subjects,
    local.create_ci_role ? [
      "repo:${var.github_repository}:ref:refs/heads/main",
      "repo:${var.github_repository}:environment:*",
    ] : [],
  )
}

data "aws_region" "current" {}

# ── image registry ─────────────────────────────────────────────────────────

module "ecr" {
  source = "git::https://github.com/patrickisgreat/tf-tools.git//modules/aws/ecr?ref=v0.3.0"

  name = var.name
  tags = local.tags
}

# ── data ───────────────────────────────────────────────────────────────────

# One table: pk = user URN, sk = setting name. See docs/architecture.md.
module "table" {
  source = "git::https://github.com/patrickisgreat/tf-tools.git//modules/aws/dynamodb-table?ref=v0.3.0"

  name      = local.prefix
  hash_key  = { name = "pk", type = "S" }
  range_key = { name = "sk", type = "S" }

  point_in_time_recovery_enabled = var.table_point_in_time_recovery
  deletion_protection_enabled    = var.table_deletion_protection
  tags                           = local.tags
}

# ── compute ────────────────────────────────────────────────────────────────

resource "aws_cloudwatch_log_group" "this" {
  name              = "/ecs/${local.prefix}"
  retention_in_days = var.log_retention_days
  tags              = local.tags
}

module "cluster" {
  source = "git::https://github.com/patrickisgreat/tf-tools.git//modules/aws/ecs-cluster?ref=v0.3.0"

  name = local.prefix
  tags = local.tags
}

module "alb" {
  source = "git::https://github.com/patrickisgreat/tf-tools.git//modules/aws/alb?ref=v0.3.0"

  name       = local.prefix
  vpc_id     = var.vpc_id
  subnet_ids = var.alb_subnet_ids

  internal            = var.alb_internal
  ingress_cidr_blocks = var.alb_ingress_cidr_blocks
  certificate_arn     = var.certificate_arn

  target_port  = var.container_port
  health_check = { path = "/-/health" }

  enable_deletion_protection = var.alb_deletion_protection
  tags                       = local.tags
}

resource "aws_security_group" "tasks" {
  name        = "${local.prefix}-tasks"
  description = "Fargate tasks for ${local.prefix}; ingress only from the load balancer."
  vpc_id      = var.vpc_id
  tags        = local.tags
}

resource "aws_vpc_security_group_ingress_rule" "from_alb" {
  security_group_id            = aws_security_group.tasks.id
  referenced_security_group_id = module.alb.security_group_id
  from_port                    = var.container_port
  to_port                      = var.container_port
  ip_protocol                  = "tcp"
  description                  = "Load balancer to service port"
}

resource "aws_vpc_security_group_egress_rule" "all" {
  security_group_id = aws_security_group.tasks.id
  ip_protocol       = "-1"
  cidr_ipv4         = "0.0.0.0/0"
  description       = "DynamoDB, the user directory and the OTLP endpoint"
}

module "execution_role" {
  source = "git::https://github.com/patrickisgreat/tf-tools.git//modules/aws/iam-role?ref=v0.3.0"

  name                  = "${local.prefix}-execution"
  description           = "Pulls the image and writes logs for ${local.prefix}."
  trusted_role_services = ["ecs-tasks.amazonaws.com"]
  managed_policy_arns   = ["arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"]
  tags                  = local.tags
}

data "aws_iam_policy_document" "task" {
  statement {
    sid = "SettingsTable"
    actions = [
      "dynamodb:GetItem",
      "dynamodb:BatchGetItem",
      "dynamodb:Query",
      "dynamodb:PutItem",
      "dynamodb:UpdateItem",
      "dynamodb:DeleteItem",
    ]
    resources = [module.table.arn, "${module.table.arn}/index/*"]
  }
}

module "task_role" {
  source = "git::https://github.com/patrickisgreat/tf-tools.git//modules/aws/iam-role?ref=v0.3.0"

  name                  = "${local.prefix}-task"
  description           = "Runtime permissions for ${local.prefix}: the settings table only."
  trusted_role_services = ["ecs-tasks.amazonaws.com"]
  inline_policies       = { settings-table = data.aws_iam_policy_document.task.json }
  tags                  = local.tags
}

locals {
  # Kept as an object (not just the encoded string) so tests can assert on the
  # fields that are known at plan time.
  container = {
    name                   = local.container_name
    image                  = "${module.ecr.repository_url}:${var.image_tag}"
    essential              = true
    readonlyRootFilesystem = true
    portMappings           = [{ containerPort = var.container_port, protocol = "tcp" }]
    environment            = [for k, v in local.environment : { name = k, value = v }]
    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.this.name
        "awslogs-region"        = data.aws_region.current.region
        "awslogs-stream-prefix" = local.container_name
      }
    }
  }

  container_definitions = jsonencode([local.container])
}

module "service" {
  source = "git::https://github.com/patrickisgreat/tf-tools.git//modules/aws/ecs-service?ref=v0.3.0"

  name        = local.prefix
  cluster_arn = module.cluster.arn

  cpu              = var.cpu
  memory           = var.memory
  cpu_architecture = var.cpu_architecture

  execution_role_arn = module.execution_role.arn
  task_role_arn      = module.task_role.arn

  container_definitions = local.container_definitions

  subnet_ids         = var.private_subnet_ids
  security_group_ids = [aws_security_group.tasks.id]

  target_group_arn  = module.alb.target_group_arn
  lb_container_name = local.container_name
  lb_container_port = var.container_port

  desired_count = var.desired_count

  # CI rolls images by registering new revisions (actions-toolkit deploy-aws-ecs).
  # Follow the newest one instead of reverting it on the next apply.
  track_latest_task_definition = true

  tags = local.tags
}

module "autoscaling" {
  source = "git::https://github.com/patrickisgreat/tf-tools.git//modules/aws/ecs-autoscaling?ref=v0.3.0"

  cluster_name = module.cluster.name
  service_name = module.service.service_name

  min_capacity       = var.min_count
  max_capacity       = var.max_count
  cpu_target_percent = var.cpu_target_percent
}

# ── CI deploy identity (GitHub OIDC, no static keys) ───────────────────────

data "aws_iam_openid_connect_provider" "github" {
  count = local.create_ci_role ? 1 : 0

  url = "https://token.actions.githubusercontent.com"
}

# Built with jsonencode rather than aws_iam_policy_document so the test suite can
# decode and inspect it under a mocked provider.
locals {
  ci_trust_policy = local.create_ci_role ? jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Action    = "sts:AssumeRoleWithWebIdentity"
      Principal = { Federated = data.aws_iam_openid_connect_provider.github[0].arn }
      Condition = {
        StringEquals = { "token.actions.githubusercontent.com:aud" = "sts.amazonaws.com" }
        StringLike   = { "token.actions.githubusercontent.com:sub" = local.github_oidc_subjects }
      }
    }]
  }) : null
}

data "aws_iam_policy_document" "ci_deploy" {
  count = local.create_ci_role ? 1 : 0

  statement {
    sid       = "EcrLogin"
    actions   = ["ecr:GetAuthorizationToken"]
    resources = ["*"]
  }
  statement {
    sid = "EcrPush"
    actions = [
      "ecr:BatchCheckLayerAvailability",
      "ecr:BatchGetImage",
      "ecr:GetDownloadUrlForLayer",
      "ecr:InitiateLayerUpload",
      "ecr:UploadLayerPart",
      "ecr:CompleteLayerUpload",
      "ecr:PutImage",
    ]
    resources = [module.ecr.repository_arn]
  }
  statement {
    sid = "EcsDescribeAndRegister"
    actions = [
      "ecs:DescribeServices",
      "ecs:DescribeTaskDefinition",
      "ecs:RegisterTaskDefinition",
    ]
    resources = ["*"]
  }
  statement {
    sid       = "EcsRollService"
    actions   = ["ecs:UpdateService"]
    resources = [module.service.service_id]
  }
  statement {
    sid       = "PassTaskRoles"
    actions   = ["iam:PassRole"]
    resources = [module.execution_role.arn, module.task_role.arn]
    condition {
      test     = "StringEquals"
      variable = "iam:PassedToService"
      values   = ["ecs-tasks.amazonaws.com"]
    }
  }
}

module "ci_role" {
  count  = local.create_ci_role ? 1 : 0
  source = "git::https://github.com/patrickisgreat/tf-tools.git//modules/aws/iam-role?ref=v0.3.0"

  name               = "${local.prefix}-github-deploy"
  description        = "Assumed by GitHub Actions in ${var.github_repository} to push images and roll ${local.prefix}."
  assume_role_policy = local.ci_trust_policy
  inline_policies    = { deploy = data.aws_iam_policy_document.ci_deploy[0].json }
  tags               = local.tags
}
