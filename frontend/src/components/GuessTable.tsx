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
