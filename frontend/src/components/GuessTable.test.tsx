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
