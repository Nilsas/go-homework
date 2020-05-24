provider "aws" {
  region = "eu-west-3"
}

resource "aws_iam_role" "iam_for_lambda" {
  name = "iam_for_lambda"

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

resource "aws_lambda_function" "func" {
  function_name = "go-homework"
  handler       = "main"
  role          = aws_iam_role.iam_for_lambda.arn
  runtime       = "go1.x"
  filename      = "main.zip"
}

