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

import { Layers, Globe, Activity, PanelLeftClose, PanelLeftOpen, Terminal } from 'lucide-react';

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
    { id: 'activity' as const, label: 'Activity', icon: Activity, count: activityCount },
  ];

  return (
    <aside
      className={cn(
        'bg-[#0b0c0e] border-r border-white/[0.06] flex flex-col justify-between select-none transition-all duration-200 z-30 flex-shrink-0 h-full',
        isCollapsed ? 'w-14' : 'w-56'
      )}
    >
      <div className="flex flex-col flex-1 overflow-y-auto">
        {/* Top Header / App Brand */}
        <div className="h-13 px-3.5 border-b border-white/[0.06] flex items-center justify-between flex-shrink-0">
          {!isCollapsed ? (
            <div className="flex items-center gap-2.5 min-w-0">
              <div className="w-6 h-6 rounded-md bg-white/[0.06] flex items-center justify-center text-zinc-200 flex-shrink-0">
                <Terminal className="w-3.5 h-3.5" strokeWidth={2} />
              </div>
              <span className="text-xs font-semibold tracking-tight text-zinc-200 truncate">GoLab</span>
            </div>
          ) : (
            <button
              onClick={onToggleCollapse}
              title="Expand sidebar"
              className="w-full flex justify-center py-1"
            >
              <div className="w-6 h-6 rounded-md bg-white/[0.06] flex items-center justify-center text-zinc-200">
                <Terminal className="w-3.5 h-3.5" strokeWidth={2} />
              </div>
            </button>
          )}

          {!isCollapsed && (
            <button
              onClick={onToggleCollapse}
              title="Collapse sidebar (Ctrl+B)"
              className="p-1 text-zinc-500 hover:text-zinc-300 hover:bg-white/[0.04] rounded-md transition-colors"
            >
              <PanelLeftClose className="w-3.5 h-3.5" strokeWidth={1.75} />
            </button>
          )}
        </div>

        {/* Navigation List */}
        <div className="p-2 space-y-1 flex-1">
          {navItems.map((item) => {
            const Icon = item.icon;
            const isActive = activeTab === item.id;

            return (
              <button
                key={item.id}
                onClick={() => onTabChange(item.id)}
                title={isCollapsed ? `${item.label} (${item.count})` : undefined}
                className={cn(
                  'w-full flex items-center gap-2.5 px-2.5 py-2 rounded-lg text-xs font-medium transition-all group',
                  isActive
                    ? 'bg-white/[0.08] text-zinc-100 font-semibold shadow-sm'
                    : 'text-zinc-400 hover:text-zinc-200 hover:bg-white/[0.03]',
                  isCollapsed && 'justify-center px-0'
                )}
              >
                <Icon
                  className={cn(
                    'w-4 h-4 flex-shrink-0 transition-colors',
                    isActive ? 'text-zinc-100' : 'text-zinc-400 group-hover:text-zinc-200'
                  )}
                  strokeWidth={1.75}
                />
                {!isCollapsed && (
                  <>
                    <span className="flex-1 text-left truncate">{item.label}</span>
                    {item.count > 0 && (
                      <span className="text-[10px] font-mono px-1.5 py-0.5 rounded-full bg-white/[0.05] text-zinc-400">
                        {item.count}
                      </span>
                    )}
                  </>
                )}
              </button>
            );
          })}
        </div>
      </div>

      {/* System Status Footer */}
      <div className="p-2.5 border-t border-white/[0.06]">
        {isCollapsed ? (
          <div className="flex flex-col items-center gap-2">
            <div
              className={cn(
                'w-2 h-2 rounded-full',
                isConnected ? 'bg-emerald-400 shadow-[0_0_6px_rgba(52,211,153,0.4)]' : 'bg-zinc-600'
              )}
              title={isConnected ? 'Connected to 127.0.0.1:9120' : 'Offline'}
            />
            <button
              onClick={onToggleCollapse}
              title="Expand sidebar (Ctrl+B)"
              className="p-1 text-zinc-500 hover:text-zinc-300 rounded transition-colors"
            >
              <PanelLeftOpen className="w-3.5 h-3.5" strokeWidth={1.75} />
            </button>
          </div>
        ) : (
          <div className="flex items-center justify-between px-1.5 py-1 text-[11px] font-mono text-zinc-500">
            <div className="flex items-center gap-2 min-w-0">
              <span
                className={cn(
                  'w-1.5 h-1.5 rounded-full flex-shrink-0',
                  isConnected ? 'bg-emerald-400 shadow-[0_0_6px_rgba(52,211,153,0.4)]' : 'bg-zinc-600'
                )}
              />
              <span className="truncate">{isConnected ? '127.0.0.1:9120' : 'Disconnected'}</span>
            </div>
          </div>
        )}
      </div>
    </aside>
  );
}
