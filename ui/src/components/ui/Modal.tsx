import React, { useEffect } from 'react';
import { X } from 'lucide-react';
import { cn } from '../../lib/utils';

export interface ModalProps {
  isOpen: boolean;
  onClose: () => void;
  title: string;
  description?: string;
  children: React.ReactNode;
  className?: string;
}

export function Modal({ isOpen, onClose, title, description, children, className }: ModalProps) {
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    if (isOpen) {
      window.addEventListener('keydown', handleKeyDown);
    }
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black/70 backdrop-blur-sm transition-opacity"
        onClick={onClose}
      />

      {/* Dialog container */}
      <div
        className={cn(
          'relative w-full max-w-md max-h-[90vh] overflow-y-auto bg-[#121215] border border-zinc-800 rounded-lg shadow-2xl p-5 z-10 flex flex-col gap-4 animate-in fade-in-0 zoom-in-95 duration-150',
          className
        )}
      >
        <div className="flex items-start justify-between gap-3">
          <div>
            <h2 className="text-sm font-semibold text-zinc-100">{title}</h2>
            {description && <p className="text-xs text-zinc-500 mt-0.5">{description}</p>}
          </div>
          <button
            onClick={onClose}
            className="text-zinc-500 hover:text-zinc-300 p-0.5 rounded transition-colors"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        <div>{children}</div>
      </div>
    </div>
  );
}
