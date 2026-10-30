import { X } from 'lucide-react';
import { useEffect, useId, useRef } from 'react';
import type { ReactNode } from 'react';
import { IconButton } from './IconButton';

interface DialogProps {
  title: string;
  onClose: () => void;
  children: ReactNode;
}

/**
 * Modal built on the native <dialog>: focus trap, Escape, and top-layer
 * stacking come from the platform. Mount it only while open so its content
 * state resets on every open.
 */
export function Dialog({ title, onClose, children }: DialogProps) {
  const ref = useRef<HTMLDialogElement>(null);
  const titleId = useId();

  useEffect(() => {
    const element = ref.current;
    if (element && !element.open) element.showModal();
  }, []);

  return (
    <dialog
      ref={ref}
      aria-labelledby={titleId}
      onClose={onClose}
      onClick={(event) => {
        if (event.target === event.currentTarget) event.currentTarget.close();
      }}
      className="m-auto w-11/12 max-w-md rounded-xl border-0 bg-surface p-0 text-fg shadow-overlay backdrop:bg-scrim"
    >
      <div className="flex h-14 items-center justify-between pl-6 pr-3">
        <h2 id={titleId} className="text-title font-semibold">
          {title}
        </h2>
        <IconButton label="Close" onClick={() => ref.current?.close()}>
          <X aria-hidden className="size-4" strokeWidth={1.75} />
        </IconButton>
      </div>
      <div className="flex flex-col gap-5 px-6 pb-6 pt-1">{children}</div>
    </dialog>
  );
}

export function DialogActions({ children }: { children: ReactNode }) {
  return <div className="-mx-6 -mb-6 mt-1 flex justify-end gap-2 rounded-b-xl border-t border-line bg-raised px-6 py-4">{children}</div>;
}
