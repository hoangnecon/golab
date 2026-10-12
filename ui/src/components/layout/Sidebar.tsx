import React from 'react';
import { Layers, Globe, Activity, PanelLeftClose, PanelLeftOpen } from 'lucide-react';
import { cn } from '../../lib/utils';

export type TabType = 'sessions' | 'profiles' | 'activity';

interface SidebarProps {
  activeTab: TabType;
  onTabChange: (tab: TabType) => void;
  isCollapsed: boolean;
  onToggleCollapse: () => void;
  sessionCount: number;
  profileCount: number;
  activityCount: number;
  isConnected: boolean;
}

export function Sidebar({
  activeTab,
  onTabChange,
  isCollapsed,
  onToggleCollapse,
  sessionCount,
  profileCount,
  activityCount,
  isConnected,
}: SidebarProps) {
  const coreNav = [
    { id: 'sessions' as const, label: 'Notebook Sessions', icon: Layers, count: sessionCount },
    { id: 'profiles' as const, label: 'Chrome Profiles', icon: Globe, count: profileCount },
  ];

  const obsNav = [
    { id: 'activity' as const, label: 'Activity Logs', icon: Activity, count: activityCount },
  ];

  return (
    <aside
      className={cn(
        'bg-[#0c0d10] border-r border-[#1e2025] flex flex-col justify-between select-none transition-all duration-200 z-30 flex-shrink-0 h-full',
        isCollapsed ? 'w-16' : 'w-60'
      )}
    >
      <div className="flex flex-col flex-1 overflow-y-auto">
        {/* Top Header / App Brand */}
        <div className="h-14 px-4 border-b border-[#1e2025] flex items-center justify-between flex-shrink-0">
          {!isCollapsed ? (
            <div className="flex items-center gap-2.5 min-w-0">
              <div className="w-7 h-7 rounded-md bg-zinc-800 border border-zinc-700 flex items-center justify-center text-xs font-mono font-bold text-white flex-shrink-0">
                GL
              </div>
              <div className="flex flex-col min-w-0">
                <span className="text-xs font-semibold tracking-tight text-zinc-100 truncate">GoLab Daemon</span>
                <span className="text-[10px] font-mono text-zinc-500 truncate">Control Plane v2</span>
              </div>
            </div>
          ) : (
            <button
              onClick={onToggleCollapse}
              title="Expand sidebar"
              className="w-full flex justify-center"
            >
              <div className="w-7 h-7 rounded-md bg-zinc-800 border border-zinc-700 flex items-center justify-center text-xs font-mono font-bold text-white">
                GL
              </div>
            </button>
          )}

          {!isCollapsed && (
            <button
              onClick={onToggleCollapse}
              title="Collapse sidebar (Ctrl+B)"
              className="p-1 text-zinc-500 hover:text-zinc-200 hover:bg-zinc-800/80 rounded transition-colors"
            >
              <PanelLeftClose className="w-4 h-4" />
            </button>
          )}
        </div>

        {/* Navigation Sections */}
        <div className="p-3 space-y-5 flex-1">
          {/* Section: Workbench */}
          <div>
            {!isCollapsed && (
              <div className="px-2 pb-1.5 text-[10px] font-mono font-medium text-zinc-500 uppercase tracking-wider">
                Workbench
              </div>
            )}
            <nav className="space-y-1">
              {coreNav.map((item) => {
                const Icon = item.icon;
                const isActive = activeTab === item.id;

                return (
                  <button
                    key={item.id}
                    onClick={() => onTabChange(item.id)}
                    title={isCollapsed ? `${item.label} (${item.count})` : undefined}
                    className={cn(
                      'w-full flex items-center gap-2.5 px-2.5 py-2 rounded-md text-xs font-medium transition-all group',
                      isActive
                        ? 'bg-zinc-800/90 text-white shadow-sm font-semibold'
                        : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900/80',
                      isCollapsed && 'justify-center px-0'
                    )}
                  >
                    <Icon className={cn('w-4 h-4 flex-shrink-0 transition-colors', isActive ? 'text-white' : 'text-zinc-400 group-hover:text-zinc-200')} />
                    {!isCollapsed && (
                      <>
                        <span className="flex-1 text-left truncate">{item.label}</span>
                        {item.count > 0 && (
                          <span className="text-[10px] font-mono px-1.5 py-0.5 rounded bg-zinc-900 text-zinc-400 border border-zinc-800">
                            {item.count}
                          </span>
                        )}
                      </>
                    )}
                  </button>
                );
              })}
            </nav>
          </div>

          {/* Section: Observability */}
          <div>
            {!isCollapsed && (
              <div className="px-2 pb-1.5 text-[10px] font-mono font-medium text-zinc-500 uppercase tracking-wider">
                Observability
              </div>
            )}
            <nav className="space-y-1">
              {obsNav.map((item) => {
                const Icon = item.icon;
                const isActive = activeTab === item.id;

                return (
                  <button
                    key={item.id}
                    onClick={() => onTabChange(item.id)}
                    title={isCollapsed ? `${item.label} (${item.count})` : undefined}
                    className={cn(
                      'w-full flex items-center gap-2.5 px-2.5 py-2 rounded-md text-xs font-medium transition-all group',
                      isActive
                        ? 'bg-zinc-800/90 text-white shadow-sm font-semibold'
                        : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900/80',
                      isCollapsed && 'justify-center px-0'
                    )}
                  >
                    <Icon className={cn('w-4 h-4 flex-shrink-0 transition-colors', isActive ? 'text-white' : 'text-zinc-400 group-hover:text-zinc-200')} />
                    {!isCollapsed && (
                      <>
                        <span className="flex-1 text-left truncate">{item.label}</span>
                        {item.count > 0 && (
                          <span className="text-[10px] font-mono px-1.5 py-0.5 rounded bg-zinc-900 text-zinc-400 border border-zinc-800">
                            {item.count}
                          </span>
                        )}
                      </>
                    )}
                  </button>
                );
              })}
            </nav>
          </div>
        </div>
      </div>

      {/* System Status Footer */}
      <div className="p-3 border-t border-[#1e2025] bg-[#090a0d]">
        {isCollapsed ? (
          <div className="flex flex-col items-center gap-2.5">
            <div
              className={cn(
                'w-2 h-2 rounded-full',
                isConnected ? 'bg-emerald-400 ring-2 ring-emerald-500/20' : 'bg-zinc-600'
              )}
              title={isConnected ? 'Daemon 9120 Online' : 'Daemon Offline'}
            />
            <button
              onClick={onToggleCollapse}
              title="Expand sidebar"
              className="p-1 text-zinc-500 hover:text-zinc-200 rounded transition-colors"
            >
              <PanelLeftOpen className="w-4 h-4" />
            </button>
          </div>
        ) : (
          <div className="space-y-1.5">
            <div className="flex items-center justify-between text-[11px] font-mono">
              <div className="flex items-center gap-2">
                <span
                  className={cn(
                    'w-2 h-2 rounded-full',
                    isConnected ? 'bg-emerald-400 ring-2 ring-emerald-500/20' : 'bg-zinc-600'
                  )}
                />
                <span className="text-zinc-300 font-medium">Core Daemon</span>
              </div>
              <span className="text-zinc-500 text-[10px]">{isConnected ? 'LIVE' : 'DOWN'}</span>
            </div>
            <div className="text-[10px] font-mono text-zinc-500 pl-4 truncate">
              127.0.0.1:9120
            </div>
          </div>
        )}
      </div>
    </aside>
  );
}
