import { Check, ChevronDown, CircleAlert, Search } from 'lucide-react';
import type { ComponentProps, ReactNode } from 'react';
import { useId } from 'react';
import { cn } from '@/shared/lib/cn';

const CONTROL =
  'h-9 w-full rounded border border-line-strong bg-canvas px-3 text-body text-fg placeholder:text-subtle transition-colors hover:border-subtle disabled:opacity-50';

interface FieldProps {
  label: string;
  /** Short qualifier shown beside the label, e.g. "Optional". */
  hint?: string;
  /** Keep the label for assistive tech but do not render it (toolbars). */
  hideLabel?: boolean;
}

function Field({
  label,
  hint,
  hideLabel,
  id,
  children,
}: FieldProps & { id: string; children: ReactNode }) {
  return (
    <div className="flex flex-col gap-1.5">
      <label
        htmlFor={id}
        className={cn(
          'flex items-baseline justify-between text-caption font-medium text-muted',
          hideLabel && 'sr-only',
        )}
      >
        <span>{label}</span>
        {hint && <span className="font-normal text-subtle">{hint}</span>}
      </label>
      {children}
    </div>
  );
}

export function Input({
  label,
  hint,
  hideLabel,
  id,
  className,
  ...props
}: FieldProps & ComponentProps<'input'>) {
  const generated = useId();
  const inputId = id ?? generated;
  return (
    <Field label={label} hint={hint} hideLabel={hideLabel} id={inputId}>
      <input id={inputId} className={cn(CONTROL, className)} {...props} />
    </Field>
  );
}

/** Search box with a leading magnifier; label stays available to assistive tech. */
export function SearchInput({
  label,
  className,
  ...props
}: { label: string } & Omit<ComponentProps<'input'>, 'type'>) {
  const id = useId();
  return (
    <Field label={label} hideLabel id={id}>
      <div className={cn('relative', className)}>
        <Search
          aria-hidden
          strokeWidth={1.75}
          className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-subtle"
        />
        <input id={id} type="search" className={cn(CONTROL, 'pl-9')} {...props} />
      </div>
    </Field>
  );
}

export function Select({
  label,
  hint,
  hideLabel,
  id,
  className,
  children,
  ...props
}: FieldProps & ComponentProps<'select'>) {
  const generated = useId();
  const selectId = id ?? generated;
  return (
    <Field label={label} hint={hint} hideLabel={hideLabel} id={selectId}>
      <div className={cn('relative', className)}>
        <select id={selectId} className={cn(CONTROL, 'appearance-none pr-9')} {...props}>
          {children}
        </select>
        <ChevronDown
          aria-hidden
          strokeWidth={1.75}
          className="pointer-events-none absolute right-3 top-1/2 size-4 -translate-y-1/2 text-subtle"
        />
      </div>
    </Field>
  );
}

export function Checkbox({
  label,
  className,
  ...props
}: { label: string } & Omit<ComponentProps<'input'>, 'type'>) {
  return (
    <label className="flex h-9 cursor-pointer items-center gap-2 text-body text-muted">
      <span className="relative flex size-4 shrink-0">
        <input
          type="checkbox"
          className={cn(
            'peer size-4 cursor-pointer appearance-none rounded border border-line-strong bg-canvas transition-colors hover:border-subtle checked:border-fg checked:bg-fg',
            className,
          )}
          {...props}
        />
        <Check
          aria-hidden
          strokeWidth={3}
          className="pointer-events-none absolute inset-0 m-auto size-3 text-canvas opacity-0 peer-checked:opacity-100"
        />
      </span>
      {label}
    </label>
  );
}

/** Inline form error. The dialog stays open; the message replaces nothing else. */
export function FormError({ children }: { children: ReactNode }) {
  return (
    <p role="alert" className="flex items-start gap-2 rounded border border-line-strong bg-raised px-3 py-2 text-caption text-fg">
      <CircleAlert aria-hidden className="mt-px size-4 shrink-0" strokeWidth={1.75} />
      <span className="min-w-0 break-words">{children}</span>
    </p>
  );
}
