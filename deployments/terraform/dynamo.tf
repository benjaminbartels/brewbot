resource "aws_dynamodb_table" "brews-table" {
  name           = "BeerBot-Brews"
  billing_mode   = "PROVISIONED"
  read_capacity  = 10
  write_capacity = 10
  hash_key       = "id"

  attribute {
    name = "id"
    type = "S"
  }

  attribute {
    name = "userId"
    type = "S"
  }

  global_secondary_index {
    name            = "byUserId"
    hash_key        = "userId"
    write_capacity  = 5
    read_capacity   = 5
    projection_type = "ALL"
  }
}

resource "aws_iam_user" "brewbot_user" {
  name = "brewbot"
}

resource "aws_iam_user_policy" "brewbot_policy" {
  name   = "brew-bot-policy"
  user   = aws_iam_user.brewbot_user.name
  policy = data.aws_iam_policy_document.brewbot_role_policy.json
}

data "aws_iam_policy_document" "brewbot_role_policy" {
  statement {
    actions = [
      "dynamodb:Scan",
      "dynamodb:Query",
      "dynamodb:GetItem",
      "dynamodb:PutItem",
      "dynamodb:DeleteItem"
    ]
    resources = [
      aws_dynamodb_table.brews-table.arn,
      "${aws_dynamodb_table.brews-table.arn}/index/*",

    ]
  }


}
