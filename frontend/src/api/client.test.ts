import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createGame, listChampions, submitGuess } from './client';

describe('api client', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn());
  });
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('listChampions GETs /api/champions', async () => {
    (fetch as any).mockResolvedValue({
      ok: true,
      json: async () => [{ id: 'ahri', name: 'Ahri' }],
    });
    const result = await listChampions();
    expect(result).toEqual([{ id: 'ahri', name: 'Ahri' }]);
    expect(fetch).toHaveBeenCalledWith(
      expect.stringContaining('/api/champions'),
      expect.any(Object),
    );
  });

  it('createGame POSTs to /api/games', async () => {
    (fetch as any).mockResolvedValue({
      ok: true,
      json: async () => ({ gameId: 'g123' }),
    });
    const result = await createGame();
    expect(result.gameId).toBe('g123');
    const call = (fetch as any).mock.calls[0];
    expect(call[1].method).toBe('POST');
  });

  it('submitGuess POSTs the championId', async () => {
    (fetch as any).mockResolvedValue({
      ok: true,
      json: async () => ({ correct: true, attemptCount: 1 }),
    });
    await submitGuess('g123', 'ahri');
    const call = (fetch as any).mock.calls[0];
    expect(call[0]).toContain('/api/games/g123/guesses');
    expect(JSON.parse(call[1].body)).toEqual({ championId: 'ahri' });
  });

  it('throws when response is not ok', async () => {
    (fetch as any).mockResolvedValue({
      ok: false,
      status: 404,
      text: async () => 'not found',
    });
    await expect(listChampions()).rejects.toThrow('HTTP 404');
  });
});
