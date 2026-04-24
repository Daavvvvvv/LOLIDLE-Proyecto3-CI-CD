import type { ChampionListItem, CreateGameResponse, GuessResponse } from './types';

const BASE = (import.meta.env.VITE_API_BASE as string | undefined) ?? 'http://localhost:8080';

async function jsonRequest<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    ...init,
    headers: { 'Content-Type': 'application/json', ...(init?.headers ?? {}) },
  });
  if (!res.ok) {
    const body = await res.text();
    throw new Error(`HTTP ${res.status}: ${body}`);
  }
  return (await res.json()) as T;
}

export function listChampions(): Promise<ChampionListItem[]> {
  return jsonRequest<ChampionListItem[]>('/api/champions');
}

export function createGame(): Promise<CreateGameResponse> {
  return jsonRequest<CreateGameResponse>('/api/games', { method: 'POST' });
}

export function submitGuess(gameId: string, championId: string): Promise<GuessResponse> {
  return jsonRequest<GuessResponse>(`/api/games/${gameId}/guesses`, {
    method: 'POST',
    body: JSON.stringify({ championId }),
  });
}
