import { useState } from 'react';
import { Button } from './Button';
import { Dialog, DialogActions } from './Dialog';

interface ConfirmRequest {
  title: string;
  message: string;
  confirmLabel: string;
  onConfirm: () => void | Promise<void>;
}

/**
 * Usage: const { confirm, dialog } = useConfirm();
 *        confirm({ ... });  and render {dialog} once in the page.
 */
export function useConfirm() {
  const [request, setRequest] = useState<ConfirmRequest | null>(null);
  const dialog = request && <ConfirmDialog {...request} onClose={() => setRequest(null)} />;
  return { confirm: setRequest, dialog };
}

function ConfirmDialog({
  title,
  message,
  confirmLabel,
  onConfirm,
  onClose,
}: ConfirmRequest & { onClose: () => void }) {
  const [pending, setPending] = useState(false);

  async function handleConfirm() {
    setPending(true);
    await onConfirm();
    onClose();
  }

  return (
    <Dialog title={title} onClose={onClose}>
      <p className="text-body text-muted">{message}</p>
      <DialogActions>
        <Button variant="ghost" onClick={onClose} disabled={pending}>
          Cancel
        </Button>
        <Button variant="primary" onClick={handleConfirm} disabled={pending}>
          {confirmLabel}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
