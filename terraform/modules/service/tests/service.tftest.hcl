# Composition tests for the user-settings service module. Runs against a mocked
# AWS provider: no credentials, nothing created. Assertions stay on values that
# are known at plan time (anything the mock would compute, such as ARNs, is
# unknown until apply). See tf-tools docs/TESTING.md for the same constraint.

mock_provider "aws" {}

variables {
  environment             = "dev"
  vpc_id                  = "vpc-0123456789abcdef0"
  private_subnet_ids      = ["subnet-aaaaaaaa", "subnet-bbbbbbbb"]
  alb_subnet_ids          = ["subnet-cccccccc", "subnet-dddddddd"]
  user_directory_base_url = "http://user-directory.internal:8080"
}

run "container_environment_matches_the_service_config" {
  command = plan

  assert {
    condition = alltrue([
      for key in ["AWS_REGION", "USER_SETTINGS_TABLE_NAME", "USER_DIRECTORY_BASE_URL", "PORT", "HOST", "LOG_LEVEL", "OTEL_SDK_DISABLED"] :
      contains(keys(local.environment), key)
    ])
    error_message = "Every variable internal/config requires must be set on the container."
  }

  assert {
    condition     = local.environment["USER_SETTINGS_TABLE_NAME"] == "user-settings-dev" && local.environment["PORT"] == "8000"
    error_message = "The service must be pointed at the table this module creates, on the port the load balancer targets."
  }
}

run "defaults_are_the_safe_ones" {
  command = plan

  assert {
    condition     = local.container.readonlyRootFilesystem == true
    error_message = "The container must run with a read-only root filesystem; the binary is static and writes nothing."
  }

  assert {
    condition     = local.environment["OTEL_SDK_DISABLED"] == "true"
    error_message = "Without an OTLP endpoint the OpenTelemetry SDK must be disabled rather than left trying to export."
  }

  assert {
    condition     = aws_vpc_security_group_ingress_rule.from_alb.from_port == 8000 && aws_vpc_security_group_ingress_rule.from_alb.to_port == 8000
    error_message = "Tasks accept traffic only on the service port."
  }

  assert {
    condition     = length(module.ci_role) == 0 && length(data.aws_iam_openid_connect_provider.github) == 0
    error_message = "No CI role is created, and no OIDC provider looked up, unless github_repository is set."
  }
}

run "otel_endpoint_enables_the_sdk" {
  command = plan

  variables {
    otel_exporter_otlp_endpoint = "http://otel-collector.internal:4318"
  }

  assert {
    condition     = local.environment["OTEL_SDK_DISABLED"] == "false" && local.environment["OTEL_EXPORTER_OTLP_ENDPOINT"] == "http://otel-collector.internal:4318"
    error_message = "Supplying an OTLP endpoint must enable the SDK and pass the endpoint through."
  }
}

run "github_repository_creates_a_scoped_oidc_role" {
  command = plan

  variables {
    github_repository = "patrickisgreat/user-settings"
  }

  assert {
    condition     = length(module.ci_role) == 1
    error_message = "Setting github_repository must create the deploy role."
  }

  assert {
    condition = contains(
      jsondecode(local.ci_trust_policy).Statement[0].Condition.StringLike["token.actions.githubusercontent.com:sub"],
      "repo:patrickisgreat/user-settings:ref:refs/heads/main",
    )
    error_message = "The trust policy must be scoped to this repository's main branch and environments, never to repo:*."
  }

  assert {
    condition     = jsondecode(local.ci_trust_policy).Statement[0].Condition.StringEquals["token.actions.githubusercontent.com:aud"] == "sts.amazonaws.com"
    error_message = "The trust policy must pin the audience to sts.amazonaws.com."
  }
}

run "rejects_unknown_environment" {
  command = plan
  variables {
    environment = "qa"
  }
  expect_failures = [var.environment]
}
