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
