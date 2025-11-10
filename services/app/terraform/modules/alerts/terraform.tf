terraform {
  required_version = "~> 1.5"

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

provider "grafana" {}
