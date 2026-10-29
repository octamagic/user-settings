terraform {
  # State lives in S3. The key is fixed here; bucket, region and locking are
  # supplied at init time so no account detail is committed:
  #
  #   terraform init -backend-config=backend.hcl        # locally (gitignored file)
  #   backend-config: ... input of terraform-plan-apply   # in CI, from repository variables
  #
  # See backend.hcl.example.
  backend "s3" {
    key = "user-settings/production/terraform.tfstate"
  }
}
