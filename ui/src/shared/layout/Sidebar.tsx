import type { LucideIcon } from 'lucide-react';
import type { ReactNode } from 'react';
import { cn } from '@/shared/lib/cn';
import { useShell } from './ShellProvider';

export interface NavItem {
  id: string;
  label: string;
  href: string;
  icon: LucideIcon;
}

interface SidebarProps {
  /** Scrollable middle region (the project tree). */
  children: ReactNode;
  /** Pinned workspace-level destinations under the tree. */
  items: readonly NavItem[];
  activeId: string;
}

/**
 * Sidebar row grid (x measured from the viewport, sidebar `w-64` with `p-2`):
 *   icon column centre x = 28   (row `px-3` → icon at 20..36)
 *   text column        x = 48   (icon + `gap-3`)
 *   trailing slot      x = 220..244 (row `pr-1` + 24px slot)
 * Every row (brand, tree, pinned nav) uses this grid and the `ROW` base class.
 */
export const ROW = 'flex h-8 items-center rounded transition-colors';

/**
 * Full-height navigation on the app background: brand row → scrolling tree →
 * pinned destinations. When closed it is removed from layout (width 0, inert).
 * The toggle lives in Shell; the brand row leaves room for it.
 */
export function Sidebar({ children, items, activeId }: SidebarProps) {
  const { sidebarOpen, toggleSidebar, dismissOnCompact } = useShell();

  return (
    <>
      {sidebarOpen && (
        <div aria-hidden className="fixed inset-0 z-30 bg-scrim md:hidden" onClick={toggleSidebar} />
      )}
      <aside
        aria-label="Primary"
        inert={!sidebarOpen}
        className={cn(
          'shrink-0 overflow-hidden bg-app transition-all duration-200 max-md:fixed max-md:inset-y-0 max-md:left-0 max-md:z-40',
          sidebarOpen ? 'w-64' : 'w-0',
        )}
      >
        <div className="flex h-full w-64 flex-col p-2">
          <div className="flex h-14 shrink-0 items-center pl-10">
            <span className="text-body font-semibold tracking-tight">GoLab</span>
          </div>
          <div className="min-h-0 flex-1 overflow-y-auto" onClick={dismissOnCompact}>
            {children}
          </div>
          <nav aria-label="Workspace" className="shrink-0 border-t border-line pt-2">
            <ul className="flex flex-col gap-0.5">
              {items.map(({ id, label, href, icon: Icon }) => {
                const active = id === activeId;
                return (
                  <li key={id}>
                    <a
                      href={href}
                      aria-current={active ? 'page' : undefined}
                      onClick={dismissOnCompact}
                      className={cn(
                        ROW,
                        'gap-3 px-3',
                        active ? 'bg-hover font-medium text-fg' : 'text-muted hover:bg-hover hover:text-fg',
                      )}
                    >
                      <Icon aria-hidden className="size-4 shrink-0" strokeWidth={1.75} />
                      {label}
                    </a>
                  </li>
                );
              })}
            </ul>
          </nav>
        </div>
      </aside>
    </>
  );
}
