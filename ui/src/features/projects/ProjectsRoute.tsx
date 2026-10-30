import type { RouteProps } from '@/shared/lib/location';
import { ProjectPage } from './ProjectPage';
import { ProjectsIndex } from './ProjectsIndex';
import { RunPage } from './RunPage';

/** `#/projects` · `#/projects/<id>` · `#/projects/<id>/runs/<run>` */
export function ProjectsRoute({ segments }: RouteProps) {
  const [projectId, second, runId] = segments;
  if (!projectId) return <ProjectsIndex />;
  if (second === 'runs' && runId) return <RunPage projectId={projectId} runId={runId} />;
  return <ProjectPage projectId={projectId} />;
}
