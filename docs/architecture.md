# Lolidle Architecture

## Overview

Lolidle is a single-player champion guessing game (Loldle.net clone) deployed
as a containerized Go backend on ECS Fargate, with a React/Vite frontend
served via S3 static website hosting. Persistence uses DynamoDB for both
game sessions and AI-generated lore cache. The Gemini API powers
post-victory lore descriptions, with the API key stored in AWS Secrets
Manager and injected into the Fargate task as an environment variable.

## Components

```
                          ┌──────────────────────────────┐
                          │  Browser (end user)          │
                          └──────────┬───────────────────┘
                                     │ HTTP GET (HTML/JS/CSS)
                                     ▼
                   ┌──────────────────────────────────────┐
                   │  S3 Static Website                   │
                   │  lolidle-<env>-frontend-<account>    │
                   │  (public read, SPA fallback on 404)  │
                   └──────────────────────────────────────┘
                                     │ XHR /api/*
                                     ▼
                   ┌──────────────────────────────────────┐
                   │  ALB  lolidle-<env>-alb              │
                   │  default listener ─▶ active TG       │
                   │  preview rule (X-Preview) ─▶ green TG│
                   └──────┬─────────────────────┬─────────┘
                          │                     │
               ┌──────────▼─────────┐ ┌─────────▼──────────┐
               │ TG blue (lolidle-  │ │ TG green (lolidle- │
               │ <env>-tg-blue)     │ │ <env>-tg-green)    │
               └──────────┬─────────┘ └─────────┬──────────┘
                          │                     │
               ┌──────────▼─────────┐ ┌─────────▼──────────┐
               │ ECS Service blue   │ │ ECS Service green  │
               │ desired=2 or 0     │ │ desired=2 or 0     │
               └──────────┬─────────┘ └─────────┬──────────┘
                          │                     │
                          └──────────┬──────────┘
                                     │
                          ┌──────────▼────────────┐
                          │ ECS Cluster (Fargate) │
                          │ lolidle-<env>-cluster │
                          └──────────┬────────────┘
                                     │ pulls image
                                     ▼
                          ┌──────────────────────┐
                          │ ECR  lolidle-backend │
                          │ (shared across envs) │
                          └──────────────────────┘

                Fargate task side-dependencies (outbound):
                          │
                          ├─────▶ DynamoDB  lolidle-<env>-sessions     (TTL)
                          ├─────▶ DynamoDB  lolidle-<env>-lore-cache   (TTL)
                          ├─────▶ Secrets Manager  lolidle/<env>/gemini-api-key
                          ├─────▶ Gemini API  (generativelanguage.googleapis.com)
                          └─────▶ CloudWatch Logs  /ecs/lolidle-<env>

                          ┌──────────────────────────────┐
                          │ CloudWatch                   │
                          │  - log group /ecs/lolidle-<env>
                          │  - dashboard  lolidle-<env>  │
                          │  - alarm  target_5xx_high    │
                          │  - alarm  p95_latency_high   │
                          └──────────────────────────────┘
```

| Component             | Service          | Purpose                                                              |
| --------------------- | ---------------- | -------------------------------------------------------------------- |
| Backend container     | ECS Fargate      | Runs the Go API server (`/api/*`)                                    |
| API entry             | ALB              | Routes HTTP, health checks, blue/green via target group swap         |
| Image registry        | ECR              | Stores Docker images tagged by commit SHA + semver (shared all envs) |
| Game state            | DynamoDB sessions| Session store so blue/green doesn't lose in-flight games             |
| Lore cache            | DynamoDB lore    | Caches Gemini responses (one entry per champion)                     |
| Secrets               | Secrets Manager  | Gemini API key, injected into Fargate as env var                     |
| Frontend hosting      | S3 website       | Public static site with SPA fallback (403/404 → index.html)          |
| Logs + metrics        | CloudWatch       | Container logs, ALB metrics, dashboard, alarms                       |
| External AI           | Gemini API       | LLM for champion lore generation                                     |

## Environments

Three environments, identical architecture, separate state and resources:

| Env     | Trigger                      | Approval     |
| ------- | ---------------------------- | ------------ |
| dev     | push to `main` (post-CI)     | none         |
| staging | after dev E2E tests pass     | none         |
| prod    | push of `v*.*.*` git tag     | required (GitHub Environment reviewer) |

## Blue/Green Deployment

Two ECS services per env (`blue`, `green`), each backed by its own ALB
target group. The listener default rule forwards to the *active* TG. A
second listener rule forwards traffic carrying `X-Preview: green` directly
to the green TG for smoke testing before swapping.

`scripts/deploy-app.sh <env> <tag>` orchestrates the swap:

1. Resolve infra IDs from the AWS API using known resource names
2. Determine active color from the listener's default action
3. Register a new task definition revision pointing at `ECR:<tag>`
4. Scale inactive service to desired=2, wait for `services-stable`
5. Smoke test via `X-Preview: green` header (only applicable when INACTIVE=green;
   blue smoke is gated by ECS health checks + ALB target health)
6. Modify listener default action to the inactive TG
7. Poll CloudWatch alarms every 10s for 5 minutes (`target_5xx_high`, `p95_latency_high`)
8. On alarm: revert listener → previous TG; else drain old service to desired=0

## Design decisions

- **S3 static website over CloudFront+OAC** — AWS Academy `voclabs` IAM role
  denies both `cloudfront:CreateOriginAccessControl` and
  `cloudfront:CreateCloudFrontOriginAccessIdentity`. Direct public-read S3
  is the Academy-safe alternative; trade-off is no HTTPS and no CDN edge.
- **Fargate over Lambda** — container model maps to industry-standard pattern,
  simpler Terraform modules for this shape of app.
- **DynamoDB over RDS** — serverless, no VPC/subnet-group plumbing, native TTL
  for session expiry, fits stateless ECS tasks and blue/green semantics.
- **Manual blue/green over CodeDeploy** — `LabRole` restricts CodeDeploy
  permissions; orchestrating the swap ourselves also surfaces every step
  transparently for the presentation.
- **Terraform apply out-of-band from CI/CD** — no remote state backend was
  provisioned (Academy scope), so the CD workflow ships *code* only; infra
  changes are deliberate manual applies from a workstation.
- **Trunk-based development + semver tags for prod** — modern industry
  practice; tags give an immutable trace from `v1.2.3` → ECR image →
  task def revision → CloudWatch logs.
