import { useControlPlane } from '@/data/ControlPlaneProvider';
import type { Session } from '@/api/types';
import { useConfirm } from '@/shared/ui/ConfirmDialog';

/**
 * Every operation on a run (a session), with the confirmations the contract
 * requires. Render `dialog` once in the page that calls this.
 */
export function useRunActions(onDetached?: () => void) {
  const { focusSession, activateSession, interruptSession, restartSession, detachSession } =
    useControlPlane();
  const { confirm, dialog } = useConfirm();

  const actions = {
    focus: (run: Session) => void focusSession(run.id),
    activate: (run: Session) => void activateSession(run.id),
    interrupt: (run: Session) => void interruptSession(run.id),
    restart: (run: Session) =>
      confirm({
        title: 'Restart kernel',
        message: `Variables and running state in “${run.name}” will be lost.`,
        confirmLabel: 'Restart',
        onConfirm: async () => {
          await restartSession(run.id);
        },
      }),
    detach: (run: Session) =>
      confirm({
        title: 'Detach run',
        message: `The browser tab for “${run.name}” will be closed.`,
        confirmLabel: 'Detach',
        onConfirm: async () => {
          if (await detachSession(run.id)) onDetached?.();
        },
      }),
  };

  return { actions, dialog };
}

export type RunActions = ReturnType<typeof useRunActions>['actions'];
