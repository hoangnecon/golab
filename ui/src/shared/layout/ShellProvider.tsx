import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react';
import type { ReactNode } from 'react';

const STORAGE_KEY = 'golab.sidebar';
const DESKTOP_QUERY = '(min-width: 768px)';

interface ShellApi {
  sidebarOpen: boolean;
  toggleSidebar: () => void;
  /** Closes the drawer after navigating on compact screens; no-op on desktop. */
  dismissOnCompact: () => void;
}

const ShellContext = createContext<ShellApi | null>(null);

function initialOpen(): boolean {
  if (!window.matchMedia(DESKTOP_QUERY).matches) return false;
  return window.localStorage.getItem(STORAGE_KEY) !== 'closed';
}

export function ShellProvider({ children }: { children: ReactNode }) {
  const [sidebarOpen, setSidebarOpen] = useState(initialOpen);

  const toggleSidebar = useCallback(() => {
    setSidebarOpen((open) => {
      window.localStorage.setItem(STORAGE_KEY, open ? 'closed' : 'open');
      return !open;
    });
  }, []);

  const dismissOnCompact = useCallback(() => {
    if (!window.matchMedia(DESKTOP_QUERY).matches) setSidebarOpen(false);
  }, []);

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'b') {
        event.preventDefault();
        toggleSidebar();
      }
    };
    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, [toggleSidebar]);

  const value = useMemo(
    () => ({ sidebarOpen, toggleSidebar, dismissOnCompact }),
    [sidebarOpen, toggleSidebar, dismissOnCompact],
  );

  return <ShellContext.Provider value={value}>{children}</ShellContext.Provider>;
}

export function useShell(): ShellApi {
  const context = useContext(ShellContext);
  if (!context) throw new Error('useShell must be used inside ShellProvider');
  return context;
}
