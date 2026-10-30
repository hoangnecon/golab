import { ChevronRight } from 'lucide-react';
import { Fragment, useState } from 'react';
import type { ActivityLog } from '@/api/types';
import { useWorkspace } from '@/data/useWorkspace';
import { runHref } from '@/domain/runs/RunsTable';
import { CallDetail } from './CallDetail';
import { NONE, formatLatency, formatTimestamp, shortId } from '@/shared/lib/format';
import { cn } from '@/shared/lib/cn';
import { Status, StatusDot } from '@/shared/ui/StatusDot';
import { TBody, TableCard, Td, Th } from '@/shared/ui/Table';

interface CallsTableProps {
  entries: ActivityLog[];
  /** Hide the run column when every row belongs to the same run. */
  hideRun?: boolean;
}

/**
 * Tool-call history. Any row expands in place to show its request and response
 * payloads (and the error for failed calls). Clicking the row or its chevron
 * toggles it; links inside the row keep their own behaviour.
 */
export function CallsTable({ entries, hideRun }: CallsTableProps) {
  const [openId, setOpenId] = useState<number | null>(null);
  const { runById } = useWorkspace();
  const columns = hideRun ? 5 : 6;

  return (
    <TableCard>
      <thead>
        <tr>
          <Th className="w-28">Result</Th>
          <Th>Tool</Th>
          {!hideRun && <Th>Run</Th>}
          <Th className="w-40">Time</Th>
          <Th className="w-24 text-right">Latency</Th>
          <Th className="w-12 pl-0">
            <span className="sr-only">Details</span>
          </Th>
        </tr>
      </thead>
      <TBody>
        {entries.map((entry) => {
          const failed = entry.status === 'error';
          const open = openId === entry.id;
          const run = runById(entry.session_id);
          return (
            <Fragment key={entry.id}>
              <tr
                onClick={(event) => {
                  if ((event.target as HTMLElement).closest('a, button')) return;
                  setOpenId(open ? null : entry.id);
                }}
                className={cn('cursor-pointer transition-colors hover:bg-hover', open && 'bg-hover')}
              >
                <Td className="whitespace-nowrap text-muted">
                  {failed ? (
                    <Status tone="bad" label="Error" className="font-medium text-fg" />
                  ) : (
                    <StatusDot tone="ok" label="Success" />
                  )}
                </Td>
                <Td className="font-mono text-caption text-fg">{entry.tool_name}</Td>
                {!hideRun && (
                  <Td className="whitespace-nowrap text-muted">
                    {run ? (
                      <a href={runHref(run)} className="transition-colors hover:text-fg">
                        {run.name}
                      </a>
                    ) : entry.session_id ? (
                      <span title="This run was detached" className="font-mono text-caption">
                        {shortId(entry.session_id)}
                      </span>
                    ) : (
                      NONE
                    )}
                  </Td>
                )}
                <Td className="whitespace-nowrap font-mono text-caption text-muted">{formatTimestamp(entry.created_at)}</Td>
                <Td className="whitespace-nowrap text-right font-mono text-caption tabular-nums text-muted">
                  {formatLatency(entry.latency_ms)}
                </Td>
                <Td className="pl-0 text-right">
                  <button
                    type="button"
                    aria-expanded={open}
                    aria-label={open ? 'Hide details' : 'Show details'}
                    title={open ? 'Hide details' : 'Show details'}
                    onClick={() => setOpenId(open ? null : entry.id)}
                    className="inline-flex size-6 items-center justify-center rounded text-muted transition-colors hover:bg-raised hover:text-fg"
                  >
                    <ChevronRight
                      aria-hidden
                      strokeWidth={1.75}
                      className={cn('size-4 transition-transform', open && 'rotate-90')}
                    />
                  </button>
                </Td>
              </tr>
              {open && (
                <tr>
                  <td colSpan={columns} className="bg-raised p-0">
                    <CallDetail entry={entry} />
                  </td>
                </tr>
              )}
            </Fragment>
          );
        })}
      </TBody>
    </TableCard>
  );
}
