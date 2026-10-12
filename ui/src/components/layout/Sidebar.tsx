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
  const navItems = [
    { id: 'sessions' as const, label: 'Sessions', icon: Layers, count: sessionCount },
    { id: 'profiles' as const, label: 'Profiles', icon: Globe, count: profileCount },
    { id: 'activity' as const, label: 'Activity Log', icon: Activity, count: activityCount },
  ];

  return (
    <aside
      className={cn(
        'bg-[#09090b] border-r border-zinc-800/80 flex flex-col justify-between select-none transition-all duration-200 z-20 flex-shrink-0',
        isCollapsed ? 'w-16' : 'w-60'
      )}
    >
      <div>
        {/* Brand header */}
        <div className="h-14 px-4 border-b border-zinc-800/80 flex items-center justify-between">
          {!isCollapsed ? (
            <div className="flex items-baseline gap-2 overflow-hidden">
              <span className="text-sm font-semibold tracking-tight text-zinc-100">GoLab</span>
              <span className="text-[11px] font-mono text-zinc-500">control</span>
            </div>
          ) : (
            <div className="w-full flex justify-center">
              <span className="text-sm font-semibold font-mono text-zinc-200">G</span>
            </div>
          )}

          <button
            onClick={onToggleCollapse}
            title={isCollapsed ? 'Expand sidebar (Ctrl+B)' : 'Collapse sidebar (Ctrl+B)'}
            className={cn(
              'p-1 text-zinc-500 hover:text-zinc-300 hover:bg-zinc-900 rounded transition-colors',
              isCollapsed && 'hidden'
            )}
          >
            <PanelLeftClose className="w-4 h-4" />
          </button>
        </div>

        {/* Nav list */}
        <nav className="p-2 space-y-1">
          {navItems.map((item) => {
            const Icon = item.icon;
            const isActive = activeTab === item.id;

            return (
              <button
                key={item.id}
                onClick={() => onTabChange(item.id)}
                title={isCollapsed ? item.label : undefined}
                className={cn(
                  'w-full flex items-center gap-2.5 px-3 py-2 rounded-md text-xs font-medium transition-colors',
                  isActive
                    ? 'bg-zinc-800 text-zinc-100 shadow-sm'
                    : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900/60',
                  isCollapsed && 'justify-center px-0'
                )}
              >
                <Icon className="w-4 h-4 flex-shrink-0" />
                {!isCollapsed && (
                  <>
                    <span className="flex-1 text-left">{item.label}</span>
                    {item.count > 0 && (
                      <span className="text-[10px] font-mono px-1.5 py-0.2 rounded bg-zinc-900 text-zinc-400 border border-zinc-800">
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

      {/* Sidebar footer */}
      <div className="p-3 border-t border-zinc-800/80">
        {isCollapsed ? (
          <div className="flex flex-col items-center gap-2">
            <div
              className={cn(
                'w-2 h-2 rounded-full',
                isConnected ? 'bg-emerald-500' : 'bg-zinc-600'
              )}
              title={isConnected ? 'Daemon 127.0.0.1:9120 connected' : 'Daemon offline'}
            />
            <button
              onClick={onToggleCollapse}
              title="Expand sidebar"
              className="p-1 text-zinc-500 hover:text-zinc-300 hover:bg-zinc-900 rounded transition-colors"
            >
              <PanelLeftOpen className="w-4 h-4" />
            </button>
          </div>
        ) : (
          <div className="flex items-center justify-between text-[11px] font-mono text-zinc-500">
            <div className="flex items-center gap-2">
              <span
                className={cn(
                  'w-1.5 h-1.5 rounded-full',
                  isConnected ? 'bg-emerald-500' : 'bg-zinc-600'
                )}
              />
              <span>127.0.0.1:9120</span>
            </div>
            <span className="text-[10px] text-zinc-600">v2.0</span>
          </div>
        )}
      </div>
    </aside>
  );
}
