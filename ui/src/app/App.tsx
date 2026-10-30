import { WifiOff } from 'lucide-react';
import { useEffect } from 'react';
import { ControlPlaneProvider, useControlPlane } from '@/data/ControlPlaneProvider';
import { ProjectsProvider } from '@/data/ProjectsProvider';
import { ProjectsNav } from '@/features/projects/ProjectsNav';
import { Shell } from '@/shared/layout/Shell';
import { ShellProvider } from '@/shared/layout/ShellProvider';
import { NoticeProvider } from '@/shared/notice/NoticeProvider';
import { NAV_ITEMS, useRoute } from './routes';

function Routed() {
  const { route, segments } = useRoute();
  const { online } = useControlPlane();
  const Current = route.component;

  useEffect(() => {
    document.title = `${route.label} · GoLab`;
  }, [route.label]);

  return (
    <Shell items={NAV_ITEMS} activeId={route.id} tree={<ProjectsNav />}>
      {!online && (
        <p role="alert" className="flex h-9 shrink-0 items-center gap-2 border-b border-line bg-raised pl-14 pr-4 text-caption font-medium text-fg md:pl-6">
          <WifiOff aria-hidden className="size-4" strokeWidth={1.75} />
          Connection to the GoLab daemon lost. Reconnecting.
        </p>
      )}
      <Current segments={segments} />
    </Shell>
  );
}

export default function App() {
  return (
    <NoticeProvider>
      <ControlPlaneProvider>
        <ProjectsProvider>
          <ShellProvider>
            <Routed />
          </ShellProvider>
        </ProjectsProvider>
      </ControlPlaneProvider>
    </NoticeProvider>
  );
}
