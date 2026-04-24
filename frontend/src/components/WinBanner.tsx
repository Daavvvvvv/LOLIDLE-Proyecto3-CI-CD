import { useEffect } from 'react';
import confetti from 'canvas-confetti';
import { getPortraitUrl } from '../api/portrait';

interface Props {
  attemptCount: number;
  championName: string;
  imageKey: string;
  version: string;
  onPlayAgain: () => void;
}

export function WinBanner({ attemptCount, championName, imageKey, version, onPlayAgain }: Props) {
  useEffect(() => {
    confetti({
      particleCount: 100,
      spread: 70,
      origin: { y: 0.6 },
      colors: ['#c8aa6e', '#f0e6d2', '#3d8bff'],
    });
  }, []);

  return (
    <div className="win-banner appearing">
      <img
        className="target-portrait"
        src={getPortraitUrl(version, imageKey)}
        alt={championName}
        width={180}
        height={180}
      />
      <h2>¡Ganaste en {attemptCount} intentos!</h2>
      <p>
        El campeón era <strong>{championName}</strong>.
      </p>
      <button onClick={onPlayAgain}>Jugar de nuevo</button>
    </div>
  );
}
