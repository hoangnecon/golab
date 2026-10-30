import { useState } from 'react';
import type { FormEvent } from 'react';
import { errorMessage } from '@/api/http';
import type { Tier } from '@/api/types';
import { useControlPlane } from '@/data/ControlPlaneProvider';
import { formatTier } from '@/shared/lib/format';
import { Button } from '@/shared/ui/Button';
import { Dialog, DialogActions } from '@/shared/ui/Dialog';
import { FormError, Input, Select } from '@/shared/ui/Field';

const TIERS: readonly Tier[] = ['free', 'pro', 'pro_plus'];

export function CreateAccountDialog({ onClose }: { onClose: () => void }) {
  const { createProfile } = useControlPlane();
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [tier, setTier] = useState<Tier>('free');
  const [pending, setPending] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setPending(true);
    setError(null);
    try {
      await createProfile({ name: name.trim(), account_email: email.trim() || undefined, tier });
      onClose();
    } catch (failure) {
      setError(errorMessage(failure));
      setPending(false);
    }
  }

  return (
    <Dialog title="New account" onClose={onClose}>
      <form onSubmit={handleSubmit} className="flex flex-col gap-4">
        <Input label="Name" required value={name} disabled={pending} onChange={(event) => setName(event.target.value)} />
        <Input
          label="Google account"
          hint="Optional"
          type="email"
          value={email}
          disabled={pending}
          onChange={(event) => setEmail(event.target.value)}
        />
        <Select label="Colab plan" value={tier} disabled={pending} onChange={(event) => setTier(event.target.value as Tier)}>
          {TIERS.map((value) => (
            <option key={value} value={value}>
              {formatTier(value)}
            </option>
          ))}
        </Select>
        {error && <FormError>{error}</FormError>}
        <DialogActions>
          <Button variant="ghost" onClick={onClose} disabled={pending}>
            Cancel
          </Button>
          <Button variant="primary" type="submit" disabled={pending || !name.trim()}>
            Create account
          </Button>
        </DialogActions>
      </form>
    </Dialog>
  );
}
