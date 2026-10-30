import { cn } from '@/shared/lib/cn';

/** Round monogram for an account: first letter of its name. */
export function Avatar({ name, className }: { name: string; className?: string }) {
  return (
    <span
      aria-hidden
      className={cn(
        'flex size-8 shrink-0 items-center justify-center rounded-full border border-line-strong bg-raised text-caption font-semibold uppercase text-fg',
        className,
      )}
    >
      {name.trim().charAt(0) || '?'}
    </span>
  );
}
