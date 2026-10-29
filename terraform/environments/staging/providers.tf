provider "aws" {
  region = var.aws_region

  default_tags {
    tags = merge(
      {
        Service     = "user-settings"
        Environment = "staging"
        ManagedBy   = "terraform"
        Repo        = "github.com/octamagic/user-settings"
      },
      var.default_tags,
    )
  }
}
