# Build the archive file to send to lambda
data "archive_file" "pkg" {
  type        = "zip"
  source_file = "pong"
  output_path = "pong_lambda_pkg.zip"
}

resource "aws_iam_role" "lambda_role" {
  name               = "pong_lambda_role"
  assume_role_policy = "${file("lambda_assumerolepolicy.json")}"
}

resource "aws_lambda_function" "lambda_pong" {
  filename         = "${data.archive_file.pkg.output_path}"
  function_name    = "pong"
  role             = "${aws_iam_role.lambda_role.arn}"
  handler          = "${data.archive_file.pkg.source_file}"
  source_code_hash = "${base64sha256(file("${data.archive_file.pkg.output_path}"))}"
  runtime          = "go1.x"

  environment {
    variables = {
      Environment = "staging"
    }
  }
}
