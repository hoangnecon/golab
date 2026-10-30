import { cn } from '@/shared/lib/cn';

interface MeterProps {
  /** 0..1 */
  value: number;
  label: string;
  className?: string;
}

/** Thin usage bar. Neutral fill; the number next to it carries the exact value. */
export function Meter({ value, label, className }: MeterProps) {
  const clamped = Math.min(1, Math.max(0, value));
  return (
    <div
      role="meter"
      aria-label={label}
      aria-valuemin={0}
      aria-valuemax={100}
      aria-valuenow={Math.round(clamped * 100)}
      className={cn('h-1 w-full overflow-hidden rounded-full bg-line', className)}
    >
      <div className="h-full rounded-full bg-fg transition-all" style={{ width: `${clamped * 100}%` }} />
    </div>
  );
}
