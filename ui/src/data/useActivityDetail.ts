import { useEffect, useState } from 'react';
import { api } from '@/api/endpoints';
import { errorMessage } from '@/api/http';
import type { ActivityDetail } from '@/api/types';

export type ActivityDetailState =
  | { status: 'loading' }
  | { status: 'error'; message: string }
  | { status: 'ready'; detail: ActivityDetail };

/** A recorded call never changes, so fetched details are kept for the page's lifetime. */
const cache = new Map<number, ActivityDetail>();

/** Request and response payloads of one tool call, fetched on demand. */
export function useActivityDetail(id: number): ActivityDetailState {
  const [state, setState] = useState<ActivityDetailState>(() => {
    const hit = cache.get(id);
    return hit ? { status: 'ready', detail: hit } : { status: 'loading' };
  });

  useEffect(() => {
    const hit = cache.get(id);
    if (hit) {
      setState({ status: 'ready', detail: hit });
      return;
    }
    let cancelled = false;
    setState({ status: 'loading' });
    api.activity
      .get(id)
      .then((detail) => {
        cache.set(id, detail);
        if (!cancelled) setState({ status: 'ready', detail });
      })
      .catch((error: unknown) => {
        if (!cancelled) setState({ status: 'error', message: errorMessage(error) });
      });
    return () => {
      cancelled = true;
    };
  }, [id]);

  return state;
}
