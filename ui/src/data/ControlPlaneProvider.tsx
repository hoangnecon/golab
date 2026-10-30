import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useReducer,
} from 'react';
import type { ReactNode } from 'react';
import { api } from '@/api/endpoints';
import { errorMessage } from '@/api/http';
import type { ActivityLog, NewProfile, NewSession, Profile, Session } from '@/api/types';
import { useNotice } from '@/shared/notice/NoticeProvider';
import { useEventStream } from './useEventStream';

interface State {
  profiles: Profile[];
  sessions: Session[];
  activity: ActivityLog[];
  loaded: boolean;
  online: boolean;
}

type Action =
  | { type: 'profiles'; profiles: Profile[] }
  | { type: 'sessions'; sessions: Session[] }
  | { type: 'activity'; activity: ActivityLog[] }
  | { type: 'loaded' }
  | { type: 'online'; online: boolean };

const initialState: State = {
  profiles: [],
  sessions: [],
  activity: [],
  loaded: false,
  online: true,
};

const MAX_ACTIVITY = 500;

/**
 * The stream only carries the latest few calls while the initial load carries
 * more, so incoming rows are merged by id instead of replacing the list.
 */
function mergeActivity(current: ActivityLog[], incoming: ActivityLog[]): ActivityLog[] {
  const byId = new Map<number, ActivityLog>();
  for (const entry of current) byId.set(entry.id, entry);
  for (const entry of incoming) byId.set(entry.id, entry);
  return [...byId.values()].sort((a, b) => b.id - a.id).slice(0, MAX_ACTIVITY);
}

function reducer(state: State, action: Action): State {
  switch (action.type) {
    case 'profiles':
      return { ...state, profiles: action.profiles };
    case 'sessions':
      return { ...state, sessions: action.sessions };
    case 'activity':
      return { ...state, activity: mergeActivity(state.activity, action.activity) };
    case 'loaded':
      return { ...state, loaded: true };
    case 'online':
      return state.online === action.online ? state : { ...state, online: action.online };
  }
}

/**
 * Form actions reject on failure so the form can show the error inline.
 * Row actions never reject: failures surface as a notice and resolve to
 * `false`, so callers can skip follow-up navigation.
 */
interface Actions {
  createProfile: (input: NewProfile) => Promise<void>;
  attachSession: (input: NewSession) => Promise<void>;
  launchProfile: (id: string) => Promise<boolean>;
  stopProfile: (id: string) => Promise<boolean>;
  deleteProfile: (id: string) => Promise<boolean>;
  focusSession: (id: string) => Promise<boolean>;
  activateSession: (id: string) => Promise<boolean>;
  interruptSession: (id: string) => Promise<boolean>;
  restartSession: (id: string) => Promise<boolean>;
  detachSession: (id: string) => Promise<boolean>;
}

type ControlPlane = State & Actions;

const ControlPlaneContext = createContext<ControlPlane | null>(null);

interface StreamTick {
  type?: string;
  sessions?: Session[] | null;
  activity?: ActivityLog[] | null;
}

export function ControlPlaneProvider({ children }: { children: ReactNode }) {
  const { notify, notifyError } = useNotice();
  const [state, dispatch] = useReducer(reducer, initialState);

  const refreshProfiles = useCallback(async () => {
    dispatch({ type: 'profiles', profiles: await api.profiles.list() });
  }, []);

  const refreshSessions = useCallback(async () => {
    dispatch({ type: 'sessions', sessions: await api.sessions.list() });
  }, []);

  useEffect(() => {
    let cancelled = false;
    Promise.all([api.profiles.list(), api.sessions.list(), api.activity.list()])
      .then(([profiles, sessions, activity]) => {
        if (cancelled) return;
        dispatch({ type: 'profiles', profiles });
        dispatch({ type: 'sessions', sessions });
        dispatch({ type: 'activity', activity });
      })
      .catch((error: unknown) => {
        if (!cancelled) notifyError(errorMessage(error));
      })
      .finally(() => {
        if (!cancelled) dispatch({ type: 'loaded' });
      });
    return () => {
      cancelled = true;
    };
  }, [notifyError]);

  const onStreamMessage = useCallback((data: unknown) => {
    const tick = data as StreamTick;
    if (tick.type !== 'tick') return;
    if (tick.sessions !== undefined) dispatch({ type: 'sessions', sessions: tick.sessions ?? [] });
    if (tick.activity !== undefined) dispatch({ type: 'activity', activity: tick.activity ?? [] });
  }, []);

  const onStreamStatus = useCallback((online: boolean) => dispatch({ type: 'online', online }), []);

  useEventStream('/api/events', onStreamMessage, onStreamStatus);

  const actions = useMemo<Actions>(() => {
    const attempt = async (task: () => Promise<unknown>, success?: string): Promise<boolean> => {
      try {
        await task();
        if (success) notify(success);
        return true;
      } catch (error) {
        notifyError(errorMessage(error));
        return false;
      }
    };

    return {
      createProfile: async (input) => {
        await api.profiles.create(input);
        await refreshProfiles();
      },
      attachSession: async (input) => {
        await api.sessions.create(input);
        await refreshSessions();
      },
      launchProfile: (id) =>
        attempt(async () => {
          await api.profiles.launch(id);
          await refreshProfiles();
        }),
      stopProfile: (id) =>
        attempt(async () => {
          await api.profiles.stop(id);
          await refreshProfiles();
        }),
      deleteProfile: (id) =>
        attempt(async () => {
          await api.profiles.remove(id);
          await refreshProfiles();
        }),
      focusSession: (id) => attempt(() => api.sessions.focus(id)),
      activateSession: (id) =>
        attempt(async () => {
          await api.sessions.activate(id);
          await refreshSessions();
        }),
      interruptSession: (id) => attempt(() => api.sessions.interrupt(id), 'Interrupt sent'),
      restartSession: (id) => attempt(() => api.sessions.restart(id), 'Kernel restarting'),
      detachSession: (id) =>
        attempt(async () => {
          await api.sessions.remove(id);
          await refreshSessions();
        }),
    };
  }, [notify, notifyError, refreshProfiles, refreshSessions]);

  const value = useMemo<ControlPlane>(() => ({ ...state, ...actions }), [state, actions]);

  return <ControlPlaneContext.Provider value={value}>{children}</ControlPlaneContext.Provider>;
}

export function useControlPlane(): ControlPlane {
  const context = useContext(ControlPlaneContext);
  if (!context) throw new Error('useControlPlane must be used inside ControlPlaneProvider');
  return context;
}
