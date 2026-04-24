import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { fetchLatestVersion, getPortraitUrl, FALLBACK_VERSION } from './portrait';

describe('portrait', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn());
  });
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('fetchLatestVersion returns n.champion when the realm call succeeds', async () => {
    (fetch as any).mockResolvedValue({
      ok: true,
      json: async () => ({ n: { champion: '14.24.1' } }),
    });
    expect(await fetchLatestVersion()).toBe('14.24.1');
  });

  it('fetchLatestVersion returns the fallback when fetch rejects', async () => {
    (fetch as any).mockRejectedValue(new Error('network down'));
    expect(await fetchLatestVersion()).toBe(FALLBACK_VERSION);
  });

  it('fetchLatestVersion returns the fallback when response is not ok', async () => {
    (fetch as any).mockResolvedValue({ ok: false, status: 500, json: async () => ({}) });
    expect(await fetchLatestVersion()).toBe(FALLBACK_VERSION);
  });

  it('fetchLatestVersion returns the fallback when n.champion is missing', async () => {
    (fetch as any).mockResolvedValue({ ok: true, json: async () => ({ n: {} }) });
    expect(await fetchLatestVersion()).toBe(FALLBACK_VERSION);
  });

  it('getPortraitUrl builds the expected Data Dragon URL', () => {
    expect(getPortraitUrl('14.24.1', 'Ahri')).toBe(
      'https://ddragon.leagueoflegends.com/cdn/14.24.1/img/champion/Ahri.png',
    );
  });
});
