import { Lock, NotebookText, Play, Plus, ShieldCheck, User } from 'lucide-react';
import { useState } from 'react';
import { useWorkspace } from '@/data/useWorkspace';
import { Page } from '@/shared/layout/Page';
import { formatRelative } from '@/shared/lib/format';
import { navigate, paths } from '@/shared/lib/paths';
import { kernelLabel, kernelTone } from '@/shared/lib/status';
import { Button } from '@/shared/ui/Button';
import { EmptyState } from '@/shared/ui/EmptyState';
import { StatusDot } from '@/shared/ui/StatusDot';
import { CreateProjectDialog } from './CreateProjectDialog';

export function ProjectsIndex() {
  const { projects, accountById } = useWorkspace();
  const [creating, setCreating] = useState(false);

  const create = (
    <Button variant="primary" size="sm" icon={Plus} onClick={() => setCreating(true)}>
      New project
    </Button>
  );

  return (
    <Page
      trail={[{ label: 'Projects' }]}
      meta={projects.length > 0 ? String(projects.length) : undefined}
      action={projects.length > 0 ? create : undefined}
    >
      {projects.length === 0 ? (
        <EmptyState
          icon={NotebookText}
          message="A project is a Colab notebook. Create one, then start runs on it with any account."
          action={create}
        />
      ) : (
        <ul className="grid grid-cols-1 gap-4 md:grid-cols-2 2xl:grid-cols-3">
          {projects.map((project) => {
            const latest = project.runs[0];
            const owner = project.ownerProfileId ? accountById(project.ownerProfileId) : undefined;
            const activeRun = project.runs.find((r) => r.is_active) ?? latest;

            return (
              <li key={project.id}>
                <a
                  href={paths.project(project.id)}
                  className="group flex h-full flex-col rounded-xl border border-line bg-surface p-4 transition-all hover:border-line-strong hover:shadow-overlay"
                >
                  <div className="flex items-start gap-3">
                    <span className="flex size-10 shrink-0 items-center justify-center rounded-lg border border-line-strong bg-raised text-primary">
                      <NotebookText aria-hidden className="size-5" strokeWidth={1.75} />
                    </span>
                    <span className="flex min-w-0 flex-col">
                      <span className="truncate font-medium text-fg group-hover:text-primary transition-colors">
                        {project.name}
                      </span>
                      <span className="truncate font-mono text-caption text-subtle">{project.notebookUrl}</span>
                    </span>
                  </div>

                  {/* Owner Profile & Permission Badges */}
                  <div className="mt-3 flex flex-wrap items-center gap-2 text-caption">
                    {owner && (
                      <span className="inline-flex items-center gap-1 rounded bg-raised px-2 py-0.5 text-muted border border-line">
                        <User aria-hidden className="size-3 text-subtle" />
                        <span className="truncate">{owner.name}</span>
                      </span>
                    )}
                    {activeRun?.permission && activeRun.permission !== 'unknown' && (
                      <span
                        className={`inline-flex items-center gap-1 rounded px-2 py-0.5 border ${
                          activeRun.permission === 'edit'
                            ? 'bg-success/10 text-success border-success/30'
                            : activeRun.permission === 'view'
                            ? 'bg-warning/10 text-warning border-warning/30'
                            : 'bg-danger/10 text-danger border-danger/30'
                        }`}
                      >
                        <ShieldCheck aria-hidden className="size-3" />
                        {activeRun.permission === 'edit' ? 'Full Access' : activeRun.permission === 'view' ? 'Read Only' : 'Denied'}
                      </span>
                    )}
                    {activeRun?.is_locked && (
                      <span className="inline-flex items-center gap-1 rounded bg-primary/10 px-2 py-0.5 text-primary border border-primary/30">
                        <Lock aria-hidden className="size-3" />
                        Locked
                      </span>
                    )}
                  </div>

                  <div className="mt-4 flex items-center gap-3 border-t border-line pt-3 text-caption text-muted">
                    <span className="tabular-nums flex items-center gap-1">
                      <Play aria-hidden className="size-3 text-subtle" />
                      {project.runs.length} {project.runs.length === 1 ? 'run' : 'runs'}
                    </span>
                    {project.runs.length > 0 && (
                      <span className="flex items-center gap-1.5">
                        {project.runs.map((run) => (
                          <StatusDot
                            key={run.id}
                            tone={kernelTone(run.kernel_state)}
                            label={`${run.name}: ${kernelLabel(run.kernel_state)}`}
                          />
                        ))}
                      </span>
                    )}
                    {latest && (
                      <span className="ml-auto text-subtle">
                        {formatRelative(latest.last_attached_at ?? latest.created_at)}
                      </span>
                    )}
                  </div>
                </a>
              </li>
            );
          })}
        </ul>
      )}
      {creating && (
        <CreateProjectDialog
          onCreated={(id) => navigate(paths.project(id))}
          onClose={() => setCreating(false)}
        />
      )}
    </Page>
  );
}
