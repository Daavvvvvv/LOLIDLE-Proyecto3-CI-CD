# UX Improvements Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add champion portraits, sequential flip-reveal animations, keyboard-navigable autocomplete, win celebration (confetti + scale-in), and overall theme polish on top of `v0.1.0-app`.

**Architecture:** Backend gains an `ImageKey` field on the Champion struct + JSON data. Frontend fetches the latest Data Dragon version at startup (with fallback), threads it through the three game components as a prop, and renders portraits via `<img>` tags pointing at Riot's CDN. Animations are pure CSS keyframes (flip-reveal, scale-in, pulse). Win celebration uses one tiny new dep: `canvas-confetti`.

**Tech Stack:** Go 1.22 (existing), React 19 + TS + Vite + Vitest (existing), `canvas-confetti` (new, ~12KB).

**Spec:** `docs/superpowers/specs/2026-04-23-ux-improvements-design.md`

---

## File Structure

### Backend (modified)
```
backend/internal/champions/store.go         # +ImageKey field on Champion struct
backend/internal/champions/champions.json   # +imageKey on each of 30 entries
backend/internal/champions/store_test.go    # +1 test asserting ImageKey loaded
backend/internal/api/handlers.go            # +ImageKey on championListItem response type
backend/internal/api/handlers_test.go       # +1 test asserting imageKey in /champions response
```

### Frontend (mix of create/modify)
```
frontend/package.json                       # +canvas-confetti dep
frontend/src/api/types.ts                   # +imageKey on Champion + ChampionListItem
frontend/src/api/portrait.ts                # NEW: fetchLatestVersion + getPortraitUrl
frontend/src/api/portrait.test.ts           # NEW: tests for the above
frontend/src/components/SearchBox.tsx       # Rewrite: portraits + keyboard nav + new prop
frontend/src/components/SearchBox.test.tsx  # Update 4 existing + 5 new tests
frontend/src/components/GuessTable.tsx      # Portraits + cell-reveal class + new prop
frontend/src/components/GuessTable.test.tsx # Update 4 existing + 1 new test
frontend/src/components/WinBanner.tsx       # Portrait + confetti + scale-in + new props
frontend/src/components/WinBanner.test.tsx  # Update 2 existing + 2 new tests + mock confetti
frontend/src/App.tsx                        # +version state + fetchLatestVersion + pass down
frontend/src/styles.css                     # Extensive rewrite for theme polish
```

---

## Task 1: Backend — add `ImageKey` field

**Files:**
- Modify: `backend/internal/champions/store.go`
- Modify: `backend/internal/champions/champions.json`
- Modify: `backend/internal/champions/store_test.go`
- Modify: `backend/internal/api/handlers.go`
- Modify: `backend/internal/api/handlers_test.go`

- [ ] **Step 1: Add a failing test in `store_test.go` asserting ImageKey is populated**

Add this function to `backend/internal/champions/store_test.go` (append; keep existing tests intact):
```go
func TestStore_ImageKey_isPopulated(t *testing.T) {
	s, _ := NewStore()
	ahri, ok := s.ByID("ahri")
	if !ok {
		t.Fatal("expected ahri")
	}
	if ahri.ImageKey != "Ahri" {
		t.Errorf("ImageKey = %q, want %q", ahri.ImageKey, "Ahri")
	}
	leeSin, _ := s.ByID("lee-sin")
	if leeSin.ImageKey != "LeeSin" {
		t.Errorf("LeeSin ImageKey = %q, want %q", leeSin.ImageKey, "LeeSin")
	}
}
```

- [ ] **Step 2: Run — should fail (Champion has no ImageKey field yet)**

```bash
cd backend && go test ./internal/champions/... -v && cd ..
```
Expected: compilation error (`ahri.ImageKey undefined`).

- [ ] **Step 3: Add `ImageKey` to the `Champion` struct**

In `backend/internal/champions/store.go`, replace the `Champion` struct definition:
```go
type Champion struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	ImageKey    string   `json:"imageKey"`
	Gender      string   `json:"gender"`
	Positions   []string `json:"positions"`
	Species     string   `json:"species"`
	Resource    string   `json:"resource"`
	RangeType   string   `json:"rangeType"`
	Regions     []string `json:"regions"`
	ReleaseYear int      `json:"releaseYear"`
}
```

- [ ] **Step 4: Replace `backend/internal/champions/champions.json` with the full 30 entries including `imageKey`**

Overwrite the file entirely:
```json
[
  {"id":"ahri","name":"Ahri","imageKey":"Ahri","gender":"Female","positions":["Mid"],"species":"Vastayan","resource":"Mana","rangeType":"Ranged","regions":["Ionia"],"releaseYear":2011},
  {"id":"yasuo","name":"Yasuo","imageKey":"Yasuo","gender":"Male","positions":["Mid","Top"],"species":"Human","resource":"Flow","rangeType":"Melee","regions":["Ionia"],"releaseYear":2013},
  {"id":"garen","name":"Garen","imageKey":"Garen","gender":"Male","positions":["Top"],"species":"Human","resource":"None","rangeType":"Melee","regions":["Demacia"],"releaseYear":2010},
  {"id":"lux","name":"Lux","imageKey":"Lux","gender":"Female","positions":["Mid","Support"],"species":"Human","resource":"Mana","rangeType":"Ranged","regions":["Demacia"],"releaseYear":2010},
  {"id":"darius","name":"Darius","imageKey":"Darius","gender":"Male","positions":["Top"],"species":"Human","resource":"None","rangeType":"Melee","regions":["Noxus"],"releaseYear":2012},
  {"id":"jinx","name":"Jinx","imageKey":"Jinx","gender":"Female","positions":["ADC"],"species":"Human","resource":"Mana","rangeType":"Ranged","regions":["Zaun"],"releaseYear":2013},
  {"id":"vi","name":"Vi","imageKey":"Vi","gender":"Female","positions":["Jungle"],"species":"Human","resource":"Mana","rangeType":"Melee","regions":["Piltover"],"releaseYear":2012},
  {"id":"caitlyn","name":"Caitlyn","imageKey":"Caitlyn","gender":"Female","positions":["ADC"],"species":"Human","resource":"Mana","rangeType":"Ranged","regions":["Piltover"],"releaseYear":2011},
  {"id":"ezreal","name":"Ezreal","imageKey":"Ezreal","gender":"Male","positions":["ADC"],"species":"Human","resource":"Mana","rangeType":"Ranged","regions":["Piltover"],"releaseYear":2010},
  {"id":"lee-sin","name":"Lee Sin","imageKey":"LeeSin","gender":"Male","positions":["Jungle"],"species":"Human","resource":"Energy","rangeType":"Melee","regions":["Ionia"],"releaseYear":2011},
  {"id":"thresh","name":"Thresh","imageKey":"Thresh","gender":"Male","positions":["Support"],"species":"Spirit","resource":"Mana","rangeType":"Melee","regions":["Shadow Isles"],"releaseYear":2013},
  {"id":"senna","name":"Senna","imageKey":"Senna","gender":"Female","positions":["Support","ADC"],"species":"Human","resource":"Mana","rangeType":"Ranged","regions":["Shadow Isles"],"releaseYear":2019},
  {"id":"lucian","name":"Lucian","imageKey":"Lucian","gender":"Male","positions":["ADC"],"species":"Human","resource":"Mana","rangeType":"Ranged","regions":["Demacia"],"releaseYear":2013},
  {"id":"akali","name":"Akali","imageKey":"Akali","gender":"Female","positions":["Mid"],"species":"Human","resource":"Energy","rangeType":"Melee","regions":["Ionia"],"releaseYear":2010},
  {"id":"zed","name":"Zed","imageKey":"Zed","gender":"Male","positions":["Mid"],"species":"Human","resource":"Energy","rangeType":"Melee","regions":["Ionia"],"releaseYear":2012},
  {"id":"teemo","name":"Teemo","imageKey":"Teemo","gender":"Male","positions":["Top"],"species":"Yordle","resource":"Mana","rangeType":"Ranged","regions":["Bandle City"],"releaseYear":2009},
  {"id":"tristana","name":"Tristana","imageKey":"Tristana","gender":"Female","positions":["ADC"],"species":"Yordle","resource":"Mana","rangeType":"Ranged","regions":["Bandle City"],"releaseYear":2009},
  {"id":"veigar","name":"Veigar","imageKey":"Veigar","gender":"Male","positions":["Mid"],"species":"Yordle","resource":"Mana","rangeType":"Ranged","regions":["Bandle City"],"releaseYear":2009},
  {"id":"kassadin","name":"Kassadin","imageKey":"Kassadin","gender":"Male","positions":["Mid"],"species":"Human","resource":"Mana","rangeType":"Melee","regions":["Void"],"releaseYear":2009},
  {"id":"chogath","name":"Cho'Gath","imageKey":"Chogath","gender":"Male","positions":["Top"],"species":"Void","resource":"Mana","rangeType":"Melee","regions":["Void"],"releaseYear":2009},
  {"id":"khazix","name":"Kha'Zix","imageKey":"Khazix","gender":"Male","positions":["Jungle"],"species":"Void","resource":"Mana","rangeType":"Melee","regions":["Void"],"releaseYear":2012},
  {"id":"reksai","name":"Rek'Sai","imageKey":"RekSai","gender":"Female","positions":["Jungle"],"species":"Void","resource":"Fury","rangeType":"Melee","regions":["Void"],"releaseYear":2014},
  {"id":"malphite","name":"Malphite","imageKey":"Malphite","gender":"Male","positions":["Top"],"species":"Golem","resource":"Mana","rangeType":"Melee","regions":["Ixtal"],"releaseYear":2009},
  {"id":"sona","name":"Sona","imageKey":"Sona","gender":"Female","positions":["Support"],"species":"Human","resource":"Mana","rangeType":"Ranged","regions":["Demacia"],"releaseYear":2010},
  {"id":"soraka","name":"Soraka","imageKey":"Soraka","gender":"Female","positions":["Support"],"species":"Celestial","resource":"Mana","rangeType":"Ranged","regions":["Targon"],"releaseYear":2009},
  {"id":"aatrox","name":"Aatrox","imageKey":"Aatrox","gender":"Male","positions":["Top"],"species":"Darkin","resource":"Blood Well","rangeType":"Melee","regions":["Runeterra"],"releaseYear":2013},
  {"id":"kayn","name":"Kayn","imageKey":"Kayn","gender":"Male","positions":["Jungle"],"species":"Human","resource":"Energy","rangeType":"Melee","regions":["Ionia"],"releaseYear":2017},
  {"id":"sett","name":"Sett","imageKey":"Sett","gender":"Male","positions":["Top"],"species":"Vastayan","resource":"Grit","rangeType":"Melee","regions":["Ionia"],"releaseYear":2020},
  {"id":"yone","name":"Yone","imageKey":"Yone","gender":"Male","positions":["Mid","Top"],"species":"Spirit","resource":"Flow","rangeType":"Melee","regions":["Ionia"],"releaseYear":2020},
  {"id":"aphelios","name":"Aphelios","imageKey":"Aphelios","gender":"Male","positions":["ADC"],"species":"Human","resource":"Mana","rangeType":"Ranged","regions":["Targon"],"releaseYear":2019}
]
```

- [ ] **Step 5: Run champions tests — should now pass**

```bash
cd backend && go test ./internal/champions/... -v && cd ..
```
Expected: 5 tests pass including the new `TestStore_ImageKey_isPopulated`.

- [ ] **Step 6: Add a failing API test asserting `/api/champions` response includes `imageKey`**

Append to `backend/internal/api/handlers_test.go`:
```go
func TestListChampions_includesImageKey(t *testing.T) {
	h := newHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/champions", nil)
	rr := httptest.NewRecorder()
	h.ListChampions(rr, req)

	var body []map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body[0]["imageKey"] == "" {
		t.Errorf("expected non-empty imageKey on first entry, got %+v", body[0])
	}
}
```

- [ ] **Step 7: Run API tests — should fail (championListItem doesn't expose ImageKey yet)**

```bash
cd backend && go test ./internal/api/... -v -run TestListChampions_includesImageKey && cd ..
```
Expected: FAIL (the field will just be "" because it's not on the struct).

- [ ] **Step 8: Add `ImageKey` to `championListItem` in `handlers.go`**

In `backend/internal/api/handlers.go`, replace the `championListItem` struct and the `ListChampions` method:
```go
type championListItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ImageKey string `json:"imageKey"`
}

func (h *Handler) ListChampions(w http.ResponseWriter, r *http.Request) {
	all := h.Champions.All()
	out := make([]championListItem, 0, len(all))
	for _, c := range all {
		out = append(out, championListItem{ID: c.ID, Name: c.Name, ImageKey: c.ImageKey})
	}
	writeJSON(w, http.StatusOK, out)
}
```

- [ ] **Step 9: Run all backend tests — all should pass**

```bash
cd backend && go test ./... -cover && cd ..
```
Expected: all packages PASS, coverage still ≥ 80% across `internal/`.

- [ ] **Step 10: Commit**

```bash
git add backend/internal/champions backend/internal/api
git commit -m "feat(backend): add ImageKey field for Data Dragon portrait URLs"
```

---

## Task 2: Frontend — types + portrait helper

**Files:**
- Modify: `frontend/src/api/types.ts`
- Create: `frontend/src/api/portrait.ts`
- Create: `frontend/src/api/portrait.test.ts`

- [ ] **Step 1: Update `frontend/src/api/types.ts` to add `imageKey`**

Replace contents:
```ts
export interface ChampionListItem {
  id: string;
  name: string;
  imageKey: string;
}

export interface Champion {
  id: string;
  name: string;
  imageKey: string;
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

- [ ] **Step 2: Write failing tests for `portrait.ts`**

Create `frontend/src/api/portrait.test.ts`:
```ts
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { fetchLatestVersion, getPortraitUrl, FALLBACK_VERSION } from './portrait';

describe('portrait', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn());
  });
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('fetchLatestVersion returns n.champion when the realm call succeeds', async () => {
    (fetch as any).mockResolvedValue({
      ok: true,
      json: async () => ({ n: { champion: '14.24.1' } }),
    });
    expect(await fetchLatestVersion()).toBe('14.24.1');
  });

  it('fetchLatestVersion returns the fallback when fetch rejects', async () => {
    (fetch as any).mockRejectedValue(new Error('network down'));
    expect(await fetchLatestVersion()).toBe(FALLBACK_VERSION);
  });

  it('fetchLatestVersion returns the fallback when response is not ok', async () => {
    (fetch as any).mockResolvedValue({ ok: false, status: 500, json: async () => ({}) });
    expect(await fetchLatestVersion()).toBe(FALLBACK_VERSION);
  });

  it('fetchLatestVersion returns the fallback when n.champion is missing', async () => {
    (fetch as any).mockResolvedValue({ ok: true, json: async () => ({ n: {} }) });
    expect(await fetchLatestVersion()).toBe(FALLBACK_VERSION);
  });

  it('getPortraitUrl builds the expected Data Dragon URL', () => {
    expect(getPortraitUrl('14.24.1', 'Ahri')).toBe(
      'https://ddragon.leagueoflegends.com/cdn/14.24.1/img/champion/Ahri.png',
    );
  });
});
```

- [ ] **Step 3: Run tests — should fail (portrait.ts doesn't exist)**

```bash
cd frontend && npm test -- portrait && cd ..
```
Expected: FAIL with "Failed to resolve import './portrait'".

- [ ] **Step 4: Implement `portrait.ts`**

Create `frontend/src/api/portrait.ts`:
```ts
export const FALLBACK_VERSION = '14.24.1';
const REALM_URL = 'https://ddragon.leagueoflegends.com/realms/na.json';

export async function fetchLatestVersion(): Promise<string> {
  try {
    const res = await fetch(REALM_URL);
    if (!res.ok) return FALLBACK_VERSION;
    const data = await res.json();
    const version = data?.n?.champion;
    return typeof version === 'string' && version.length > 0 ? version : FALLBACK_VERSION;
  } catch {
    return FALLBACK_VERSION;
  }
}

export function getPortraitUrl(version: string, imageKey: string): string {
  return `https://ddragon.leagueoflegends.com/cdn/${version}/img/champion/${imageKey}.png`;
}
```

- [ ] **Step 5: Run tests — 5 should pass**

```bash
cd frontend && npm test -- portrait && cd ..
```
Expected: 5 tests PASS.

- [ ] **Step 6: Confirm the updated types compile**

```bash
cd frontend && npx tsc --noEmit && cd ..
```
Expected: no errors. (The existing test fixture in GuessTable.test.tsx uses a `Champion` literal without `imageKey` now — that's fine, TypeScript only checks when code references this field. If TS complains, we fix it in the dedicated task for that component.)

If `tsc --noEmit` errors because existing test fixtures lack `imageKey`, that's expected and will be fixed in Tasks 3-5 when those files get rewritten. Proceed; the build script will re-run cleanly at the end of the plan.

- [ ] **Step 7: Commit**

```bash
git add frontend/src/api
git commit -m "feat(frontend): add imageKey types and Data Dragon version helper"
```

---

## Task 3: SearchBox — keyboard navigation + portraits

**Files:**
- Modify: `frontend/src/components/SearchBox.tsx`
- Modify: `frontend/src/components/SearchBox.test.tsx`

- [ ] **Step 1: Rewrite `SearchBox.test.tsx` with existing + new tests**

Overwrite `frontend/src/components/SearchBox.test.tsx`:
```tsx
import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { SearchBox } from './SearchBox';

const champions = [
  { id: 'ahri', name: 'Ahri', imageKey: 'Ahri' },
  { id: 'akali', name: 'Akali', imageKey: 'Akali' },
  { id: 'aatrox', name: 'Aatrox', imageKey: 'Aatrox' },
  { id: 'yasuo', name: 'Yasuo', imageKey: 'Yasuo' },
];

const defaultProps = {
  champions,
  excludedIds: new Set<string>(),
  onSelect: () => {},
  version: '14.24.1',
};

describe('SearchBox — filtering', () => {
  it('shows no suggestions when query is empty', () => {
    render(<SearchBox {...defaultProps} />);
    expect(screen.queryByRole('option')).toBeNull();
  });

  it('filters champions by prefix (case-insensitive)', () => {
    render(<SearchBox {...defaultProps} />);
    fireEvent.change(screen.getByRole('textbox'), { target: { value: 'a' } });
    expect(screen.getByText('Ahri')).toBeInTheDocument();
    expect(screen.getByText('Akali')).toBeInTheDocument();
    expect(screen.getByText('Aatrox')).toBeInTheDocument();
    expect(screen.queryByText('Yasuo')).toBeNull();
  });

  it('excludes already-guessed champions', () => {
    render(<SearchBox {...defaultProps} excludedIds={new Set(['ahri'])} />);
    fireEvent.change(screen.getByRole('textbox'), { target: { value: 'a' } });
    expect(screen.queryByText('Ahri')).toBeNull();
    expect(screen.getByText('Akali')).toBeInTheDocument();
  });

  it('calls onSelect when a suggestion is clicked', () => {
    const onSelect = vi.fn();
    render(<SearchBox {...defaultProps} onSelect={onSelect} />);
    fireEvent.change(screen.getByRole('textbox'), { target: { value: 'yas' } });
    fireEvent.click(screen.getByText('Yasuo'));
    expect(onSelect).toHaveBeenCalledWith('yasuo');
  });
});

describe('SearchBox — portraits', () => {
  it('renders a portrait img for each suggestion', () => {
    render(<SearchBox {...defaultProps} />);
    fireEvent.change(screen.getByRole('textbox'), { target: { value: 'a' } });
    const imgs = screen.getAllByRole('img');
    expect(imgs.length).toBe(3);
    expect(imgs[0].getAttribute('src')).toContain('14.24.1/img/champion/Ahri.png');
  });
});

describe('SearchBox — keyboard navigation', () => {
  it('Enter selects the first option when nothing is highlighted', () => {
    const onSelect = vi.fn();
    render(<SearchBox {...defaultProps} onSelect={onSelect} />);
    const input = screen.getByRole('textbox');
    fireEvent.change(input, { target: { value: 'a' } });
    fireEvent.keyDown(input, { key: 'Enter' });
    expect(onSelect).toHaveBeenCalledWith('ahri');
  });

  it('ArrowDown then Enter selects the second option', () => {
    const onSelect = vi.fn();
    render(<SearchBox {...defaultProps} onSelect={onSelect} />);
    const input = screen.getByRole('textbox');
    fireEvent.change(input, { target: { value: 'a' } });
    fireEvent.keyDown(input, { key: 'ArrowDown' });
    fireEvent.keyDown(input, { key: 'ArrowDown' });
    fireEvent.keyDown(input, { key: 'Enter' });
    expect(onSelect).toHaveBeenCalledWith('akali');
  });

  it('ArrowDown cycles back to the first option at the end', () => {
    const onSelect = vi.fn();
    render(<SearchBox {...defaultProps} onSelect={onSelect} />);
    const input = screen.getByRole('textbox');
    fireEvent.change(input, { target: { value: 'a' } });
    // 3 matches total (ahri, akali, aatrox). Press Down 4 times: 0 → 1 → 2 → 0
    fireEvent.keyDown(input, { key: 'ArrowDown' });
    fireEvent.keyDown(input, { key: 'ArrowDown' });
    fireEvent.keyDown(input, { key: 'ArrowDown' });
    fireEvent.keyDown(input, { key: 'ArrowDown' });
    fireEvent.keyDown(input, { key: 'Enter' });
    expect(onSelect).toHaveBeenCalledWith('ahri');
  });

  it('ArrowUp from null highlights the last option', () => {
    const onSelect = vi.fn();
    render(<SearchBox {...defaultProps} onSelect={onSelect} />);
    const input = screen.getByRole('textbox');
    fireEvent.change(input, { target: { value: 'a' } });
    fireEvent.keyDown(input, { key: 'ArrowUp' });
    fireEvent.keyDown(input, { key: 'Enter' });
    expect(onSelect).toHaveBeenCalledWith('aatrox');
  });

  it('Escape clears the query and closes the dropdown', () => {
    render(<SearchBox {...defaultProps} />);
    const input = screen.getByRole('textbox') as HTMLInputElement;
    fireEvent.change(input, { target: { value: 'a' } });
    expect(screen.getAllByRole('option').length).toBeGreaterThan(0);
    fireEvent.keyDown(input, { key: 'Escape' });
    expect(input.value).toBe('');
    expect(screen.queryByRole('option')).toBeNull();
  });
});
```

- [ ] **Step 2: Run SearchBox tests — should fail (version prop not recognized; keyboard handlers not implemented)**

```bash
cd frontend && npm test -- SearchBox && cd ..
```
Expected: FAIL — either TS errors about unknown `version` prop or runtime errors from keyboard tests.

- [ ] **Step 3: Rewrite `SearchBox.tsx`**

Overwrite `frontend/src/components/SearchBox.tsx`:
```tsx
import { useEffect, useMemo, useRef, useState } from 'react';
import type { ChampionListItem } from '../api/types';
import { getPortraitUrl } from '../api/portrait';

interface Props {
  champions: ChampionListItem[];
  excludedIds: Set<string>;
  onSelect: (championId: string) => void;
  disabled?: boolean;
  version: string;
}

export function SearchBox({ champions, excludedIds, onSelect, disabled, version }: Props) {
  const [query, setQuery] = useState('');
  const [highlightedIndex, setHighlightedIndex] = useState<number | null>(null);
  const listRef = useRef<HTMLUListElement | null>(null);

  const matches = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return [];
    return champions
      .filter((c) => !excludedIds.has(c.id) && c.name.toLowerCase().startsWith(q))
      .slice(0, 8);
  }, [champions, excludedIds, query]);

  useEffect(() => {
    setHighlightedIndex(null);
  }, [query]);

  useEffect(() => {
    if (highlightedIndex === null || !listRef.current) return;
    const el = listRef.current.children[highlightedIndex] as HTMLElement | undefined;
    el?.scrollIntoView({ block: 'nearest' });
  }, [highlightedIndex]);

  function handleSelect(id: string) {
    onSelect(id);
    setQuery('');
    setHighlightedIndex(null);
  }

  function handleKeyDown(e: React.KeyboardEvent<HTMLInputElement>) {
    if (matches.length === 0) return;
    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault();
        setHighlightedIndex((i) => (i === null ? 0 : (i + 1) % matches.length));
        break;
      case 'ArrowUp':
        e.preventDefault();
        setHighlightedIndex((i) =>
          i === null ? matches.length - 1 : (i - 1 + matches.length) % matches.length,
        );
        break;
      case 'Enter':
        e.preventDefault();
        handleSelect(matches[highlightedIndex ?? 0].id);
        break;
      case 'Escape':
        e.preventDefault();
        setQuery('');
        setHighlightedIndex(null);
        break;
    }
  }

  return (
    <div className="search-box">
      <input
        type="text"
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        onKeyDown={handleKeyDown}
        placeholder="Buscar campeón..."
        disabled={disabled}
        aria-label="Buscar campeón"
      />
      {matches.length > 0 && (
        <ul role="listbox" ref={listRef}>
          {matches.map((c, idx) => {
            const isHighlighted = idx === highlightedIndex;
            return (
              <li
                key={c.id}
                role="option"
                aria-selected={isHighlighted}
                className={isHighlighted ? 'highlighted' : ''}
                onMouseEnter={() => setHighlightedIndex(idx)}
                onClick={() => handleSelect(c.id)}
              >
                <img
                  className="option-portrait"
                  src={getPortraitUrl(version, c.imageKey)}
                  alt=""
                  width={32}
                  height={32}
                  loading="lazy"
                />
                <span>{c.name}</span>
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}
```

- [ ] **Step 4: Run SearchBox tests — all 10 should pass**

```bash
cd frontend && npm test -- SearchBox && cd ..
```
Expected: 10 tests PASS (4 filtering + 1 portraits + 5 keyboard).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/SearchBox.tsx frontend/src/components/SearchBox.test.tsx
git commit -m "feat(frontend): SearchBox keyboard nav + portraits"
```

---

## Task 4: GuessTable — portraits + sequential flip-reveal

**Files:**
- Modify: `frontend/src/components/GuessTable.tsx`
- Modify: `frontend/src/components/GuessTable.test.tsx`

- [ ] **Step 1: Rewrite `GuessTable.test.tsx` with existing (updated) + new test**

Overwrite `frontend/src/components/GuessTable.test.tsx`:
```tsx
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { GuessTable } from './GuessTable';
import type { GuessResponse } from '../api/types';

const sampleGuess: GuessResponse = {
  guess: {
    id: 'yasuo',
    name: 'Yasuo',
    imageKey: 'Yasuo',
    gender: 'Male',
    positions: ['Mid', 'Top'],
    species: 'Human',
    resource: 'Flow',
    rangeType: 'Melee',
    regions: ['Ionia'],
    releaseYear: 2013,
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

const defaultProps = { version: '14.24.1' };

describe('GuessTable', () => {
  it('renders nothing when there are no guesses', () => {
    const { container } = render(<GuessTable guesses={[]} {...defaultProps} />);
    expect(container.querySelector('table')).toBeNull();
  });

  it('renders a row per guess with the champion name', () => {
    render(<GuessTable guesses={[sampleGuess]} {...defaultProps} />);
    expect(screen.getByText('Yasuo')).toBeInTheDocument();
  });

  it('applies status classes per attribute cell', () => {
    const { container } = render(<GuessTable guesses={[sampleGuess]} {...defaultProps} />);
    expect(container.querySelectorAll('.cell-match').length).toBeGreaterThan(0);
    expect(container.querySelector('.cell-partial')).not.toBeNull();
    expect(container.querySelector('.cell-nomatch')).not.toBeNull();
  });

  it('shows year with up arrow when status is higher', () => {
    render(<GuessTable guesses={[sampleGuess]} {...defaultProps} />);
    expect(screen.getByText(/2013.*⬆/)).toBeInTheDocument();
  });

  it('renders a portrait img in the first column for each guess', () => {
    render(<GuessTable guesses={[sampleGuess]} {...defaultProps} />);
    const img = screen.getByRole('img');
    expect(img.getAttribute('src')).toContain('14.24.1/img/champion/Yasuo.png');
  });
});
```

- [ ] **Step 2: Run — should fail (version prop + portrait not implemented)**

```bash
cd frontend && npm test -- GuessTable && cd ..
```
Expected: FAIL.

- [ ] **Step 3: Rewrite `GuessTable.tsx`**

Overwrite `frontend/src/components/GuessTable.tsx`:
```tsx
import type { Champion, Feedback, GuessResponse } from '../api/types';
import { getPortraitUrl } from '../api/portrait';

interface Props {
  guesses: GuessResponse[];
  version: string;
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

export function GuessTable({ guesses, version }: Props) {
  if (guesses.length === 0) return null;
  const newestIndex = guesses.length - 1;
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
        {guesses.map((g, rowIdx) => {
          const isNewest = rowIdx === newestIndex;
          const revealClass = isNewest ? ' cell-reveal' : '';
          return (
            <tr key={rowIdx}>
              <td
                className={`cell-portrait${revealClass}`}
                style={isNewest ? { animationDelay: '0ms' } : undefined}
              >
                <img
                  src={getPortraitUrl(version, g.guess.imageKey)}
                  alt={g.guess.name}
                  width={56}
                  height={56}
                  loading="lazy"
                />
                <div className="champ-name">{g.guess.name}</div>
              </td>
              {ATTRIBUTES.map((a, colIdx) => {
                const status = g.feedback[a].status;
                return (
                  <td
                    key={a}
                    className={`cell cell-${status}${revealClass}`}
                    style={isNewest ? { animationDelay: `${(colIdx + 1) * 120}ms` } : undefined}
                  >
                    {cellValue(a, g.guess, status)}
                  </td>
                );
              })}
            </tr>
          );
        })}
      </tbody>
    </table>
  );
}
```

- [ ] **Step 4: Run GuessTable tests — all 5 should pass**

```bash
cd frontend && npm test -- GuessTable && cd ..
```
Expected: 5 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/GuessTable.tsx frontend/src/components/GuessTable.test.tsx
git commit -m "feat(frontend): GuessTable portrait column + sequential flip class"
```

---

## Task 5: WinBanner — confetti + portrait + scale-in

**Files:**
- Modify: `frontend/package.json` (install dep)
- Modify: `frontend/src/components/WinBanner.tsx`
- Modify: `frontend/src/components/WinBanner.test.tsx`

- [ ] **Step 1: Install `canvas-confetti`**

```bash
cd frontend && npm install canvas-confetti && npm install -D @types/canvas-confetti && cd ..
```
Expected: package added. `package.json` gains `"canvas-confetti": "^1.x"` and `"@types/canvas-confetti": "^1.x"` in devDependencies.

- [ ] **Step 2: Rewrite `WinBanner.test.tsx`**

Overwrite `frontend/src/components/WinBanner.test.tsx`:
```tsx
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { WinBanner } from './WinBanner';

const confettiMock = vi.fn();
vi.mock('canvas-confetti', () => ({
  default: (...args: unknown[]) => confettiMock(...args),
}));

const defaultProps = {
  attemptCount: 5,
  championName: 'Ahri',
  imageKey: 'Ahri',
  version: '14.24.1',
  onPlayAgain: () => {},
};

describe('WinBanner', () => {
  beforeEach(() => {
    confettiMock.mockClear();
  });

  it('shows attempt count and champion name', () => {
    render(<WinBanner {...defaultProps} />);
    expect(screen.getByText(/5 intentos/)).toBeInTheDocument();
    expect(screen.getByText(/Ahri/)).toBeInTheDocument();
  });

  it('calls onPlayAgain when button is clicked', () => {
    const onPlayAgain = vi.fn();
    render(<WinBanner {...defaultProps} attemptCount={3} championName="Yasuo" onPlayAgain={onPlayAgain} />);
    fireEvent.click(screen.getByRole('button', { name: /jugar de nuevo/i }));
    expect(onPlayAgain).toHaveBeenCalled();
  });

  it('renders a portrait img with the imageKey and version', () => {
    render(<WinBanner {...defaultProps} />);
    const img = screen.getByRole('img');
    expect(img.getAttribute('src')).toContain('14.24.1/img/champion/Ahri.png');
  });

  it('fires confetti on mount', () => {
    render(<WinBanner {...defaultProps} />);
    expect(confettiMock).toHaveBeenCalledTimes(1);
    const opts = confettiMock.mock.calls[0][0];
    expect(opts.particleCount).toBeGreaterThan(0);
  });
});
```

- [ ] **Step 3: Run WinBanner tests — should fail (imageKey/version props not accepted; confetti not called)**

```bash
cd frontend && npm test -- WinBanner && cd ..
```
Expected: FAIL.

- [ ] **Step 4: Rewrite `WinBanner.tsx`**

Overwrite `frontend/src/components/WinBanner.tsx`:
```tsx
import { useEffect } from 'react';
import confetti from 'canvas-confetti';
import { getPortraitUrl } from '../api/portrait';

interface Props {
  attemptCount: number;
  championName: string;
  imageKey: string;
  version: string;
  onPlayAgain: () => void;
}

export function WinBanner({ attemptCount, championName, imageKey, version, onPlayAgain }: Props) {
  useEffect(() => {
    confetti({
      particleCount: 100,
      spread: 70,
      origin: { y: 0.6 },
      colors: ['#c8aa6e', '#f0e6d2', '#3d8bff'],
    });
  }, []);

  return (
    <div className="win-banner appearing">
      <img
        className="target-portrait"
        src={getPortraitUrl(version, imageKey)}
        alt={championName}
        width={180}
        height={180}
      />
      <h2>¡Ganaste en {attemptCount} intentos!</h2>
      <p>
        El campeón era <strong>{championName}</strong>.
      </p>
      <button onClick={onPlayAgain}>Jugar de nuevo</button>
    </div>
  );
}
```

- [ ] **Step 5: Run WinBanner tests — all 4 should pass**

```bash
cd frontend && npm test -- WinBanner && cd ..
```
Expected: 4 tests PASS.

- [ ] **Step 6: Commit**

```bash
git add frontend/package.json frontend/package-lock.json frontend/src/components/WinBanner.tsx frontend/src/components/WinBanner.test.tsx
git commit -m "feat(frontend): WinBanner with portrait, scale-in class, and confetti"
```

---

## Task 6: App.tsx wiring + full CSS rewrite + final integration

**Files:**
- Modify: `frontend/src/App.tsx`
- Modify: `frontend/src/styles.css`

- [ ] **Step 1: Update `App.tsx` to fetch version and thread it through**

Overwrite `frontend/src/App.tsx`:
```tsx
import { useEffect, useState } from 'react';
import './styles.css';
import { createGame, listChampions, submitGuess } from './api/client';
import { FALLBACK_VERSION, fetchLatestVersion } from './api/portrait';
import type { ChampionListItem, GuessResponse } from './api/types';
import { SearchBox } from './components/SearchBox';
import { GuessTable } from './components/GuessTable';
import { WinBanner } from './components/WinBanner';

export function App() {
  const [champions, setChampions] = useState<ChampionListItem[]>([]);
  const [version, setVersion] = useState<string>(FALLBACK_VERSION);
  const [gameId, setGameId] = useState<string | null>(null);
  const [guesses, setGuesses] = useState<GuessResponse[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchLatestVersion().then(setVersion);
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
          version={version}
        />
      )}
      {won && lastGuess && (
        <WinBanner
          attemptCount={lastGuess.attemptCount}
          championName={lastGuess.guess.name}
          imageKey={lastGuess.guess.imageKey}
          version={version}
          onPlayAgain={startNewGame}
        />
      )}
      <GuessTable guesses={guesses} version={version} />
    </div>
  );
}
```

- [ ] **Step 2: Overwrite `frontend/src/styles.css` with the full polish styles**

```css
* { box-sizing: border-box; }

body {
  margin: 0;
  font-family: system-ui, -apple-system, sans-serif;
  background:
    radial-gradient(ellipse at top, #1a1d2e 0%, #0e1014 60%);
  background-attachment: fixed;
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
  text-shadow: 0 0 20px rgba(200, 170, 110, 0.4);
}

header p {
  margin: 0.5rem 0 0;
  color: #888;
}

/* --- Search box --- */
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
  transition: box-shadow 150ms ease, border-color 150ms ease;
}

.search-box input:focus {
  border-color: #c8aa6e;
  box-shadow: 0 0 0 2px rgba(200, 170, 110, 0.35);
}

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
  max-height: 320px;
  overflow-y: auto;
  z-index: 10;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);
}

.search-box li {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 0.5rem 1rem;
  cursor: pointer;
  border-left: 3px solid transparent;
  transition: background 100ms ease, border-color 100ms ease;
}

.search-box li.highlighted,
.search-box li:hover {
  background: #2a2e36;
  border-left-color: #c8aa6e;
}

.search-box li img.option-portrait {
  width: 32px;
  height: 32px;
  border-radius: 4px;
  object-fit: cover;
  flex-shrink: 0;
}

/* --- Guess table --- */
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
  letter-spacing: 0.05em;
}

.guess-table td {
  padding: 0.5rem;
  text-align: center;
  background: #1a1d23;
  border-radius: 4px;
  font-size: 0.9rem;
  min-height: 72px;
  vertical-align: middle;
}

.guess-table td.cell-portrait {
  padding: 0.25rem;
}

.guess-table td.cell-portrait img {
  display: block;
  width: 56px;
  height: 56px;
  border-radius: 4px;
  margin: 0 auto;
  object-fit: cover;
}

.guess-table td.cell-portrait .champ-name {
  font-size: 0.8rem;
  margin-top: 4px;
  color: #e6e6e6;
}

.cell-match { background: #2d6a4f !important; color: white; }
.cell-partial { background: #c08b00 !important; color: white; }
.cell-nomatch { background: #6a1f1f !important; color: white; }
.cell-higher, .cell-lower { background: #6a1f1f !important; color: white; }

/* --- Flip reveal animation (newest row only) --- */
.cell-reveal {
  animation: flip-reveal 500ms ease-out both;
  backface-visibility: hidden;
  transform-style: preserve-3d;
}

@keyframes flip-reveal {
  0%   { transform: rotateX(0deg);   }
  50%  { transform: rotateX(90deg);  }
  100% { transform: rotateX(0deg);   }
}

/* --- Win banner --- */
.win-banner {
  text-align: center;
  padding: 2rem;
  background: #1a1d23;
  border-radius: 8px;
  margin: 1rem auto;
  max-width: 480px;
  box-shadow: 0 0 40px rgba(200, 170, 110, 0.2);
}

.win-banner.appearing {
  animation: scale-in 350ms cubic-bezier(0.34, 1.56, 0.64, 1) both;
}

@keyframes scale-in {
  from { transform: scale(0.7); opacity: 0; }
  to   { transform: scale(1);   opacity: 1; }
}

.win-banner img.target-portrait {
  width: 180px;
  height: 180px;
  border: 4px solid #c8aa6e;
  border-radius: 8px;
  margin-bottom: 1rem;
  box-shadow: 0 0 30px rgba(200, 170, 110, 0.45);
  object-fit: cover;
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
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%   { box-shadow: 0 0 0 0   rgba(200, 170, 110, 0.7); }
  70%  { box-shadow: 0 0 0 12px rgba(200, 170, 110, 0);   }
  100% { box-shadow: 0 0 0 0   rgba(200, 170, 110, 0);    }
}

/* --- Error banner --- */
.error {
  background: #6a1f1f;
  color: white;
  padding: 0.75rem 1rem;
  border-radius: 4px;
  margin-bottom: 1rem;
}
```

- [ ] **Step 3: Run full frontend test suite — all should pass**

```bash
cd frontend && npm test && cd ..
```
Expected: `Test Files 5 passed (5)`, total count should be `4 (client) + 5 (portrait) + 10 (SearchBox) + 5 (GuessTable) + 4 (WinBanner) = 28 tests PASS`.

- [ ] **Step 4: Verify production build still compiles**

```bash
cd frontend && npm run build && cd ..
```
Expected: `✓ built in ...ms`, no errors.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/App.tsx frontend/src/styles.css
git commit -m "feat(frontend): App wiring for version + full theme polish CSS"
```

- [ ] **Step 6: Rebuild backend binary and restart servers for manual smoke**

Kill any running backend and restart with the new Champion struct compiled in:
```bash
taskkill //F //IM server.exe 2>/dev/null || true
cd backend && go build -o server.exe ./cmd/server && cd ..
```

Start backend (use `run_in_background: true`):
```bash
cd backend && PORT=8080 ./server.exe
```

Start frontend dev server (use `run_in_background: true`):
```bash
cd frontend && npm run dev
```

Wait 3-4 seconds then verify:
```bash
curl -s http://localhost:8080/api/champions | python -c "import sys, json; d=json.load(sys.stdin); print(f'{len(d)} champs, first: {d[0]}')"
```
Expected: `30 champs, first: {'id': 'ahri', 'name': 'Ahri', 'imageKey': 'Ahri'}`.

- [ ] **Step 7: Manual browser check**

Open `http://localhost:5173` and verify:

1. Header "LOLIDLE" has a subtle gold glow.
2. Background has a subtle radial gradient (not flat black).
3. Type "a" in the search box. Suggestions appear with champion portraits (32x32) on the left.
4. Press ArrowDown — first suggestion gets a gold left border and darker background. Press again — second.
5. Press Enter — that champion is submitted as a guess.
6. The new row animates: cells flip in sequence from left to right over ~1.3 seconds.
7. The first column of the new row shows a 56x56 portrait with the champion name below.
8. Type a letter and press Escape — the search box clears.
9. Keep guessing until you win. The WinBanner scales in with overshoot, shows the target champion's 180x180 portrait, fires a confetti burst, and the "Jugar de nuevo" button pulses.
10. Click "Jugar de nuevo" — table clears, new game starts.

If anything breaks, fix it and re-run the tests before committing.

---

## Self-Review

### Spec coverage
| Spec section | Plan task |
|---|---|
| Backend `ImageKey` struct field + JSON + `championListItem` | Task 1 |
| Frontend types `imageKey` on Champion + ChampionListItem | Task 2 |
| `portrait.ts` with `fetchLatestVersion` + `getPortraitUrl` | Task 2 |
| SearchBox portraits + keyboard nav (arrow keys, Enter, Escape) + mouse hover sync | Task 3 |
| SearchBox auto-scroll on keyboard nav | Task 3 (`scrollIntoView` call in useEffect) |
| SearchBox version prop wiring | Task 3 + Task 6 |
| GuessTable portrait column | Task 4 |
| GuessTable sequential flip reveal on newest row only | Task 4 (`cell-reveal` class conditional) + Task 6 (CSS keyframe) |
| WinBanner portrait + scale-in + confetti + pulse button | Task 5 (component) + Task 6 (CSS) |
| Theme polish (gradient bg, gold glow, focus rings, hover states) | Task 6 CSS |
| New dep `canvas-confetti` | Task 5 |
| Fallback on Data Dragon fetch failure | Task 2 (`FALLBACK_VERSION` used in both portrait.ts and App.tsx) |
| 8 new tests + updated existing tests | Covered across Tasks 2-5 |

No gaps.

### Placeholder scan
No "TBD", "TODO", "implement later", or vague "add validation" instructions. Every step has literal code or an exact command.

### Type consistency
- Backend `Champion` struct → JSON tag `imageKey` → TS `Champion.imageKey` ✓
- `championListItem.ImageKey` → JSON tag `imageKey` → TS `ChampionListItem.imageKey` ✓
- `getPortraitUrl(version, imageKey)` signature used consistently in SearchBox, GuessTable, WinBanner ✓
- `FALLBACK_VERSION` exported from `portrait.ts`, imported by `App.tsx` ✓
- Confetti module default export mocked with matching signature ✓

No mismatches.

---

## Roadmap impact

After this plan ships, the Lolidle app has the visual polish needed for a demo-quality presentation. The next plan (future spec) covers CI/CD: Dockerfiles, GitHub Actions pipelines, Terraform for AWS, two environments, smoke tests, and the three required pipeline modifications.
