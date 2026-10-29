# Infrastructure

Terraform for running the service on AWS: ECR, a DynamoDB table, an ECS Fargate
service behind an internal ALB with CPU autoscaling, CloudWatch logs, and a GitHub
OIDC role the deploy workflow assumes. Everything is composed from modules in
[tf-tools](https://github.com/octamagic/tf-tools), pinned to a release tag.

```
terraform/
  modules/service/        one composition: all of the above for one environment
    tests/                terraform test suite against a mocked provider
  environments/
    dev/ staging/ production/   roots: sizing differs, inputs come from variables
```

## Layout rules

- `modules/service` owns the shape. Environments only set sizing and supply
  account-specific inputs (VPC, subnets, upstream URL, certificate). A diff between
  two environment `main.tf` files is a diff between the environments.
- No account IDs, hostnames or ARNs are committed. The state bucket is supplied at
  `init`; network inputs come from a gitignored `environment.auto.tfvars` locally and
  the `TF_SECRET_VARS` secret in CI. Examples sit next to each root.
- tf-tools sources are pinned to one tag everywhere. Bump every `?ref=` together.

## Local workflow

```bash
make tf-fmt
make tf-lint                      # tflint
make tf-test                      # validate every root + terraform test, against a
                                  # local tf-tools checkout (TF_TOOLS_DIR, default ~/code/tf-tools)

cd terraform/environments/dev
cp backend.hcl.example backend.hcl && cp environment.auto.tfvars.example environment.auto.tfvars
terraform init -backend-config=backend.hcl
terraform plan
```

`make tf-test` rewrites the git sources to the local checkout so a tf-tools change can
be exercised before it is released. CI uses the pinned tag.

## Deploys and drift

CI deploys by registering a new task-definition revision with the new image digest
(actions-toolkit `deploy-aws-ecs`). The service module sets
`track_latest_task_definition = true`, so Terraform follows that revision instead of
rolling it back on the next apply, while a revision Terraform registers itself (an env
var change, say) still wins because it is newer. The `image_tag` variable therefore only
matters on the very first apply.

## Bootstrap, once per account

1. An S3 bucket for state, and the GitHub OIDC provider
   (`token.actions.githubusercontent.com`) in IAM.
2. A role for Terraform itself with rights to manage this stack, trusted by this
   repository through OIDC. Its ARN is the `AWS_TERRAFORM_ROLE_ARN` repository variable.
3. First `plan`/`apply` from a laptop, or through the Terraform workflow once the
   variables in `.github/workflows/terraform.yml` are set.
4. Copy each environment's `ci_role_arn` output into `AWS_DEPLOY_ROLE_ARN_<ENV>`; the
   Deploy workflow starts working.

## Pinned version

`?ref=v0.3.0` requires two tf-tools changes that are on a branch at the time of
writing: the `aws/dynamodb-table` module and `track_latest_task_definition` on
`aws/ecs-service`. Until that tag exists, `make tf-test` against the local checkout is
the way to validate; CI's `checks` job will fail on `init` until the tag is cut.
