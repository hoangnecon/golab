import {
  AppWindow,
  CircleCheck,
  Cpu,
  Disc,
  Lock,
  NotebookText,
  Play,
  Power,
  RotateCcw,
  Save,
  ShieldCheck,
  Square,
  Unlock,
  Unplug,
  User,
} from 'lucide-react';
import { useCallback, useEffect, useState } from 'react';
import { api } from '@/api/endpoints';
import type { AcceleratorOption, RunningCellInfo } from '@/api/types';
import { useWorkspace } from '@/data/useWorkspace';
import { CallsTable } from '@/domain/logs/CallsTable';
import { useRunActions } from '@/domain/runs/useRunActions';
import { Detail, Metrics, Panel, PanelActions, PanelBody, Section } from '@/shared/layout/Detail';
import { Page } from '@/shared/layout/Page';
import { NONE, formatRam, formatRelative, formatVram, shortId, vramUsedFraction } from '@/shared/lib/format';
import { navigate, paths } from '@/shared/lib/paths';
import { kernelLabel, kernelTone } from '@/shared/lib/status';
import { useNotice } from '@/shared/notice/NoticeProvider';
import { Button } from '@/shared/ui/Button';
import { Dialog, DialogActions } from '@/shared/ui/Dialog';
import { EmptyState } from '@/shared/ui/EmptyState';
import { Meter } from '@/shared/ui/Meter';
import { PropertyList } from '@/shared/ui/PropertyList';
import { Status } from '@/shared/ui/StatusDot';

interface RunPageProps {
  projectId: string;
  runId: string;
}

const LINK = 'truncate transition-colors hover:text-muted';

export function RunPage({ projectId, runId }: RunPageProps) {
  const { projectById, runById, accountById, callsOf } = useWorkspace();
  const { notify, notifyError } = useNotice();
  const run = runById(runId);
  const project = projectById(projectId);
  const { actions, dialog } = useRunActions(() => navigate(paths.project(projectId)));

  // Running Cells Live Stream State
  const [runningCells, setRunningCells] = useState<RunningCellInfo[]>([]);
  const [isLocked, setIsLocked] = useState(run?.is_locked ?? false);
  const [accelerators, setAccelerators] = useState<AcceleratorOption[]>([]);
  const [highRam, setHighRam] = useState(false);
  const [accModalOpen, setAccModalOpen] = useState(false);
  const [loadingAcc, setLoadingAcc] = useState(false);
  const [saving, setSaving] = useState(false);
  const [connecting, setConnecting] = useState(false);

  useEffect(() => {
    if (run) setIsLocked(run.is_locked ?? false);
  }, [run]);

  // Poll Running Cells every 1.5 seconds if session is active
  const pollRunning = useCallback(async () => {
    if (!run) return;
    try {
      const report = await api.sessions.runningCells(run.id);
      setRunningCells(report.running_cells ?? []);
    } catch (_) {}
  }, [run]);

  useEffect(() => {
    pollRunning();
    const timer = setInterval(pollRunning, 1500);
    return () => clearInterval(timer);
  }, [pollRunning]);

  // Load accelerators when modal opens
  const loadAccelerators = async () => {
    if (!run) return;
    setLoadingAcc(true);
    setAccModalOpen(true);
    try {
      const res = await api.sessions.accelerators(run.id);
      setAccelerators(res.accelerators ?? []);
      setHighRam(res.high_ram);
    } catch (err) {
      notifyError(err instanceof Error ? err.message : 'Could not fetch accelerators');
    } finally {
      setLoadingAcc(false);
    }
  };

  const selectAccelerator = async (val: string) => {
    if (!run) return;
    try {
      await api.sessions.setAccelerator(run.id, val, highRam);
      notify('Hardware accelerator updated. Reconnect runtime to apply.');
      setAccModalOpen(false);
    } catch (err) {
      notifyError(err instanceof Error ? err.message : 'Failed to set accelerator');
    }
  };

  const handleConnect = async () => {
    if (!run) return;
    setConnecting(true);
    try {
      await api.sessions.connect(run.id);
      notify('Connecting runtime container...');
    } catch (err) {
      notifyError(err instanceof Error ? err.message : 'Failed to connect');
    } finally {
      setConnecting(false);
    }
  };

  const handleDisconnect = async () => {
    if (!run) return;
    try {
      await api.sessions.disconnect(run.id);
      notify('Colab VM released and disconnected.');
    } catch (err) {
      notifyError(err instanceof Error ? err.message : 'Failed to disconnect');
    }
  };

  const handleSave = async () => {
    if (!run) return;
    setSaving(true);
    try {
      await api.sessions.save(run.id);
      notify('Notebook and outputs saved to Google Drive (Ctrl+S).');
    } catch (err) {
      notifyError(err instanceof Error ? err.message : 'Failed to save');
    } finally {
      setSaving(false);
    }
  };

  const handleToggleLock = async () => {
    if (!run) return;
    const next = !isLocked;
    setIsLocked(next);
    try {
      await api.sessions.lock(run.id, next);
      notify(next ? 'Session locked: Agent restricted to this notebook.' : 'Session lock removed.');
    } catch (err) {
      setIsLocked(!next);
      notifyError('Failed to toggle session lock');
    }
  };

  const handleRunAll = async () => {
    if (!run) return;
    try {
      await api.sessions.runall(run.id);
      notify('Run All triggered: executing all cells.');
    } catch (err) {
      notifyError(err instanceof Error ? err.message : 'Failed to run all cells');
    }
  };

  if (!run) {
    return (
      <Page
        trail={[
          { label: 'Projects', href: paths.projects },
          ...(project ? [{ label: project.name, href: paths.project(project.id) }] : []),
          { label: 'Run not found' },
        ]}
      >
        <EmptyState icon={NotebookText} message="This run does not exist or was detached." />
      </Page>
    );
  }

  const account = accountById(run.profile_id);
  const owner = project?.ownerProfileId ? accountById(project.ownerProfileId) : account;
  const calls = callsOf([run]);
  const vram = formatVram(run.gpu_memory_free_mb, run.gpu_memory_total_mb);
  const used = vramUsedFraction(run.gpu_memory_free_mb, run.gpu_memory_total_mb);
  const activeCell = runningCells.find((c) => c.is_running) ?? runningCells[0];

  return (
    <Page
      trail={[
        { label: 'Projects', href: paths.projects },
        { label: project?.name ?? 'Project', href: paths.project(projectId) },
        { label: run.name },
      ]}
      action={
        <div className="flex items-center gap-2">
          <Button
            variant={isLocked ? 'primary' : 'secondary'}
            size="sm"
            icon={isLocked ? Lock : Unlock}
            onClick={handleToggleLock}
          >
            {isLocked ? 'Locked to Agent' : 'Lock Session'}
          </Button>
          <Button
            variant="secondary"
            size="sm"
            icon={Save}
            disabled={saving}
            onClick={handleSave}
          >
            {saving ? 'Saving...' : 'Save (Ctrl+S)'}
          </Button>
        </div>
      }
    >
      <div className="flex flex-col gap-8">
        {/* Live Running Cell Stream Console Card */}
        {activeCell && (
          <div className="flex flex-col rounded-xl border border-line-strong bg-surface p-4 shadow-overlay">
            <div className="flex items-center justify-between pb-3 border-b border-line">
              <div className="flex items-center gap-2">
                <span className="relative flex size-3">
                  {activeCell.is_running ? (
                    <span className="relative inline-flex size-3 rounded-full bg-success"></span>
                  ) : (
                    <span className="inline-flex size-3 rounded-full bg-line-strong"></span>
                  )}
                </span>
                <span className="font-medium text-fg">
                  {activeCell.is_running ? 'Executing Cell in Progress' : 'Last Executed Cell Output'}
                </span>
                <span className="rounded bg-raised px-2 py-0.5 font-mono text-caption text-primary border border-line">
                  Cell #{activeCell.index} ({activeCell.cell_id})
                </span>
              </div>
              <div className="flex items-center gap-2 font-mono text-caption text-muted">
                <span>Elapsed: {activeCell.elapsed_seconds.toFixed(1)}s</span>
                <Button variant="ghost" size="sm" icon={Square} onClick={() => actions.interrupt(run)}>
                  Stop
                </Button>
              </div>
            </div>
            {activeCell.stdout ? (
              <div className="mt-3 max-h-56 overflow-y-auto rounded-lg bg-canvas p-3 font-mono text-caption text-fg border border-line leading-relaxed whitespace-pre-wrap">
                {activeCell.stdout}
              </div>
            ) : (
              <div className="mt-3 text-caption text-subtle italic">Waiting for stdout stream...</div>
            )}
          </div>
        )}

        <Metrics
          items={[
            { label: 'Kernel', value: <Status tone={kernelTone(run.kernel_state)} label={kernelLabel(run.kernel_state)} /> },
            {
              label: 'Accelerator',
              value: (
                <div className="flex items-center gap-2">
                  <span>{run.gpu_name ?? (run.kernel_state === 'connected' ? 'CPU' : NONE)}</span>
                  <button
                    type="button"
                    onClick={loadAccelerators}
                    className="text-caption text-primary hover:underline cursor-pointer"
                  >
                    Change
                  </button>
                </div>
              ),
            },
            {
              label: 'VRAM free',
              value: vram ?? NONE,
              footer: used !== null ? <Meter value={used} label="VRAM in use" /> : undefined,
            },
            { label: 'RAM free', value: formatRam(run.ram_free_gb) },
            { label: 'Drive', value: run.drive_mounted ? 'Mounted' : 'Not mounted' },
          ]}
        />
        <Detail
          main={
            <Section title="Tool calls" count={calls.length}>
              {calls.length === 0 ? (
                <EmptyState message="No tool calls recorded for this run." />
              ) : (
                <CallsTable entries={calls} hideRun />
              )}
            </Section>
          }
          rail={
            <>
              <Section title="Session Context">
                <Panel>
                  <PanelBody>
                    <PropertyList
                      items={[
                        {
                          term: 'Owner Profile',
                          value: owner ? (
                            <span className="flex items-center gap-1.5 font-medium text-fg">
                              <User aria-hidden className="size-3.5 text-primary" />
                              {owner.name}
                            </span>
                          ) : (
                            NONE
                          ),
                        },
                        {
                          term: 'Running Account',
                          value: account ? (
                            <a href={paths.account(account.id)} className={LINK}>
                              {account.name}
                            </a>
                          ) : (
                            NONE
                          ),
                        },
                        {
                          term: 'Access Level',
                          value: (
                            <span
                              className={`inline-flex items-center gap-1 rounded px-2 py-0.5 text-caption font-medium ${
                                run.permission === 'edit'
                                  ? 'bg-success/15 text-success'
                                  : run.permission === 'view'
                                  ? 'bg-warning/15 text-warning'
                                  : 'bg-danger/15 text-danger'
                              }`}
                            >
                              <ShieldCheck aria-hidden className="size-3" />
                              {run.permission === 'edit' ? 'Full Access' : run.permission === 'view' ? 'Read Only' : 'Denied'}
                            </span>
                          ),
                        },
                        {
                          term: 'Agent Lock',
                          value: isLocked ? (
                            <span className="text-primary font-medium flex items-center gap-1">
                              <Lock aria-hidden className="size-3" /> Exclusive Target
                            </span>
                          ) : (
                            'Unlocked'
                          ),
                        },
                        {
                          term: 'Notebook',
                          value: (
                            <a href={run.notebook_url} target="_blank" rel="noreferrer" className={`font-mono text-caption ${LINK}`}>
                              {run.notebook_url}
                            </a>
                          ),
                        },
                        { term: 'Attached', value: formatRelative(run.last_attached_at ?? run.created_at) },
                        { term: 'ID', value: <span className="font-mono text-caption">{shortId(run.id)}</span> },
                      ]}
                    />
                  </PanelBody>
                </Panel>
              </Section>
              <Section title="Runtime Operations">
                <Panel>
                  <PanelActions>
                    <Button variant="menu" icon={Power} disabled={connecting} onClick={handleConnect}>
                      {connecting ? 'Connecting...' : 'Connect Hosted Runtime'}
                    </Button>
                    <Button variant="menu" icon={Play} onClick={handleRunAll}>
                      Run All Cells
                    </Button>
                    <Button variant="menu" icon={Save} disabled={saving} onClick={handleSave}>
                      Save Notebook (Ctrl+S)
                    </Button>
                    <Button variant="menu" icon={AppWindow} onClick={() => actions.focus(run)}>
                      Focus browser tab
                    </Button>
                    <Button variant="menu" icon={CircleCheck} disabled={run.is_active} onClick={() => actions.activate(run)}>
                      {run.is_active ? 'Default target for agent' : 'Make default target'}
                    </Button>
                    <Button variant="menu" icon={Square} onClick={() => actions.interrupt(run)}>
                      Interrupt kernel
                    </Button>
                    <Button variant="menu" icon={RotateCcw} onClick={() => actions.restart(run)}>
                      Restart kernel
                    </Button>
                  </PanelActions>
                  <PanelActions separated>
                    <Button variant="menu" icon={Disc} onClick={handleDisconnect}>
                      Disconnect & Release VM
                    </Button>
                    <Button variant="menu" icon={Unplug} onClick={() => actions.detach(run)}>
                      Detach run
                    </Button>
                  </PanelActions>
                </Panel>
              </Section>
            </>
          }
        />
      </div>
      {dialog}

      {/* Hardware Accelerator Modal */}
      {accModalOpen && (
        <Dialog title="Change Hardware Accelerator" onClose={() => setAccModalOpen(false)}>
          {loadingAcc ? (
            <div className="p-6 text-center text-subtle">Inspecting account accelerators from Colab...</div>
          ) : (
            <div className="flex flex-col gap-4">
              <p className="text-caption text-muted">
                Select hardware accelerator based on your Google account quota (Free vs Pro/Pro+):
              </p>
              <div className="grid grid-cols-1 gap-2">
                {accelerators.map((acc) => (
                  <button
                    key={acc.value}
                    type="button"
                    disabled={acc.disabled}
                    onClick={() => selectAccelerator(acc.value)}
                    className={`flex items-center justify-between rounded-lg border p-3 text-left transition-all ${
                      acc.disabled
                        ? 'opacity-40 cursor-not-allowed border-line bg-canvas'
                        : acc.checked
                        ? 'border-primary bg-primary/10 text-primary'
                        : 'border-line bg-surface hover:border-line-strong'
                    }`}
                  >
                    <div className="flex items-center gap-2">
                      <Cpu className="size-4" />
                      <span className="font-medium">{acc.name}</span>
                    </div>
                    {acc.disabled ? (
                      <span className="text-caption text-subtle">Requires Upgrade</span>
                    ) : acc.checked ? (
                      <span className="text-caption text-primary font-medium">Active</span>
                    ) : (
                      <span className="text-caption text-muted">Select</span>
                    )}
                  </button>
                ))}
              </div>
              <DialogActions>
                <Button variant="ghost" onClick={() => setAccModalOpen(false)}>
                  Close
                </Button>
              </DialogActions>
            </div>
          )}
        </Dialog>
      )}
    </Page>
  );
}
