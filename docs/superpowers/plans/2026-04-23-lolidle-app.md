# Lolidle App Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a working local Lolidle clone (LoL Classic mode, freeplay, no DB, no auth) — Go backend + React/TS frontend — ready to be wrapped in CI/CD in a follow-up plan.

**Architecture:** Monorepo with `backend/` (Go 1.22 + chi, in-memory session store, embedded `champions.json`) and `frontend/` (Vite + React + TypeScript). Stateless API, 4 endpoints, 7 attribute comparison.

**Tech Stack:** Go 1.22, `github.com/go-chi/chi/v5`, React 18, Vite, TypeScript, Vitest, React Testing Library.

**Spec:** `docs/superpowers/specs/2026-04-23-lolidle-design.md`

---

## File Structure

### Backend (`backend/`)
```
backend/
├── cmd/server/main.go              # entrypoint, wires deps + chi router
├── internal/
│   ├── champions/
│   │   ├── champions.json          # embedded data (~30 champions)
│   │   ├── store.go                # Champion type + Store{All, ByID, Random}
│   │   └── store_test.go
│   ├── game/
│   │   ├── compare.go              # pure Compare(guess, target) → Feedback
│   │   └── compare_test.go
│   ├── session/
│   │   ├── store.go                # in-memory Store with TTL
│   │   └── store_test.go
│   └── api/
│       ├── handlers.go             # 4 HTTP handlers + Handler struct
│       └── handlers_test.go
├── go.mod
└── go.sum
```

### Frontend (`frontend/`)
```
frontend/
├── index.html
├── package.json
├── tsconfig.json
├── vite.config.ts
├── src/
│   ├── main.tsx                    # React mount
│   ├── App.tsx                     # top-level state + composition
│   ├── styles.css                  # plain CSS
│   ├── api/
│   │   ├── types.ts                # shared TS types matching backend JSON
│   │   ├── client.ts               # listChampions, createGame, submitGuess
│   │   └── client.test.ts
│   └── components/
│       ├── SearchBox.tsx           # autocomplete input
│       ├── SearchBox.test.tsx
│       ├── GuessTable.tsx          # rows of past guesses with colored cells
│       ├── GuessTable.test.tsx
│       ├── WinBanner.tsx           # "you won in N tries"
│       └── WinBanner.test.tsx
```

### Top level
```
.gitignore
README.md
```

---

## Task 1: Initialize repo and Go module

**Files:**
- Create: `.gitignore`
- Create: `README.md`
- Create: `backend/go.mod`

- [ ] **Step 1: Initialize git**

Run from `D:/Programming/EAFIT/devops`:
```bash
git init
git branch -M main
```

- [ ] **Step 2: Create `.gitignore`**

```gitignore
# Go
backend/server
backend/server.exe
*.test
*.out
coverage.txt

# Node
frontend/node_modules
frontend/dist
.npm

# OS / IDE
.DS_Store
Thumbs.db
.idea/
.vscode/
*.log
```

- [ ] **Step 3: Create `README.md` skeleton**

```markdown
# Lolidle

LoL champion guessing game (Loldle clone, Classic mode, freeplay).

## Run locally

### Backend
```bash
cd backend
go run ./cmd/server
# listens on :8080
```

### Frontend
```bash
cd frontend
npm install
npm run dev
# opens http://localhost:5173
```
```

- [ ] **Step 4: Initialize Go module**

```bash
mkdir -p backend/cmd/server backend/internal/champions backend/internal/game backend/internal/session backend/internal/api
cd backend
go mod init lolidle/backend
go get github.com/go-chi/chi/v5
go get github.com/go-chi/chi/v5/middleware
cd ..
```

- [ ] **Step 5: Verify Go module resolves**

```bash
cd backend && go mod tidy && cd ..
```
Expected: no errors, `go.sum` is created.

- [ ] **Step 6: Commit**

```bash
git add .
git commit -m "chore: initialize repo with go module and gitignore"
```

---

## Task 2: Curate champion data and build the Store

**Files:**
- Create: `backend/internal/champions/champions.json`
- Create: `backend/internal/champions/store.go`
- Test: `backend/internal/champions/store_test.go`

- [ ] **Step 1: Write the failing test**

Create `backend/internal/champions/store_test.go`:
```go
package champions

import (
	"testing"
)

func TestStore_All_returnsAllChampions(t *testing.T) {
	s, err := NewStore()
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	all := s.All()
	if len(all) < 20 {
		t.Errorf("expected at least 20 champions, got %d", len(all))
	}
}

func TestStore_ByID_returnsChampionWhenFound(t *testing.T) {
	s, _ := NewStore()
	c, ok := s.ByID("ahri")
	if !ok {
		t.Fatal("expected to find ahri")
	}
	if c.Name != "Ahri" {
		t.Errorf("expected name Ahri, got %s", c.Name)
	}
}

func TestStore_ByID_returnsFalseWhenNotFound(t *testing.T) {
	s, _ := NewStore()
	if _, ok := s.ByID("nonexistent"); ok {
		t.Error("expected ok=false for unknown id")
	}
}

func TestStore_Random_returnsChampionFromList(t *testing.T) {
	s, _ := NewStore()
	c := s.Random()
	if c.ID == "" {
		t.Error("expected non-empty ID")
	}
	if _, ok := s.ByID(c.ID); !ok {
		t.Error("expected random champion to exist in store")
	}
}
```

- [ ] **Step 2: Run test — fails because `champions.json` and `store.go` don't exist**

```bash
cd backend && go test ./internal/champions/... -v && cd ..
```
Expected: FAIL (no files).

- [ ] **Step 3: Create `champions.json` with 30 curated champions**

Create `backend/internal/champions/champions.json`:
```json
[
  {"id":"ahri","name":"Ahri","gender":"Female","positions":["Mid"],"species":"Vastayan","resource":"Mana","rangeType":"Ranged","regions":["Ionia"],"releaseYear":2011},
  {"id":"yasuo","name":"Yasuo","gender":"Male","positions":["Mid","Top"],"species":"Human","resource":"Flow","rangeType":"Melee","regions":["Ionia"],"releaseYear":2013},
  {"id":"garen","name":"Garen","gender":"Male","positions":["Top"],"species":"Human","resource":"None","rangeType":"Melee","regions":["Demacia"],"releaseYear":2010},
  {"id":"lux","name":"Lux","gender":"Female","positions":["Mid","Support"],"species":"Human","resource":"Mana","rangeType":"Ranged","regions":["Demacia"],"releaseYear":2010},
  {"id":"darius","name":"Darius","gender":"Male","positions":["Top"],"species":"Human","resource":"None","rangeType":"Melee","regions":["Noxus"],"releaseYear":2012},
  {"id":"jinx","name":"Jinx","gender":"Female","positions":["ADC"],"species":"Human","resource":"Mana","rangeType":"Ranged","regions":["Zaun"],"releaseYear":2013},
  {"id":"vi","name":"Vi","gender":"Female","positions":["Jungle"],"species":"Human","resource":"Mana","rangeType":"Melee","regions":["Piltover"],"releaseYear":2012},
  {"id":"caitlyn","name":"Caitlyn","gender":"Female","positions":["ADC"],"species":"Human","resource":"Mana","rangeType":"Ranged","regions":["Piltover"],"releaseYear":2011},
  {"id":"ezreal","name":"Ezreal","gender":"Male","positions":["ADC"],"species":"Human","resource":"Mana","rangeType":"Ranged","regions":["Piltover"],"releaseYear":2010},
  {"id":"lee-sin","name":"Lee Sin","gender":"Male","positions":["Jungle"],"species":"Human","resource":"Energy","rangeType":"Melee","regions":["Ionia"],"releaseYear":2011},
  {"id":"thresh","name":"Thresh","gender":"Male","positions":["Support"],"species":"Spirit","resource":"Mana","rangeType":"Melee","regions":["Shadow Isles"],"releaseYear":2013},
  {"id":"senna","name":"Senna","gender":"Female","positions":["Support","ADC"],"species":"Human","resource":"Mana","rangeType":"Ranged","regions":["Shadow Isles"],"releaseYear":2019},
  {"id":"lucian","name":"Lucian","gender":"Male","positions":["ADC"],"species":"Human","resource":"Mana","rangeType":"Ranged","regions":["Demacia"],"releaseYear":2013},
  {"id":"akali","name":"Akali","gender":"Female","positions":["Mid"],"species":"Human","resource":"Energy","rangeType":"Melee","regions":["Ionia"],"releaseYear":2010},
  {"id":"zed","name":"Zed","gender":"Male","positions":["Mid"],"species":"Human","resource":"Energy","rangeType":"Melee","regions":["Ionia"],"releaseYear":2012},
  {"id":"teemo","name":"Teemo","gender":"Male","positions":["Top"],"species":"Yordle","resource":"Mana","rangeType":"Ranged","regions":["Bandle City"],"releaseYear":2009},
  {"id":"tristana","name":"Tristana","gender":"Female","positions":["ADC"],"species":"Yordle","resource":"Mana","rangeType":"Ranged","regions":["Bandle City"],"releaseYear":2009},
  {"id":"veigar","name":"Veigar","gender":"Male","positions":["Mid"],"species":"Yordle","resource":"Mana","rangeType":"Ranged","regions":["Bandle City"],"releaseYear":2009},
  {"id":"kassadin","name":"Kassadin","gender":"Male","positions":["Mid"],"species":"Human","resource":"Mana","rangeType":"Melee","regions":["Void"],"releaseYear":2009},
  {"id":"chogath","name":"Cho'Gath","gender":"Male","positions":["Top"],"species":"Void","resource":"Mana","rangeType":"Melee","regions":["Void"],"releaseYear":2009},
  {"id":"khazix","name":"Kha'Zix","gender":"Male","positions":["Jungle"],"species":"Void","resource":"Mana","rangeType":"Melee","regions":["Void"],"releaseYear":2012},
  {"id":"reksai","name":"Rek'Sai","gender":"Female","positions":["Jungle"],"species":"Void","resource":"Fury","rangeType":"Melee","regions":["Void"],"releaseYear":2014},
  {"id":"malphite","name":"Malphite","gender":"Male","positions":["Top"],"species":"Golem","resource":"Mana","rangeType":"Melee","regions":["Ixtal"],"releaseYear":2009},
  {"id":"sona","name":"Sona","gender":"Female","positions":["Support"],"species":"Human","resource":"Mana","rangeType":"Ranged","regions":["Demacia"],"releaseYear":2010},
  {"id":"soraka","name":"Soraka","gender":"Female","positions":["Support"],"species":"Celestial","resource":"Mana","rangeType":"Ranged","regions":["Targon"],"releaseYear":2009},
  {"id":"aatrox","name":"Aatrox","gender":"Male","positions":["Top"],"species":"Darkin","resource":"Blood Well","rangeType":"Melee","regions":["Runeterra"],"releaseYear":2013},
  {"id":"kayn","name":"Kayn","gender":"Male","positions":["Jungle"],"species":"Human","resource":"Energy","rangeType":"Melee","regions":["Ionia"],"releaseYear":2017},
  {"id":"sett","name":"Sett","gender":"Male","positions":["Top"],"species":"Vastayan","resource":"Grit","rangeType":"Melee","regions":["Ionia"],"releaseYear":2020},
  {"id":"yone","name":"Yone","gender":"Male","positions":["Mid","Top"],"species":"Spirit","resource":"Flow","rangeType":"Melee","regions":["Ionia"],"releaseYear":2020},
  {"id":"aphelios","name":"Aphelios","gender":"Male","positions":["ADC"],"species":"Human","resource":"Mana","rangeType":"Ranged","regions":["Targon"],"releaseYear":2019}
]
```

- [ ] **Step 4: Implement `store.go`**

Create `backend/internal/champions/store.go`:
```go
package champions

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math/rand/v2"
)

//go:embed champions.json
var rawData []byte

type Champion struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Gender      string   `json:"gender"`
	Positions   []string `json:"positions"`
	Species     string   `json:"species"`
	Resource    string   `json:"resource"`
	RangeType   string   `json:"rangeType"`
	Regions     []string `json:"regions"`
	ReleaseYear int      `json:"releaseYear"`
}

type Store struct {
	list []Champion
	byID map[string]Champion
}

func NewStore() (*Store, error) {
	var list []Champion
	if err := json.Unmarshal(rawData, &list); err != nil {
		return nil, fmt.Errorf("unmarshal champions.json: %w", err)
	}
	if len(list) == 0 {
		return nil, fmt.Errorf("champions.json is empty")
	}
	byID := make(map[string]Champion, len(list))
	for _, c := range list {
		byID[c.ID] = c
	}
	return &Store{list: list, byID: byID}, nil
}

func (s *Store) All() []Champion {
	return s.list
}

func (s *Store) ByID(id string) (Champion, bool) {
	c, ok := s.byID[id]
	return c, ok
}

func (s *Store) Random() Champion {
	return s.list[rand.IntN(len(s.list))]
}
```

- [ ] **Step 5: Run tests — should pass**

```bash
cd backend && go test ./internal/champions/... -v && cd ..
```
Expected: all 4 tests PASS.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/champions backend/go.mod backend/go.sum
git commit -m "feat(champions): add embedded data store with 30 champions"
```

---

## Task 3: Pure game comparison logic

**Files:**
- Create: `backend/internal/game/compare.go`
- Test: `backend/internal/game/compare_test.go`

- [ ] **Step 1: Write the failing tests (table-driven)**

Create `backend/internal/game/compare_test.go`:
```go
package game

import (
	"testing"

	"lolidle/backend/internal/champions"
)

func ch(id string, gender string, positions []string, species, resource, rangeType string, regions []string, year int) champions.Champion {
	return champions.Champion{
		ID: id, Name: id, Gender: gender, Positions: positions, Species: species,
		Resource: resource, RangeType: rangeType, Regions: regions, ReleaseYear: year,
	}
}

func TestCompare_singleAttributes(t *testing.T) {
	target := ch("ahri", "Female", []string{"Mid"}, "Vastayan", "Mana", "Ranged", []string{"Ionia"}, 2011)

	tests := []struct {
		name     string
		guess    champions.Champion
		wantGen  Status
		wantSpec Status
		wantRes  Status
		wantRng  Status
	}{
		{
			name:     "all match",
			guess:    target,
			wantGen:  StatusMatch, wantSpec: StatusMatch, wantRes: StatusMatch, wantRng: StatusMatch,
		},
		{
			name:     "all different",
			guess:    ch("garen", "Male", []string{"Top"}, "Human", "None", "Melee", []string{"Demacia"}, 2010),
			wantGen:  StatusNoMatch, wantSpec: StatusNoMatch, wantRes: StatusNoMatch, wantRng: StatusNoMatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fb, _ := Compare(tt.guess, target)
			if fb.Gender.Status != tt.wantGen {
				t.Errorf("Gender = %s, want %s", fb.Gender.Status, tt.wantGen)
			}
			if fb.Species.Status != tt.wantSpec {
				t.Errorf("Species = %s, want %s", fb.Species.Status, tt.wantSpec)
			}
			if fb.Resource.Status != tt.wantRes {
				t.Errorf("Resource = %s, want %s", fb.Resource.Status, tt.wantRes)
			}
			if fb.RangeType.Status != tt.wantRng {
				t.Errorf("RangeType = %s, want %s", fb.RangeType.Status, tt.wantRng)
			}
		})
	}
}

func TestCompare_multiAttributes(t *testing.T) {
	target := ch("yasuo", "Male", []string{"Mid", "Top"}, "Human", "Flow", "Melee", []string{"Ionia"}, 2013)

	tests := []struct {
		name      string
		guess     champions.Champion
		wantPos   Status
		wantReg   Status
	}{
		{
			name:    "exact match positions and regions",
			guess:   ch("yone", "Male", []string{"Mid", "Top"}, "Spirit", "Flow", "Melee", []string{"Ionia"}, 2020),
			wantPos: StatusMatch, wantReg: StatusMatch,
		},
		{
			name:    "partial positions, exact region",
			guess:   ch("akali", "Female", []string{"Mid"}, "Human", "Energy", "Melee", []string{"Ionia"}, 2010),
			wantPos: StatusPartial, wantReg: StatusMatch,
		},
		{
			name:    "no match positions, no match region",
			guess:   ch("garen", "Male", []string{"Jungle"}, "Human", "None", "Melee", []string{"Demacia"}, 2010),
			wantPos: StatusNoMatch, wantReg: StatusNoMatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fb, _ := Compare(tt.guess, target)
			if fb.Positions.Status != tt.wantPos {
				t.Errorf("Positions = %s, want %s", fb.Positions.Status, tt.wantPos)
			}
			if fb.Regions.Status != tt.wantReg {
				t.Errorf("Regions = %s, want %s", fb.Regions.Status, tt.wantReg)
			}
		})
	}
}

func TestCompare_releaseYear(t *testing.T) {
	target := ch("ahri", "Female", []string{"Mid"}, "Vastayan", "Mana", "Ranged", []string{"Ionia"}, 2011)

	tests := []struct {
		name      string
		guessYear int
		want      Status
	}{
		{"same year", 2011, StatusMatch},
		{"target older than guess", 2020, StatusLower},
		{"target newer than guess", 2009, StatusHigher},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guess := target
			guess.ReleaseYear = tt.guessYear
			fb, _ := Compare(guess, target)
			if fb.ReleaseYear.Status != tt.want {
				t.Errorf("ReleaseYear = %s, want %s", fb.ReleaseYear.Status, tt.want)
			}
		})
	}
}

func TestCompare_correctFlag(t *testing.T) {
	target := ch("ahri", "Female", []string{"Mid"}, "Vastayan", "Mana", "Ranged", []string{"Ionia"}, 2011)

	_, correct := Compare(target, target)
	if !correct {
		t.Error("expected correct=true when guessing the target")
	}

	other := ch("yasuo", "Male", []string{"Mid"}, "Human", "Flow", "Melee", []string{"Ionia"}, 2013)
	_, correct = Compare(other, target)
	if correct {
		t.Error("expected correct=false for different champion")
	}
}
```

- [ ] **Step 2: Run tests — should fail (no compare.go yet)**

```bash
cd backend && go test ./internal/game/... -v && cd ..
```
Expected: FAIL.

- [ ] **Step 3: Implement `compare.go`**

Create `backend/internal/game/compare.go`:
```go
package game

import "lolidle/backend/internal/champions"

type Status string

const (
	StatusMatch   Status = "match"
	StatusPartial Status = "partial"
	StatusNoMatch Status = "nomatch"
	StatusHigher  Status = "higher" // target year > guess year
	StatusLower   Status = "lower"  // target year < guess year
)

type AttributeFeedback struct {
	Status Status `json:"status"`
}

type Feedback struct {
	Gender      AttributeFeedback `json:"gender"`
	Positions   AttributeFeedback `json:"positions"`
	Species     AttributeFeedback `json:"species"`
	Resource    AttributeFeedback `json:"resource"`
	RangeType   AttributeFeedback `json:"rangeType"`
	Regions     AttributeFeedback `json:"regions"`
	ReleaseYear AttributeFeedback `json:"releaseYear"`
}

func Compare(guess, target champions.Champion) (Feedback, bool) {
	fb := Feedback{
		Gender:      compareSingle(guess.Gender, target.Gender),
		Positions:   compareMulti(guess.Positions, target.Positions),
		Species:     compareSingle(guess.Species, target.Species),
		Resource:    compareSingle(guess.Resource, target.Resource),
		RangeType:   compareSingle(guess.RangeType, target.RangeType),
		Regions:     compareMulti(guess.Regions, target.Regions),
		ReleaseYear: compareYear(guess.ReleaseYear, target.ReleaseYear),
	}
	return fb, guess.ID == target.ID
}

func compareSingle(g, t string) AttributeFeedback {
	if g == t {
		return AttributeFeedback{Status: StatusMatch}
	}
	return AttributeFeedback{Status: StatusNoMatch}
}

func compareMulti(g, t []string) AttributeFeedback {
	gs := toSet(g)
	ts := toSet(t)

	if len(gs) == len(ts) {
		equal := true
		for k := range gs {
			if _, ok := ts[k]; !ok {
				equal = false
				break
			}
		}
		if equal {
			return AttributeFeedback{Status: StatusMatch}
		}
	}

	for k := range gs {
		if _, ok := ts[k]; ok {
			return AttributeFeedback{Status: StatusPartial}
		}
	}

	return AttributeFeedback{Status: StatusNoMatch}
}

func compareYear(g, t int) AttributeFeedback {
	switch {
	case g == t:
		return AttributeFeedback{Status: StatusMatch}
	case t > g:
		return AttributeFeedback{Status: StatusHigher}
	default:
		return AttributeFeedback{Status: StatusLower}
	}
}

func toSet(s []string) map[string]struct{} {
	m := make(map[string]struct{}, len(s))
	for _, v := range s {
		m[v] = struct{}{}
	}
	return m
}
```

- [ ] **Step 4: Run tests — should pass**

```bash
cd backend && go test ./internal/game/... -v && cd ..
```
Expected: all PASS.

- [ ] **Step 5: Check coverage**

```bash
cd backend && go test ./internal/game/... -cover && cd ..
```
Expected: ≥ 90% coverage on `internal/game`.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/game
git commit -m "feat(game): add Compare with single, multi, and year feedback"
```

---

## Task 4: In-memory session store with TTL

**Files:**
- Create: `backend/internal/session/store.go`
- Test: `backend/internal/session/store_test.go`

- [ ] **Step 1: Write the failing tests**

Create `backend/internal/session/store_test.go`:
```go
package session

import (
	"testing"
	"time"
)

func TestStore_Create_returnsGameWithUniqueID(t *testing.T) {
	s := NewStore(time.Minute)
	g1 := s.Create("ahri")
	g2 := s.Create("ahri")
	if g1.ID == "" || g2.ID == "" {
		t.Fatal("expected non-empty IDs")
	}
	if g1.ID == g2.ID {
		t.Error("expected unique IDs across Create calls")
	}
	if g1.TargetID != "ahri" {
		t.Errorf("TargetID = %s, want ahri", g1.TargetID)
	}
}

func TestStore_Get_returnsCreatedGame(t *testing.T) {
	s := NewStore(time.Minute)
	g := s.Create("yasuo")

	got, ok := s.Get(g.ID)
	if !ok {
		t.Fatal("expected to find created game")
	}
	if got.TargetID != "yasuo" {
		t.Errorf("TargetID = %s, want yasuo", got.TargetID)
	}
}

func TestStore_Get_returnsFalseForUnknownID(t *testing.T) {
	s := NewStore(time.Minute)
	if _, ok := s.Get("nonexistent"); ok {
		t.Error("expected ok=false for unknown id")
	}
}

func TestStore_Get_returnsFalseForExpiredGame(t *testing.T) {
	s := NewStore(10 * time.Millisecond)
	g := s.Create("ahri")

	time.Sleep(20 * time.Millisecond)

	if _, ok := s.Get(g.ID); ok {
		t.Error("expected expired game to be evicted")
	}
}

func TestStore_recordsAttemptsAndWin(t *testing.T) {
	s := NewStore(time.Minute)
	g := s.Create("ahri")

	g.Attempts++
	g.Attempts++
	g.Won = true

	got, _ := s.Get(g.ID)
	if got.Attempts != 2 {
		t.Errorf("Attempts = %d, want 2", got.Attempts)
	}
	if !got.Won {
		t.Error("expected Won=true")
	}
}
```

- [ ] **Step 2: Run tests — should fail**

```bash
cd backend && go test ./internal/session/... -v && cd ..
```
Expected: FAIL.

- [ ] **Step 3: Implement `store.go`**

Create `backend/internal/session/store.go`:
```go
package session

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type Game struct {
	ID           string
	TargetID     string
	Attempts     int
	Won          bool
	LastAccessed time.Time
}

type Store struct {
	mu    sync.Mutex
	games map[string]*Game
	ttl   time.Duration
}

func NewStore(ttl time.Duration) *Store {
	return &Store{
		games: make(map[string]*Game),
		ttl:   ttl,
	}
}

func (s *Store) Create(targetID string) *Game {
	s.mu.Lock()
	defer s.mu.Unlock()
	g := &Game{
		ID:           newID(),
		TargetID:     targetID,
		LastAccessed: time.Now(),
	}
	s.games[g.ID] = g
	return g
}

func (s *Store) Get(id string) (*Game, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	g, ok := s.games[id]
	if !ok {
		return nil, false
	}
	if time.Since(g.LastAccessed) > s.ttl {
		delete(s.games, id)
		return nil, false
	}
	g.LastAccessed = time.Now()
	return g, true
}

func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
```

- [ ] **Step 4: Run tests — should pass**

```bash
cd backend && go test ./internal/session/... -v && cd ..
```
Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/session
git commit -m "feat(session): in-memory game store with TTL eviction"
```

---

## Task 5: API handlers — `/health` and `/champions`

**Files:**
- Create: `backend/internal/api/handlers.go`
- Test: `backend/internal/api/handlers_test.go`

- [ ] **Step 1: Write failing tests for `/health` and `/champions`**

Create `backend/internal/api/handlers_test.go`:
```go
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"lolidle/backend/internal/champions"
	"lolidle/backend/internal/session"
)

func newHandler(t *testing.T) *Handler {
	t.Helper()
	cs, err := champions.NewStore()
	if err != nil {
		t.Fatalf("champions store: %v", err)
	}
	return &Handler{
		Champions: cs,
		Sessions:  session.NewStore(time.Minute),
	}
}

func TestHealth_returnsOK(t *testing.T) {
	h := newHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rr := httptest.NewRecorder()
	h.Health(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("status field = %s, want ok", body["status"])
	}
}

func TestListChampions_returnsIDAndName(t *testing.T) {
	h := newHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/champions", nil)
	rr := httptest.NewRecorder()
	h.ListChampions(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	var body []map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body) < 20 {
		t.Errorf("expected at least 20 champions, got %d", len(body))
	}
	if body[0]["id"] == "" || body[0]["name"] == "" {
		t.Errorf("expected non-empty id and name, got %+v", body[0])
	}
}
```

- [ ] **Step 2: Run tests — should fail (no handlers.go)**

```bash
cd backend && go test ./internal/api/... -v && cd ..
```
Expected: FAIL.

- [ ] **Step 3: Implement `handlers.go` with the two endpoints**

Create `backend/internal/api/handlers.go`:
```go
package api

import (
	"encoding/json"
	"net/http"

	"lolidle/backend/internal/champions"
	"lolidle/backend/internal/session"
)

type Handler struct {
	Champions *champions.Store
	Sessions  *session.Store
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type championListItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (h *Handler) ListChampions(w http.ResponseWriter, r *http.Request) {
	all := h.Champions.All()
	out := make([]championListItem, 0, len(all))
	for _, c := range all {
		out = append(out, championListItem{ID: c.ID, Name: c.Name})
	}
	writeJSON(w, http.StatusOK, out)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
```

- [ ] **Step 4: Run tests — should pass**

```bash
cd backend && go test ./internal/api/... -v && cd ..
```
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/api
git commit -m "feat(api): /health and /champions endpoints"
```

---

## Task 6: API handlers — `/games` (create) and `/games/:id/guesses` (submit)

**Files:**
- Modify: `backend/internal/api/handlers.go`
- Modify: `backend/internal/api/handlers_test.go`

- [ ] **Step 1: Add failing tests for both endpoints**

Append to `backend/internal/api/handlers_test.go`:
```go

import (
	// (already imported above; add these to the existing import block):
	// "bytes"
	// "github.com/go-chi/chi/v5"
)

func TestCreateGame_returnsGameID(t *testing.T) {
	h := newHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/api/games", nil)
	rr := httptest.NewRecorder()
	h.CreateGame(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", rr.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["gameId"] == "" {
		t.Error("expected non-empty gameId")
	}
}

func TestSubmitGuess_returnsFeedbackForKnownGame(t *testing.T) {
	h := newHandler(t)

	// Create game directly via Sessions to control the target
	g := h.Sessions.Create("ahri")

	body := bytes.NewBufferString(`{"championId":"yasuo"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/games/"+g.ID+"/guesses", body)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("gameId", g.ID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.SubmitGuess(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", rr.Code, rr.Body.String())
	}

	var resp struct {
		Correct      bool `json:"correct"`
		AttemptCount int  `json:"attemptCount"`
		Guess        struct {
			ID string `json:"id"`
		} `json:"guess"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Correct {
		t.Error("expected correct=false (yasuo != ahri)")
	}
	if resp.AttemptCount != 1 {
		t.Errorf("AttemptCount = %d, want 1", resp.AttemptCount)
	}
	if resp.Guess.ID != "yasuo" {
		t.Errorf("Guess.ID = %s, want yasuo", resp.Guess.ID)
	}
}

func TestSubmitGuess_correctChampionWinsGame(t *testing.T) {
	h := newHandler(t)
	g := h.Sessions.Create("ahri")

	body := bytes.NewBufferString(`{"championId":"ahri"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/games/"+g.ID+"/guesses", body)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("gameId", g.ID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.SubmitGuess(rr, req)

	var resp struct{ Correct bool `json:"correct"` }
	json.Unmarshal(rr.Body.Bytes(), &resp)
	if !resp.Correct {
		t.Error("expected correct=true for matching guess")
	}
}

func TestSubmitGuess_returns404ForUnknownGame(t *testing.T) {
	h := newHandler(t)
	body := bytes.NewBufferString(`{"championId":"ahri"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/games/nonexistent/guesses", body)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("gameId", "nonexistent")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.SubmitGuess(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rr.Code)
	}
}

func TestSubmitGuess_returns400ForUnknownChampion(t *testing.T) {
	h := newHandler(t)
	g := h.Sessions.Create("ahri")
	body := bytes.NewBufferString(`{"championId":"nonexistent"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/games/"+g.ID+"/guesses", body)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("gameId", g.ID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.SubmitGuess(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}

func TestSubmitGuess_returns409WhenAlreadyWon(t *testing.T) {
	h := newHandler(t)
	g := h.Sessions.Create("ahri")
	g.Won = true

	body := bytes.NewBufferString(`{"championId":"ahri"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/games/"+g.ID+"/guesses", body)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("gameId", g.ID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.SubmitGuess(rr, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("status = %d, want 409", rr.Code)
	}
}
```

Update the imports at the top of `backend/internal/api/handlers_test.go` to include:
```go
import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"lolidle/backend/internal/champions"
	"lolidle/backend/internal/session"
)
```

- [ ] **Step 2: Run tests — should fail**

```bash
cd backend && go test ./internal/api/... -v && cd ..
```
Expected: FAIL (handlers don't exist yet).

- [ ] **Step 3: Add the two handlers to `handlers.go`**

Append to `backend/internal/api/handlers.go`:
```go

import (
	// add to existing import block:
	// "github.com/go-chi/chi/v5"
	// "lolidle/backend/internal/game"
)

type createGameResponse struct {
	GameID string `json:"gameId"`
}

func (h *Handler) CreateGame(w http.ResponseWriter, r *http.Request) {
	target := h.Champions.Random()
	g := h.Sessions.Create(target.ID)
	writeJSON(w, http.StatusCreated, createGameResponse{GameID: g.ID})
}

type guessRequest struct {
	ChampionID string `json:"championId"`
}

type guessResponse struct {
	Guess        champions.Champion `json:"guess"`
	Feedback     game.Feedback      `json:"feedback"`
	Correct      bool               `json:"correct"`
	AttemptCount int                `json:"attemptCount"`
}

func (h *Handler) SubmitGuess(w http.ResponseWriter, r *http.Request) {
	gameID := chi.URLParam(r, "gameId")

	g, ok := h.Sessions.Get(gameID)
	if !ok {
		writeError(w, http.StatusNotFound, "game not found or expired")
		return
	}
	if g.Won {
		writeError(w, http.StatusConflict, "game already won")
		return
	}

	var req guessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	guess, ok := h.Champions.ByID(req.ChampionID)
	if !ok {
		writeError(w, http.StatusBadRequest, "unknown champion")
		return
	}

	target, _ := h.Champions.ByID(g.TargetID)
	fb, correct := game.Compare(guess, target)
	g.Attempts++
	if correct {
		g.Won = true
	}

	writeJSON(w, http.StatusOK, guessResponse{
		Guess:        guess,
		Feedback:     fb,
		Correct:      correct,
		AttemptCount: g.Attempts,
	})
}
```

Update the imports at the top of `backend/internal/api/handlers.go` to:
```go
import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"lolidle/backend/internal/champions"
	"lolidle/backend/internal/game"
	"lolidle/backend/internal/session"
)
```

- [ ] **Step 4: Run tests — should pass**

```bash
cd backend && go test ./internal/api/... -v && cd ..
```
Expected: all PASS.

- [ ] **Step 5: Run full backend coverage**

```bash
cd backend && go test ./... -cover && cd ..
```
Expected: ≥ 80% coverage on all `internal/` packages.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/api
git commit -m "feat(api): /games create + guess endpoints with full error handling"
```

---

## Task 7: Wire backend `main.go` and run a smoke test

**Files:**
- Create: `backend/cmd/server/main.go`

- [ ] **Step 1: Implement `main.go`**

Create `backend/cmd/server/main.go`:
```go
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"lolidle/backend/internal/api"
	"lolidle/backend/internal/champions"
	"lolidle/backend/internal/session"
)

func main() {
	cs, err := champions.NewStore()
	if err != nil {
		log.Fatalf("load champions: %v", err)
	}
	ss := session.NewStore(30 * time.Minute)

	h := &api.Handler{Champions: cs, Sessions: ss}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)

	r.Get("/api/health", h.Health)
	r.Get("/api/champions", h.ListChampions)
	r.Post("/api/games", h.CreateGame)
	r.Post("/api/games/{gameId}/guesses", h.SubmitGuess)

	addr := ":8080"
	if v := os.Getenv("PORT"); v != "" {
		addr = ":" + v
	}
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	allowed := os.Getenv("CORS_ORIGIN")
	if allowed == "" {
		allowed = "http://localhost:5173"
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", allowed)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
```

- [ ] **Step 2: Build the binary**

```bash
cd backend && go build -o server ./cmd/server && cd ..
```
Expected: no errors, `backend/server` (or `server.exe` on Windows) created.

- [ ] **Step 3: Run the server in the background**

```bash
cd backend && ./server &
sleep 1
```
On Windows: `cd backend && start /b server.exe` or run it in another terminal.

- [ ] **Step 4: Smoke test the endpoints with curl**

```bash
curl -s http://localhost:8080/api/health
# expected: {"status":"ok"}

curl -s http://localhost:8080/api/champions | head -c 200
# expected: JSON array starting with [{"id":"ahri","name":"Ahri"}, ...

GAME=$(curl -s -X POST http://localhost:8080/api/games | python -c "import sys, json; print(json.load(sys.stdin)['gameId'])")
echo "Game: $GAME"

curl -s -X POST http://localhost:8080/api/games/$GAME/guesses \
  -H "Content-Type: application/json" \
  -d '{"championId":"ahri"}'
# expected: full JSON with guess, feedback, correct, attemptCount
```
Expected: all four calls return valid JSON. Stop the server when done.

- [ ] **Step 5: Commit**

```bash
git add backend/cmd
git commit -m "feat(server): wire main with chi router, CORS, and middleware"
```

---

## Task 8: Initialize frontend with Vite + React + TypeScript

**Files:**
- Create: `frontend/package.json`
- Create: `frontend/tsconfig.json`
- Create: `frontend/vite.config.ts`
- Create: `frontend/index.html`
- Create: `frontend/src/main.tsx`

- [ ] **Step 1: Create the project with Vite**

```bash
npm create vite@latest frontend -- --template react-ts
cd frontend && npm install && cd ..
```

- [ ] **Step 2: Add Vitest and React Testing Library**

```bash
cd frontend
npm install -D vitest @vitest/coverage-v8 @testing-library/react @testing-library/jest-dom jsdom
cd ..
```

- [ ] **Step 3: Update `frontend/vite.config.ts` to enable Vitest**

Replace contents with:
```ts
/// <reference types="vitest" />
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  server: { port: 5173 },
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./src/test-setup.ts'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'lcov'],
    },
  },
});
```

- [ ] **Step 4: Create `frontend/src/test-setup.ts`**

```ts
import '@testing-library/jest-dom';
```

- [ ] **Step 5: Update `frontend/package.json` scripts**

Edit the `"scripts"` block to:
```json
"scripts": {
  "dev": "vite",
  "build": "tsc -b && vite build",
  "lint": "eslint .",
  "preview": "vite preview",
  "test": "vitest run",
  "test:watch": "vitest",
  "coverage": "vitest run --coverage"
}
```

- [ ] **Step 6: Verify the scaffold runs**

```bash
cd frontend && npm run build && cd ..
```
Expected: builds without errors.

- [ ] **Step 7: Commit**

```bash
git add frontend
git commit -m "chore(frontend): scaffold Vite + React + TS + Vitest"
```

---

## Task 9: Frontend API client and types

**Files:**
- Create: `frontend/src/api/types.ts`
- Create: `frontend/src/api/client.ts`
- Test: `frontend/src/api/client.test.ts`

- [ ] **Step 1: Create shared types**

Create `frontend/src/api/types.ts`:
```ts
export interface ChampionListItem {
  id: string;
  name: string;
}

export interface Champion {
  id: string;
  name: string;
  gender: string;
  positions: string[];
  species: string;
  resource: string;
  rangeType: string;
  regions: string[];
  releaseYear: number;
}

export type AttributeStatus = 'match' | 'partial' | 'nomatch' | 'higher' | 'lower';

export interface AttributeFeedback {
  status: AttributeStatus;
}

export interface Feedback {
  gender: AttributeFeedback;
  positions: AttributeFeedback;
  species: AttributeFeedback;
  resource: AttributeFeedback;
  rangeType: AttributeFeedback;
  regions: AttributeFeedback;
  releaseYear: AttributeFeedback;
}

export interface CreateGameResponse {
  gameId: string;
}

export interface GuessResponse {
  guess: Champion;
  feedback: Feedback;
  correct: boolean;
  attemptCount: number;
}
```

- [ ] **Step 2: Write the failing client tests**

Create `frontend/src/api/client.test.ts`:
```ts
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createGame, listChampions, submitGuess } from './client';

describe('api client', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn());
  });
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('listChampions GETs /api/champions', async () => {
    (fetch as any).mockResolvedValue({
      ok: true,
      json: async () => [{ id: 'ahri', name: 'Ahri' }],
    });
    const result = await listChampions();
    expect(result).toEqual([{ id: 'ahri', name: 'Ahri' }]);
    expect(fetch).toHaveBeenCalledWith(
      expect.stringContaining('/api/champions'),
      expect.any(Object),
    );
  });

  it('createGame POSTs to /api/games', async () => {
    (fetch as any).mockResolvedValue({
      ok: true,
      json: async () => ({ gameId: 'g123' }),
    });
    const result = await createGame();
    expect(result.gameId).toBe('g123');
    const call = (fetch as any).mock.calls[0];
    expect(call[1].method).toBe('POST');
  });

  it('submitGuess POSTs the championId', async () => {
    (fetch as any).mockResolvedValue({
      ok: true,
      json: async () => ({ correct: true, attemptCount: 1 }),
    });
    await submitGuess('g123', 'ahri');
    const call = (fetch as any).mock.calls[0];
    expect(call[0]).toContain('/api/games/g123/guesses');
    expect(JSON.parse(call[1].body)).toEqual({ championId: 'ahri' });
  });

  it('throws when response is not ok', async () => {
    (fetch as any).mockResolvedValue({
      ok: false,
      status: 404,
      text: async () => 'not found',
    });
    await expect(listChampions()).rejects.toThrow('HTTP 404');
  });
});
```

- [ ] **Step 3: Run tests — should fail (no client.ts)**

```bash
cd frontend && npm test -- src/api && cd ..
```
Expected: FAIL.

- [ ] **Step 4: Implement `client.ts`**

Create `frontend/src/api/client.ts`:
```ts
import type { ChampionListItem, CreateGameResponse, GuessResponse } from './types';

const BASE = (import.meta.env.VITE_API_BASE as string | undefined) ?? 'http://localhost:8080';

async function jsonRequest<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    ...init,
    headers: { 'Content-Type': 'application/json', ...(init?.headers ?? {}) },
  });
  if (!res.ok) {
    const body = await res.text();
    throw new Error(`HTTP ${res.status}: ${body}`);
  }
  return (await res.json()) as T;
}

export function listChampions(): Promise<ChampionListItem[]> {
  return jsonRequest<ChampionListItem[]>('/api/champions');
}

export function createGame(): Promise<CreateGameResponse> {
  return jsonRequest<CreateGameResponse>('/api/games', { method: 'POST' });
}

export function submitGuess(gameId: string, championId: string): Promise<GuessResponse> {
  return jsonRequest<GuessResponse>(`/api/games/${gameId}/guesses`, {
    method: 'POST',
    body: JSON.stringify({ championId }),
  });
}
```

- [ ] **Step 5: Run tests — should pass**

```bash
cd frontend && npm test -- src/api && cd ..
```
Expected: 4 tests PASS.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/api
git commit -m "feat(frontend): typed API client with fetch and tests"
```

---

## Task 10: SearchBox component

**Files:**
- Create: `frontend/src/components/SearchBox.tsx`
- Test: `frontend/src/components/SearchBox.test.tsx`

- [ ] **Step 1: Write failing tests**

Create `frontend/src/components/SearchBox.test.tsx`:
```tsx
import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { SearchBox } from './SearchBox';

const champions = [
  { id: 'ahri', name: 'Ahri' },
  { id: 'akali', name: 'Akali' },
  { id: 'yasuo', name: 'Yasuo' },
];

describe('SearchBox', () => {
  it('shows no suggestions when query is empty', () => {
    render(<SearchBox champions={champions} excludedIds={new Set()} onSelect={() => {}} />);
    expect(screen.queryByRole('option')).toBeNull();
  });

  it('filters champions by prefix (case-insensitive)', () => {
    render(<SearchBox champions={champions} excludedIds={new Set()} onSelect={() => {}} />);
    fireEvent.change(screen.getByRole('textbox'), { target: { value: 'a' } });
    expect(screen.getByText('Ahri')).toBeInTheDocument();
    expect(screen.getByText('Akali')).toBeInTheDocument();
    expect(screen.queryByText('Yasuo')).toBeNull();
  });

  it('excludes already-guessed champions', () => {
    render(
      <SearchBox
        champions={champions}
        excludedIds={new Set(['ahri'])}
        onSelect={() => {}}
      />,
    );
    fireEvent.change(screen.getByRole('textbox'), { target: { value: 'a' } });
    expect(screen.queryByText('Ahri')).toBeNull();
    expect(screen.getByText('Akali')).toBeInTheDocument();
  });

  it('calls onSelect when a suggestion is clicked', () => {
    const onSelect = vi.fn();
    render(<SearchBox champions={champions} excludedIds={new Set()} onSelect={onSelect} />);
    fireEvent.change(screen.getByRole('textbox'), { target: { value: 'yas' } });
    fireEvent.click(screen.getByText('Yasuo'));
    expect(onSelect).toHaveBeenCalledWith('yasuo');
  });
});
```

- [ ] **Step 2: Run — should fail**

```bash
cd frontend && npm test -- SearchBox && cd ..
```
Expected: FAIL.

- [ ] **Step 3: Implement `SearchBox.tsx`**

Create `frontend/src/components/SearchBox.tsx`:
```tsx
import { useMemo, useState } from 'react';
import type { ChampionListItem } from '../api/types';

interface Props {
  champions: ChampionListItem[];
  excludedIds: Set<string>;
  onSelect: (championId: string) => void;
  disabled?: boolean;
}

export function SearchBox({ champions, excludedIds, onSelect, disabled }: Props) {
  const [query, setQuery] = useState('');

  const matches = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return [];
    return champions
      .filter((c) => !excludedIds.has(c.id) && c.name.toLowerCase().startsWith(q))
      .slice(0, 8);
  }, [champions, excludedIds, query]);

  function handleSelect(id: string) {
    onSelect(id);
    setQuery('');
  }

  return (
    <div className="search-box">
      <input
        type="text"
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        placeholder="Buscar campeón..."
        disabled={disabled}
        aria-label="Buscar campeón"
      />
      {matches.length > 0 && (
        <ul role="listbox">
          {matches.map((c) => (
            <li key={c.id} role="option" aria-selected="false" onClick={() => handleSelect(c.id)}>
              {c.name}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
```

- [ ] **Step 4: Run — should pass**

```bash
cd frontend && npm test -- SearchBox && cd ..
```
Expected: 4 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/SearchBox.tsx frontend/src/components/SearchBox.test.tsx
git commit -m "feat(frontend): SearchBox component with prefix autocomplete"
```

---

## Task 11: GuessTable component

**Files:**
- Create: `frontend/src/components/GuessTable.tsx`
- Test: `frontend/src/components/GuessTable.test.tsx`

- [ ] **Step 1: Write failing tests**

Create `frontend/src/components/GuessTable.test.tsx`:
```tsx
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { GuessTable } from './GuessTable';
import type { GuessResponse } from '../api/types';

const sampleGuess: GuessResponse = {
  guess: {
    id: 'yasuo', name: 'Yasuo', gender: 'Male', positions: ['Mid', 'Top'],
    species: 'Human', resource: 'Flow', rangeType: 'Melee', regions: ['Ionia'], releaseYear: 2013,
  },
  feedback: {
    gender: { status: 'match' },
    positions: { status: 'partial' },
    species: { status: 'nomatch' },
    resource: { status: 'nomatch' },
    rangeType: { status: 'match' },
    regions: { status: 'match' },
    releaseYear: { status: 'higher' },
  },
  correct: false,
  attemptCount: 1,
};

describe('GuessTable', () => {
  it('renders nothing when there are no guesses', () => {
    const { container } = render(<GuessTable guesses={[]} />);
    expect(container.querySelector('table')).toBeNull();
  });

  it('renders a row per guess with the champion name', () => {
    render(<GuessTable guesses={[sampleGuess]} />);
    expect(screen.getByText('Yasuo')).toBeInTheDocument();
  });

  it('applies status classes per attribute cell', () => {
    const { container } = render(<GuessTable guesses={[sampleGuess]} />);
    expect(container.querySelectorAll('.cell-match').length).toBeGreaterThan(0);
    expect(container.querySelector('.cell-partial')).not.toBeNull();
    expect(container.querySelector('.cell-nomatch')).not.toBeNull();
  });

  it('shows year with up arrow when status is higher', () => {
    render(<GuessTable guesses={[sampleGuess]} />);
    expect(screen.getByText(/2013.*⬆/)).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run — should fail**

```bash
cd frontend && npm test -- GuessTable && cd ..
```
Expected: FAIL.

- [ ] **Step 3: Implement `GuessTable.tsx`**

Create `frontend/src/components/GuessTable.tsx`:
```tsx
import type { Champion, Feedback, GuessResponse } from '../api/types';

interface Props {
  guesses: GuessResponse[];
}

const ATTRIBUTES: Array<keyof Feedback> = [
  'gender',
  'positions',
  'species',
  'resource',
  'rangeType',
  'regions',
  'releaseYear',
];

const HEADERS: Record<keyof Feedback, string> = {
  gender: 'Gender',
  positions: 'Position',
  species: 'Species',
  resource: 'Resource',
  rangeType: 'Range',
  regions: 'Region',
  releaseYear: 'Year',
};

function cellValue(attr: keyof Feedback, guess: Champion, status: string): string {
  if (attr === 'releaseYear') {
    if (status === 'higher') return `${guess.releaseYear} ⬆️`;
    if (status === 'lower') return `${guess.releaseYear} ⬇️`;
    return String(guess.releaseYear);
  }
  const v = guess[attr] as string | string[];
  return Array.isArray(v) ? v.join(', ') : v;
}

export function GuessTable({ guesses }: Props) {
  if (guesses.length === 0) return null;
  return (
    <table className="guess-table">
      <thead>
        <tr>
          <th>Champion</th>
          {ATTRIBUTES.map((a) => (
            <th key={a}>{HEADERS[a]}</th>
          ))}
        </tr>
      </thead>
      <tbody>
        {guesses.map((g, i) => (
          <tr key={i}>
            <td>{g.guess.name}</td>
            {ATTRIBUTES.map((a) => {
              const status = g.feedback[a].status;
              return (
                <td key={a} className={`cell cell-${status}`}>
                  {cellValue(a, g.guess, status)}
                </td>
              );
            })}
          </tr>
        ))}
      </tbody>
    </table>
  );
}
```

- [ ] **Step 4: Run — should pass**

```bash
cd frontend && npm test -- GuessTable && cd ..
```
Expected: 4 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/GuessTable.tsx frontend/src/components/GuessTable.test.tsx
git commit -m "feat(frontend): GuessTable with colored cells and year arrows"
```

---

## Task 12: WinBanner + App composition + CSS

**Files:**
- Create: `frontend/src/components/WinBanner.tsx`
- Test: `frontend/src/components/WinBanner.test.tsx`
- Create/Modify: `frontend/src/App.tsx`
- Create/Modify: `frontend/src/main.tsx`
- Create: `frontend/src/styles.css`

- [ ] **Step 1: Write failing test for WinBanner**

Create `frontend/src/components/WinBanner.test.tsx`:
```tsx
import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { WinBanner } from './WinBanner';

describe('WinBanner', () => {
  it('shows attempt count and champion name', () => {
    render(<WinBanner attemptCount={5} championName="Ahri" onPlayAgain={() => {}} />);
    expect(screen.getByText(/5 intentos/)).toBeInTheDocument();
    expect(screen.getByText(/Ahri/)).toBeInTheDocument();
  });

  it('calls onPlayAgain when button is clicked', () => {
    const onPlayAgain = vi.fn();
    render(<WinBanner attemptCount={3} championName="Yasuo" onPlayAgain={onPlayAgain} />);
    fireEvent.click(screen.getByRole('button', { name: /jugar de nuevo/i }));
    expect(onPlayAgain).toHaveBeenCalled();
  });
});
```

- [ ] **Step 2: Run — should fail**

```bash
cd frontend && npm test -- WinBanner && cd ..
```
Expected: FAIL.

- [ ] **Step 3: Implement `WinBanner.tsx`**

Create `frontend/src/components/WinBanner.tsx`:
```tsx
interface Props {
  attemptCount: number;
  championName: string;
  onPlayAgain: () => void;
}

export function WinBanner({ attemptCount, championName, onPlayAgain }: Props) {
  return (
    <div className="win-banner">
      <h2>¡Ganaste en {attemptCount} intentos!</h2>
      <p>
        El campeón era <strong>{championName}</strong>.
      </p>
      <button onClick={onPlayAgain}>Jugar de nuevo</button>
    </div>
  );
}
```

- [ ] **Step 4: Run — should pass**

```bash
cd frontend && npm test -- WinBanner && cd ..
```
Expected: 2 tests PASS.

- [ ] **Step 5: Replace `frontend/src/App.tsx`**

```tsx
import { useEffect, useState } from 'react';
import './styles.css';
import { createGame, listChampions, submitGuess } from './api/client';
import type { ChampionListItem, GuessResponse } from './api/types';
import { SearchBox } from './components/SearchBox';
import { GuessTable } from './components/GuessTable';
import { WinBanner } from './components/WinBanner';

export function App() {
  const [champions, setChampions] = useState<ChampionListItem[]>([]);
  const [gameId, setGameId] = useState<string | null>(null);
  const [guesses, setGuesses] = useState<GuessResponse[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    listChampions().then(setChampions).catch((e) => setError(String(e)));
    void startNewGame();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  async function startNewGame() {
    setGuesses([]);
    setError(null);
    try {
      const { gameId } = await createGame();
      setGameId(gameId);
    } catch (e) {
      setError(String(e));
    }
  }

  async function handleGuess(championId: string) {
    if (!gameId) return;
    try {
      const result = await submitGuess(gameId, championId);
      setGuesses((prev) => [...prev, result]);
    } catch (e) {
      setError(String(e));
    }
  }

  const lastGuess = guesses[guesses.length - 1];
  const won = lastGuess?.correct ?? false;
  const guessedIds = new Set(guesses.map((g) => g.guess.id));

  return (
    <div className="app">
      <header>
        <h1>LOLIDLE</h1>
        <p>Adivina el campeón</p>
      </header>
      {error && <div className="error">{error}</div>}
      {!won && (
        <SearchBox
          champions={champions}
          excludedIds={guessedIds}
          onSelect={handleGuess}
          disabled={!gameId}
        />
      )}
      {won && lastGuess && (
        <WinBanner
          attemptCount={lastGuess.attemptCount}
          championName={lastGuess.guess.name}
          onPlayAgain={startNewGame}
        />
      )}
      <GuessTable guesses={guesses} />
    </div>
  );
}
```

- [ ] **Step 6: Replace `frontend/src/main.tsx`**

```tsx
import React from 'react';
import ReactDOM from 'react-dom/client';
import { App } from './App';

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
```

- [ ] **Step 7: Create `frontend/src/styles.css`**

```css
* { box-sizing: border-box; }

body {
  margin: 0;
  font-family: system-ui, -apple-system, sans-serif;
  background: #0e1014;
  color: #e6e6e6;
  min-height: 100vh;
}

.app {
  max-width: 1100px;
  margin: 0 auto;
  padding: 2rem 1rem;
}

header {
  text-align: center;
  margin-bottom: 2rem;
}

header h1 {
  margin: 0;
  font-size: 3rem;
  letter-spacing: 0.15em;
  color: #c8aa6e;
}

header p {
  margin: 0.5rem 0 0;
  color: #888;
}

.search-box {
  position: relative;
  max-width: 480px;
  margin: 0 auto 1.5rem;
}

.search-box input {
  width: 100%;
  padding: 0.75rem 1rem;
  font-size: 1rem;
  background: #1a1d23;
  border: 1px solid #333;
  color: #e6e6e6;
  border-radius: 6px;
  outline: none;
}

.search-box input:focus { border-color: #c8aa6e; }

.search-box ul {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  margin: 4px 0 0;
  padding: 0;
  list-style: none;
  background: #1a1d23;
  border: 1px solid #333;
  border-radius: 6px;
  max-height: 240px;
  overflow-y: auto;
  z-index: 10;
}

.search-box li {
  padding: 0.5rem 1rem;
  cursor: pointer;
}

.search-box li:hover { background: #2a2e36; }

.guess-table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 4px;
  margin-top: 1rem;
}

.guess-table th {
  padding: 0.5rem;
  text-align: center;
  color: #888;
  font-weight: 500;
  font-size: 0.85rem;
  text-transform: uppercase;
}

.guess-table td {
  padding: 0.75rem 0.5rem;
  text-align: center;
  background: #1a1d23;
  border-radius: 4px;
  font-size: 0.9rem;
}

.cell-match { background: #2d6a4f !important; color: white; }
.cell-partial { background: #c08b00 !important; color: white; }
.cell-nomatch { background: #6a1f1f !important; color: white; }
.cell-higher, .cell-lower { background: #6a1f1f !important; color: white; }

.win-banner {
  text-align: center;
  padding: 2rem;
  background: #1a1d23;
  border-radius: 8px;
  margin: 1rem auto;
  max-width: 480px;
}

.win-banner button {
  margin-top: 1rem;
  padding: 0.6rem 1.5rem;
  background: #c8aa6e;
  color: #0e1014;
  border: none;
  border-radius: 4px;
  font-weight: 600;
  cursor: pointer;
  font-size: 1rem;
}

.error {
  background: #6a1f1f;
  color: white;
  padding: 0.75rem 1rem;
  border-radius: 4px;
  margin-bottom: 1rem;
}
```

- [ ] **Step 8: Run all frontend tests**

```bash
cd frontend && npm test && cd ..
```
Expected: all tests PASS.

- [ ] **Step 9: Commit**

```bash
git add frontend/src
git commit -m "feat(frontend): WinBanner, App composition, and styling"
```

---

## Task 13: End-to-end manual smoke test + README

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Run the backend**

In one terminal:
```bash
cd backend && go run ./cmd/server
```
Expected: log line `listening on :8080`.

- [ ] **Step 2: Run the frontend**

In another terminal:
```bash
cd frontend && npm run dev
```
Expected: Vite prints `Local: http://localhost:5173/`.

- [ ] **Step 3: Open the browser and play one full game**

Navigate to `http://localhost:5173`.

Verify:
1. Header "LOLIDLE" renders
2. Search box accepts text and shows autocomplete with up to 8 suggestions
3. Clicking a suggestion submits the guess
4. A row appears in the table with colored cells (mix of green/yellow/red)
5. Year cell shows the year + ⬆️ or ⬇️ when not matching
6. Already-guessed champions disappear from the suggestions
7. Eventually win — WinBanner appears, search box hides
8. Click "Jugar de nuevo" — table clears, new game starts

If anything breaks, fix it before committing. Common issues:
- CORS error in console → check `CORS_ORIGIN` env var on backend matches frontend URL
- 404 on POST /api/games → check backend is running on :8080
- Type error → check `frontend/src/api/types.ts` matches backend JSON shape

- [ ] **Step 4: Update `README.md` with full instructions**

Replace contents:
```markdown
# Lolidle

LoL champion guessing game (Loldle clone, Classic mode, freeplay) built for the EAFIT DevOps final project.

## Stack

- **Backend:** Go 1.22 + chi
- **Frontend:** React 18 + Vite + TypeScript
- **Data:** ~30 curated champions embedded in the Go binary

## Run locally

### Backend

```bash
cd backend
go run ./cmd/server
# listens on :8080, CORS allows http://localhost:5173 by default
```

Override with env vars: `PORT`, `CORS_ORIGIN`.

### Frontend

```bash
cd frontend
npm install
npm run dev
# opens http://localhost:5173
```

## Test

```bash
# Backend
cd backend && go test ./... -cover

# Frontend
cd frontend && npm test
cd frontend && npm run coverage
```

## API endpoints

| Method | Path | Body | Response |
|---|---|---|---|
| `GET` | `/api/health` | — | `{"status":"ok"}` |
| `GET` | `/api/champions` | — | `[{id, name}, ...]` |
| `POST` | `/api/games` | — | `{gameId}` |
| `POST` | `/api/games/:gameId/guesses` | `{championId}` | `{guess, feedback, correct, attemptCount}` |

## Roadmap

- [x] App layer (this plan)
- [ ] CI pipeline (GitHub Actions: build → test → release)
- [ ] CD pipeline (Terraform → AWS, two environments, smoke tests, rollback)
```

- [ ] **Step 5: Commit**

```bash
git add README.md
git commit -m "docs: full local setup and API reference in README"
```

- [ ] **Step 6: Tag the milestone**

```bash
git tag -a v0.1.0-app -m "App layer complete; CI/CD next"
```

---

## Self-review

**Spec coverage:**
- ✅ 2 components (backend Go + frontend React) — Tasks 1–12
- ✅ Embedded `champions.json` with `//go:embed` — Task 2
- ✅ `chi` router + middleware — Task 7
- ✅ In-memory session store with 30-min TTL — Task 4
- ✅ All 4 endpoints (`/health`, `/champions`, `/games`, `/games/:id/guesses`) — Tasks 5, 6
- ✅ 7-attribute comparison with match/partial/nomatch + higher/lower for year — Task 3
- ✅ Stateless API, no auth, freeplay only, no extra hints — Tasks 4, 6
- ✅ React + Vite + TS frontend with autocomplete, colored table, win banner — Tasks 8, 10, 11, 12
- ✅ Tests at every layer with TDD — every task
- ✅ Backend coverage ≥ 80% target — Task 6
- ✅ E2E smoke test (manual) — Task 13
- ✅ README with run instructions — Task 13

**Placeholder scan:** none.

**Type consistency:**
- Go `Status` constants match TS `AttributeStatus` union (match, partial, nomatch, higher, lower) ✓
- Go `Champion` struct fields match TS `Champion` interface ✓
- Go `Feedback` struct keys match TS `Feedback` interface ✓
- Go `guessResponse` JSON keys match TS `GuessResponse` interface ✓
- `gameId` field name consistent across Go and TS ✓

No issues found.

---

## What's next

After this plan ships and the local app works end-to-end, the follow-up plan covers:

- Multi-stage Dockerfiles (`distroless` for Go, `nginx` for the SPA)
- GitHub Actions CI: lint + test + build + release artifacts
- Terraform for AWS (ECR + ECS Fargate or Lambda, ALB, CloudWatch)
- Two environments (dev + prod) with promotion gates
- Pre/Post-prod smoke tests
- Manual rollback (panic button)
- The 3 required pipeline/architecture modifications (TBD: blue/green, canary, manual approvals, security scanning, observability — pick 3)
