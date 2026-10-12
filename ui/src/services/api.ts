import { Profile, Session, ActivityLog } from '../types';

export class ApiError extends Error {
  constructor(public status: number, message: string) {
    super(message);
    this.name = 'ApiError';
  }
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(path, {
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
    ...options,
  });

  if (!res.ok) {
    let errorMessage = `HTTP ${res.status}`;
    try {
      const data = await res.json();
      if (data.message) errorMessage = data.message;
      else if (data.error) errorMessage = data.error;
    } catch {
      // ignore parse error
    }
    throw new ApiError(res.status, errorMessage);
  }

  return res.json();
}

export const api = {
  // Profiles
  async getProfiles(): Promise<Profile[]> {
    const res = await request<{ profiles: Profile[] }>('/api/profiles');
    return res.profiles || [];
  },
  async createProfile(data: { name: string; account_email?: string; tier: string }): Promise<Profile> {
    return request<Profile>('/api/profiles', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },
  async launchProfile(id: string): Promise<void> {
    await request(`/api/profiles/${id}/launch`, { method: 'POST' });
  },
  async stopProfile(id: string): Promise<void> {
    await request(`/api/profiles/${id}/stop`, { method: 'POST' });
  },
  async deleteProfile(id: string): Promise<void> {
    await request(`/api/profiles/${id}`, { method: 'DELETE' });
  },

  // Sessions
  async getSessions(): Promise<Session[]> {
    const res = await request<{ sessions: Session[] }>('/api/sessions');
    return res.sessions || [];
  },
  async createSession(data: { profile_id: string; notebook_url: string; name?: string }): Promise<Session> {
    return request<Session>('/api/sessions', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },
  async focusSession(id: string): Promise<void> {
    await request(`/api/sessions/${id}/focus`, { method: 'POST' });
  },
  async interruptSession(id: string): Promise<void> {
    await request(`/api/sessions/${id}/interrupt`, { method: 'POST' });
  },
  async restartSession(id: string): Promise<void> {
    await request(`/api/sessions/${id}/restart`, { method: 'POST' });
  },
  async deleteSession(id: string): Promise<void> {
    await request(`/api/sessions/${id}`, { method: 'DELETE' });
  },

  // Activity
  async getActivity(): Promise<ActivityLog[]> {
    const res = await request<{ activity: ActivityLog[] }>('/api/activity');
    return res.activity || [];
  },
};
