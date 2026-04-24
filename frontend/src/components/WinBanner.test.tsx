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

  it('renders lore when provided', () => {
    render(<WinBanner {...defaultProps} lore="Ahri es una vastaya nine-tailed." />);
    expect(screen.getByText(/vastaya nine-tailed/)).toBeInTheDocument();
  });

  it('does not render lore blockquote when lore is empty', () => {
    const { container } = render(<WinBanner {...defaultProps} />);
    expect(container.querySelector('.champion-lore')).toBeNull();
  });
});
