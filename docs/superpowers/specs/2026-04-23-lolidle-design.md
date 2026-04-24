# Lolidle — Design Spec

**Date:** 2026-04-23
**Author:** David Vélez
**Context:** Final project for EAFIT DevOps course. App layer first; CI/CD pipeline (GitHub Actions + AWS) designed in a follow-up spec.

## Goal

Build a self-hosted clone of [Loldle.net](https://loldle.net) **Classic mode** — a daily-style guessing game where the player tries to identify a League of Legends champion by submitting guesses and receiving per-attribute feedback (match / partial / no match).

Scope is intentionally narrow: **freeplay only** (each new game picks a random champion), **no persistence**, **no auth**, **no daily challenge**. The simplicity is deliberate so the bulk of the project effort can go into the CI/CD pipeline (which is what the assignment evaluates).

## Non-goals

- Daily challenge logic (everyone-gets-the-same-champion-today)
- User accounts, login, or leaderboards
- Streaks, stats, or any cross-session persistence
- Additional game modes (Quote, Ability, Splash, Emoji)
- Hints triggered after N failed attempts — feedback comes only from attribute comparison
- Mobile-first design (desktop is primary; mobile should not break but is not optimized)

## Components

The repo is a monorepo with two cleanly separated components, satisfying the assignment's "≥ 2 components" requirement.

```
┌─────────────┐     HTTP/JSON      ┌──────────────────┐
│  Frontend   │  ───────────────►  │     Backend      │
│ React+Vite  │                     │   Go + chi       │
│ TypeScript  │  ◄───────────────  │   stateless API  │
└─────────────┘                     └──────────────────┘
                                            │
                                  //go:embed │
                                            ▼
                                   ┌──────────────────┐
                                   │ champions.json   │
                                   │ (~170 champions) │
                                   └──────────────────┘
```

### Backend (`backend/`)

- **Language:** Go (1.22+)
- **Router:** [`chi`](https://github.com/go-chi/chi) — lightweight, idiomatic, fully `http.Handler`-compatible. Trade-off vs stdlib: chi gives us middleware (CORS, logging, recovery) with one-line wiring, which we'd otherwise hand-roll.
- **State:** In-memory `map[gameID]GameState` guarded by `sync.RWMutex`, with TTL eviction (30 minutes). No external dependencies.
- **Data:** A curated `data/champions.json` file embedded into the binary via `//go:embed`. Build is reproducible; runtime has no external network dependency.

### Frontend (`frontend/`)

- **Language:** TypeScript
- **Build tool:** Vite
- **UI framework:** React 18
- **HTTP:** native `fetch`
- **Styling:** plain CSS modules (no Tailwind/UI-kit — keep it tiny)

## Game flow

1. User opens the page.
2. Frontend calls `POST /api/games` → receives `{gameId}`.
3. Frontend calls `GET /api/champions` (once, cached) for the autocomplete list of `{id, name}`.
4. User types in the search box; autocomplete suggests champions.
5. User submits a guess → `POST /api/games/:id/guesses` with `{championId}`.
6. Backend compares the guess against the target and returns per-attribute feedback.
7. Frontend appends a row to the guesses table with colored cells per attribute.
8. Loop steps 4–7 until the response has `correct: true`.
9. Win banner appears with attempt count and a "Play again" button (which calls `POST /api/games` again).

## Attributes and feedback rules

Seven attributes per champion, matching standard Loldle:

| Attribute | Type | Feedback rule |
|---|---|---|
| `gender` | single value | `match` / `nomatch` |
| `positions` | multi-value (1–2 of: Top, Jungle, Mid, ADC, Support) | `match` (sets equal) / `partial` (non-empty intersection) / `nomatch` |
| `species` | single value (Human, Yordle, Void, Vastayan, …) | `match` / `nomatch` |
| `resource` | single value (Mana, Energy, Fury, None, …) | `match` / `nomatch` |
| `rangeType` | single value (Melee, Ranged) | `match` / `nomatch` |
| `regions` | multi-value (1–N of: Demacia, Noxus, Ionia, …) | `match` / `partial` / `nomatch` |
| `releaseYear` | integer (e.g., 2009–2024) | `match` if equal; otherwise `lower` if target year < guess year, `higher` if target year > guess year |

**No additional hints are revealed after N failed attempts.** The only information the player sees is the per-attribute feedback above.

There is **no attempt limit**. The game ends only when all attributes match.

## API

All endpoints return JSON. CORS is enabled for the frontend origin (configurable via env var, default `http://localhost:5173`).

### `GET /api/health`

Healthcheck. Returns `{"status":"ok"}`. Used by smoke tests in the CD pipeline (later spec).

### `GET /api/champions`

Returns a minimal list for the autocomplete:

```json
[
  {"id": "ahri", "name": "Ahri"},
  {"id": "yasuo", "name": "Yasuo"},
  ...
]
```

### `POST /api/games`

Creates a new game session. The server picks a random champion as the target.

**Response:**
```json
{
  "gameId": "01HXY..."
}
```

The `gameId` is a UUID/ULID. The target champion is **not** disclosed.

### `POST /api/games/:gameId/guesses`

Submits a guess.

**Request:**
```json
{ "championId": "yasuo" }
```

**Response:**
```json
{
  "guess": {
    "id": "yasuo",
    "name": "Yasuo",
    "gender": "Male",
    "positions": ["Mid", "Top"],
    "species": "Human",
    "resource": "Flow",
    "rangeType": "Melee",
    "regions": ["Ionia"],
    "releaseYear": 2013
  },
  "feedback": {
    "gender":      {"status": "match"},
    "positions":   {"status": "partial"},
    "species":     {"status": "match"},
    "resource":    {"status": "nomatch"},
    "rangeType":   {"status": "match"},
    "regions":     {"status": "nomatch"},
    "releaseYear": {"status": "higher"}
  },
  "correct": false,
  "attemptCount": 3
}
```

**Errors:**
- `404` if `gameId` does not exist or has expired
- `400` if `championId` is unknown
- `409` if the game is already won (frontend should redirect to "Play again")

## Frontend layout

A single page, no routing required.

```
┌─────────────────────────────────────────────────┐
│                  LOLIDLE                        │
│            Adivina el campeón                   │
│                                                 │
│  ┌──────────────────────────────────┐           │
│  │ Buscar campeón...                │ ←autocomp │
│  └──────────────────────────────────┘           │
│                                                 │
│  ┌──────┬───────┬─────────┬──────┬───┬──────┐  │
│  │Champ │Gender │Position │Spec. │...│ Year │  │
│  ├──────┼───────┼─────────┼──────┼───┼──────┤  │
│  │Yasuo │  🟩   │   🟨   │  🟥  │...│  ⬇️  │  │
│  │Ahri  │  🟩   │   🟩   │  🟩  │...│  ⬆️  │  │
│  └──────┴───────┴─────────┴──────┴───┴──────┘  │
└─────────────────────────────────────────────────┘
```

When the player wins, a banner replaces the search box: "¡Ganaste en N intentos!" + a "Jugar de nuevo" button.

## Repo layout

```
devops/
├── backend/
│   ├── cmd/server/main.go           # entrypoint, wires router
│   ├── internal/
│   │   ├── game/                    # pure game logic (compare, evaluate)
│   │   │   ├── compare.go
│   │   │   └── compare_test.go
│   │   ├── champions/               # load + index embedded JSON
│   │   │   ├── store.go
│   │   │   └── store_test.go
│   │   ├── session/                 # in-memory game store with TTL
│   │   │   ├── store.go
│   │   │   └── store_test.go
│   │   └── api/                     # HTTP handlers
│   │       ├── handlers.go
│   │       └── handlers_test.go
│   ├── data/
│   │   └── champions.json
│   ├── go.mod
│   ├── go.sum
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   │   ├── SearchBox.tsx
│   │   │   ├── GuessTable.tsx
│   │   │   └── WinBanner.tsx
│   │   ├── api/client.ts
│   │   ├── App.tsx
│   │   └── main.tsx
│   ├── index.html
│   ├── package.json
│   ├── vite.config.ts
│   ├── tsconfig.json
│   └── Dockerfile
├── docs/
│   └── superpowers/specs/
│       └── 2026-04-23-lolidle-design.md
└── README.md
```

## Champion data

The seeded `champions.json` will contain a curated subset of ~30–40 popular champions for the first iteration (enough to play meaningfully and demonstrate every attribute kind). The full ~170-champion set can be added later — same shape, no code changes needed.

Schema per entry:

```json
{
  "id": "ahri",
  "name": "Ahri",
  "gender": "Female",
  "positions": ["Mid"],
  "species": "Vastayan",
  "resource": "Mana",
  "rangeType": "Ranged",
  "regions": ["Ionia"],
  "releaseYear": 2011
}
```

Source of truth: Riot's Data Dragon for names/positions, Wiki for the metadata Data Dragon doesn't expose (gender, species, exact region, release year). Curated by hand into the JSON.

## Testing strategy (app layer)

| Layer | Tool | What it covers |
|---|---|---|
| Backend unit | `go test` + table-driven tests | `internal/game.Compare` for every attribute kind and every feedback status; `internal/session` TTL eviction |
| Backend handler | `net/http/httptest` | Full request/response cycle for each endpoint; happy path + error paths (`404`, `400`, `409`) |
| Backend coverage | `go test -coverprofile` | Target ≥ 80% on `internal/` packages |
| Backend lint | `go vet` + `staticcheck` | Static analysis |
| Frontend unit | Vitest + React Testing Library | `compareGuess` rendering, `SearchBox` autocomplete, `GuessTable` row rendering |
| Frontend lint | ESLint + Prettier | Style + common bugs |
| Frontend type | `tsc --noEmit` | Type checking |

E2E tests (Playwright) are deferred to the CI/CD spec, where they belong as the post-deploy smoke tests.

## What's deferred to a follow-up spec

- Dockerfiles content (multi-stage builds, distroless base for Go, nginx for the SPA)
- GitHub Actions CI pipeline (Checkout → Build → Test → Release)
- AWS CD pipeline (IaC with Terraform, two environments, blue/green or canary, smoke tests, manual rollback button)
- The three "modifications/additions" required by the assignment

These will live in `docs/superpowers/specs/2026-04-23-lolidle-cicd-design.md` once the app layer is working locally.

## Open risks

- **Time:** assignment is due 2026-04-25 (2 days). The app must be done by end of day 2026-04-24 to leave a full day for CI/CD. If app slips, we cut frontend polish first (the table works, just less pretty).
- **Champion data curation:** hand-curating 30 champions takes ~1 hour. If shorter is acceptable, start with 15 of the most iconic.
- **No DB means no real persistence:** if the grader expects to see a database in the architecture, we'd need to revisit. The assignment text says "≥ 2 components" — the FE/BE split satisfies this literally.
