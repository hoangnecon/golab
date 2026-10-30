import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState } from 'react';
import type { ReactNode } from 'react';
import { Check, CircleAlert } from 'lucide-react';

type NoticeKind = 'info' | 'error';

interface Notice {
  id: number;
  kind: NoticeKind;
  message: string;
}

interface NoticeApi {
  notify: (message: string) => void;
  notifyError: (message: string) => void;
}

const NoticeContext = createContext<NoticeApi | null>(null);

const LIFETIME_MS: Record<NoticeKind, number> = { info: 4000, error: 8000 };
const MAX_VISIBLE = 3;

export function NoticeProvider({ children }: { children: ReactNode }) {
  const [notices, setNotices] = useState<Notice[]>([]);
  const nextId = useRef(0);

  const push = useCallback((kind: NoticeKind, message: string) => {
    const id = nextId.current++;
    setNotices((current) => [...current.slice(-(MAX_VISIBLE - 1)), { id, kind, message }]);
  }, []);

  const dismiss = useCallback((id: number) => {
    setNotices((current) => current.filter((notice) => notice.id !== id));
  }, []);

  const api = useMemo<NoticeApi>(
    () => ({
      notify: (message) => push('info', message),
      notifyError: (message) => push('error', message),
    }),
    [push],
  );

  return (
    <NoticeContext.Provider value={api}>
      {children}
      <div
        aria-live="polite"
        className="pointer-events-none fixed bottom-4 right-4 z-50 flex max-w-sm flex-col items-end gap-2 max-sm:left-4"
      >
        {notices.map((notice) => (
          <NoticeItem key={notice.id} notice={notice} onDismiss={dismiss} />
        ))}
      </div>
    </NoticeContext.Provider>
  );
}

function NoticeItem({ notice, onDismiss }: { notice: Notice; onDismiss: (id: number) => void }) {
  const { id, kind } = notice;

  useEffect(() => {
    const timer = setTimeout(() => onDismiss(id), LIFETIME_MS[kind]);
    return () => clearTimeout(timer);
  }, [id, kind, onDismiss]);

  const Icon = kind === 'error' ? CircleAlert : Check;
  return (
    <button
      type="button"
      onClick={() => onDismiss(id)}
      className="pointer-events-auto flex items-start gap-3 rounded-lg bg-fg px-4 py-3 text-left text-body text-canvas shadow-overlay"
    >
      <Icon aria-label={kind === 'error' ? 'Error' : 'Done'} className="mt-0.5 size-4 shrink-0" strokeWidth={2} />
      <span className="min-w-0 break-words">{notice.message}</span>
    </button>
  );
}

export function useNotice(): NoticeApi {
  const context = useContext(NoticeContext);
  if (!context) throw new Error('useNotice must be used inside NoticeProvider');
  return context;
}
