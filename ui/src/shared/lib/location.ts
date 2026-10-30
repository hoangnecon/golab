import { useMemo, useSyncExternalStore } from 'react';

export interface RouteLocation {
  segments: string[];
  params: URLSearchParams;
}

/** Props every routed page receives: the path segments after the section, e.g. ['p1', 'runs', 's1', 'logs']. */
export interface RouteProps {
  segments: string[];
}

function safeDecode(value: string): string {
  try {
    return decodeURIComponent(value);
  } catch {
    return value;
  }
}

export function parseHash(hash: string): RouteLocation {
  const [path = '', query = ''] = hash.replace(/^#\/?/, '').split('?');
  return {
    segments: path.split('/').filter(Boolean).map(safeDecode),
    params: new URLSearchParams(query),
  };
}

function subscribe(onChange: () => void) {
  window.addEventListener('hashchange', onChange);
  return () => window.removeEventListener('hashchange', onChange);
}

const getHash = () => window.location.hash;

/** Hash routing keeps the page across reloads and needs no server fallback. */
export function useLocation(): RouteLocation {
  const hash = useSyncExternalStore(subscribe, getHash);
  return useMemo(() => parseHash(hash), [hash]);
}

/**
 * Updates query params in place without adding history entries; an empty value
 * removes the key. Synchronous (replaceState + synthetic hashchange) because
 * controlled inputs bound to the URL must see the new value within the same
 * event, otherwise React reverts what the user typed.
 */
export function replaceParams(update: Record<string, string>) {
  const { segments, params } = parseHash(window.location.hash);
  for (const [key, value] of Object.entries(update)) {
    if (value) params.set(key, value);
    else params.delete(key);
  }
  const query = params.toString();
  const path = `#/${segments.map(encodeURIComponent).join('/')}`;
  window.history.replaceState(null, '', query ? `${path}?${query}` : path);
  window.dispatchEvent(new HashChangeEvent('hashchange'));
}
