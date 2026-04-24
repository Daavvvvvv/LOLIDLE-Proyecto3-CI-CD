import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { WinBanner } from './WinBanner';

describe('WinBanner', () => {
  it('shows attempt count and champion name', () => {
    render(<WinBanner attemptCount={5} championName="Ahri" onPlayAgain={() => {}} />);
    expect(screen.getByText(/5 intentos/)).toBeInTheDocument();
    expect(screen.getByText(/Ahri/)).toBeInTheDocument();
  });

  it('calls onPlayAgain when button is clicked', () => {
    const onPlayAgain = vi.fn();
    render(<WinBanner attemptCount={3} championName="Yasuo" onPlayAgain={onPlayAgain} />);
    fireEvent.click(screen.getByRole('button', { name: /jugar de nuevo/i }));
    expect(onPlayAgain).toHaveBeenCalled();
  });
});
