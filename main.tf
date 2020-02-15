locals {
  name = "${var.db_cluster_name}-lambda-audit-snapshots"
}

resource "aws_iam_role" "this" {
  name = local.name

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_policy" "this" {
  name = local.name
  path = "/"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
     {
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:aws:logs:*:*:*",
      "Effect": "Allow"
     },
     {
      "Effect": "Allow",
      "Action": [
        "rds:CreateDBClusterSnapshot",
        "rds:DescribeDBClusterSnapshots",
        "rds:DescribeDBClusterSnapshotAttributes",
        "rds:ModifyDBClusterSnapshotAttribute",
        "rds:CopyDBClusterSnapshot"
      ],
      "Resource": ["*"]
    },
    {
      "Effect": "Allow",
      "Action": [
        "sns:Publish"
      ],
      "Resource": ["*"]
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "this" {
  role       = aws_iam_role.this.name
  policy_arn = aws_iam_policy.this.arn
}

resource "aws_cloudwatch_log_group" "this" {
  name              = "/aws/lambda/local.name"
  retention_in_days = 14
}

resource "aws_lambda_function" "this" {
  filename      = "${path.module}/lambda.zip"
  function_name = local.name
  role          = aws_iam_role.this.arn
  handler       = "main"

  timeout = "900"

  source_code_hash = filebase64sha256("${path.module}/lambda.zip")

  runtime = "go1.x"

  environment {
    variables = {
      REGION            = var.region
      CLUSTER_ID        = var.db_cluster_name
      TARGET_ACCOUNT_ID = var.target_account
      SNS_TOPIC_ARN     = var.target_sns_topic_arn
    }
  }
}

resource "aws_cloudwatch_event_rule" "trigger" {
  name        = local.name
  description = "Trigger creation of RDS snapshot on schedule"

  schedule_expression = var.lambda_schedule
}

resource "aws_cloudwatch_event_target" "this" {
  rule = aws_cloudwatch_event_rule.trigger.name
  arn  = aws_lambda_function.this.arn
}

resource "aws_lambda_permission" "this" {
  statement_id  = "${aws_lambda_function.this.function_name}-AllowExecutionFromCloudWatch"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.this.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.trigger.arn
}

