import { Play, Square, Trash2, UserRound } from 'lucide-react';
import { useControlPlane } from '@/data/ControlPlaneProvider';
import { useWorkspace } from '@/data/useWorkspace';
import { CallsTable } from '@/domain/logs/CallsTable';
import { RunsTable } from '@/domain/runs/RunsTable';
import { Detail, Metrics, Panel, PanelActions, PanelBody, Section } from '@/shared/layout/Detail';
import { Page } from '@/shared/layout/Page';
import { NONE, formatTier, formatTimestamp } from '@/shared/lib/format';
import { navigate, paths } from '@/shared/lib/paths';
import { Button } from '@/shared/ui/Button';
import { useConfirm } from '@/shared/ui/ConfirmDialog';
import { EmptyState } from '@/shared/ui/EmptyState';
import { PropertyList } from '@/shared/ui/PropertyList';
import { Status } from '@/shared/ui/StatusDot';

/** The daemon seeds this account at startup (cmd/server/main.go); the UI does not offer to delete it. */
const PROTECTED_ACCOUNT = 'default';

export function AccountPage({ accountId }: { accountId: string }) {
  const { accountById, callsOf } = useWorkspace();
  const { sessions, launchProfile, stopProfile, deleteProfile } = useControlPlane();
  const { confirm, dialog: confirmDialog } = useConfirm();
  const account = accountById(accountId);

  if (!account) {
    return (
      <Page trail={[{ label: 'Accounts', href: paths.accounts }, { label: 'Not found' }]}>
        <EmptyState icon={UserRound} message="This account does not exist." />
      </Page>
    );
  }

  const runs = sessions.filter((run) => run.profile_id === account.id);
  const calls = callsOf(runs);

  function remove() {
    if (!account) return;
    confirm({
      title: 'Delete account',
      message: `“${account.name}” and its Chrome profile data will be removed. Its ${runs.length} run(s) are detached.`,
      confirmLabel: 'Delete',
      onConfirm: async () => {
        if (await deleteProfile(account.id)) navigate(paths.accounts);
      },
    });
  }

  return (
    <Page trail={[{ label: 'Accounts', href: paths.accounts }, { label: account.name }]}>
      <div className="flex flex-col gap-8">
        <Metrics
          items={[
            {
              label: 'Browser',
              value: <Status tone={account.is_running ? 'ok' : 'idle'} label={account.is_running ? 'Running' : 'Stopped'} />,
            },
            { label: 'Plan', value: formatTier(account.tier) },
            { label: 'Runs', value: runs.length },
            { label: 'Tool calls', value: calls.length },
          ]}
        />
        <Detail
          main={
            <>
              <Section title="Runs" count={runs.length}>
                {runs.length === 0 ? (
                  <EmptyState message="This account has no runs. Start one from a project." />
                ) : (
                  <RunsTable runs={runs} showProject hideAccount />
                )}
              </Section>
              <Section title="Tool calls" count={calls.length}>
                {calls.length === 0 ? (
                  <EmptyState message="No tool calls recorded for this account's current runs." />
                ) : (
                  <CallsTable entries={calls} />
                )}
              </Section>
            </>
          }
          rail={
            <>
              <Section title="Details">
                <Panel>
                  <PanelBody>
                    <PropertyList
                      items={[
                        { term: 'Google account', value: account.account_email || NONE },
                        { term: 'Debug port', value: account.cdp_port || NONE },
                        {
                          term: 'Profile directory',
                          value: <span className="truncate font-mono text-caption">{account.user_data_dir}</span>,
                        },
                        { term: 'Created', value: formatTimestamp(account.created_at) },
                      ]}
                    />
                  </PanelBody>
                </Panel>
              </Section>
              <Section title="Controls">
                <Panel>
                  <PanelActions>
                    {account.is_running ? (
                      <Button variant="menu" icon={Square} onClick={() => void stopProfile(account.id)}>
                        Stop browser
                      </Button>
                    ) : (
                      <Button variant="menu" icon={Play} onClick={() => void launchProfile(account.id)}>
                        Launch browser
                      </Button>
                    )}
                  </PanelActions>
                  <PanelActions separated>
                    <Button
                      variant="menu"
                      icon={Trash2}
                      disabled={account.id === PROTECTED_ACCOUNT}
                      title={account.id === PROTECTED_ACCOUNT ? 'The default account cannot be deleted' : undefined}
                      onClick={remove}
                    >
                      Delete account
                    </Button>
                  </PanelActions>
                </Panel>
              </Section>
            </>
          }
        />
      </div>
      {confirmDialog}
    </Page>
  );
}
