import { cn } from '@/shared/lib/cn';

export type Tone = 'ok' | 'warn' | 'bad' | 'idle';

/**
 * Monochrome palette, so state is carried by shape, never by hue:
 * ok   solid disc          (ready, running, success)
 * warn hollow ring         (busy, starting: work in progress)
 * bad  solid diamond       (error, dead)
 * idle dim disc            (stopped, unknown)
 */
const SHAPE: Record<Tone, string> = {
  ok: 'rounded-full bg-fg',
  warn: 'rounded-full border-2 border-fg',
  bad: 'rotate-45 rounded-none bg-fg',
  idle: 'rounded-full bg-line-strong',
};

interface StatusDotProps {
  tone: Tone;
  /** Spoken label; shape alone never carries meaning. */
  label: string;
  className?: string;
}

export function StatusDot({ tone, label, className }: StatusDotProps) {
  return (
    <span
      role="img"
      aria-label={label}
      title={label}
      className={cn('inline-flex size-2 shrink-0', SHAPE[tone], className)}
    />
  );
}

/** Dot plus its label, for places with room for text. */
export function Status({ tone, label, className }: StatusDotProps) {
  return (
    <span className={cn('inline-flex items-center gap-2', className)}>
      <StatusDot tone={tone} label={label} />
      <span aria-hidden>{label}</span>
    </span>
  );
}
