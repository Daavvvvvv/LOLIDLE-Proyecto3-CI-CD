resource "aws_dynamodb_table" "sessions" {
  name         = "lolidle-${var.environment}-sessions"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "gameId"

  attribute {
    name = "gameId"
    type = "S"
  }

  ttl {
    attribute_name = "expiresAt"
    enabled        = true
  }
}

resource "aws_dynamodb_table" "lore_cache" {
  name         = "lolidle-${var.environment}-lore-cache"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "championId"

  attribute {
    name = "championId"
    type = "S"
  }
}
