# Lolidle

LoL champion guessing game (Loldle clone, Classic mode, freeplay) built for the EAFIT DevOps final project.

## Stack

- **Backend:** Go 1.22 + chi (in-memory session store, embedded champion data)
- **Frontend:** React 19 + Vite + TypeScript
- **Data:** 30 curated LoL champions embedded in the Go binary via `//go:embed`

## Run locally

### Backend

```bash
cd backend
go run ./cmd/server
# listens on :8080
# CORS allows http://localhost:5173 by default
```

Env var overrides:
- `PORT` — change listening port
- `CORS_ORIGIN` — change allowed frontend origin

### Frontend

```bash
cd frontend
npm install
npm run dev
# opens http://localhost:5173
```

If you change the backend port, set `VITE_API_BASE` accordingly:
```bash
VITE_API_BASE=http://localhost:9000 npm run dev
```

## Test

```bash
# Backend (4 packages, ~94% avg coverage)
cd backend && go test ./... -cover

# Frontend (14 tests across 4 files)
cd frontend && npm test
cd frontend && npm run coverage
```

## API endpoints

| Method | Path | Body | Response |
|---|---|---|---|
| `GET` | `/api/health` | — | `{"status":"ok"}` |
| `GET` | `/api/champions` | — | `[{id, name}, ...]` |
| `POST` | `/api/games` | — | `{gameId}` (201 Created) |
| `POST` | `/api/games/:gameId/guesses` | `{championId}` | `{guess, feedback, correct, attemptCount}` |

Error responses:
- `404` — game not found or expired
- `409` — game already won
- `400` — invalid body or unknown champion

## Game rules (Classic mode)

Pick a champion. The 7 attribute cells turn:
- 🟩 **green** if your champion matches the target on that attribute
- 🟨 **yellow** if there's a partial overlap (multi-value attributes only: positions, regions)
- 🟥 **red** if no match
- ⬆️ / ⬇️ for the **release year** column, indicating the target is newer/older

No attempt limit. No additional hints. Win by guessing the exact champion.

## Project structure

```
.
├── backend/
│   ├── cmd/server/main.go           # entrypoint, chi router, CORS, middleware
│   ├── internal/
│   │   ├── champions/               # embedded JSON store (All, ByID, Random)
│   │   ├── game/                    # pure Compare(guess, target) → Feedback
│   │   ├── session/                 # in-memory game store with TTL
│   │   └── api/                     # HTTP handlers
│   └── go.mod
├── frontend/
│   ├── src/
│   │   ├── api/                     # types + typed fetch client
│   │   ├── components/              # SearchBox, GuessTable, WinBanner
│   │   ├── App.tsx                  # state + composition
│   │   ├── main.tsx                 # React mount
│   │   └── styles.css
│   └── package.json
└── docs/superpowers/
    ├── specs/                       # design specs
    └── plans/                       # implementation plans
```

## Roadmap

- [x] **App layer** — Go backend + React frontend, fully tested locally
- [ ] **CI pipeline** — GitHub Actions: lint → test → build → release artifacts
- [ ] **CD pipeline** — Terraform → AWS (ECR + ECS Fargate or Lambda), 2 environments (dev + prod), smoke tests pre/post-deploy, manual rollback button
- [ ] **3 pipeline/architecture modifications** required by the assignment (TBD: blue/green, security scanning, observability, etc.)
