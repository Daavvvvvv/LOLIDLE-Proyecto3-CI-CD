import { useEffect, useState } from "react";
import "./styles.css";
import { createGame, listChampions, submitGuess } from "./api/client";
import { FALLBACK_VERSION, fetchLatestVersion } from "./api/portrait";
import type { ChampionListItem, GuessResponse } from "./api/types";
import { SearchBox } from "./components/SearchBox";
import { GuessTable } from "./components/GuessTable";
import { WinBanner } from "./components/WinBanner";

export function App() {
  const [champions, setChampions] = useState<ChampionListItem[]>([]);
  const [version, setVersion] = useState<string>(FALLBACK_VERSION);
  const [gameId, setGameId] = useState<string | null>(null);
  const [guesses, setGuesses] = useState<GuessResponse[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchLatestVersion().then(setVersion);
    listChampions()
      .then(setChampions)
      .catch((e) => setError(String(e)));
    void startNewGame();
    // eslint-disable-next-line react-hooks/exhaustive-deps
    console.log("App mounted, starting new game and fetching champions");
  }, []);

  async function startNewGame() {
    setGuesses([]);
    setError(null);
    try {
      const { gameId } = await createGame();
      setGameId(gameId);
    } catch (e) {
      setError(String(e));
    }
  }

  async function handleGuess(championId: string) {
    if (!gameId) return;
    try {
      const result = await submitGuess(gameId, championId);
      setGuesses((prev) => [...prev, result]);
    } catch (e) {
      setError(String(e));
    }
  }

  const lastGuess = guesses[guesses.length - 1];
  const won = lastGuess?.correct ?? false;
  const guessedIds = new Set(guesses.map((g) => g.guess.id));

  return (
    <div className="app">
      <header>
        <h1>LOLIDLE</h1>
        <p>Adivina el campeón</p>
      </header>
      {error && <div className="error">{error}</div>}
      {!won && (
        <SearchBox
          champions={champions}
          excludedIds={guessedIds}
          onSelect={handleGuess}
          disabled={!gameId}
          version={version}
        />
      )}
      {won && lastGuess && (
        <WinBanner
          attemptCount={lastGuess.attemptCount}
          championName={lastGuess.guess.name}
          imageKey={lastGuess.guess.imageKey}
          version={version}
          onPlayAgain={startNewGame}
          lore={lastGuess.lore}
        />
      )}
      <GuessTable guesses={guesses} version={version} />
    </div>
  );
}
