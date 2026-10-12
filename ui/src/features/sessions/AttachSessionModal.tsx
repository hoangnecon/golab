import React, { useState } from 'react';
import { Modal } from '../../components/ui/Modal';
import { Input } from '../../components/ui/Input';
import { Select } from '../../components/ui/Select';
import { Button } from '../../components/ui/Button';
import { Profile } from '../../types';

interface AttachSessionModalProps {
  isOpen: boolean;
  onClose: () => void;
  profiles: Profile[];
  defaultProfileId: string;
  onSubmit: (data: { profile_id: string; notebook_url: string; name?: string }) => Promise<void>;
}

export function AttachSessionModal({
  isOpen,
  onClose,
  profiles,
  defaultProfileId,
  onSubmit,
}: AttachSessionModalProps) {
  const [url, setUrl] = useState('');
  const [name, setName] = useState('');
  const [profileId, setProfileId] = useState(defaultProfileId || (profiles[0]?.id ?? ''));
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!url.trim() || !profileId) return;
    setIsSubmitting(true);
    try {
      await onSubmit({
        profile_id: profileId,
        notebook_url: url.trim(),
        name: name.trim() || undefined,
      });
      setUrl('');
      setName('');
      onClose();
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Attach Colab Notebook"
      description="Connect GoLab directly to an open or new Google Colab notebook tab."
    >
      <form onSubmit={handleSubmit} className="space-y-4">
        <Input
          label="Notebook URL"
          type="url"
          required
          placeholder="https://colab.research.google.com/drive/..."
          value={url}
          onChange={(e) => setUrl(e.target.value)}
          className="font-mono text-xs"
        />

        <Input
          label="Display Name (Optional)"
          type="text"
          placeholder="e.g. Model Training v1"
          value={name}
          onChange={(e) => setName(e.target.value)}
        />

        <Select
          label="Target Profile"
          value={profileId}
          onChange={(e) => setProfileId(e.target.value)}
        >
          {profiles.map((p) => (
            <option key={p.id} value={p.id}>
              {p.name} ({p.tier.toUpperCase()} - Port {p.cdp_port})
            </option>
          ))}
        </Select>

        <div className="flex justify-end gap-2 pt-2 border-t border-zinc-800">
          <Button variant="ghost" size="sm" type="button" onClick={onClose} disabled={isSubmitting}>
            Cancel
          </Button>
          <Button variant="primary" size="sm" type="submit" disabled={isSubmitting}>
            {isSubmitting ? 'Attaching...' : 'Attach Notebook'}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
