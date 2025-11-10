data "aws_iam_policy_document" "dynamo_table_r_data_export_policy" {
  statement {
    sid    = "AllowDynamoDBERxportAction"
    effect = "Allow"
    actions = [
      "dynamodb:Batch*",
      "dynamodb:Get*",
      "dynamodb:Query",
      "dynamodb:Scan",
      "dynamodb:ExportTableToPointInTime",
      "dynamodb:DescribeTable",
      "dynamodb:ListTables",
      "dynamodb:ListBackups",
      "dynamodb:DescribeBackup",
      "dynamodb:ListExports"
    ]
    resources = [
      aws_dynamodb_table.user_settings.arn,
      "${aws_dynamodb_table.user_settings.arn}/*",
    ]
  }
}

# This allows gcp to assume the given role
data "aws_iam_policy_document" "gcp_data_export_policy" {
  statement {
    sid    = "AllowAssumeRoleGCP"
    effect = "Allow"

    principals {
      type        = "Federated"
      identifiers = ["accounts.google.com"]
    }

    actions = [
      "sts:AssumeRoleWithWebIdentity",
    ]

    condition {
      test     = "StringEquals"
      variable = "accounts.google.com:sub"
      values   = var.gcp_data_accounts
    }
  }

  # This allows the RDS export service
  # to 'impersonate' the role when exporting
  statement {
    sid    = "AllowAssumeRdsExport"
    effect = "Allow"

    actions = [
      "sts:AssumeRole",
    ]

    principals {
      type        = "Service"
      identifiers = ["export.rds.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "user_settings_data_export_role" {
  name               = "${var.system}-data-export-${local.aws_region}-${local.environment}"
  assume_role_policy = data.aws_iam_policy_document.gcp_data_export_policy.json
}

resource "aws_iam_role_policy" "dynamo_export_data_policy" {
  name   = "dynamo_table_r_data_export_policy"
  role   = aws_iam_role.user_settings_data_export_role.id
  policy = data.aws_iam_policy_document.dynamo_table_r_data_export_policy.json
}
