output "alb_dns_name" {
  value = aws_lb.this.dns_name
}

output "alb_arn" {
  value = aws_lb.this.arn
}

output "listener_arn" {
  value = aws_lb_listener.http.arn
}

output "tg_blue_arn" {
  value = aws_lb_target_group.blue.arn
}

output "tg_green_arn" {
  value = aws_lb_target_group.green.arn
}

output "alb_sg_id" {
  value = aws_security_group.alb.id
}

output "alb_arn_suffix" {
  value = aws_lb.this.arn_suffix
}

output "tg_blue_arn_suffix" {
  value = aws_lb_target_group.blue.arn_suffix
}

output "tg_green_arn_suffix" {
  value = aws_lb_target_group.green.arn_suffix
}
