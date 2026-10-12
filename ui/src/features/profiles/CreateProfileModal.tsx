import React, { useState } from 'react';
import { Modal } from '../../components/ui/Modal';
import { Input } from '../../components/ui/Input';
import { Select } from '../../components/ui/Select';
import { Button } from '../../components/ui/Button';

interface CreateProfileModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: { name: string; account_email?: string; tier: string }) => Promise<void>;
}

export function CreateProfileModal({ isOpen, onClose, onSubmit }: CreateProfileModalProps) {
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [tier, setTier] = useState('free');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;
    setIsSubmitting(true);
    try {
      await onSubmit({
        name: name.trim(),
        account_email: email.trim() || undefined,
        tier,
      });
      setName('');
      setEmail('');
      setTier('free');
      onClose();
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Add Chrome Profile"
      description="Creates an isolated profile folder and assigns an independent CDP debugging port."
    >
      <form onSubmit={handleSubmit} className="space-y-4">
        <Input
          label="Profile Name"
          type="text"
          required
          placeholder="e.g. Work Account"
          value={name}
          onChange={(e) => setName(e.target.value)}
        />

        <Input
          label="Account Email (Optional)"
          type="email"
          placeholder="user@gmail.com"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
        />

        <Select label="Colab Subscription Tier" value={tier} onChange={(e) => setTier(e.target.value)}>
          <option value="free">Free</option>
          <option value="pro">Pro</option>
          <option value="pro_plus">Pro+</option>
        </Select>

        <div className="flex justify-end gap-2 pt-2 border-t border-zinc-800">
          <Button variant="ghost" size="sm" type="button" onClick={onClose} disabled={isSubmitting}>
            Cancel
          </Button>
          <Button variant="primary" size="sm" type="submit" disabled={isSubmitting}>
            {isSubmitting ? 'Creating...' : 'Create Profile'}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
