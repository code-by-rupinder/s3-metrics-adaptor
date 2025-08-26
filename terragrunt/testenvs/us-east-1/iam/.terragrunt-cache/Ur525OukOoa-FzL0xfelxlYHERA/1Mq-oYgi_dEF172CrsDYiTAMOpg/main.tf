# IAM Role Module
resource "aws_iam_role" "this" {
  name               = var.role_name
  assume_role_policy = var.assume_role_policy
  description        = var.description
  tags               = var.tags
}

resource "aws_iam_role_policy" "inline" {
  count  = length(var.inline_policies)
  name   = var.inline_policies[count.index].name
  role   = aws_iam_role.this.id
  policy = var.inline_policies[count.index].policy
}

resource "aws_iam_role_policy_attachment" "managed" {
  count      = length(var.managed_policy_arns)
  role       = aws_iam_role.this.name
  policy_arn = var.managed_policy_arns[count.index]
}
