import type { LucideIcon } from 'lucide-react';
import type { ComponentProps } from 'react';
import { cn } from '@/shared/lib/cn';

/**
 * primary   filled foreground (white on dark, black on light); one per view
 * secondary hairline outline on the surface
 * ghost     text only, for tertiary actions
 * menu      full-width, left-aligned row inside a `PanelActions` list
 */
export type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'menu';
export type ButtonSize = 'sm' | 'md';

const BASE =
  'inline-flex shrink-0 select-none items-center gap-2 whitespace-nowrap rounded text-body font-medium transition-colors disabled:pointer-events-none disabled:opacity-40';

const SIZE: Record<ButtonSize, string> = {
  sm: 'h-8 px-3',
  md: 'h-9 px-4',
};

const VARIANT: Record<ButtonVariant, string> = {
  primary: 'justify-center bg-fg text-canvas hover:bg-fg/85',
  secondary: 'justify-center border border-line-strong bg-surface text-fg hover:bg-hover',
  ghost: 'justify-center text-muted hover:bg-hover hover:text-fg',
  menu: 'w-full justify-start text-fg hover:bg-hover',
};

/** Menu rows ignore `size` padding so their icons align with panel text (see PanelActions). */
const MENU_PADDING = 'px-2';

const ICON: Record<ButtonVariant, string> = {
  primary: '',
  secondary: 'text-muted',
  ghost: '',
  menu: 'text-muted',
};

interface Shared {
  variant?: ButtonVariant;
  size?: ButtonSize;
  icon?: LucideIcon;
}

function classes(variant: ButtonVariant, size: ButtonSize, className?: string) {
  return cn(BASE, SIZE[size], VARIANT[variant], variant === 'menu' && MENU_PADDING, className);
}

export function Button({
  variant = 'secondary',
  size = 'md',
  icon: Icon,
  type = 'button',
  className,
  children,
  ...props
}: Shared & ComponentProps<'button'>) {
  return (
    <button type={type} className={classes(variant, size, className)} {...props}>
      {Icon && <Icon aria-hidden className={cn('size-4 shrink-0', ICON[variant])} strokeWidth={1.75} />}
      {children}
    </button>
  );
}

/** Navigation that looks like a Button (e.g. opening the notebook in a new tab). */
export function ButtonLink({
  variant = 'secondary',
  size = 'md',
  icon: Icon,
  className,
  children,
  ...props
}: Shared & ComponentProps<'a'>) {
  return (
    <a className={classes(variant, size, className)} {...props}>
      {Icon && <Icon aria-hidden className={cn('size-4 shrink-0', ICON[variant])} strokeWidth={1.75} />}
      {children}
    </a>
  );
}
