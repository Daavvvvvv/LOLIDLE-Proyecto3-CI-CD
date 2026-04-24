export const FALLBACK_VERSION = '14.24.1';
const REALM_URL = 'https://ddragon.leagueoflegends.com/realms/na.json';

export async function fetchLatestVersion(): Promise<string> {
  try {
    const res = await fetch(REALM_URL);
    if (!res.ok) return FALLBACK_VERSION;
    const data = await res.json();
    const version = data?.n?.champion;
    return typeof version === 'string' && version.length > 0 ? version : FALLBACK_VERSION;
  } catch {
    return FALLBACK_VERSION;
  }
}

export function getPortraitUrl(version: string, imageKey: string): string {
  return `https://ddragon.leagueoflegends.com/cdn/${version}/img/champion/${imageKey}.png`;
}
