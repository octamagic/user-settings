terraform {
  # Root configs pin the provider and commit .terraform.lock.hcl; the modules
  # they call use permissive constraints.
  required_version = ">= 1.9.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
  }
}
