import type { ReactNode } from 'react';

export interface Property {
  term: string;
  value: ReactNode;
}

/** Label/value rows separated by hairlines; values align right. Horizontal padding comes from the container. */
export function PropertyList({ items }: { items: readonly Property[] }) {
  return (
    <dl className="divide-y divide-line">
      {items.map(({ term, value }) => (
        <div key={term} className="flex h-11 items-center justify-between gap-6">
          <dt className="shrink-0 text-body text-muted">{term}</dt>
          <dd className="m-0 flex min-w-0 justify-end truncate text-right text-body text-fg">{value}</dd>
        </div>
      ))}
    </dl>
  );
}
