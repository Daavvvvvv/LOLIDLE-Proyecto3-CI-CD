output "alb_url" {
  value = "http://${module.alb.alb_dns_name}"
}

output "frontend_url" {
  value = module.frontend.cloudfront_url
}

output "frontend_bucket" {
  value = module.frontend.bucket_name
}

output "cf_distribution_id" {
  value = module.frontend.cloudfront_distribution_id
}

output "ecr_repository" {
  value = data.aws_ecr_repository.shared.repository_url
}

output "dashboard_url" {
  value = module.observability.dashboard_url
}

output "cluster_name" {
  value = module.ecs.cluster_name
}

output "service_blue" {
  value = module.ecs.service_blue
}

output "service_green" {
  value = module.ecs.service_green
}

output "listener_arn" {
  value = module.alb.listener_arn
}

output "tg_blue_arn" {
  value = module.alb.tg_blue_arn
}

output "tg_green_arn" {
  value = module.alb.tg_green_arn
}

output "sessions_table" {
  value = module.dynamodb.sessions_table_name
}

output "lore_cache_table" {
  value = module.dynamodb.lore_cache_table_name
}
