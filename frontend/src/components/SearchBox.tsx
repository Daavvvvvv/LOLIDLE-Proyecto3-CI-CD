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
    if (el && typeof el.scrollIntoView === 'function') {
      el.scrollIntoView({ block: 'nearest' });
    }
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
                  alt={c.name}
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
