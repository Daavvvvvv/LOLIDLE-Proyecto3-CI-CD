output "log_group_name" {
  value = aws_cloudwatch_log_group.ecs.name
}

output "alarm_5xx_name" {
  value = aws_cloudwatch_metric_alarm.target_5xx_high.alarm_name
}

output "alarm_latency_name" {
  value = aws_cloudwatch_metric_alarm.p95_latency_high.alarm_name
}

output "dashboard_url" {
  value = "https://us-east-1.console.aws.amazon.com/cloudwatch/home?region=us-east-1#dashboards:name=${aws_cloudwatch_dashboard.main.dashboard_name}"
}
