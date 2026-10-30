import type { ComponentProps } from 'react';
import { cn } from '@/shared/lib/cn';

export type IconButtonSize = 'sm' | 'md';

const BASE =
  'inline-flex shrink-0 items-center justify-center rounded text-muted transition-colors hover:bg-hover hover:text-fg disabled:pointer-events-none disabled:opacity-40';

/** sm (24px) sits in the trailing slot of 32px sidebar rows; md (32px) everywhere else. */
const SIZE: Record<IconButtonSize, string> = {
  sm: 'size-6',
  md: 'size-8',
};

interface IconButtonProps extends Omit<ComponentProps<'button'>, 'aria-label'> {
  /** Required: icon-only controls need an accessible name. Also used as tooltip. */
  label: string;
  size?: IconButtonSize;
}

export function IconButton({ label, size = 'md', type = 'button', className, children, ...props }: IconButtonProps) {
  return (
    <button type={type} aria-label={label} title={label} className={cn(BASE, SIZE[size], className)} {...props}>
      {children}
    </button>
  );
}

interface IconLinkProps extends Omit<ComponentProps<'a'>, 'aria-label'> {
  label: string;
  size?: IconButtonSize;
}

/** Same look as IconButton, for icon-only navigation. */
export function IconLink({ label, size = 'md', className, children, ...props }: IconLinkProps) {
  return (
    <a aria-label={label} title={label} className={cn(BASE, SIZE[size], className)} {...props}>
      {children}
    </a>
  );
}
