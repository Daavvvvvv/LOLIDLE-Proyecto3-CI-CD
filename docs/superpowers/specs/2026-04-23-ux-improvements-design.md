# UX/UI Improvements — Design Spec

**Date:** 2026-04-23
**Author:** David Vélez
**Context:** Polish iteration on top of `v0.1.0-app`. The base app works end-to-end (backend + frontend) but is visually plain. This spec adds the visual identity and keyboard-first interaction that separate a "demo clone" from a "professional clone" of Loldle.

## Goal

Bring the Lolidle clone closer to the look-and-feel of [loldle.net](https://loldle.net):
1. **Champion portraits** everywhere a champion is shown (autocomplete, table, win banner).
2. **Sequential flip-reveal** animation per cell when a new guess lands.
3. **Keyboard-first autocomplete** (arrows + Enter + Escape).
4. **Win celebration** with confetti + scale-in banner.
5. **Background and theme polish** (radial gradient, glows, hover/focus states).

## Non-goals

- New game modes (still Classic / freeplay only)
- Backend logic changes beyond adding the `imageKey` field to the Champion struct
- Mobile-first responsive design (desktop remains primary; layout shouldn't break on mobile but isn't optimized)
- Riot API integration beyond the public Data Dragon CDN
- Persisting the chosen Data Dragon version anywhere

## Components

### 1. Champion data — add `imageKey`

**Backend (`backend/internal/champions/`):**

- `Champion` struct gets a new field:
  ```go
  ImageKey string `json:"imageKey"`
  ```
- `champions.json` is updated so each of the 30 entries has the matching Data Dragon key. Mapping:

  | id (our slug) | imageKey (Riot) |
  |---|---|
  | ahri | Ahri |
  | yasuo | Yasuo |
  | garen | Garen |
  | lux | Lux |
  | darius | Darius |
  | jinx | Jinx |
  | vi | Vi |
  | caitlyn | Caitlyn |
  | ezreal | Ezreal |
  | lee-sin | LeeSin |
  | thresh | Thresh |
  | senna | Senna |
  | lucian | Lucian |
  | akali | Akali |
  | zed | Zed |
  | teemo | Teemo |
  | tristana | Tristana |
  | veigar | Veigar |
  | kassadin | Kassadin |
  | chogath | Chogath |
  | khazix | Khazix |
  | reksai | RekSai |
  | malphite | Malphite |
  | sona | Sona |
  | soraka | Soraka |
  | aatrox | Aatrox |
  | kayn | Kayn |
  | sett | Sett |
  | yone | Yone |
  | aphelios | Aphelios |

The field flows automatically through `/api/champions` (we update the `championListItem` to include it) and `/api/games/:id/guesses` (since the response embeds the full `Champion`).

### 2. Frontend — Data Dragon version + portrait helper

New file: `frontend/src/api/portrait.ts`

```ts
const FALLBACK_VERSION = '14.24.1';
const REALM_URL = 'https://ddragon.leagueoflegends.com/realms/na.json';

export async function fetchLatestVersion(): Promise<string> {
  try {
    const res = await fetch(REALM_URL);
    if (!res.ok) return FALLBACK_VERSION;
    const data = await res.json();
    return data?.n?.champion ?? FALLBACK_VERSION;
  } catch {
    return FALLBACK_VERSION;
  }
}

export function getPortraitUrl(version: string, imageKey: string): string {
  return `https://ddragon.leagueoflegends.com/cdn/${version}/img/champion/${imageKey}.png`;
}
```

- Fetches from `realms/na.json` once at app startup. The relevant field is `n.champion` (e.g. `"14.24.1"`).
- Falls back to a hardcoded version if the fetch fails (Riot CDN down, network blocked, etc.) — never leaves the user with broken images because of an external dependency.
- Riot's CDN has permissive CORS so the browser can fetch directly.

### 3. SearchBox — keyboard nav + portraits

Behavior changes:
- **Suggestions render with portraits**: each `<li>` shows `<img src={portraitUrl} width=32 height=32>` to the left of the name. `loading="lazy"`. If the image fails (`onError`), replace with a styled letter circle showing the first letter of the name.
- **Highlighted state**: a single index `highlightedIndex` (or `null`) tracks which suggestion is "selected" by keyboard. Visual: background `#2a2e36` + 3px gold left border.
- **Mouse hover** updates `highlightedIndex` so mouse and keyboard share the same notion of "selected".
- **Keyboard handlers** (on the `<input>`):
  - `ArrowDown` → `setHighlightedIndex((i) => i === null ? 0 : (i + 1) % matches.length)`
  - `ArrowUp` → `setHighlightedIndex((i) => i === null ? matches.length - 1 : (i - 1 + matches.length) % matches.length)`
  - `Enter` → if `matches.length > 0`, select `matches[highlightedIndex ?? 0]`
  - `Escape` → `setHighlightedIndex(null)` and clear query
  - All keyboard handlers `preventDefault()` to keep focus on the input
- **Reset highlighted index** to `null` whenever `query` changes (so typing fresh letters starts you back at "no selection, Enter picks first").
- **Auto-scroll**: if highlighted option falls outside the visible dropdown, scroll it into view (`option.scrollIntoView({ block: 'nearest' })`).

Props stay the same (`champions`, `excludedIds`, `onSelect`, `disabled`) plus a new prop `version: string` for portrait URLs.

### 4. GuessTable — sequential flip + portraits

Behavior changes:
- **First column** changes from a plain `<td>{name}</td>` to a stacked layout: portrait on top (56x56) + name below.
- **Sequential flip-reveal** animates the newest row only:
  - Newest row index = `guesses.length - 1`. When rendering each row, set `const isNewest = rowIndex === guesses.length - 1`.
  - For each cell in that row, conditionally apply class `cell-reveal` and `style={{ animationDelay: \`${cellIndex * 120}ms\` }}` only when `isNewest` is true. Older rows render with no animation classes.
  - cellIndex enumeration: 0 for the portrait+name cell, 1-7 for the seven attribute cells.
  - Keyframe rotates the cell `rotateX(0) → rotateX(90deg) → rotateX(0)` over 500ms with `animation-fill-mode: both`. Since the keyframes start and end at the identity transform, removing the class on subsequent renders (when this row is no longer newest) leaves no visual artifact.
  - Total reveal time for one row: `7 * 120ms + 500ms = ~1.34s`.
- React `key` for rows is the row index, which is stable for append-only data — React reuses existing row elements and only mounts the new one, so the animation fires exactly once per new guess.

Props stay the same (`guesses`) plus a new prop `version: string`.

### 5. WinBanner — scale-in + confetti + portrait

Behavior changes:
- Banner appears with a CSS animation: `transform: scale(0.7) → scale(1)` and `opacity: 0 → 1`, 350ms cubic-bezier(0.34, 1.56, 0.64, 1) (gentle overshoot).
- Above the text, render the **target champion's portrait** at 180x180 with a gold border.
- On mount, fire confetti via `canvas-confetti`:
  ```ts
  import confetti from 'canvas-confetti';
  useEffect(() => {
    confetti({
      particleCount: 100,
      spread: 70,
      origin: { y: 0.6 },
      colors: ['#c8aa6e', '#f0e6d2', '#3d8bff'], // LoL gold + light gold + blue
    });
  }, []);
  ```
- "Jugar de nuevo" button gets a subtle pulse: `box-shadow` keyframe oscillating between `0 0 0 0 rgba(200,170,110,0.7)` and `0 0 0 12px rgba(200,170,110,0)`, infinite, 2s.

Props gain `imageKey` and `version` (so the banner can build the portrait URL).

### 6. CSS / theme polish

`frontend/src/styles.css` is rewritten to add:

- `body` background: `radial-gradient(ellipse at top, #1a1d2e 0%, #0e1014 60%)` instead of flat black.
- `header h1`: text-shadow `0 0 20px rgba(200, 170, 110, 0.4)`.
- `.search-box input:focus`: gold ring `box-shadow: 0 0 0 2px #c8aa6e`.
- `.search-box li.highlighted`: bg `#2a2e36`, `border-left: 3px solid #c8aa6e`, padding adjusted so width doesn't jump.
- `.search-box li img.option-portrait`: 32x32, rounded 4px, margin-right 12px, vertical-align middle.
- `.search-box li .option-fallback`: same dimensions, bg `#3a3e46`, white centered initial.
- `.guess-table td .cell-portrait img`: 56x56, rounded 4px, display block.
- `.cell-reveal`: `animation: flip-reveal 500ms ease-out both;`
- `@keyframes flip-reveal { 0% { transform: rotateX(0); } 50% { transform: rotateX(90deg); } 100% { transform: rotateX(0); } }`
- `.win-banner.appearing`: `animation: scale-in 350ms cubic-bezier(0.34, 1.56, 0.64, 1) both;`
- `@keyframes scale-in { from { transform: scale(0.7); opacity: 0; } to { transform: scale(1); opacity: 1; } }`
- `.win-banner img.target-portrait`: 180x180, `border: 4px solid #c8aa6e`, `border-radius: 8px`, box-shadow gold glow.
- `.win-banner button`: existing styles + `animation: pulse 2s infinite;`
- `@keyframes pulse { 0% { box-shadow: 0 0 0 0 rgba(200,170,110,0.7); } 70% { box-shadow: 0 0 0 12px rgba(200,170,110,0); } 100% { box-shadow: 0 0 0 0 rgba(200,170,110,0); } }`

### 7. App.tsx — wiring the new state

```tsx
const [version, setVersion] = useState<string>('14.24.1');

useEffect(() => {
  fetchLatestVersion().then(setVersion);
  // existing listChampions + startNewGame calls
}, []);
```

Pass `version` down to SearchBox, GuessTable, WinBanner.

## Tests

### New tests

`SearchBox.test.tsx`:
- `Enter selects first option when nothing highlighted`
- `ArrowDown highlights next, cycles at end`
- `ArrowUp highlights previous, cycles at start`
- `Escape clears query and highlighted state`
- `renders portrait img per option (with version prop)`

`GuessTable.test.tsx`:
- `renders portrait img in first column for each guess`

`WinBanner.test.tsx`:
- `renders target champion portrait with version`
- `calls confetti on mount` (mock the canvas-confetti default export)

### Existing tests — small updates required

The existing tests need minor tweaks because some component props change:

- **SearchBox**: gains required `version` prop. Each existing test in `SearchBox.test.tsx` adds `version="14.24.1"` to its `<SearchBox ... />` invocation. No other behavior change needed.
- **GuessTable**: gains required `version` prop. Same trivial update across the 4 existing tests.
- **WinBanner**: gains required `imageKey` and `version` props. The 2 existing tests pass `imageKey="Ahri"` and `version="14.24.1"`. Also wrap the test file with `vi.mock('canvas-confetti', () => ({ default: vi.fn() }))` at the top so mounting the component doesn't try to draw a real canvas.

After those tweaks, existing assertions still pass (filter logic, click handler, status classes, attempt count rendering — all preserved).

Mocking `canvas-confetti`:
```ts
vi.mock('canvas-confetti', () => ({ default: vi.fn() }));
```

Animations are CSS-only or trigger fire-and-forget side effects (confetti). They don't change the DOM in ways that break Testing Library queries.

## Dependencies

- **New runtime dep**: `canvas-confetti` (~12KB min+gzip).
- **No new dev deps** required — Vitest + RTL handle the new tests.

## File-by-file changes

```
backend/internal/champions/champions.json       # +imageKey on each entry
backend/internal/champions/store.go             # +ImageKey field on Champion struct
backend/internal/api/handlers.go                # championListItem gains ImageKey

frontend/package.json                           # +canvas-confetti
frontend/src/api/types.ts                       # +imageKey on Champion + ChampionListItem
frontend/src/api/portrait.ts                    # NEW: fetchLatestVersion + getPortraitUrl
frontend/src/api/portrait.test.ts               # NEW: tests with mocked fetch
frontend/src/components/SearchBox.tsx           # rewrite: portraits + keyboard nav + version prop
frontend/src/components/SearchBox.test.tsx     # +5 new tests
frontend/src/components/GuessTable.tsx          # +portrait column + cell-reveal class + version prop
frontend/src/components/GuessTable.test.tsx    # +1 new test
frontend/src/components/WinBanner.tsx           # +confetti + portrait + scale-in + version prop
frontend/src/components/WinBanner.test.tsx     # +mock confetti + 2 new tests
frontend/src/App.tsx                            # +version state + fetchLatestVersion
frontend/src/styles.css                         # extensive rewrite for theme polish
```

## Backwards compatibility

- The new `imageKey` field on `Champion` is additive; old API consumers that ignore unknown fields keep working.
- The new `version` prop on components is required — every render site (only one: App.tsx) must pass it. App.tsx initializes it to the fallback string before the fetch resolves, so nothing renders with `undefined`.

## Risks

- **Riot CDN down or rate-limited**: handled by the fallback version + `<img onError>` fallback.
- **Performance of 30 portrait fetches**: dropdown shows max 8 at a time; lazy loading + browser cache mean the first guess triggers ~8 image loads in parallel, which is fine.
- **Test flakiness with timer-driven animations**: the flip animation is purely CSS, so it doesn't affect tests. The confetti call is a side effect; we mock the module.
- **Bundle size**: canvas-confetti adds ~12KB. Acceptable.

## Out of scope (deferred)

- Animating the search dropdown opening/closing (could be added later but not critical for the polish goal).
- Particle background effects.
- Champion ability icons or splash arts.
- Multiple Data Dragon regions/realms — `na.json` is a fine global default.
