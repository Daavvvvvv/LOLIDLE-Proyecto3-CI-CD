# Lolidle Operational Runbook

## Refreshing AWS Academy credentials (every session — REQUIRED)

Vocareum credentials expire every ~4 hours. Before any terraform or CI run:

1. Open https://awsacademy.instructure.com/
2. Start the Learner Lab → "AWS Details" → "AWS CLI" → "Show"
3. Copy the three lines (`aws_access_key_id`, `aws_secret_access_key`, `aws_session_token`)
4. **Local** — write them to `~/.aws/credentials`:
   ```
   [default]
   aws_access_key_id=ASIA...
   aws_secret_access_key=...
   aws_session_token=...
   ```
   Verify: `aws sts get-caller-identity`
5. **GitHub Actions** — update repo secrets at Settings → Secrets and variables → Actions:
   - `AWS_ACCESS_KEY_ID`
   - `AWS_SECRET_ACCESS_KEY`
   - `AWS_SESSION_TOKEN`

Any workflow that ran with expired credentials needs a re-run after refreshing.

## One-time GitHub setup

Done once per repo; required before the first CD run can hit prod:

1. Settings → Environments → New environment → `prod`
2. Under "Deployment protection rules", check "Required reviewers" and add
   yourself. This makes the prod workflow pause for manual approval.
3. Optionally: add a deployment branch rule allowing only `v*.*.*` tag refs.

## Common operations

### Deploy the current `main` to dev + staging
```bash
git push origin main
```
CI runs → if green, CD triggers → `deploy-app.sh` blue/greens dev → Playwright
E2E → `deploy-app.sh` blue/greens staging. Watch it in the GitHub Actions tab.

### Release to prod
```bash
git tag -a v1.2.3 -m "Release 1.2.3"
git push origin v1.2.3
```
This triggers `cd-prod.yml`. Approve the deployment in the Actions UI when
prompted (pauses on the `prod` Environment reviewer gate).

### Roll back to a previous image
1. Confirm the image tag exists in ECR:
   ```bash
   aws ecr describe-images --repository-name lolidle-backend \
     --query 'imageDetails[*].imageTags' --region us-east-1
   ```
2. Actions tab → "Panic Rollback" → Run workflow
3. Pick environment + paste the target tag → Run
4. Wait for completion; verify with
   `curl http://$(aws elbv2 describe-load-balancers --names lolidle-prod-alb --query 'LoadBalancers[0].DNSName' --output text)/api/health`

### Scale an env down to save budget (between demos)
```bash
for svc in blue green; do
  aws ecs update-service --cluster lolidle-dev-cluster \
    --service lolidle-dev-$svc --desired-count 0 --region us-east-1
done
```
Services re-scale on the next `deploy-app.sh` or `terraform apply`.

### Bring up prod for the first time (not applied by default)
```bash
cd infra/envs/prod
cp terraform.tfvars.example terraform.tfvars
# edit terraform.tfvars: paste real gemini_api_key
terraform init
terraform apply -auto-approve
```
Then build + push a bootstrap image (same process as dev bootstrap, see
Task 9 in the plan) so ECS can pull on first launch.

### Tear everything down (end of demo)
```bash
for env in prod staging dev; do
  (cd infra/envs/$env && terraform destroy -auto-approve) || true
done
```
Order matters: prod → staging → dev (dev owns the shared ECR repo that
staging/prod reference via data source). If destroy gets stuck on a
frontend S3 bucket, empty it first:
`aws s3 rm s3://<bucket>/ --recursive`.

## Troubleshooting

| Symptom | Likely cause | Fix |
| --- | --- | --- |
| CI `Configure AWS creds` fails | Academy session expired | Refresh creds + update GitHub secrets |
| ECS tasks crashloop | Bad image or missing env var | `aws logs tail /ecs/lolidle-<env> --follow` |
| 503 from ALB | All targets unhealthy (tasks still pulling or crashing) | Wait 60s; check target group health in EC2 → Target Groups |
| `terraform init` fails on provider download | IPv6 connection to `releases.hashicorp.com` blocked by ISP | Use `-plugin-dir=$HOME/.terraform.d/plugins` with a locally-downloaded provider (see below) |
| `terraform apply` says `AccessDenied cloudfront:Create*` | `voclabs` role restrictions | Frontend module already uses S3-only to avoid this; make sure your local branch is up to date |
| Frontend shows CORS errors from browser | `CORS_ORIGIN` env on backend doesn't match the S3 website URL | Re-apply terraform (it re-injects `cors_origin` from `module.frontend.cloudfront_url`) and redeploy the backend |
| Lore never appears in WinBanner | Gemini secret missing, API quota exceeded, or cache miss failing silently | Check Secrets Manager; `aws logs tail /ecs/lolidle-<env> --follow \| grep -i gemini` |
| Panic rollback says "image not found" | That tag was never pushed to ECR | `aws ecr describe-images --repository-name lolidle-backend --query 'imageDetails[*].imageTags'` |

### Terraform IPv6 workaround (local-only)

If `terraform init` hangs or errors on `wsarecv: An existing connection was forcibly closed`:

```bash
MIRROR="$HOME/.terraform.d/plugins/registry.terraform.io/hashicorp/aws/5.100.0/windows_amd64"
mkdir -p "$MIRROR" && cd "$MIRROR"
curl -4 -sS -o p.zip "https://releases.hashicorp.com/terraform-provider-aws/5.100.0/terraform-provider-aws_5.100.0_windows_amd64.zip"
unzip -o p.zip && rm p.zip
cd /d/Programming/EAFIT/devops/infra/envs/dev
terraform init -plugin-dir="$HOME/.terraform.d/plugins"
```

(`-4` forces IPv4; the issue is IPv6 connectivity to CloudFront from some
ISPs). GitHub Actions' ubuntu runners don't have this problem.
