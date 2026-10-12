import React from 'react';
import { cn } from '../../lib/utils';

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'outline' | 'ghost' | 'destructive';
  size?: 'sm' | 'md' | 'icon';
}

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = 'secondary', size = 'sm', disabled, children, ...props }, ref) => {
    const variantStyles = {
      primary: 'bg-zinc-100 text-zinc-950 hover:bg-zinc-200 font-medium shadow-sm active:scale-[0.98]',
      secondary: 'bg-zinc-900 hover:bg-zinc-800 text-zinc-200 border border-zinc-800 hover:border-zinc-700',
      outline: 'bg-transparent hover:bg-zinc-900 text-zinc-300 border border-zinc-800 hover:border-zinc-700',
      ghost: 'bg-transparent hover:bg-zinc-900 text-zinc-400 hover:text-zinc-100',
      destructive: 'bg-red-950/60 hover:bg-red-900/60 text-red-300 border border-red-800/50',
    };

    const sizeStyles = {
      sm: 'h-7 px-2.5 text-xs rounded-md gap-1.5',
      md: 'h-9 px-3.5 text-sm rounded-md gap-2',
      icon: 'h-7 w-7 p-0 rounded-md justify-center',
    };

    return (
      <button
        ref={ref}
        disabled={disabled}
        className={cn(
          'inline-flex items-center justify-center font-sans transition-colors focus:outline-none focus:ring-1 focus:ring-zinc-600 disabled:opacity-40 disabled:pointer-events-none select-none cursor-pointer',
          variantStyles[variant],
          sizeStyles[size],
          className
        )}
        {...props}
      >
        {children}
      </button>
    );
  }
);
Button.displayName = 'Button';
