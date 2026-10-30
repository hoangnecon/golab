import { ChevronRight, NotebookText, Plus } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useWorkspace } from '@/data/useWorkspace';
import { ROW } from '@/shared/layout/Sidebar';
import { cn } from '@/shared/lib/cn';
import { useLocation } from '@/shared/lib/location';
import { navigate, paths } from '@/shared/lib/paths';
import { kernelLabel, kernelTone } from '@/shared/lib/status';
import { IconButton } from '@/shared/ui/IconButton';
import { StatusDot } from '@/shared/ui/StatusDot';
import { CreateProjectDialog } from './CreateProjectDialog';

/**
 * Sidebar tree on the sidebar row grid: icons centred on x = 28, every label
 * starts at x = 48, and the trailing 24px slot holds the chevron (projects),
 * the status shape (runs) or the add button (header). Selection mirrors the URL.
 */
export function ProjectsNav() {
  const { projects } = useWorkspace();
  const { segments } = useLocation();
  const [creating, setCreating] = useState(false);
  const [collapsed, setCollapsed] = useState<ReadonlySet<string>>(new Set());

  const inProjects = segments[0] === 'projects';
  const projectId = inProjects ? segments[1] : undefined;
  const runId = inProjects && segments[2] === 'runs' ? segments[3] : undefined;

  useEffect(() => {
    if (!projectId) return;
    setCollapsed((current) => {
      if (!current.has(projectId)) return current;
      const next = new Set(current);
      next.delete(projectId);
      return next;
    });
  }, [projectId]);

  function toggle(id: string) {
    setCollapsed((current) => {
      const next = new Set(current);
      if (!next.delete(id)) next.add(id);
      return next;
    });
  }

  return (
    <>
      <div className={cn(ROW, 'justify-between pl-3 pr-1')}>
        <a
          href={paths.projects}
          aria-current={inProjects && !projectId ? 'page' : undefined}
          className="text-caption font-medium text-subtle transition-colors hover:text-fg"
        >
          Projects
        </a>
        <IconButton size="sm" label="New project" onClick={() => setCreating(true)}>
          <Plus aria-hidden className="size-4" strokeWidth={1.75} />
        </IconButton>
      </div>

      {projects.length === 0 ? (
        <p className="py-2 pl-10 pr-3 text-caption text-subtle">No projects yet</p>
      ) : (
        <ul className="flex flex-col gap-0.5">
          {projects.map((project) => {
            const expanded = !collapsed.has(project.id);
            const selected = project.id === projectId && !runId;
            return (
              <li key={project.id}>
                <div className={cn(ROW, 'pr-1 hover:bg-hover', selected && 'bg-hover')}>
                  <a
                    href={paths.project(project.id)}
                    aria-current={selected ? 'page' : undefined}
                    className={cn(
                      'flex h-full min-w-0 flex-1 items-center gap-3 pl-3',
                      selected ? 'font-medium text-fg' : 'text-muted hover:text-fg',
                    )}
                  >
                    <NotebookText aria-hidden className="size-4 shrink-0" strokeWidth={1.75} />
                    <span className="truncate">{project.name}</span>
                  </a>
                  <IconButton
                    size="sm"
                    aria-expanded={expanded}
                    label={`${expanded ? 'Collapse' : 'Expand'} ${project.name}`}
                    onClick={() => toggle(project.id)}
                  >
                    <ChevronRight
                      aria-hidden
                      strokeWidth={1.75}
                      className={cn('size-4 transition-transform', expanded && 'rotate-90')}
                    />
                  </IconButton>
                </div>
                {expanded && project.runs.length > 0 && (
                  <ul className="relative flex flex-col gap-0.5 pt-0.5">
                    <span aria-hidden className="absolute bottom-1 left-5 top-1 w-px bg-line" />
                    {project.runs.map((run) => {
                      const current = run.id === runId;
                      return (
                        <li key={run.id}>
                          <a
                            href={paths.run(project.id, run.id)}
                            aria-current={current ? 'page' : undefined}
                            className={cn(
                              ROW,
                              'pl-10 pr-1',
                              current ? 'bg-hover font-medium text-fg' : 'text-muted hover:bg-hover hover:text-fg',
                            )}
                          >
                            <span className="min-w-0 flex-1 truncate">{run.name}</span>
                            <span className="flex size-6 shrink-0 items-center justify-center">
                              <StatusDot tone={kernelTone(run.kernel_state)} label={kernelLabel(run.kernel_state)} />
                            </span>
                          </a>
                        </li>
                      );
                    })}
                  </ul>
                )}
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
    </>
  );
}
