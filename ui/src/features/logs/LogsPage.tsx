import { Activity, Pause, Play } from 'lucide-react';
import { useState } from 'react';
import type { ActivityLog } from '@/api/types';
import { useControlPlane } from '@/data/ControlPlaneProvider';
import { CallsTable } from '@/domain/logs/CallsTable';
import { Page } from '@/shared/layout/Page';
import { replaceParams, useLocation } from '@/shared/lib/location';
import { Button } from '@/shared/ui/Button';
import { EmptyState } from '@/shared/ui/EmptyState';
import { Checkbox, SearchInput, Select } from '@/shared/ui/Field';

/** Tool-call history across every run. Filters live in the URL so a view can be shared and survives reload. */
export function LogsPage() {
  const { activity, sessions, profiles } = useControlPlane();
  const { params } = useLocation();
  const [frozen, setFrozen] = useState<ActivityLog[] | null>(null);

  const q = params.get('q') ?? '';
  const tool = params.get('tool') ?? '';
  const run = params.get('run') ?? '';
  const account = params.get('account') ?? '';
  const errorsOnly = params.get('errors') === '1';

  const source = frozen ?? activity;
  const tools = [...new Set(source.map((entry) => entry.tool_name))].sort();
  const accountOf = new Map(sessions.map((session) => [session.id, session.profile_id]));
  const needle = q.trim().toLowerCase();

  const visible = source.filter(
    (entry) =>
      (!tool || entry.tool_name === tool) &&
      (!run || entry.session_id === run) &&
      (!account || (entry.session_id !== undefined && accountOf.get(entry.session_id) === account)) &&
      (!errorsOnly || entry.status === 'error') &&
      (!needle ||
        entry.tool_name.toLowerCase().includes(needle) ||
        (entry.error_message ?? '').toLowerCase().includes(needle)),
  );

  return (
    <Page
      trail={[{ label: 'Logs' }]}
      meta={source.length > 0 ? `${visible.length} / ${source.length}` : undefined}
      action={
        <Button
          size="sm"
          icon={frozen ? Play : Pause}
          onClick={() => setFrozen(frozen ? null : activity)}
        >
          {frozen ? 'Resume' : 'Pause'}
        </Button>
      }
    >
      <div className="flex flex-col gap-4">
        <div className="flex flex-wrap items-end gap-3">
          <SearchInput
            label="Search tool or error"
            placeholder="Search tool or error"
            value={q}
            onChange={(event) => replaceParams({ q: event.target.value })}
            className="w-64"
          />
          <Select label="Tool" hideLabel value={tool} onChange={(event) => replaceParams({ tool: event.target.value })} className="w-48">
            <option value="">All tools</option>
            {tools.map((name) => (
              <option key={name} value={name}>
                {name}
              </option>
            ))}
          </Select>
          <Select label="Run" hideLabel value={run} onChange={(event) => replaceParams({ run: event.target.value })} className="w-48">
            <option value="">All runs</option>
            {sessions.map((session) => (
              <option key={session.id} value={session.id}>
                {session.name}
              </option>
            ))}
          </Select>
          <Select label="Account" hideLabel value={account} onChange={(event) => replaceParams({ account: event.target.value })} className="w-48">
            <option value="">All accounts</option>
            {profiles.map((profile) => (
              <option key={profile.id} value={profile.id}>
                {profile.name}
              </option>
            ))}
          </Select>
          <Checkbox
            label="Errors only"
            checked={errorsOnly}
            onChange={(event) => replaceParams({ errors: event.target.checked ? '1' : '' })}
          />
        </div>
        {visible.length === 0 ? (
          <EmptyState icon={Activity} message={source.length === 0 ? 'No tool calls yet.' : 'No calls match these filters.'} />
        ) : (
          <CallsTable entries={visible} />
        )}
      </div>
    </Page>
  );
}
