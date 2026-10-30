import type { Tone } from '@/shared/ui/StatusDot';

/**
 * Colab kernel states are not specified by the backend API (`kernel_state` is a
 * free string), so this maps the common Jupyter names and treats anything else
 * as neutral rather than guessing.
 */
export function kernelTone(state: string | undefined): Tone {
  switch (state?.toLowerCase()) {
    case 'idle':
      return 'ok';
    case 'busy':
    case 'starting':
    case 'restarting':
    case 'connecting':
      return 'warn';
    case 'dead':
    case 'disconnected':
    case 'error':
      return 'bad';
    default:
      return 'idle';
  }
}

export function kernelLabel(state: string | undefined): string {
  if (!state) return 'Unknown';
  return state.charAt(0).toUpperCase() + state.slice(1);
}
