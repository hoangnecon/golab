import { Copy } from 'lucide-react';
import type { ActivityLog, Payload } from '@/api/types';
import { useActivityDetail } from '@/data/useActivityDetail';
import { formatBytes } from '@/shared/lib/format';
import { useNotice } from '@/shared/notice/NoticeProvider';
import { CodeBlock } from '@/shared/ui/CodeBlock';
import { FormError } from '@/shared/ui/Field';
import { IconButton } from '@/shared/ui/IconButton';

/** Truncated bodies are a prefix of the document, not JSON, so they show as text. */
function payloadText(payload: Payload): string {
  if (payload.truncated && typeof payload.body === 'string') return payload.body;
  return JSON.stringify(payload.body, null, 2) ?? '';
}

interface PayloadPanelProps {
  title: 'Request' | 'Response';
  payload: Payload | null;
}

function PayloadPanel({ title, payload }: PayloadPanelProps) {
  const { notify, notifyError } = useNotice();
  const text = payload ? payloadText(payload) : '';

  const copy = () => {
    navigator.clipboard.writeText(text).then(
      () => notify(`${title} copied`),
      () => notifyError('Clipboard is not available'),
    );
  };

  return (
    <section aria-label={title} className="min-w-0 overflow-hidden rounded-lg border border-line bg-surface">
      <header className="flex h-10 items-center gap-2 border-b border-line pl-4 pr-2">
        <h3 className="text-caption font-medium text-fg">{title}</h3>
        {payload && (
          <span className="text-caption text-subtle">
            {payload.truncated
              ? `First ${formatBytes(new TextEncoder().encode(text).length)} of ${formatBytes(payload.bytes)}`
              : formatBytes(payload.bytes)}
          </span>
        )}
        {payload && (
          <IconButton label={`Copy ${title.toLowerCase()}`} size="sm" onClick={copy} className="ml-auto">
            <Copy aria-hidden className="size-4" strokeWidth={1.75} />
          </IconButton>
        )}
      </header>
      {payload ? (
        <CodeBlock code={text} plain={payload.truncated} label={`${title} body`} />
      ) : (
        <p className="p-4 text-caption text-subtle">Not recorded</p>
      )}
    </section>
  );
}

/**
 * Expanded view of one tool call: error first, then request and response side
 * by side. Rendered inside a table cell: `w-0 min-w-full` stops long payload
 * lines from widening the table, so they wrap inside the panels instead.
 */
export function CallDetail({ entry }: { entry: ActivityLog }) {
  const state = useActivityDetail(entry.id);

  return (
    <div className="flex w-0 min-w-full flex-col gap-4 p-4">
      {entry.status === 'error' && (
        <div className="flex flex-col gap-1">
          <span className="text-caption font-medium text-fg">Error</span>
          <p className="break-words font-mono text-caption text-fg">{entry.error_message || 'No error message'}</p>
        </div>
      )}
      {state.status === 'loading' && <p className="text-caption text-subtle">Loading payload</p>}
      {state.status === 'error' && <FormError>{state.message}</FormError>}
      {state.status === 'ready' &&
        (state.detail.has_payload ? (
          <div className="grid gap-4 lg:grid-cols-2">
            <PayloadPanel title="Request" payload={state.detail.request} />
            <PayloadPanel title="Response" payload={state.detail.response} />
          </div>
        ) : (
          <p className="text-caption text-subtle">Payloads were not recorded for this call.</p>
        ))}
    </div>
  );
}
