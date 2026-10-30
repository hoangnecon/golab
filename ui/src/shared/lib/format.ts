import type { Tier } from '@/api/types';

export const NONE = '—';

const TIER_LABEL: Record<Tier, string> = {
  free: 'Free',
  pro: 'Pro',
  pro_plus: 'Pro+',
};

export function formatTier(tier: Tier): string {
  return TIER_LABEL[tier] ?? tier;
}

const positive = (value: number | undefined): value is number =>
  value !== undefined && value > 0;

export function formatRam(freeGb: number | undefined): string {
  return positive(freeGb) ? `${freeGb.toFixed(1)} GB` : NONE;
}

const gb = (mb: number) => (mb / 1024).toFixed(1);

/** "free / total GB", or just "free GB" when the total is unknown. */
export function formatVram(freeMb: number | undefined, totalMb: number | undefined): string | null {
  if (!positive(freeMb)) return null;
  return positive(totalMb) ? `${gb(freeMb)} / ${gb(totalMb)} GB` : `${gb(freeMb)} GB`;
}

/** Fraction of VRAM in use (0..1), or null when free/total are unknown. */
export function vramUsedFraction(
  freeMb: number | undefined,
  totalMb: number | undefined,
): number | null {
  if (freeMb === undefined || !positive(totalMb)) return null;
  return Math.min(1, Math.max(0, (totalMb - freeMb) / totalMb));
}

export function shortId(id: string): string {
  return id.slice(0, 8);
}

export function formatLatency(ms: number): string {
  return `${ms} ms`;
}

/** "812 B", "4.2 KB", "1.3 MB" (binary units). */
export function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

const timeFormat = new Intl.DateTimeFormat('en', {
  hour: '2-digit',
  minute: '2-digit',
  second: '2-digit',
  hourCycle: 'h23',
});
const dateFormat = new Intl.DateTimeFormat('en', { month: 'short', day: 'numeric' });

/** Time of day for today, "date time" otherwise. */
export function formatTimestamp(iso: string): string {
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return NONE;
  const time = timeFormat.format(date);
  return date.toDateString() === new Date().toDateString()
    ? time
    : `${dateFormat.format(date)} ${time}`;
}

const relativeFormat = new Intl.RelativeTimeFormat('en', { numeric: 'auto', style: 'short' });

const UNITS: ReadonlyArray<[Intl.RelativeTimeFormatUnit, number]> = [
  ['day', 86_400],
  ['hour', 3_600],
  ['minute', 60],
];

/** "3 min. ago", "yesterday"; falls back to the clock time for under a minute. */
export function formatRelative(iso: string): string {
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return NONE;
  const seconds = Math.round((date.getTime() - Date.now()) / 1000);
  for (const [unit, size] of UNITS) {
    if (Math.abs(seconds) >= size) return relativeFormat.format(Math.round(seconds / size), unit);
  }
  return relativeFormat.format(0, 'second');
}
