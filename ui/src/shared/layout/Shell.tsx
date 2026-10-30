import { PanelLeftClose, PanelLeftOpen } from 'lucide-react';
import type { ReactNode } from 'react';
import { cn } from '@/shared/lib/cn';
import { IconButton } from '@/shared/ui/IconButton';
import { Sidebar } from './Sidebar';
import type { NavItem } from './Sidebar';
import { useShell } from './ShellProvider';

interface ShellProps {
  items: readonly NavItem[];
  activeId: string;
  tree: ReactNode;
  children: ReactNode;
}

/**
 * App frame on the `app` background: sidebar column + an inset, rounded
 * content frame (8px gutter). Geometry (see DESIGN.md § Layout grid):
 * - the content frame starts at y = 8, its 56px header is centred on y = 36;
 * - the sidebar brand row spans y = 8..64, also centred on y = 36;
 * - the single toggle (32px) is fixed at left 12 / top 20, centred on (28, 36),
 *   which is also the centre of every sidebar icon column. It never moves.
 */
export function Shell({ items, activeId, tree, children }: ShellProps) {
  const { sidebarOpen, toggleSidebar } = useShell();
  const Icon = sidebarOpen ? PanelLeftClose : PanelLeftOpen;

  return (
    <div className="flex h-dvh overflow-hidden bg-app text-fg">
      <Sidebar items={items} activeId={activeId}>
        {tree}
      </Sidebar>
      <div className={cn('flex min-w-0 flex-1 flex-col p-2 transition-all duration-200', sidebarOpen && 'md:pl-0')}>
        <main className="relative flex min-h-0 flex-1 flex-col overflow-hidden rounded-xl bg-canvas ring-1 ring-line">
          {children}
        </main>
      </div>
      <IconButton
        label={sidebarOpen ? 'Hide sidebar' : 'Show sidebar'}
        onClick={toggleSidebar}
        className="fixed left-3 top-5 z-50"
      >
        <Icon aria-hidden className="size-4" strokeWidth={1.75} />
      </IconButton>
    </div>
  );
}
