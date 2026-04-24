#!/usr/bin/env bash
set -euo pipefail

# Blue/green deployment orchestrator for Lolidle backend.
#
# Usage: deploy-app.sh <env> <image-tag>
#   env       — dev | staging | prod
#   image-tag — git SHA or semver tag already pushed to ECR

ENV=${1:?env required (dev|staging|prod)}
IMAGE_TAG=${2:?image tag required (git sha or semver)}

echo "==> Deploying $IMAGE_TAG to $ENV"

# Resolve resource identifiers from AWS API using known names. This keeps
# the script self-contained (no dependency on local terraform state), which
# is what CI/CD runners need — infra is managed separately (apply from a
# workstation or a dedicated infra workflow).
ALB_DNS=$(aws elbv2 describe-load-balancers \
  --names "lolidle-$ENV-alb" \
  --query 'LoadBalancers[0].DNSName' --output text)
ALB_URL="http://$ALB_DNS"
ALB_ARN=$(aws elbv2 describe-load-balancers \
  --names "lolidle-$ENV-alb" \
  --query 'LoadBalancers[0].LoadBalancerArn' --output text)
LISTENER_ARN=$(aws elbv2 describe-listeners \
  --load-balancer-arn "$ALB_ARN" \
  --query 'Listeners[?Port==`80`].ListenerArn | [0]' --output text)
TG_BLUE_ARN=$(aws elbv2 describe-target-groups \
  --names "lolidle-$ENV-tg-blue" \
  --query 'TargetGroups[0].TargetGroupArn' --output text)
TG_GREEN_ARN=$(aws elbv2 describe-target-groups \
  --names "lolidle-$ENV-tg-green" \
  --query 'TargetGroups[0].TargetGroupArn' --output text)
CLUSTER="lolidle-$ENV-cluster"
SVC_BLUE="lolidle-$ENV-blue"
SVC_GREEN="lolidle-$ENV-green"
ECR_REPO=$(aws ecr describe-repositories \
  --repository-names lolidle-backend \
  --query 'repositories[0].repositoryUri' --output text)

CURRENT_TG=$(aws elbv2 describe-listeners \
  --listener-arns "$LISTENER_ARN" \
  --query 'Listeners[0].DefaultActions[0].TargetGroupArn' \
  --output text)

if [[ "$CURRENT_TG" == "$TG_BLUE_ARN" ]]; then
  ACTIVE="blue"
  INACTIVE="green"
  ACTIVE_TG="$TG_BLUE_ARN"
  INACTIVE_TG="$TG_GREEN_ARN"
  ACTIVE_SVC="$SVC_BLUE"
  INACTIVE_SVC="$SVC_GREEN"
else
  ACTIVE="green"
  INACTIVE="blue"
  ACTIVE_TG="$TG_GREEN_ARN"
  INACTIVE_TG="$TG_BLUE_ARN"
  ACTIVE_SVC="$SVC_GREEN"
  INACTIVE_SVC="$SVC_BLUE"
fi

echo "==> Active: $ACTIVE, deploying to $INACTIVE"

TASK_DEF_FAMILY="lolidle-$ENV"
CURRENT_TD=$(aws ecs describe-task-definition \
  --task-definition "$TASK_DEF_FAMILY" \
  --query 'taskDefinition' --output json)

TMP_TD=$(mktemp)
trap 'rm -f "$TMP_TD"' EXIT
echo "$CURRENT_TD" | jq --arg IMAGE "$ECR_REPO:$IMAGE_TAG" '
  .containerDefinitions[0].image = $IMAGE |
  {family, networkMode, containerDefinitions, requiresCompatibilities, cpu, memory, executionRoleArn, taskRoleArn}
' > "$TMP_TD"

NEW_TD_ARN=$(aws ecs register-task-definition \
  --cli-input-json "file://$TMP_TD" \
  --query 'taskDefinition.taskDefinitionArn' --output text)
echo "==> Registered task def: $NEW_TD_ARN"

aws ecs update-service \
  --cluster "$CLUSTER" \
  --service "$INACTIVE_SVC" \
  --task-definition "$NEW_TD_ARN" \
  --desired-count 2 > /dev/null

echo "==> Waiting for $INACTIVE_SVC to be stable..."
aws ecs wait services-stable --cluster "$CLUSTER" --services "$INACTIVE_SVC"

# Smoke test against the inactive target group via X-Preview header.
# The preview listener rule routes X-Preview=green to the green TG, so this
# validation is only meaningful when INACTIVE=green. When INACTIVE=blue we
# fall back to validating via the primary endpoint after service-stable
# (ECS health checks + ALB target health already gate readiness).
echo "==> Smoke testing $INACTIVE"
SMOKE_OK=true
if [[ "$INACTIVE" == "green" ]]; then
  for endpoint in "/api/health" "/api/champions"; do
    CODE=$(curl -s -o /dev/null -w "%{http_code}" \
      -H "X-Preview: green" \
      --max-time 10 \
      "$ALB_URL$endpoint" || echo "000")
    if [[ "$CODE" != "200" ]]; then
      echo "    FAIL: $endpoint returned $CODE"
      SMOKE_OK=false
    else
      echo "    OK: $endpoint"
    fi
  done
else
  echo "    (skipping preview smoke: INACTIVE=blue, no preview rule for blue)"
fi

if [[ "$SMOKE_OK" != "true" ]]; then
  echo "==> Smoke test failed; aborting deploy without swap"
  aws ecs update-service \
    --cluster "$CLUSTER" --service "$INACTIVE_SVC" \
    --desired-count 0 > /dev/null
  exit 1
fi

echo "==> Swapping listener default to $INACTIVE_TG"
aws elbv2 modify-listener \
  --listener-arn "$LISTENER_ARN" \
  --default-actions Type=forward,TargetGroupArn="$INACTIVE_TG" > /dev/null

echo "==> Observation window: 5 minutes monitoring CloudWatch alarms"
ROLLBACK_NEEDED=false
for i in {1..30}; do
  ALARM_STATE=$(aws cloudwatch describe-alarms \
    --alarm-names "lolidle-$ENV-5xx-rate-high" "lolidle-$ENV-p95-latency-high" \
    --query 'MetricAlarms[?StateValue==`ALARM`].AlarmName' --output text)
  if [[ -n "$ALARM_STATE" ]]; then
    echo "    !! ALARM TRIGGERED: $ALARM_STATE"
    ROLLBACK_NEEDED=true
    break
  fi
  echo "    ($i/30) alarms quiet"
  sleep 10
done

if [[ "$ROLLBACK_NEEDED" == "true" ]]; then
  echo "==> ROLLBACK: reverting listener to $ACTIVE_TG"
  aws elbv2 modify-listener \
    --listener-arn "$LISTENER_ARN" \
    --default-actions Type=forward,TargetGroupArn="$ACTIVE_TG" > /dev/null
  exit 1
fi

echo "==> Draining old service $ACTIVE_SVC (desired count -> 0)"
aws ecs update-service \
  --cluster "$CLUSTER" --service "$ACTIVE_SVC" \
  --desired-count 0 > /dev/null

echo "==> Deploy successful: $ACTIVE -> $INACTIVE"
