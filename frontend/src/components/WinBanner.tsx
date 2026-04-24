interface Props {
  attemptCount: number;
  championName: string;
  onPlayAgain: () => void;
}

export function WinBanner({ attemptCount, championName, onPlayAgain }: Props) {
  return (
    <div className="win-banner">
      <h2>¡Ganaste en {attemptCount} intentos!</h2>
      <p>
        El campeón era <strong>{championName}</strong>.
      </p>
      <button onClick={onPlayAgain}>Jugar de nuevo</button>
    </div>
  );
}
