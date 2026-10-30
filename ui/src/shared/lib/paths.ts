const enc = encodeURIComponent;

/** Single place that knows URL shapes. Links and navigation must use these builders. */
export const paths = {
  projects: '#/projects',
  project: (id: string) => `#/projects/${enc(id)}`,
  run: (projectId: string, runId: string) => `#/projects/${enc(projectId)}/runs/${enc(runId)}`,
  accounts: '#/accounts',
  account: (id: string) => `#/accounts/${enc(id)}`,
  logs: (query: Record<string, string> = {}) => {
    const qs = new URLSearchParams(query).toString();
    return qs ? `#/logs?${qs}` : '#/logs';
  },
} as const;

export function navigate(hash: string) {
  window.location.hash = hash;
}
