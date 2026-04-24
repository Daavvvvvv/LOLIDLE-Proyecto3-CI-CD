variable "environment" {
  type = string
}

variable "vpc_id" {
  type = string
}

variable "subnet_ids" {
  type = list(string)
}

variable "alb_sg_id" {
  type = string
}

variable "tg_blue_arn" {
  type = string
}

variable "tg_green_arn" {
  type = string
}

variable "image_uri" {
  type = string
}

variable "sessions_table_arn" {
  type = string
}

variable "sessions_table_name" {
  type = string
}

variable "lore_cache_table_arn" {
  type = string
}

variable "lore_cache_table_name" {
  type = string
}

variable "gemini_secret_arn" {
  type = string
}

variable "cors_origin" {
  type = string
}

variable "log_group_name" {
  type = string
}
