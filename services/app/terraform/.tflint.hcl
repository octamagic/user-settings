plugin "terraform" {
  enabled = true
  preset  = "recommended"
}

plugin "aws" {
    enabled = true
    deep_check = false
    version = "0.31.0"
    source  = "github.com/terraform-linters/tflint-ruleset-aws"
}
