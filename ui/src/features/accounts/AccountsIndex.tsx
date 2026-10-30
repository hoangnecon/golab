import { Play, Plus, Square, UserRound } from 'lucide-react';
import { useState } from 'react';
import { useControlPlane } from '@/data/ControlPlaneProvider';
import { Page } from '@/shared/layout/Page';
import { NONE, formatRelative, formatTier } from '@/shared/lib/format';
import { paths } from '@/shared/lib/paths';
import { Avatar } from '@/shared/ui/Avatar';
import { Button } from '@/shared/ui/Button';
import { EmptyState } from '@/shared/ui/EmptyState';
import { IconButton } from '@/shared/ui/IconButton';
import { Status } from '@/shared/ui/StatusDot';
import { TBody, TableCard, Td, Th } from '@/shared/ui/Table';
import { CreateAccountDialog } from './CreateAccountDialog';

export function AccountsIndex() {
  const { profiles, sessions, launchProfile, stopProfile } = useControlPlane();
  const [creating, setCreating] = useState(false);

  const create = (
    <Button variant="primary" size="sm" icon={Plus} onClick={() => setCreating(true)}>
      New account
    </Button>
  );

  return (
    <Page
      trail={[{ label: 'Accounts' }]}
      meta={profiles.length > 0 ? String(profiles.length) : undefined}
      action={profiles.length > 0 ? create : undefined}
    >
      {profiles.length === 0 ? (
        <EmptyState
          icon={UserRound}
          message="An account is a Chrome profile signed in to Google. Runs use it to open notebooks."
          action={create}
        />
      ) : (
        <TableCard>
          <thead>
            <tr>
              <Th>Account</Th>
              <Th className="w-32">Plan</Th>
              <Th className="w-32">Browser</Th>
              <Th className="w-24 text-right">Runs</Th>
              <Th className="w-36 text-right">Created</Th>
              <Th className="w-14 pl-0">
                <span className="sr-only">Browser control</span>
              </Th>
            </tr>
          </thead>
          <TBody>
            {profiles.map((account) => (
              <tr key={account.id} className="transition-colors hover:bg-hover">
                <Td className="py-2">
                  <a href={paths.account(account.id)} className="flex min-w-0 items-center gap-3">
                    <Avatar name={account.name} />
                    <span className="flex min-w-0 flex-col">
                      <span className="truncate font-medium text-fg hover:underline">{account.name}</span>
                      <span className="truncate text-caption text-subtle">{account.account_email || NONE}</span>
                    </span>
                  </a>
                </Td>
                <Td className="text-muted">{formatTier(account.tier)}</Td>
                <Td className="whitespace-nowrap text-muted">
                  <Status tone={account.is_running ? 'ok' : 'idle'} label={account.is_running ? 'Running' : 'Stopped'} />
                </Td>
                <Td className="text-right tabular-nums text-muted">
                  {sessions.filter((run) => run.profile_id === account.id).length}
                </Td>
                <Td className="whitespace-nowrap text-right text-muted">{formatRelative(account.created_at)}</Td>
                <Td className="pl-0 text-right">
                  {account.is_running ? (
                    <IconButton label={`Stop ${account.name} browser`} onClick={() => void stopProfile(account.id)}>
                      <Square aria-hidden className="size-4" strokeWidth={1.75} />
                    </IconButton>
                  ) : (
                    <IconButton label={`Launch ${account.name} browser`} onClick={() => void launchProfile(account.id)}>
                      <Play aria-hidden className="size-4" strokeWidth={1.75} />
                    </IconButton>
                  )}
                </Td>
              </tr>
            ))}
          </TBody>
        </TableCard>
      )}
      {creating && <CreateAccountDialog onClose={() => setCreating(false)} />}
    </Page>
  );
}
