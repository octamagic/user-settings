terraform {
  required_version = "~> 1.5"

  backend "s3" {
    key     = "services/user-settings"
    region  = "eu-central-1"
    encrypt = "true"
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    pagerduty = {
      source  = "pagerduty/pagerduty"
      version = "~> 3.21.0"
    }
    grafana = {
      source  = "grafana/grafana"
      version = "~> 4.10"
    }
  }
}

provider "aws" {
  region = "eu-central-1"
  default_tags {
    tags = local.tags
  }
}

provider "pagerduty" {}

provider "grafana" {}
