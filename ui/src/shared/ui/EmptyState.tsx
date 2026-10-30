import type { LucideIcon } from 'lucide-react';
import type { ReactNode } from 'react';

interface EmptyStateProps {
  message: string;
  icon?: LucideIcon;
  /** The one next step, e.g. a primary Button. */
  action?: ReactNode;
}

export function EmptyState({ message, icon: Icon, action }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center gap-4 rounded-lg border border-dashed border-line-strong px-6 py-14 text-center">
      {Icon && (
        <span className="flex size-10 items-center justify-center rounded-full border border-line-strong bg-raised text-muted">
          <Icon aria-hidden className="size-5" strokeWidth={1.5} />
        </span>
      )}
      <p className="text-body text-muted">{message}</p>
      {action}
    </div>
  );
}
