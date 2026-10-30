import type { Session } from '@/api/types';
import { projectIdFor } from '@/data/projects';
import { useWorkspace } from '@/data/useWorkspace';
import { NONE, formatRelative, formatVram, vramUsedFraction } from '@/shared/lib/format';
import { paths } from '@/shared/lib/paths';
import { kernelLabel, kernelTone } from '@/shared/lib/status';
import { Meter } from '@/shared/ui/Meter';
import { Status } from '@/shared/ui/StatusDot';
import { TBody, TableCard, Td, Th } from '@/shared/ui/Table';

interface RunsTableProps {
  runs: Session[];
  /** Show the project column (used where runs of several projects mix). */
  showProject?: boolean;
  /** Hide the account column when every row belongs to the same account. */
  hideAccount?: boolean;
}

export function runHref(run: Session): string {
  return paths.run(projectIdFor(run.notebook_url), run.id);
}

const LINK = 'transition-colors hover:text-fg';

/** Runs with their live hardware state, one line per row. Controls live on the run page. */
export function RunsTable({ runs, showProject, hideAccount }: RunsTableProps) {
  const { accountById, projectOfRun } = useWorkspace();

  return (
    <TableCard>
      <thead>
        <tr>
          <Th>Run</Th>
          <Th className="w-28">Kernel</Th>
          {showProject && <Th>Project</Th>}
          {!hideAccount && <Th>Account</Th>}
          <Th>GPU</Th>
          <Th className="w-56">VRAM free</Th>
          <Th className="w-32 text-right">Attached</Th>
        </tr>
      </thead>
      <TBody>
        {runs.map((run) => {
          const account = accountById(run.profile_id);
          const project = projectOfRun(run);
          const vram = formatVram(run.gpu_memory_free_mb, run.gpu_memory_total_mb);
          const used = vramUsedFraction(run.gpu_memory_free_mb, run.gpu_memory_total_mb);
          return (
            <tr key={run.id} className="transition-colors hover:bg-hover">
              <Td className="whitespace-nowrap">
                <a href={runHref(run)} className="flex min-w-0 items-baseline gap-2">
                  <span className="truncate font-medium text-fg hover:underline">{run.name}</span>
                  {run.is_active && <span className="shrink-0 text-caption text-subtle">Active</span>}
                </a>
              </Td>
              <Td className="whitespace-nowrap text-muted">
                <Status tone={kernelTone(run.kernel_state)} label={kernelLabel(run.kernel_state)} />
              </Td>
              {showProject && (
                <Td className="whitespace-nowrap text-muted">
                  {project ? (
                    <a href={paths.project(project.id)} className={LINK}>
                      {project.name}
                    </a>
                  ) : (
                    NONE
                  )}
                </Td>
              )}
              {!hideAccount && (
                <Td className="whitespace-nowrap text-muted">
                  {account ? (
                    <a href={paths.account(account.id)} className={LINK}>
                      {account.name}
                    </a>
                  ) : (
                    NONE
                  )}
                </Td>
              )}
              <Td className="whitespace-nowrap text-muted">{run.gpu_name ?? NONE}</Td>
              <Td className="text-muted">
                {vram ? (
                  <div className="flex items-center gap-3">
                    <span className="w-24 shrink-0 whitespace-nowrap tabular-nums">{vram}</span>
                    {used !== null && <Meter value={used} label="VRAM in use" className="w-20" />}
                  </div>
                ) : (
                  NONE
                )}
              </Td>
              <Td className="whitespace-nowrap text-right text-muted">
                {formatRelative(run.last_attached_at ?? run.created_at)}
              </Td>
            </tr>
          );
        })}
      </TBody>
    </TableCard>
  );
}
