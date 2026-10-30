import { Fragment } from 'react';
import type { ReactNode } from 'react';
import { cn } from '@/shared/lib/cn';
import { useShell } from './ShellProvider';

export interface Crumb {
  label: string;
  /** Omit on the last crumb: it is the page title. */
  href?: string;
}

interface PageProps {
  /** Location trail. The last crumb is the page's <h1>. */
  trail: readonly Crumb[];
  /** Quiet count next to the title. */
  meta?: string;
  /** At most one primary action (the create button of the page). */
  action?: ReactNode;
  children: ReactNode;
}

/**
 * 56px header (trail + one action), scrolling body. Left padding clears the
 * fixed sidebar toggle when the sidebar is closed (and always on compact screens).
 */
export function Page({ trail, meta, action, children }: PageProps) {
  const { sidebarOpen } = useShell();
  const current = trail[trail.length - 1];
  const parents = trail.slice(0, -1);

  return (
    <>
      <header
        className={cn(
          'flex h-14 shrink-0 items-center gap-3 border-b border-line pl-14 pr-4 transition-all duration-200',
          sidebarOpen && 'md:pl-6',
        )}
      >
        <nav aria-label="Breadcrumb" className="flex min-w-0 flex-1 items-center gap-2 text-body">
          {parents.map((crumb) => (
            <Fragment key={crumb.label}>
              <a href={crumb.href} className="max-w-48 shrink-0 truncate text-muted transition-colors hover:text-fg">
                {crumb.label}
              </a>
              <span aria-hidden className="text-line-strong">
                /
              </span>
            </Fragment>
          ))}
          <h1 className="truncate font-medium text-fg">{current?.label}</h1>
          {meta && <span className="shrink-0 tabular-nums text-subtle">{meta}</span>}
        </nav>
        {action}
      </header>
      <div className="min-h-0 flex-1 overflow-y-auto p-6 max-md:p-4">{children}</div>
    </>
  );
}
