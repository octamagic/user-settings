resource "aws_dynamodb_table" "user_settings" {
  name             = var.system
  hash_key         = local.user_settings_table.hash_key
  range_key        = local.user_settings_table.range_key
  billing_mode     = local.user_settings_table.billing_mode
  read_capacity    = local.user_settings_table.read_capacity
  write_capacity   = local.user_settings_table.write_capacity
  stream_enabled   = local.user_settings_table.stream_enabled
  stream_view_type = local.user_settings_table.stream_view_type

  point_in_time_recovery {
    enabled = local.user_settings_table.point_in_time_recovery
  }

  attribute {
    name = local.user_settings_table.hash_key
    type = "S"
  }
  attribute {
    name = local.user_settings_table.range_key
    type = "S"
  }

  tags = local.tags
}

resource "aws_iam_policy" "dynamo_table_rw" {
  name = "${var.system}-table-rw"
  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow",
        Action = [
          "dynamodb:Batch*",
          "dynamodb:ConditionCheckItem",
          "dynamodb:DeleteItem",
          "dynamodb:Get*",
          "dynamodb:PutItem",
          "dynamodb:Query",
          "dynamodb:Scan",
          "dynamodb:UpdateItem",
        ],
        Resource = [
          aws_dynamodb_table.user_settings.arn,
          "${aws_dynamodb_table.user_settings.arn}/*",
        ]
      },
      {
        Effect = "Allow",
        Action = [
          "dynamodb:List*",
          "dynamodb:Describe*",
        ],
        Resource = "*"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "main" {
  role       = "${var.system}-${local.aws_region}-${local.environment}"
  policy_arn = aws_iam_policy.dynamo_table_rw.arn
}
