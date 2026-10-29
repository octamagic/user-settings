#!/usr/bin/env bash
# Validate and test the Terraform against a local tf-tools checkout instead of the
# pinned release tag. Useful while a tf-tools change this repo needs is unreleased.
#
#   TF_TOOLS_DIR=~/code/tf-tools scripts/tf-local.sh
#
# Copies terraform/ to a temp dir, rewrites every tf-tools git source to a path
# under $TF_TOOLS_DIR, then runs init/validate on each root and `terraform test`
# on the composition module. Prints the temp dir so lock files can be copied back.
set -euo pipefail

TF_TOOLS_DIR="${TF_TOOLS_DIR:-$HOME/code/tf-tools}"
[ -d "$TF_TOOLS_DIR/modules/aws" ] || { echo "TF_TOOLS_DIR=$TF_TOOLS_DIR has no modules/aws" >&2; exit 1; }

work="$(mktemp -d)"
cp -R terraform "$work/"
find "$work/terraform" -name '*.tf' -exec sed -i.bak -E \
  "s#git::https://github.com/octamagic/tf-tools.git//modules/aws/([a-z-]+)\?ref=[^\"]+#${TF_TOOLS_DIR}/modules/aws/\1#g" {} +
find "$work/terraform" -name '*.bak' -delete

for dir in "$work"/terraform/modules/service "$work"/terraform/environments/*; do
  echo "==> validate ${dir#"$work"/}"
  (cd "$dir" && terraform init -backend=false -input=false -no-color >/dev/null && terraform validate -no-color)
done

echo "==> test terraform/modules/service"
(cd "$work/terraform/modules/service" && terraform test -no-color)

echo "work dir: $work"
