import { ExternalLink, NotebookText, Play, Trash2 } from 'lucide-react';
import { useState } from 'react';
import { useControlPlane } from '@/data/ControlPlaneProvider';
import { useProjects } from '@/data/ProjectsProvider';
import { useWorkspace } from '@/data/useWorkspace';
import { CallsTable } from '@/domain/logs/CallsTable';
import { RunsTable } from '@/domain/runs/RunsTable';
import { Detail, Metrics, Panel, PanelActions, PanelBody, Section } from '@/shared/layout/Detail';
import { Page } from '@/shared/layout/Page';
import { NONE, formatRelative, formatTimestamp } from '@/shared/lib/format';
import { navigate, paths } from '@/shared/lib/paths';
import { Button, ButtonLink } from '@/shared/ui/Button';
import { useConfirm } from '@/shared/ui/ConfirmDialog';
import { EmptyState } from '@/shared/ui/EmptyState';
import { PropertyList } from '@/shared/ui/PropertyList';
import { StartRunDialog } from './StartRunDialog';

export function ProjectPage({ projectId }: { projectId: string }) {
  const { projectById, callsOf } = useWorkspace();
  const { profiles, attachSession, detachSession } = useControlPlane();
  const { saved, removeProject } = useProjects();
  const { confirm, dialog: confirmDialog } = useConfirm();
  const [starting, setStarting] = useState(false);
  const project = projectById(projectId);

  if (!project) {
    return (
      <Page trail={[{ label: 'Projects', href: paths.projects }, { label: 'Not found' }]}>
        <EmptyState icon={NotebookText} message="This project does not exist." />
      </Page>
    );
  }

  const calls = callsOf(project.runs);
  const savedAt = saved.find((entry) => entry.id === project.id)?.created_at;
  const latest = project.runs[0];

  function remove() {
    if (!project) return;
    const count = project.runs.length;
    confirm({
      title: 'Delete project',
      message:
        count > 0
          ? `“${project.name}” and its ${count} ${count === 1 ? 'run' : 'runs'} will be detached. Browser tabs close.`
          : `“${project.name}” will be removed from this browser.`,
      confirmLabel: 'Delete',
      onConfirm: async () => {
        for (const run of project.runs) await detachSession(run.id);
        removeProject(project.id);
        navigate(paths.projects);
      },
    });
  }

  const start = (
    <Button variant="primary" size="sm" icon={Play} onClick={() => setStarting(true)}>
      Start run
    </Button>
  );

  return (
    <Page trail={[{ label: 'Projects', href: paths.projects }, { label: project.name }]} action={start}>
      <div className="flex flex-col gap-8">
        <Metrics
          items={[
            { label: 'Runs', value: project.runs.length },
            { label: 'Busy kernels', value: project.runs.filter((run) => run.kernel_state === 'busy').length },
            { label: 'Tool calls', value: calls.length },
            { label: 'Failed calls', value: calls.filter((entry) => entry.status === 'error').length },
          ]}
        />
        <Detail
          main={
            <>
              <Section title="Runs" count={project.runs.length}>
                {project.runs.length === 0 ? (
                  <EmptyState icon={Play} message="No runs yet. Pick an account to open this notebook." action={start} />
                ) : (
                  <RunsTable runs={project.runs} />
                )}
              </Section>
              <Section title="Tool calls" count={calls.length}>
                {calls.length === 0 ? (
                  <EmptyState message="No tool calls recorded for this project." />
                ) : (
                  <CallsTable entries={calls} />
                )}
              </Section>
            </>
          }
          rail={
            <>
              <Section title="Notebook">
                <Panel>
                  <PanelBody>
                    <PropertyList
                      items={[
                        {
                          term: 'URL',
                          value: (
                            <a
                              href={project.notebookUrl}
                              target="_blank"
                              rel="noreferrer"
                              className="truncate font-mono text-caption transition-colors hover:text-muted"
                            >
                              {project.notebookUrl}
                            </a>
                          ),
                        },
                        { term: 'Saved', value: savedAt ? formatTimestamp(savedAt) : 'Not saved' },
                        {
                          term: 'Last attached',
                          value: latest ? formatRelative(latest.last_attached_at ?? latest.created_at) : NONE,
                        },
                      ]}
                    />
                  </PanelBody>
                </Panel>
              </Section>
              <Section title="Controls">
                <Panel>
                  <PanelActions>
                    <ButtonLink
                      variant="menu"
                      icon={ExternalLink}
                      href={project.notebookUrl}
                      target="_blank"
                      rel="noreferrer"
                    >
                      Open in Colab
                    </ButtonLink>
                  </PanelActions>
                  <PanelActions separated>
                    <Button variant="menu" icon={Trash2} onClick={remove}>
                      Delete project
                    </Button>
                  </PanelActions>
                </Panel>
              </Section>
            </>
          }
        />
      </div>
      {starting && (
        <StartRunDialog
          project={project}
          accounts={profiles}
          onSubmit={attachSession}
          onClose={() => setStarting(false)}
        />
      )}
      {confirmDialog}
    </Page>
  );
}
