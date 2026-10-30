import type { ComponentProps, ReactNode } from 'react';
import { cn } from '@/shared/lib/cn';

/**
 * Rounded table frame. Rows are separated by `divide-y` on <tbody>, so the last
 * row never draws a border against the frame. Scrolls sideways on narrow screens.
 */
export function TableCard({ children }: { children: ReactNode }) {
  return (
    <div className="overflow-x-auto rounded-lg border border-line bg-surface">
      <table className="w-full border-collapse text-left text-body">{children}</table>
    </div>
  );
}

export function TBody({ children }: { children: ReactNode }) {
  return <tbody className="divide-y divide-line">{children}</tbody>;
}

export function Th({ className, ...props }: ComponentProps<'th'>) {
  return (
    <th
      scope="col"
      className={cn(
        'h-10 whitespace-nowrap border-b border-line px-4 text-caption font-medium text-subtle',
        className,
      )}
      {...props}
    />
  );
}

export function Td({ className, ...props }: ComponentProps<'td'>) {
  return <td className={cn('h-12 px-4 align-middle', className)} {...props} />;
}
