import type { ActivityLog, Profile, Session } from '@/api/types';
import { useControlPlane } from './ControlPlaneProvider';
import { useProjects } from './ProjectsProvider';
import { projectIdFor } from './projects';

export interface Project {
  id: string;
  name: string;
  notebookUrl: string;
  ownerProfileId?: string;
  /** Sessions opened on this notebook, newest attach first (backend order). */
  runs: Session[];
  saved: boolean;
}

export interface Workspace {
  projects: Project[];
  projectById: (id: string | undefined) => Project | undefined;
  projectOfRun: (run: Session) => Project | undefined;
  accountById: (id: string | undefined) => Profile | undefined;
  runById: (id: string | undefined) => Session | undefined;
  /** Calls whose session is one of `runs`. Orphaned calls (session detached) cannot be attributed. */
  callsOf: (runs: Session[]) => ActivityLog[];
}

/** Read-only derived view; all mutations stay in ControlPlaneProvider / ProjectsProvider. */
export function useWorkspace(): Workspace {
  const { sessions, profiles, activity } = useControlPlane();
  const { saved } = useProjects();

  const byId = new Map<string, Project>();
  for (const project of saved) {
    byId.set(project.id, {
      id: project.id,
      name: project.name,
      notebookUrl: project.notebook_url,
      ownerProfileId: project.owner_profile_id,
      runs: [],
      saved: true,
    });
  }
  for (const session of sessions) {
    const id = projectIdFor(session.notebook_url);
    const existing = byId.get(id);
    if (existing) existing.runs.push(session);
    else {
      byId.set(id, {
        id,
        name: session.name,
        notebookUrl: session.notebook_url,
        ownerProfileId: session.owner_profile_id || session.profile_id,
        runs: [session],
        saved: false,
      });
    }
  }
  const projects = [...byId.values()];

  return {
    projects,
    projectById: (id) => (id ? byId.get(id) : undefined),
    projectOfRun: (run) => byId.get(projectIdFor(run.notebook_url)),
    accountById: (id) => profiles.find((profile) => profile.id === id),
    runById: (id) => sessions.find((session) => session.id === id),
    callsOf: (runs) => {
      const ids = new Set(runs.map((run) => run.id));
      return activity.filter((entry) => entry.session_id !== undefined && ids.has(entry.session_id));
    },
  };
}
