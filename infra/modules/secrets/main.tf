resource "aws_secretsmanager_secret" "gemini" {
  name                    = "lolidle/${var.environment}/gemini-api-key"
  recovery_window_in_days = 0 # Academy: allow immediate recreation if needed
}

resource "aws_secretsmanager_secret_version" "gemini" {
  secret_id     = aws_secretsmanager_secret.gemini.id
  secret_string = var.gemini_api_key
}
