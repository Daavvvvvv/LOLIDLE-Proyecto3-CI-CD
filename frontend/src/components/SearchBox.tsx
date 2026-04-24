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
