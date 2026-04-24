terraform {
  required_version = ">= 1.5.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = "us-east-1"
  default_tags {
    tags = {
      Project     = "lolidle"
      Environment = "staging"
      ManagedBy   = "terraform"
    }
  }
}

locals {
  environment = "staging"
}

data "aws_vpc" "default" {
  default = true
}

data "aws_subnets" "default" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default.id]
  }
}

# ECR is shared across envs — dev owns it, staging/prod reference it
data "aws_ecr_repository" "shared" {
  name = "lolidle-backend"
}

module "dynamodb" {
  source      = "../../modules/dynamodb"
  environment = local.environment
}

module "secrets" {
  source         = "../../modules/secrets"
  environment    = local.environment
  gemini_api_key = var.gemini_api_key
}

module "frontend" {
  source      = "../../modules/frontend"
  environment = local.environment
}

module "alb" {
  source      = "../../modules/alb"
  environment = local.environment
  vpc_id      = data.aws_vpc.default.id
  subnet_ids  = data.aws_subnets.default.ids
}

module "observability" {
  source              = "../../modules/observability"
  environment         = local.environment
  alb_arn_suffix      = module.alb.alb_arn_suffix == "" ? "placeholder" : module.alb.alb_arn_suffix
  tg_blue_arn_suffix  = module.alb.tg_blue_arn_suffix
  tg_green_arn_suffix = module.alb.tg_green_arn_suffix
  cluster_name        = "lolidle-${local.environment}-cluster"
  service_blue        = "lolidle-${local.environment}-blue"
  service_green       = "lolidle-${local.environment}-green"
}

module "ecs" {
  source                = "../../modules/ecs-service"
  environment           = local.environment
  vpc_id                = data.aws_vpc.default.id
  subnet_ids            = data.aws_subnets.default.ids
  alb_sg_id             = module.alb.alb_sg_id
  tg_blue_arn           = module.alb.tg_blue_arn
  tg_green_arn          = module.alb.tg_green_arn
  image_uri             = "${data.aws_ecr_repository.shared.repository_url}:${var.image_tag}"
  sessions_table_arn    = module.dynamodb.sessions_table_arn
  sessions_table_name   = module.dynamodb.sessions_table_name
  lore_cache_table_arn  = module.dynamodb.lore_cache_table_arn
  lore_cache_table_name = module.dynamodb.lore_cache_table_name
  gemini_secret_arn     = module.secrets.gemini_secret_arn
  cors_origin           = module.frontend.cloudfront_url
  log_group_name        = module.observability.log_group_name
}
