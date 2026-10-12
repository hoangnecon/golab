import React from 'react';
import { Plus, PanelLeftOpen } from 'lucide-react';
import { Button } from '../ui/Button';
import { TabType } from './Sidebar';

interface HeaderProps {
  activeTab: TabType;
  isCollapsed: boolean;
  onToggleCollapse: () => void;
  onPrimaryAction: () => void;
  statusMessage: string | null;
  isConnected: boolean;
}

export function Header({
  activeTab,
  isCollapsed,
  onToggleCollapse,
  onPrimaryAction,
  statusMessage,
  isConnected,
}: HeaderProps) {
  const titles: Record<TabType, string> = {
    sessions: 'Sessions',
    profiles: 'Profiles',
    activity: 'Activity',
  };

  const actionLabels: Record<TabType, string | null> = {
    sessions: 'Attach Notebook',
    profiles: 'New Profile',
    activity: null,
  };

  const title = titles[activeTab];
  const actionLabel = actionLabels[activeTab];

  return (
    <header className="h-13 px-4 sm:px-6 border-b border-white/[0.06] flex items-center justify-between bg-[#0b0c0e]/80 backdrop-blur-md sticky top-0 z-20 flex-shrink-0 gap-3">
      {/* Left side: Title & Sidebar toggle */}
      <div className="flex items-center gap-3 min-w-0">
        {isCollapsed && (
          <button
            onClick={onToggleCollapse}
            title="Expand sidebar (Ctrl+B)"
            className="p-1 text-zinc-500 hover:text-zinc-300 hover:bg-white/[0.04] rounded-md transition-colors flex-shrink-0"
          >
            <PanelLeftOpen className="w-3.5 h-3.5" strokeWidth={1.75} />
          </button>
        )}

        <div className="flex items-center gap-2.5 min-w-0">
          <h1 className="text-xs font-semibold text-zinc-200 truncate">{title}</h1>
          {statusMessage && (
            <span className="hidden sm:inline-block text-[11px] font-mono text-zinc-400 bg-white/[0.04] px-2 py-0.5 rounded border border-white/[0.06] truncate">
              {statusMessage}
            </span>
          )}
        </div>
      </div>

      {/* Right side: Action */}
      <div className="flex items-center gap-2.5 flex-shrink-0">
        {actionLabel && (
          <Button variant="primary" size="sm" onClick={onPrimaryAction}>
            <Plus className="w-3.5 h-3.5" strokeWidth={2} />
            <span>{actionLabel}</span>
          </Button>
        )}
      </div>
    </header>
  );
}
