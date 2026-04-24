output "cluster_name" {
  value = aws_ecs_cluster.this.name
}

output "service_blue" {
  value = aws_ecs_service.blue.name
}

output "service_green" {
  value = aws_ecs_service.green.name
}

output "task_def_family" {
  value = aws_ecs_task_definition.this.family
}
