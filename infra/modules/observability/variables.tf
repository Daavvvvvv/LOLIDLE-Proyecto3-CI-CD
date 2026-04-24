variable "environment" {
  type = string
}

variable "alb_arn_suffix" {
  type        = string
  description = "Last part of ALB ARN, used by CloudWatch metrics"
}

variable "tg_blue_arn_suffix" {
  type = string
}

variable "tg_green_arn_suffix" {
  type = string
}

variable "cluster_name" {
  type = string
}

variable "service_blue" {
  type = string
}

variable "service_green" {
  type = string
}
