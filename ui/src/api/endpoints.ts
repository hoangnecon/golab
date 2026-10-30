import { request } from './http';
import type {
  ActivityDetail,
  ActivityLog,
  NewProfile,
  NewProject,
  NewSession,
  Profile,
  Project,
  RunningCellsReport,
  RuntimeConfigReport,
  Session,
} from './types';

const post = (path: string) => request<unknown>(path, { method: 'POST' });

export const api = {
  projects: {
    async list(): Promise<Project[]> {
      const body = await request<{ projects: Project[] | null }>('/api/projects');
      return body.projects ?? [];
    },
    create: (input: NewProject) =>
      request<Project>('/api/projects', { method: 'POST', body: JSON.stringify(input) }),
    get: (id: string) => request<Project>(`/api/projects/${id}`),
    update: (id: string, input: Partial<NewProject>) =>
      request<Project>(`/api/projects/${id}`, { method: 'PUT', body: JSON.stringify(input) }),
    remove: (id: string) => request<unknown>(`/api/projects/${id}`, { method: 'DELETE' }),
    launch: (id: string, profileId?: string) =>
      request<Session>(`/api/projects/${id}/launch`, {
        method: 'POST',
        body: JSON.stringify({ profile_id: profileId }),
      }),
  },
  profiles: {
    async list(): Promise<Profile[]> {
      const body = await request<{ profiles: Profile[] | null }>('/api/profiles');
      return body.profiles ?? [];
    },
    create: (input: NewProfile) =>
      request<Profile>('/api/profiles', { method: 'POST', body: JSON.stringify(input) }),
    launch: (id: string) => post(`/api/profiles/${id}/launch`),
    stop: (id: string) => post(`/api/profiles/${id}/stop`),
    remove: (id: string) => request<unknown>(`/api/profiles/${id}`, { method: 'DELETE' }),
  },
  sessions: {
    async list(): Promise<Session[]> {
      const body = await request<{ sessions: Session[] | null }>('/api/sessions');
      return body.sessions ?? [];
    },
    create: (input: NewSession) =>
      request<Session>('/api/sessions', { method: 'POST', body: JSON.stringify(input) }),
    focus: (id: string) => post(`/api/sessions/${id}/focus`),
    activate: (id: string) => post(`/api/sessions/${id}/activate`),
    interrupt: (id: string) => post(`/api/sessions/${id}/interrupt`),
    restart: (id: string) => post(`/api/sessions/${id}/restart`),
    remove: (id: string) => request<unknown>(`/api/sessions/${id}`, { method: 'DELETE' }),
    lock: (id: string, locked: boolean) =>
      request<{ is_locked: boolean }>(`/api/sessions/${id}/lock`, {
        method: 'POST',
        body: JSON.stringify({ locked }),
      }),
    connect: (id: string) => post(`/api/sessions/${id}/connect`),
    disconnect: (id: string) => post(`/api/sessions/${id}/disconnect`),
    save: (id: string) => post(`/api/sessions/${id}/save`),
    runall: (id: string) => post(`/api/sessions/${id}/runall`),
    runningCells: (id: string) => request<RunningCellsReport>(`/api/sessions/${id}/running-cells`),
    accelerators: (id: string) => request<RuntimeConfigReport>(`/api/sessions/${id}/accelerators`),
    setAccelerator: (id: string, accelerator: string, high_ram: boolean) =>
      request<unknown>(`/api/sessions/${id}/accelerator`, {
        method: 'POST',
        body: JSON.stringify({ accelerator, high_ram }),
      }),
  },
  activity: {
    async list(): Promise<ActivityLog[]> {
      const body = await request<{ activity: ActivityLog[] | null }>('/api/activity');
      return body.activity ?? [];
    },
    get: (id: number) => request<ActivityDetail>(`/api/activity/${id}`),
  },
} as const;
