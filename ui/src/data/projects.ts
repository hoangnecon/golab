/**
 * The backend has no "project" concept: a project is a notebook URL, and a run
 * is a session opened on it. Projects are therefore UI-side: saved ones live in
 * localStorage, and any session whose notebook has no saved project appears as
 * an implicit one so nothing is ever hidden.
 */
export interface StoredProject {
  id: string;
  name: string;
  notebookUrl: string;
  createdAt: string;
}

const STORAGE_KEY = 'golab.projects';

/** Origin + path only: fragments (#scrollTo=…) and query strings do not identify a notebook. */
export function normalizeNotebookUrl(raw: string): string | null {
  try {
    const url = new URL(raw.trim());
    if (url.protocol !== 'https:' && url.protocol !== 'http:') return null;
    return `${url.origin}${url.pathname.replace(/\/+$/, '')}`;
  } catch {
    return null;
  }
}

/** FNV-1a 32-bit: a stable, URL-safe id derived from the notebook, so sessions and projects join without a table. */
export function projectIdFor(notebookUrl: string): string {
  const normalized = normalizeNotebookUrl(notebookUrl) ?? notebookUrl;
  let hash = 0x811c9dc5;
  for (let index = 0; index < normalized.length; index += 1) {
    hash ^= normalized.charCodeAt(index);
    hash = Math.imul(hash, 0x01000193);
  }
  return (hash >>> 0).toString(36);
}

function isStoredProject(value: unknown): value is StoredProject {
  if (typeof value !== 'object' || value === null) return false;
  const record = value as Record<string, unknown>;
  return (
    typeof record.id === 'string' &&
    typeof record.name === 'string' &&
    typeof record.notebookUrl === 'string' &&
    typeof record.createdAt === 'string'
  );
}

export function loadProjects(): StoredProject[] {
  try {
    const parsed: unknown = JSON.parse(window.localStorage.getItem(STORAGE_KEY) ?? '[]');
    return Array.isArray(parsed) ? parsed.filter(isStoredProject) : [];
  } catch {
    return [];
  }
}

export function saveProjects(projects: StoredProject[]): void {
  window.localStorage.setItem(STORAGE_KEY, JSON.stringify(projects));
}
