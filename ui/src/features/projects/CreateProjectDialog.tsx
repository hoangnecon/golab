import { useState } from 'react';
import type { FormEvent } from 'react';
import { useControlPlane } from '@/data/ControlPlaneProvider';
import { useProjects } from '@/data/ProjectsProvider';
import { Button } from '@/shared/ui/Button';
import { Dialog, DialogActions } from '@/shared/ui/Dialog';
import { FormError, Input, Select } from '@/shared/ui/Field';

interface CreateProjectDialogProps {
  onCreated: (projectId: string) => void;
  onClose: () => void;
}

export function CreateProjectDialog({ onCreated, onClose }: CreateProjectDialogProps) {
  const { createProject } = useProjects();
  const { profiles } = useControlPlane();
  const [name, setName] = useState('');
  const [url, setUrl] = useState('');
  const [ownerProfileId, setOwnerProfileId] = useState(profiles[0]?.id ?? 'default');
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSubmitting(true);
    setError(null);
    try {
      const project = await createProject(name, url, ownerProfileId);
      onClose();
      onCreated(project.id);
    } catch (failure) {
      setError(failure instanceof Error ? failure.message : 'Could not create the project');
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Dialog title="New project" onClose={onClose}>
      <form onSubmit={handleSubmit} className="flex flex-col gap-4">
        <Input label="Project Name" required value={name} onChange={(event) => setName(event.target.value)} placeholder="e.g. LLM Fine-tuning Notebook" />
        <Input
          label="Notebook URL"
          type="url"
          required
          value={url}
          onChange={(event) => {
            setUrl(event.target.value);
            setError(null);
          }}
          placeholder="https://colab.research.google.com/drive/…"
          className="font-mono"
        />
        {profiles.length > 0 && (
          <Select
            label="Owner Google Profile"
            value={ownerProfileId}
            onChange={(event) => setOwnerProfileId(event.target.value)}
          >
            {profiles.map((p) => (
              <option key={p.id} value={p.id}>
                {p.name} ({p.account_email || p.tier})
              </option>
            ))}
          </Select>
        )}
        {error && <FormError>{error}</FormError>}
        <DialogActions>
          <Button variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button variant="primary" type="submit" disabled={!name.trim() || !url.trim() || submitting}>
            {submitting ? 'Creating...' : 'Create project'}
          </Button>
        </DialogActions>
      </form>
    </Dialog>
  );
}
