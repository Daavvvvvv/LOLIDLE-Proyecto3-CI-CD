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
