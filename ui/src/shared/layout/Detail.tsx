import type { ReactNode } from 'react';
import { cn } from '@/shared/lib/cn';

/**
 * Detail-page body. Two columns at xl: `main` (2/3) and a rail (1/3). Both
 * columns are stacks of `Section`s, so their first headings share one baseline
 * and their cards share one top edge. On narrower screens the rail comes first.
 */
export function Detail({ main, rail }: { main: ReactNode; rail: ReactNode }) {
  return (
    <div className="grid grid-cols-1 items-start gap-x-6 gap-y-8 xl:grid-cols-3">
      <div className="flex min-w-0 flex-col gap-8 xl:col-span-2">{main}</div>
      <div className="flex min-w-0 flex-col gap-8 max-xl:order-first">{rail}</div>
    </div>
  );
}

interface SectionProps {
  title: string;
  /** Quiet count next to the title. */
  count?: number;
  children: ReactNode;
}

/** Heading row (32px) + content. The only way to title a block of a page. */
export function Section({ title, count, children }: SectionProps) {
  return (
    <section className="flex min-w-0 flex-col gap-2">
      <h2 className="flex h-8 items-center gap-2 text-body font-medium text-fg">
        {title}
        {count !== undefined && <span className="font-normal tabular-nums text-subtle">{count}</span>}
      </h2>
      {children}
    </section>
  );
}

/** Rounded card. Content padding is `px-4`; `PanelActions` reaches the same 16px text edge. */
export function Panel({ children, className }: { children: ReactNode; className?: string }) {
  return <div className={cn('overflow-hidden rounded-lg border border-line bg-surface', className)}>{children}</div>;
}

export function PanelBody({ children }: { children: ReactNode }) {
  return <div className="px-4">{children}</div>;
}

/**
 * Vertical list of `menu` buttons. `p-2` + the button's `px-2` puts icons on
 * the same 16px edge as `PanelBody` text. `separated` draws a hairline above
 * (used for the destructive group, which always comes last).
 */
export function PanelActions({ children, separated }: { children: ReactNode; separated?: boolean }) {
  return (
    <div className={cn('flex flex-col gap-0.5 p-2', separated && 'border-t border-line')}>{children}</div>
  );
}

export interface Metric {
  label: string;
  value: ReactNode;
  /** Optional line under the value, e.g. a Meter. */
  footer?: ReactNode;
}

const COLUMNS: Record<number, string> = {
  2: 'lg:grid-cols-2',
  3: 'lg:grid-cols-3',
  4: 'lg:grid-cols-4',
  5: 'lg:grid-cols-5',
};

/**
 * Key numbers of an entity as equal cells in one rounded card, separated by
 * 1px gutters (the card background shows through `gap-px`). Two columns below
 * `lg`, one row at `lg`+. With an odd count the last cell spans the row on
 * narrow screens so no empty gutter-coloured cell appears.
 */
export function Metrics({ items }: { items: readonly Metric[] }) {
  const odd = items.length % 2 === 1;
  return (
    <dl
      className={cn(
        'grid grid-cols-2 gap-px overflow-hidden rounded-lg border border-line bg-line',
        COLUMNS[items.length],
      )}
    >
      {items.map(({ label, value, footer }, index) => (
        <div
          key={label}
          className={cn(
            'flex min-w-0 flex-col gap-1 bg-surface px-4 pb-3 pt-4',
            odd && index === items.length - 1 && 'max-lg:col-span-2',
          )}
        >
          <dt className="text-caption text-subtle">{label}</dt>
          <dd className="m-0 flex h-7 min-w-0 items-center truncate text-title font-medium text-fg">{value}</dd>
          <div className="flex h-3 items-center">{footer}</div>
        </div>
      ))}
    </dl>
  );
}
