import React from 'react';
import { cn } from '../../lib/utils';

export interface BadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
  variant?: 'neutral' | 'success' | 'warning' | 'error';
  dot?: boolean;
}

export function Badge({ className, variant = 'neutral', dot = false, children, ...props }: BadgeProps) {
  const variants = {
    neutral: 'bg-zinc-900 text-zinc-400 border-zinc-800',
    success: 'bg-emerald-950/40 text-emerald-400 border-emerald-800/40',
    warning: 'bg-amber-950/40 text-amber-400 border-amber-800/40',
    error: 'bg-red-950/40 text-red-400 border-red-800/40',
  };

  const dotColors = {
    neutral: 'bg-zinc-500',
    success: 'bg-emerald-500',
    warning: 'bg-amber-400 animate-pulse',
    error: 'bg-red-500',
  };

  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 px-2 py-0.5 rounded text-[11px] font-mono border whitespace-nowrap flex-shrink-0 select-none',
        variants[variant],
        className
      )}
      {...props}
    >
      {dot && <span className={cn('w-1.5 h-1.5 rounded-full flex-shrink-0', dotColors[variant])} />}
      {children}
    </span>
  );
}
