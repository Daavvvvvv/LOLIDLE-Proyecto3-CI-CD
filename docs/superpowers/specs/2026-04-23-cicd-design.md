# CI/CD Design Spec — Lolidle

**Date:** 2026-04-23
**Author:** David Vélez
**Context:** Final project for the EAFIT DevOps course, presented 2026-04-25. Builds on top of `v0.1.0-app` (Go + React Loldle clone with portraits, animations, keyboard nav). This spec covers the cloud architecture and CI/CD pipeline that fulfills the assignment rubric.

## Goal

Ship a complete CI/CD pipeline on AWS that satisfies every line item of the assignment rubric:

- CI: Checkout → Build → Test (Static, Unit, Coverage) → Release artifact
- CD: IaC → Deploy Infra (≥2 envs) → Deploy App (≥2 envs) → Tests Pre-Prod → Tests Post-Prod → Rollback (auto on smoke fail + manual panic button)
- 3 modifications/additions at pipeline or architecture level
- Software architecture diagram with named components
- Branching strategy with documented justification
- Working demo in each environment

The implementation must work within AWS Academy Learner Lab constraints (short-lived credentials, `LabRole` only, no Route53, ~$50-100 budget).

## Non-goals

- Multi-region deployment
- Custom domain names (Academy can't reliably issue ACM certs without Route53; we use the AWS-provided `*.cloudfront.net` and `*.execute-api...amazonaws.com` style URLs)
- 24/7 uptime — the architecture is built for the demo session and reproducible re-deploys, not for permanent operation
- WAF, GuardDuty, Inspector, or other AWS security services beyond the basics
- Cost monitoring beyond the Academy's built-in budget
- Mobile-first responsive frontend
- Internationalization
- Authentication / user accounts (still freeplay, no login)

## AWS Academy constraints (acknowledge upfront)

- **Sessions die after ~4 hours.** Each new session generates new `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_SESSION_TOKEN`. We must **rotate GitHub repository secrets** at the start of each session for the pipeline's CD jobs to work. This is a documented manual step in the runbook.
- **Only `LabRole` available.** We cannot create custom IAM users. We *can* create IAM roles via Terraform as long as their trust relationship is for AWS services (e.g., `ecs-tasks.amazonaws.com`). Pipeline assumes those custom roles via the lab session credentials.
- **Region: `us-east-1`** only.
- **No Route53 hosted zones, no ACM custom certs.** All public URLs use AWS-provided defaults.
- **Demo plan:** the entire infra is brought up in one Academy session, the demo runs, and we can `terraform destroy` afterwards. The graders see screenshots / recordings if they review later, plus the GitHub Actions run history (which persists indefinitely).

## Architecture

```
                                  ┌──────────────────┐
                                  │  Browser (User)  │
                                  └────────┬─────────┘
                                           │ HTTPS
                          ┌────────────────┴────────────────┐
                          ▼                                 ▼
                 ┌────────────────┐                ┌──────────────────┐
                 │  CloudFront    │                │     ALB          │
                 │  (frontend CDN)│                │  (backend entry) │
                 └────────┬───────┘                └────────┬─────────┘
                          │                                 │
                          ▼                                 ▼
                 ┌────────────────┐                ┌─────────────────────┐
                 │      S3        │                │  ECS Fargate        │
                 │  (dist/ React) │                │  Service "BLUE"     │
                 └────────────────┘                │  + tasks (Go API)   │
                                                   └────────┬────────────┘
                                                            │
                                  ┌─────────────────────────┼──────────┐
                                  │                         │          │
                                  ▼                         ▼          ▼
                         ┌──────────────────┐      ┌────────────┐  ┌──────────────┐
                         │  ECS Service     │      │  DynamoDB  │  │   Secrets    │
                         │  "GREEN"         │      │ sessions + │  │   Manager    │
                         │  (during deploy) │      │ lore cache │  │ (Gemini key) │
                         └──────────────────┘      └────────────┘  └──────┬───────┘
                                                                          │
                                                                          ▼
                                                                  ┌──────────────┐
                                                                  │  Gemini API  │
                                                                  │ (genera lore)│
                                                                  └──────────────┘

  Cross-cutting observability:
    CloudWatch Logs (containers + ALB access logs)
    CloudWatch Metrics (ECS, ALB, custom)
    CloudWatch Dashboard (single pane of glass per env)
    CloudWatch Alarms (5xx > 5%, p95 > 2s → auto-rollback trigger)
```

**Component count: 12** (ECR, ECS Cluster, ECS Service Blue, ECS Service Green, 2 Target Groups, ALB, S3, CloudFront, DynamoDB, Secrets Manager, CloudWatch composite, Gemini external).

**3 environments**: `dev`, `staging`, `prod`. Each is an independent Terraform stack (separate state file under `infra/envs/<env>/`) but shares the same module library under `infra/modules/`. Resources are namespaced by env (e.g., `lolidle-dev-cluster`, `lolidle-prod-cluster`).

## Per-component specs

### ECR (Elastic Container Registry)

- One repository: `lolidle-backend`
- Image tags: `<git-sha>` and `<semver-tag-when-applicable>`
- Lifecycle policy: keep latest 20 images, expire older
- Image scanning enabled (AWS native, in addition to CI's Trivy scan)

### ECS Fargate

**Cluster:** one per environment, `lolidle-{env}-cluster`. Capacity providers: `FARGATE` only (no spot to keep simplicity).

**Task Definition:**
- CPU: 256 (0.25 vCPU), Memory: 512 MiB
- Container: `lolidle-backend:<tag>` from ECR
- Port mapping: 8080
- Env vars: `PORT=8080`, `CORS_ORIGIN=<cloudfront-url>`, `ENV={env}`
- Secrets injected from Secrets Manager: `GEMINI_API_KEY` (via `secrets` block in container def, AWS handles the fetch)
- Logging driver: `awslogs` → `/ecs/lolidle-{env}` log group
- Task IAM role: permissions to read DynamoDB tables `lolidle-{env}-*`, read Secrets Manager `lolidle-{env}/*`, write CloudWatch metrics in custom namespace

**Services:** Two per environment — `lolidle-{env}-blue` and `lolidle-{env}-green`. Both registered against separate target groups behind the same ALB. Desired count: 2 each (for ALB health redundancy). The "active" service has its target group attached to the listener default rule; the "idle" service has its target group accessible only via header-routed rule for smoke tests.

### ALB + Target Groups + Listener Rules

**ALB:** `lolidle-{env}-alb`, internet-facing, two public subnets in different AZs, security group allowing 80 from 0.0.0.0/0.

**Target groups:** `lolidle-{env}-tg-blue`, `lolidle-{env}-tg-green`. Each:
- Protocol: HTTP, port 8080
- Health check path: `/api/health`, interval 15s, healthy threshold 2, unhealthy 3
- Deregistration delay: 30s

**Listener (port 80):**
- Default rule: forward 100% to active target group (initially blue)
- Rule with priority 100: condition `http-header X-Preview = green`, forward to green TG → used for smoke tests against the inactive deployment without affecting real traffic

### DynamoDB

**Tables (one set per env):**

`lolidle-{env}-sessions`
- Partition key: `gameId` (String)
- TTL attribute: `expiresAt` (Number, Unix epoch) → DynamoDB auto-deletes expired sessions
- Billing: PAY_PER_REQUEST (on-demand)
- Item schema: `{ gameId, targetId, attempts, won, lastAccessed, expiresAt }`

`lolidle-{env}-lore-cache`
- Partition key: `championId` (String)
- Billing: PAY_PER_REQUEST
- Item schema: `{ championId, lore, generatedAt }`
- No TTL — lore is stable across champion lifetime; we manually invalidate via `terraform taint` if we want regen

### Secrets Manager

Secret: `lolidle/{env}/gemini-api-key`
- Plain string value (the API key from Google AI Studio)
- ECS task role gets `secretsmanager:GetSecretValue` on this ARN only
- ECS injects it as env var `GEMINI_API_KEY` at task start (no Go code needs to call AWS SDK for this; ECS handles it)

### S3 + CloudFront (frontend)

**S3 bucket:** `lolidle-{env}-frontend`
- Static website hosting disabled (CloudFront fronts it)
- Bucket policy: only `cloudfront.amazonaws.com` (Origin Access Control) can read
- Versioning: enabled (so pipeline can roll back frontend with a previous object version)

**CloudFront distribution:** one per env
- Origin: the S3 bucket via OAC
- Default root object: `index.html`
- Error responses: 404/403 → return `index.html` with 200 (SPA routing fallback)
- Default cache behavior: cache `*.js`/`*.css` for 1 year (Vite hashes filenames), cache `index.html` for 0 (always fresh)
- Default certificate (`*.cloudfront.net`)

### CloudWatch

**Log Groups:** `/ecs/lolidle-{env}` (retention: 7 days for dev/staging, 30 for prod)

**Dashboard per env:** widgets for:
- ALB request count + 5xx rate
- ALB target health (per target group)
- ECS task CPU and memory utilization (blue + green)
- p50/p95/p99 latency from ALB
- Custom metric: Gemini API call success rate (emitted by Go app via `cloudwatch:PutMetricData`)
- Custom metric: DynamoDB cache hit rate
- Recent CloudWatch Logs Insights query (top 5xx by path in last 1h)

**Alarms (per env, used for auto-rollback):**
- `lolidle-{env}-5xx-rate-high`: `5xx_rate > 5%` for 2 consecutive 1-minute periods
- `lolidle-{env}-p95-latency-high`: `p95 > 2000ms` for 2 consecutive 1-minute periods
- `lolidle-{env}-target-unhealthy`: `unhealthy_host_count > 0` for 1 minute

These alarms are queried by the CD pipeline during the 5-minute observation window after a blue/green swap.

## Backend code changes

### `internal/session/store.go` — switch to DynamoDB

Replace the in-memory `Store` with a DynamoDB-backed implementation that satisfies the same interface:

```go
type Store interface {
    Create(targetID string) (*Game, error)
    Get(id string) (*Game, error)
    Update(g *Game) error
}
```

Two implementations:
- `MemoryStore` — keep for local dev and tests (no AWS dependency)
- `DynamoDBStore` — production, used in deployed envs

Selection at startup based on env var `STORE_BACKEND` (`memory` | `dynamodb`).

DynamoDB writes use TTL attribute `expiresAt = now + 30min` so abandoned sessions auto-clean.

### `internal/lore/service.go` — Gemini client + cache

New package. Public interface:

```go
type Service interface {
    Generate(ctx context.Context, championID string) (string, error) // returns "" on error or skip
}
```

Implementation: `GeminiService`
- On `Generate(championID)`:
  1. Check DynamoDB cache `lore-cache` table for `championId`
  2. If hit: return cached string
  3. If miss: call Gemini API (`https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent`) with prompt:
     > "Write a brief 2-3 sentence flavor description of the League of Legends champion {name}, focusing on their lore: who they are, where they come from, and what they're known for. Spanish language. No spoilers about specific gameplay mechanics."
  4. On success: cache the response in DynamoDB, return it
  5. On any error (timeout, rate limit, secret missing, API down): return `("", nil)` — caller treats empty string as "no lore" and omits from response

Constructor takes `geminiAPIKey string` (read once from `os.Getenv("GEMINI_API_KEY")` injected by ECS), `dynamoClient *dynamodb.Client`, `cacheTableName string`. If `geminiAPIKey == ""` (e.g., local dev without secret), the service is a no-op that always returns `("", nil)`.

### `internal/api/handlers.go` — include lore in win response

When `correct=true` in `SubmitGuess` response, also call `loreService.Generate(ctx, target.ID)` and include `"lore": "..."` in the response. The frontend already receives the `Champion` shape; we add a top-level `lore string` field on the response struct.

If lore generation takes > 3 seconds, the request times out the lore call but still returns the rest of the response (`lore: ""`).

### `frontend/src/components/WinBanner.tsx` — render lore if present

Add optional prop `lore?: string`. If non-empty, render below the "El campeón era X" line as italicized flavor text in a styled `<blockquote>`. App.tsx passes `lastGuess.lore`.

### `Dockerfile` (new, in `backend/Dockerfile`)

Multi-stage:
```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /server /server
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/server"]
```

Final image size: ~15-20 MB. Distroless has no shell (security), runs as non-root.

### Tests for new code

- `lore.GeminiService` unit tests with mocked HTTP client (use `httptest.NewServer`) and mocked DynamoDB (use `aws-sdk-go-v2-mock` or interface-driven mock)
- `session.DynamoDBStore` integration tests using DynamoDB Local (Docker) — runs in CI as a service container
- `session.MemoryStore` tests stay (they are the existing tests, unchanged)
- API handler tests get an additional case: `TestSubmitGuess_includesLoreOnWin_whenServiceReturnsText` and `TestSubmitGuess_omitsLoreOnError_whenServiceReturnsEmpty`

## CI Pipeline (`.github/workflows/ci.yml`)

Triggers: every push to any branch, every PR to main.

```yaml
jobs:
  backend:
    steps:
      - checkout
      - setup-go 1.22
      - cache go modules
      - run: go fmt -l . | (! grep .)        # fail on any unformatted files
      - run: go vet ./...
      - run: staticcheck ./... (install via go install)
      - run: gosec -severity high ./...      # SAST → mod #2
      - run: go test ./... -race -coverprofile=coverage.out
      - run: go tool cover -func=coverage.out | awk '/total/ {if ($3+0 < 80) exit 1}'
      - upload coverage as artifact

  frontend:
    steps:
      - checkout
      - setup-node 22
      - cache npm
      - run: npm ci
      - run: npm run lint
      - run: npx tsc --noEmit
      - run: npm run coverage              # vitest with v8 coverage
      - run: npm audit --audit-level=high  # → mod #2
      - run: npm run build
      - upload dist/ as artifact

  docker:
    needs: [backend]
    steps:
      - checkout
      - setup buildx
      - run: hadolint backend/Dockerfile         # → mod #2
      - build (load into local docker, no push yet)
      - run: trivy image --severity CRITICAL --exit-code 1 lolidle-backend:test  # → mod #2
      - if push to main:
          - configure AWS creds from secrets (academy-rotated)
          - login to ECR
          - re-tag and push as <sha> + (if tag event) <semver>
```

CI is the gate for merge. PRs cannot be merged if any job fails.

## CD Pipelines

Three workflow files separating concerns:

### `.github/workflows/cd-dev-staging.yml`

Triggers: push to `main` (after CI succeeds via `workflow_run` dependency).

```yaml
jobs:
  deploy-dev:
    environment: dev
    steps:
      - checkout
      - configure AWS creds
      - terraform init -chdir=infra/envs/dev
      - terraform apply -auto-approve  # IaC → rubric ✅
      - run deploy-app.sh dev <sha>     # see "Deploy script" below
      - smoke test: curl ALB /api/health, post a guess, verify response
      - if any failure: rollback (auto)

  e2e-dev:
    needs: deploy-dev
    steps:
      - install playwright
      - run E2E suite against dev URL
      - upload trace as artifact

  deploy-staging:
    needs: e2e-dev
    environment: staging
    steps: (same shape as deploy-dev but for staging env)

  e2e-staging:
    needs: deploy-staging
    steps: (same as e2e-dev but staging URL)
```

### `.github/workflows/cd-prod.yml`

Triggers: push of git tag matching `v*.*.*`.

```yaml
jobs:
  deploy-prod:
    environment: prod    # GitHub Environment with required reviewer = manual approval gate
    steps:
      - checkout (the tagged commit)
      - configure AWS creds
      - terraform init -chdir=infra/envs/prod
      - terraform apply -auto-approve
      - run deploy-app.sh prod <semver-tag>
      - smoke + observation window
      - on failure: auto-rollback to previous task definition revision
```

The `environment: prod` declaration makes GitHub block the job until a designated reviewer manually approves. This is the **manual approval gate** required by mod #1.

### `.github/workflows/panic-rollback.yml`

Triggers: `workflow_dispatch` only (manual button in GitHub UI).

Inputs:
- `environment` (dev | staging | prod) — required
- `target_version` (e.g., `v1.2.3` or git SHA) — required, dropdown of recent tags

Steps:
1. Configure AWS creds
2. Look up the ECR image tagged `target_version` — fail loudly if not found
3. Force a new ECS task definition revision pointing at that image
4. Update the active service to use the new task definition
5. Wait for steady state
6. Smoke test the endpoint
7. Notify (Slack webhook is out of scope; we'll print a clear summary in the workflow run)

This is the **botón de pánico manual** required by the rubric.

### `deploy-app.sh` — the blue/green orchestration script

Used by all CD workflows. Pseudocode:

```bash
#!/usr/bin/env bash
set -euo pipefail
ENV=$1
IMAGE_TAG=$2

# 1. Determine current active color (blue or green) from the listener default rule
ACTIVE=$(aws elbv2 describe-rules ... | jq -r '.[] ... extract target group color')
INACTIVE=$([[ "$ACTIVE" = "blue" ]] && echo "green" || echo "blue")

# 2. Update the inactive service's task definition to the new image
aws ecs register-task-definition --image lolidle-backend:$IMAGE_TAG ...
aws ecs update-service --service lolidle-$ENV-$INACTIVE --task-definition <new-arn>

# 3. Wait for inactive service to be stable
aws ecs wait services-stable --services lolidle-$ENV-$INACTIVE

# 4. Smoke test against inactive via header-routed listener rule
curl -H "X-Preview: green" $ALB_URL/api/health  # health
curl -X POST -H "X-Preview: green" $ALB_URL/api/games  # create game
# ... small functional smoke

# 5. Swap default rule → inactive becomes new active
aws elbv2 modify-rule --rule-arn <default-rule> --actions ...$INACTIVE...

# 6. Observation window (5 min): poll CloudWatch alarms
for i in {1..30}; do
  STATE=$(aws cloudwatch describe-alarms --alarm-names lolidle-$ENV-5xx-rate-high lolidle-$ENV-p95-latency-high \
    --query 'MetricAlarms[?StateValue==`ALARM`].AlarmName' --output text)
  if [[ -n "$STATE" ]]; then
    echo "ALARM TRIGGERED: $STATE — rolling back"
    aws elbv2 modify-rule --rule-arn <default-rule> --actions ...$ACTIVE...  # revert
    exit 1
  fi
  sleep 10
done

# 7. Drain old service (set desired count to 0)
aws ecs update-service --service lolidle-$ENV-$ACTIVE --desired-count 0

echo "Deploy successful: $ACTIVE → $INACTIVE"
```

This script implements the blue/green pattern manually using only basic ECS + ALB primitives, satisfying mod #1 without depending on CodeDeploy.

## Frontend deploy (parallel to backend deploys)

Triggered as a sub-job of each CD workflow, after backend smoke passes:

```bash
aws s3 sync frontend/dist/ s3://lolidle-$ENV-frontend/ --delete
aws cloudfront create-invalidation --distribution-id $CF_ID --paths "/*"
```

S3 versioning is enabled, so the panic-rollback workflow can `aws s3api list-object-versions` and restore a previous frontend bundle if needed.

## The 3 Modifications (explicit, with deliverables)

### Modification 1: Blue/Green deployment with manual orchestration + automated rollback on metric breach

**What's modified at architecture level:** Two parallel ECS services per environment behind one ALB with two target groups. Listener default rule controls traffic direction. Header-routed preview rule allows zero-tprafic smoke testing of the new deployment.

**What's modified at pipeline level:** `deploy-app.sh` orchestrates the entire blue/green dance. CloudWatch alarms are queried during the post-swap observation window; if any fire, the script reverts the listener rule.

**Demo evidence:**
- Show ALB target groups in AWS console with both blue and green at different points
- Trigger a deploy in real time, screen-record the listener rule swap
- Demonstrate auto-rollback by deploying a deliberately broken image (returns 500 on `/api/health`) — observe the listener revert within 60 seconds

**Manual panic button:** `panic-rollback.yml` workflow_dispatch.

### Modification 2: DevSecOps — multi-layer security scanning in CI

**What's modified at pipeline level:** Four separate scanning steps in CI that block merges:

1. `gosec -severity high ./...` — Go SAST
2. `npm audit --audit-level=high` — frontend dependency vulns
3. `hadolint backend/Dockerfile` — Dockerfile lint
4. `trivy image --severity CRITICAL --exit-code 1` — image scan

PRs that introduce a CRITICAL CVE or HIGH dep vuln cannot be merged. Reports are uploaded as workflow artifacts for review.

**Demo evidence:**
- Show the CI run for a clean PR (all 4 checks green)
- Open a PR that downgrades a dep to a known-vulnerable version, show the failed scan
- Show the artifact reports

### Modification 3: Full observability + AI lore generation with Secrets Manager

**What's added at architecture level:**
- DynamoDB tables for sessions and lore cache (replaces in-memory store)
- Secrets Manager for Gemini API key
- Gemini API external integration
- CloudWatch Dashboard with custom metrics (Gemini call success rate, cache hit rate)
- Structured JSON logging (Go `slog`)

**What's added at pipeline level:**
- Smoke tests verify lore generation works end-to-end (DynamoDB cache + Gemini round-trip)
- CloudWatch alarms feed into the auto-rollback decision

**Demo evidence:**
- Open the CloudWatch Dashboard live during demo
- Win a game, show the AI-generated lore appearing in the WinBanner
- Open Secrets Manager in console to show the Gemini key is stored (value redacted)
- Check CloudWatch Logs Insights to show JSON structured logs queryable

## Branching Strategy → Pipeline mapping (for the rubric's "explanation of versioning strategy" item)

**Strategy: Trunk-based development with semver tag-based prod releases.**

Justification:
- Trunk-based eliminates long-lived feature branches and merge debt; aligns with continuous-delivery best practice (used by Google, Meta, Netflix)
- Solo project with high iteration velocity benefits from short-lived feature branches that merge to main multiple times per day
- Semver tags (`v1.0.0`, `v1.1.0`, ...) provide an immutable, human-meaningful artifact identity: `v1.2.3` in GitHub == image `lolidle-backend:v1.2.3` in ECR == specific task definition revision == specific entries in CloudWatch logs and X-Ray traces

**Mapping:**

| Event | Branch | Triggers | Deploys to |
|---|---|---|---|
| Push to feature branch | `feature/*` | CI (tests + scans, no deploy) | nothing |
| PR opened/updated | `feature/*` → `main` | CI | nothing |
| Merge to main (after PR) | `main` | CI + push image to ECR with tag `<sha>` + auto-deploy chain | dev → e2e → staging → e2e |
| Push tag `v*.*.*` | tagged commit | Re-tag image in ECR with `<semver>` + deploy gated by manual approval | prod |
| Manual `workflow_dispatch` panic-rollback | n/a | Re-deploy specified prior version | any env |

## Repo structure (after implementation)

```
lolidle/
├── backend/
│   ├── cmd/server/main.go               # MODIFIED: choose store backend, init lore service
│   ├── internal/
│   │   ├── api/handlers.go              # MODIFIED: include lore in win response
│   │   ├── champions/...                # unchanged
│   │   ├── game/...                     # unchanged
│   │   ├── session/
│   │   │   ├── store.go                 # MODIFIED: introduce Store interface
│   │   │   ├── memory.go                # MOVED: existing impl renamed
│   │   │   ├── dynamodb.go              # NEW
│   │   │   └── *_test.go                # MODIFIED: tests for both impls
│   │   ├── lore/
│   │   │   ├── service.go               # NEW
│   │   │   ├── gemini.go                # NEW
│   │   │   ├── cache.go                 # NEW: DynamoDB cache layer
│   │   │   └── service_test.go          # NEW
│   │   └── observability/
│   │       └── logger.go                # NEW: slog JSON setup
│   ├── Dockerfile                       # NEW: multi-stage distroless
│   └── ...
├── frontend/
│   ├── src/
│   │   ├── api/types.ts                 # MODIFIED: add lore?: string to GuessResponse
│   │   ├── components/WinBanner.tsx     # MODIFIED: render lore if present
│   │   └── components/WinBanner.test.tsx # MODIFIED: test lore rendering
│   └── ...
├── infra/                               # NEW
│   ├── modules/
│   │   ├── ecs-service/                 # blue + green services + task def + IAM
│   │   ├── alb/                         # ALB + target groups + listener + rules
│   │   ├── frontend/                    # S3 + CloudFront + OAC
│   │   ├── dynamodb/                    # both tables
│   │   ├── secrets/                     # secret + access policy
│   │   ├── observability/               # log group + dashboard + alarms
│   │   └── ecr/                         # repository + lifecycle policy
│   ├── envs/
│   │   ├── dev/main.tf                  # composes modules with env="dev"
│   │   ├── staging/main.tf              # env="staging"
│   │   └── prod/main.tf                 # env="prod"
│   └── shared/
│       ├── backend.tf                   # remote state backend (S3 + DynamoDB lock)
│       └── providers.tf
├── scripts/
│   ├── deploy-app.sh                    # NEW: blue/green orchestrator
│   └── refresh-academy-creds.sh         # NEW: helper for the recurrent secret rotation
├── e2e/                                  # NEW
│   ├── playwright.config.ts
│   └── specs/play-a-game.spec.ts
├── .github/workflows/
│   ├── ci.yml                           # NEW
│   ├── cd-dev-staging.yml               # NEW
│   ├── cd-prod.yml                      # NEW
│   └── panic-rollback.yml               # NEW
└── docs/
    ├── architecture.md                  # NEW: the diagram + per-component explanation, for the presentation
    ├── runbook.md                       # NEW: how to refresh Academy creds, deploy, rollback, troubleshoot
    └── presentation.md                  # NEW: speaker notes / demo script for the 25-Apr presentation
```

## Tests across environments

| Test type | Where it runs | What it validates |
|---|---|---|
| Static analysis (go vet, eslint, tsc, gosec, hadolint) | CI (every push/PR) | Style + security findings without running the app |
| Unit tests (Go test, vitest) | CI | Pure logic correctness |
| Coverage threshold (80%) | CI | Prevents un-tested code from merging |
| Dep vuln scanning (npm audit, trivy) | CI | Supply chain safety |
| Image vulnerability scan | CI (Trivy) + AWS-native (ECR scan-on-push) | OS and base image CVEs |
| Smoke tests (curl/api functional) | CD pipelines, **inactive (preview) target group** before swap | Catches gross failures before traffic shift |
| E2E (Playwright) | CD pipelines, after deploy to dev and staging, runs against the live URL | User-flow correctness in real environment |
| Observation window monitoring | CD pipelines, post-swap | Catches regressions caught by metrics, not by tests |
| Manual exploratory | demo session in prod | The grader's eyeballs |

## Risks and mitigations

| Risk | Mitigation |
|---|---|
| AWS Academy credentials expire mid-deploy | Document refresh procedure; pipeline logs make the failure obvious; restart deploy after refresh |
| Academy quota exhausted before demo | Use Fargate scale-to-zero between sessions; `terraform destroy` non-prod envs after each test; monitor Lab budget UI |
| Gemini API rate limit / quota exhaustion during demo | Cache aggressively in DynamoDB (each champion looked up at most once per env); fallback to no-lore response gracefully |
| Blue/Green script breaks mid-execution leaving inconsistent ALB state | Idempotent script — every run determines current state from AWS, doesn't assume; documented manual recovery steps in runbook |
| Required IAM permission missing in `LabRole` | Identified by experimentation in early implementation tasks; document workarounds (e.g., reduce scope, use alternate approach) |
| Time exhaustion before the 25-Apr presentation | Implement in priority order: 1 env first (dev only) → CI green → basic CD → blue/green → 2 more envs → 3 mods. Each step is a presentable artifact even if subsequent steps slip |
| Frontend SPA routing breaks behind CloudFront | CloudFront error-document config returns index.html with 200 for 403/404 (standard SPA pattern) |

## Out of scope / future work

- WAF in front of ALB
- AWS Inspector / GuardDuty
- Dependabot / Renovate for automated dep updates
- Cost anomaly detection
- Multi-region (DR)
- Custom domain + ACM cert (would need Route53)
- Canary deployment via gradual weighted target groups (currently we do a single swap; canary is mod-future-work)
- Performance regression testing in pipeline (e.g., k6 load tests)
- SBOM generation and image signing with Sigstore/Cosign
- OIDC-based AWS auth from GitHub (Academy doesn't support cross-account trust easily)
- Secrets rotation automation
- Chaos engineering / fault injection

These are listed because they would naturally be the "next steps" question from the grader; we acknowledge them and explain why they're deferred.

## Presentation deliverables (what the rubric grades)

1. ✅ **Software artifact explanation** — covered by `v0.1.0-app` (Loldle game)
2. ✅ **Branching strategy + justification** — section above
3. ✅ **Pipeline diagram + each step + tool** — to be drawn in `docs/architecture.md`
4. ✅ **Versioning ↔ pipeline integration** — section above (mapping table)
5. ✅ **Architecture diagram** — to be drawn (Excalidraw or Mermaid) and embedded in `docs/architecture.md`
6. ✅ **Tests explanation per env** — section above (table)
7. ✅ **Pipeline execution evidence** — GitHub Actions logs (will accumulate naturally)
8. ✅ **Evidence of 3 modifications** — each demo'd live + screenshots
9. ✅ **Demo per environment** — open dev / staging / prod URLs in browser
10. ✅ **Challenges & learnings** — captured in `docs/presentation.md`
