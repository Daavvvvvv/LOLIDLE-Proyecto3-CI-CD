output "sessions_table_name" {
  value = aws_dynamodb_table.sessions.name
}
output "sessions_table_arn" {
  value = aws_dynamodb_table.sessions.arn
}
output "lore_cache_table_name" {
  value = aws_dynamodb_table.lore_cache.name
}
output "lore_cache_table_arn" {
  value = aws_dynamodb_table.lore_cache.arn
}
