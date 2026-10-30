import { useState } from 'react';
import type { FormEvent } from 'react';
import { errorMessage } from '@/api/http';
import type { NewSession, Profile } from '@/api/types';
import type { Project } from '@/data/useWorkspace';
import { paths } from '@/shared/lib/paths';
import { Button } from '@/shared/ui/Button';
import { Dialog, DialogActions } from '@/shared/ui/Dialog';
import { FormError, Input, Select } from '@/shared/ui/Field';

interface StartRunDialogProps {
  project: Project;
  accounts: Profile[];
  onSubmit: (input: NewSession) => Promise<void>;
  onClose: () => void;
}

/** A run = this project's notebook opened under one account. */
export function StartRunDialog({ project, accounts, onSubmit, onClose }: StartRunDialogProps) {
  const preferred = accounts.find((account) => account.is_running) ?? accounts[0];
  const [accountId, setAccountId] = useState(preferred?.id ?? '');
  const [name, setName] = useState('');
  const [pending, setPending] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setPending(true);
    setError(null);
    try {
      await onSubmit({
        profile_id: accountId,
        notebook_url: project.notebookUrl,
        name: name.trim() || undefined,
      });
      onClose();
    } catch (failure) {
      setError(errorMessage(failure));
      setPending(false);
    }
  }

  return (
    <Dialog title="Start run" onClose={onClose}>
      <form onSubmit={handleSubmit} className="flex flex-col gap-4">
        <Select
          label="Account"
          required
          value={accountId}
          onChange={(event) => setAccountId(event.target.value)}
          disabled={accounts.length === 0 || pending}
        >
          {accounts.map((account) => (
            <option key={account.id} value={account.id}>
              {account.name}
              {account.is_running ? ' · browser running' : ''}
            </option>
          ))}
        </Select>
        {accounts.length === 0 && (
          <p className="text-caption text-muted">
            A run needs an account.{' '}
            <a href={paths.accounts} className="font-medium text-fg underline underline-offset-2" onClick={onClose}>
              Create one
            </a>
          </p>
        )}
        <Input
          label="Run name"
          hint="Optional"
          value={name}
          disabled={pending}
          onChange={(event) => setName(event.target.value)}
        />
        {pending && (
          <p role="status" className="text-caption text-muted">
            Opening the notebook. Launching the browser can take a few seconds.
          </p>
        )}
        {error && <FormError>{error}</FormError>}
        <DialogActions>
          <Button variant="ghost" onClick={onClose} disabled={pending}>
            Cancel
          </Button>
          <Button variant="primary" type="submit" disabled={pending || !accountId}>
            Start run
          </Button>
        </DialogActions>
      </form>
    </Dialog>
  );
}
