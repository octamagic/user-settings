---
name: tf-bump
description: Bump the pinned tf-tools release or the AWS provider constraint in terraform/ and revalidate. Use when asked to update the tf-tools ref, bump Terraform modules, or change provider versions.
argument-hint: <tf-tools tag, e.g. v0.4.0>
disable-model-invocation: true
---

Move every tf-tools pin to `$ARGUMENTS` and prove the Terraform still validates.

1. Replace every `?ref=` in `terraform/modules/service/main.tf` with `$ARGUMENTS`. They move
   together; a partial bump is a bug.
2. If the provider constraint in `terraform/environments/*/versions.tf` or
   `terraform/modules/service/versions.tf` changes, delete
   `terraform/environments/*/.terraform.lock.hcl` first. A stale lock file makes `init` fail with
   "Failed to query available provider packages".
3. Run `make tf-test`. It validates every root and runs the composition tests against the checkout
   in `TF_TOOLS_DIR` (default `~/code/tf-tools`); point that at the new tag.
4. Copy the regenerated `.terraform.lock.hcl` files back from the work dir the script prints,
   one per environment.
5. Run `make tf-fmt && make tf-lint`.
6. Update the "Pinned version" note in `terraform/README.md`. Commit with a `chore:` prefix
   naming the old and new tags.
