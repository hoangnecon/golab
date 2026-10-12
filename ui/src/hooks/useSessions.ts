import { useState, useCallback } from 'react';
import { Session } from '../types';
import { api } from '../services/api';

export function useSessions(onFeedback?: (msg: string) => void) {
  const [sessions, setSessions] = useState<Session[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(false);

  const loadSessions = useCallback(async () => {
    setIsLoading(true);
    try {
      const data = await api.getSessions();
      setSessions(data);
    } catch (err) {
      console.error('Failed to load sessions:', err);
    } finally {
      setIsLoading(false);
    }
  }, []);

  const attachSession = useCallback(
    async (data: { profile_id: string; notebook_url: string; name?: string }) => {
      try {
        await api.createSession(data);
        await loadSessions();
        onFeedback?.('Notebook attached successfully');
      } catch (err: any) {
        alert('Failed to attach: ' + (err.message || err));
      }
    },
    [loadSessions, onFeedback]
  );

  const interruptSession = useCallback(
    async (id: string) => {
      try {
        await api.interruptSession(id);
        onFeedback?.('SIGINT signal dispatched to kernel');
      } catch (err: any) {
        alert('Interrupt failed: ' + (err.message || err));
      }
    },
    [onFeedback]
  );

  const restartSession = useCallback(
    async (id: string) => {
      try {
        await api.restartSession(id);
        onFeedback?.('Kernel restart initiated');
      } catch (err: any) {
        alert('Restart failed: ' + (err.message || err));
      }
    },
    [onFeedback]
  );

  const focusSession = useCallback(
    async (id: string) => {
      try {
        await api.focusSession(id);
        onFeedback?.('Focused Chrome tab');
      } catch (err: any) {
        console.error('Focus failed:', err);
      }
    },
    [onFeedback]
  );

  const detachSession = useCallback(
    async (id: string) => {
      try {
        await api.deleteSession(id);
        await loadSessions();
        onFeedback?.('Session detached');
      } catch (err: any) {
        console.error('Detach failed:', err);
      }
    },
    [loadSessions, onFeedback]
  );

  return {
    sessions,
    setSessions,
    isLoading,
    loadSessions,
    attachSession,
    interruptSession,
    restartSession,
    focusSession,
    detachSession,
  };
}
