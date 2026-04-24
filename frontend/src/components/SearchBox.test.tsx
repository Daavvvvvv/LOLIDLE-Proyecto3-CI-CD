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
