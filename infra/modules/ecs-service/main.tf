resource "aws_ecs_cluster" "this" {
  name = "lolidle-${var.environment}-cluster"
}

resource "aws_security_group" "tasks" {
  name        = "lolidle-${var.environment}-tasks-sg"
  description = "Allow ALB to reach Fargate tasks"
  vpc_id      = var.vpc_id

  ingress {
    from_port       = 8080
    to_port         = 8080
    protocol        = "tcp"
    security_groups = [var.alb_sg_id]
  }
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# In Academy, use the LabRole as the execution role
data "aws_iam_role" "lab_role" {
  name = "LabRole"
}

resource "aws_ecs_task_definition" "this" {
  family                   = "lolidle-${var.environment}"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = data.aws_iam_role.lab_role.arn
  task_role_arn            = data.aws_iam_role.lab_role.arn

  container_definitions = jsonencode([{
    name      = "lolidle-backend"
    image     = var.image_uri
    essential = true
    portMappings = [{
      containerPort = 8080
      protocol      = "tcp"
    }]
    environment = [
      { name = "PORT", value = "8080" },
      { name = "STORE_BACKEND", value = "dynamodb" },
      { name = "AWS_REGION", value = "us-east-1" },
      { name = "SESSIONS_TABLE", value = var.sessions_table_name },
      { name = "LORE_CACHE_TABLE", value = var.lore_cache_table_name },
      { name = "CORS_ORIGIN", value = var.cors_origin },
      { name = "ENV", value = var.environment },
    ]
    secrets = [
      { name = "GEMINI_API_KEY", valueFrom = var.gemini_secret_arn },
    ]
    logConfiguration = {
      logDriver = "awslogs"
      options = {
        awslogs-group         = var.log_group_name
        awslogs-region        = "us-east-1"
        awslogs-stream-prefix = "ecs"
      }
    }
  }])
}

resource "aws_ecs_service" "blue" {
  name                              = "lolidle-${var.environment}-blue"
  cluster                           = aws_ecs_cluster.this.id
  task_definition                   = aws_ecs_task_definition.this.arn
  desired_count                     = 2
  launch_type                       = "FARGATE"
  health_check_grace_period_seconds = 60

  network_configuration {
    subnets          = var.subnet_ids
    security_groups  = [aws_security_group.tasks.id]
    assign_public_ip = true
  }

  load_balancer {
    target_group_arn = var.tg_blue_arn
    container_name   = "lolidle-backend"
    container_port   = 8080
  }

  deployment_circuit_breaker {
    enable   = true
    rollback = true
  }

  lifecycle {
    ignore_changes = [task_definition, desired_count] # pipeline manages these
  }
}

resource "aws_ecs_service" "green" {
  name                              = "lolidle-${var.environment}-green"
  cluster                           = aws_ecs_cluster.this.id
  task_definition                   = aws_ecs_task_definition.this.arn
  desired_count                     = 0 # green starts off
  launch_type                       = "FARGATE"
  health_check_grace_period_seconds = 60

  network_configuration {
    subnets          = var.subnet_ids
    security_groups  = [aws_security_group.tasks.id]
    assign_public_ip = true
  }

  load_balancer {
    target_group_arn = var.tg_green_arn
    container_name   = "lolidle-backend"
    container_port   = 8080
  }

  deployment_circuit_breaker {
    enable   = true
    rollback = true
  }

  lifecycle {
    ignore_changes = [task_definition, desired_count]
  }
}
