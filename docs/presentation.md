# Presentation Speaker Notes — Lolidle Final Project

Total target time: ~25 minutes + Q&A.

## Slide 1: Software Artifact (3 min)

- Loldle clone for League of Legends, Classic mode
- Two components: Go backend + React/TS frontend
- Medium complexity:
  - 7-attribute comparison logic (region, role, gender, resource, range, class, release year)
  - Autocomplete with keyboard nav (↑/↓/Enter/Escape)
  - Sequential flip animations on guess reveal
  - Post-victory AI-generated lore via Gemini, cached in DynamoDB
- Tech picks reflect real-world trade-offs: Go for a small static binary,
  React/Vite for fast dev loop, DynamoDB for serverless persistence.

## Slide 2: Branching Strategy (2 min)

- **Trunk-based** with semver tag-based prod releases
- Justification:
  - Reduces merge debt vs GitFlow
  - Aligns with continuous delivery (Google, Meta, Netflix)
  - Tags = immutable trace from `v1.2.3` → ECR image → task def → CloudWatch logs
- Mapping to pipelines:
  - `feature/*` → CI only, no deploy
  - merge to `main` → CI + ECR push + auto-deploy dev → E2E → auto-deploy staging
  - `git tag v*.*.*` on main → CD prod with required-reviewer gate

## Slide 3: Pipeline Diagram (4 min)

Show the architecture diagram from `architecture.md`. Walk through:

- **CI** (every push/PR): 3 parallel jobs
  - backend: gofmt, vet, staticcheck, gosec (SAST), tests with 80% coverage, build
  - frontend: eslint, tsc, vitest coverage, npm audit, build
  - docker: hadolint, buildx build, Trivy CRITICAL scan, push to ECR *if main*
- **CD dev+staging** (on CI success, main): `deploy-app.sh dev SHA` → Playwright E2E → `deploy-app.sh staging SHA`
- **CD prod** (on `v*.*.*` tag): manual approval via GitHub Environment, then `deploy-app.sh prod v*` + smoke
- **Panic rollback** (manual workflow_dispatch): env + target_version inputs → `deploy-app.sh` to the older image

## Slide 4: Architecture Diagram (3 min)

Walk through components in `architecture.md` (ECR → ECS blue/green → ALB →
S3 → DynamoDB → Secrets → CloudWatch → Gemini). Trace one guess through
the stack end-to-end: browser → S3 HTML → XHR to ALB → listener →
active TG → Fargate task → DynamoDB Get/Update → response.

## Slide 5: Tests per Environment (2 min)

| Stage            | Tests                                                              |
| ---------------- | ------------------------------------------------------------------ |
| CI (every push)  | Static analysis, unit tests + 80% coverage gate, security scans    |
| Pre-swap         | Smoke test via `X-Preview: green` header against inactive target   |
| Post-swap        | 5-min observation window on CloudWatch alarms → auto-rollback      |
| E2E (post-dev)   | Playwright against dev S3 URL; staging deploy gated on success     |
| Post-prod deploy | `smoke-prod` job curls `/api/health` + `/api/champions` from prod  |

## Slide 6: 3 Modifications (5 min — the big one)

### Mod 1: Blue/Green with auto-rollback on metric breach

- Two ECS services per env, each on its own ALB target group
- Custom orchestration (`scripts/deploy-app.sh`):
  - Detects active color, deploys to inactive, smoke tests, swaps, observes,
    auto-reverts on alarm
- 5-min observation window polls `target_5xx_high` (>5 errors/min) and
  `p95_latency_high` (>1000ms for 2 consecutive 60s periods)
- **Demo:** tail CloudWatch logs during a deploy → show listener swap
  → show alarm dimensions in dashboard

### Mod 2: DevSecOps — multi-layer security in CI

- `gosec -severity high` (Go SAST)
- `npm audit --audit-level=high` (frontend deps; HIGH+ blocks)
- `hadolint` (Dockerfile lint)
- `Trivy` (container image scan; CRITICAL blocks)
- Secrets Manager for Gemini key (never in repo, never in env.production)
- **Demo:** show clean PR (all green) and a deliberately-vulnerable PR
  (failing scan with report artifact)

### Mod 3: Observability stack + AI integration with secure secrets

- CloudWatch Dashboard with 4 panels:
  - ALB request count by target group
  - 5xx error counts by target group
  - p95 response time (active TG)
  - ECS CPU+memory utilization per service (blue + green)
- Structured JSON logging from Go using `slog` with `service=lolidle-backend` attribute
- Gemini API integration: cache-aside pattern in DynamoDB, graceful
  degradation (`("", nil)` on any Gemini error so a win never blocks)
- API key stored in AWS Secrets Manager, fetched by the Fargate task at
  startup via `secrets` field in task definition
- **Demo:** open dashboard live, win a game, show AI lore rendered in
  WinBanner, open Secrets Manager entry with value redacted

## Slide 7: Pipeline Evidence (2 min)

- GitHub Actions tab → recent successful CI runs with all 3 jobs green
- Drill into a CD run → show the sequence deploy-dev → e2e-dev → deploy-staging
- Show a prod workflow paused at the reviewer gate, then approved
- Show the `Panic Rollback` `workflow_dispatch` UI with the two inputs

## Slide 8: Demo per environment (3 min)

- Open dev URL → play a game → show portrait flip + lore
- Open staging URL → same (but maybe different tag deployed)
- Open prod URL → same
- Highlight: same code, same architecture, separate state + resources

## Slide 9: Challenges + Learnings (3 min)

- **AWS Academy session expiration** → manual creds refresh per session
  (operational learning: parameterize auth so rotation is cheap)
- **CloudFront blocked by voclabs IAM** → redesigned frontend to S3 static
  website hosting (constraint-driven design)
- **Stateful blue/green** required externalizing session state to DynamoDB
  so in-flight games survive the listener swap
- **IAM `LabRole` shaped which patterns are practical** (e.g., no CodeDeploy,
  no ACM cert for custom domain)
- **Manual blue/green** required understanding exactly what CodeDeploy
  does under the hood — gained deeper ops knowledge vs. click-ops
- **Cost discipline** matters even with Academy credit budget — no NAT
  gateway, one ALB per env, prod deferred until needed

## Q&A prep

- **Why not CodeDeploy?** → Academy IAM friction; manual gives more
  visibility into every orchestration step, which is useful for grading
- **Why not multi-region?** → out of scope; single-user academic demo
- **Why no remote terraform state?** → Academy scope; local state is
  checked in via the operator (me) and CI deploys code only, not infra.
  Remote state (S3 + DDB lock) is the first thing I'd add for a real org
- **What's the next step?** → WAF, OIDC GitHub→AWS auth instead of long-
  lived keys, canary instead of blue/green, Dependabot for deps,
  remote terraform state with Atlantis
