import { Activity, UserRound } from 'lucide-react';
import type { ComponentType } from 'react';
import { AccountsRoute } from '@/features/accounts/AccountsRoute';
import { LogsPage } from '@/features/logs/LogsPage';
import { ProjectsRoute } from '@/features/projects/ProjectsRoute';
import type { NavItem } from '@/shared/layout/Sidebar';
import { useLocation } from '@/shared/lib/location';
import type { RouteProps } from '@/shared/lib/location';

interface Route {
  id: 'projects' | 'accounts' | 'logs';
  label: string;
  component: ComponentType<RouteProps>;
}

/** Top-level sections. Projects is the default and is rendered by the sidebar tree, not a pinned link. */
export const ROUTES: readonly Route[] = [
  { id: 'projects', label: 'Projects', component: ProjectsRoute },
  { id: 'accounts', label: 'Accounts', component: AccountsRoute },
  { id: 'logs', label: 'Logs', component: LogsPage },
];

const PINNED: Record<string, NavItem['icon']> = { accounts: UserRound, logs: Activity };

/** Pinned sidebar destinations under the project tree. */
export const NAV_ITEMS: readonly NavItem[] = ROUTES.filter((route) => route.id in PINNED).map(
  ({ id, label }) => ({ id, label, icon: PINNED[id]!, href: `#/${id}` }),
);

/** Resolves `#/<section>/…rest`; unknown or empty sections fall back to Projects. */
export function useRoute(): { route: Route } & RouteProps {
  const { segments } = useLocation();
  const [section, ...rest] = segments;
  const route = ROUTES.find((candidate) => candidate.id === section);
  return route ? { route, segments: rest } : { route: ROUTES[0]!, segments: [] };
}
