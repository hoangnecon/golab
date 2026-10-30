import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react';
import type { ReactNode } from 'react';
import { api } from '@/api/endpoints';
import type { Project } from '@/api/types';
import { normalizeNotebookUrl } from './projects';

interface ProjectsApi {
  saved: Project[];
  createProject: (name: string, notebookUrl: string, ownerProfileId: string) => Promise<Project>;
  removeProject: (id: string) => Promise<void>;
  refreshProjects: () => Promise<void>;
}

const ProjectsContext = createContext<ProjectsApi | null>(null);

export function ProjectsProvider({ children }: { children: ReactNode }) {
  const [saved, setSaved] = useState<Project[]>([]);

  const refreshProjects = useCallback(async () => {
    try {
      const list = await api.projects.list();
      setSaved(list);
    } catch (_) {}
  }, []);

  useEffect(() => {
    refreshProjects();
  }, [refreshProjects]);

  const createProject = useCallback(
    async (name: string, notebookUrl: string, ownerProfileId: string) => {
      const normalized = normalizeNotebookUrl(notebookUrl);
      if (!normalized) throw new Error('Enter a valid notebook URL');
      if (saved.some((p) => p.notebook_url === notebookUrl.trim())) {
        throw new Error('A project with this notebook already exists');
      }
      const created = await api.projects.create({
        name: name.trim(),
        notebook_url: notebookUrl.trim(),
        owner_profile_id: ownerProfileId,
      });
      await refreshProjects();
      return created;
    },
    [saved, refreshProjects],
  );

  const removeProject = useCallback(
    async (id: string) => {
      await api.projects.remove(id);
      await refreshProjects();
    },
    [refreshProjects],
  );

  const value = useMemo(
    () => ({ saved, createProject, removeProject, refreshProjects }),
    [saved, createProject, removeProject, refreshProjects],
  );

  return <ProjectsContext.Provider value={value}>{children}</ProjectsContext.Provider>;
}

export function useProjects(): ProjectsApi {
  const context = useContext(ProjectsContext);
  if (!context) throw new Error('useProjects must be used inside ProjectsProvider');
  return context;
}
