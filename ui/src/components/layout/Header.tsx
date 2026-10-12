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
    sessions: 'Active Sessions',
    profiles: 'Chrome Profiles',
    activity: 'Activity Stream',
  };

  const actionLabels: Record<TabType, string | null> = {
    sessions: 'Attach',
    profiles: 'New Profile',
    activity: null,
  };

  const actionLabel = actionLabels[activeTab];

  return (
    <header className="h-14 px-3 sm:px-6 border-b border-zinc-800/80 flex items-center justify-between bg-[#09090b]/80 backdrop-blur-md sticky top-0 z-10 flex-shrink-0 gap-2">
      {/* Left section */}
      <div className="flex items-center gap-2 sm:gap-3 min-w-0 flex-1">
        <button
          onClick={onToggleCollapse}
          title={isCollapsed ? 'Expand sidebar (Ctrl+B)' : 'Collapse sidebar (Ctrl+B)'}
          className="p-1.5 text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900 rounded-md transition-colors flex-shrink-0"
        >
          <PanelLeftOpen className="w-4 h-4" />
        </button>

        <h1 className="text-xs font-semibold text-zinc-100 truncate">{titles[activeTab]}</h1>

        {statusMessage && (
          <span className="hidden sm:inline-block text-[11px] font-mono text-zinc-400 bg-zinc-800/90 px-2 py-0.5 rounded border border-zinc-700/60 truncate max-w-[200px]">
            {statusMessage}
          </span>
        )}
      </div>

      {/* Right section */}
      <div className="flex items-center gap-2 sm:gap-3 flex-shrink-0">
        <div className="flex items-center gap-1.5 px-2 py-1 rounded bg-zinc-900 border border-zinc-800 text-[11px] font-mono text-zinc-400">
          <span className={`w-1.5 h-1.5 rounded-full flex-shrink-0 ${isConnected ? 'bg-emerald-500' : 'bg-zinc-600'}`} />
          <span className="hidden sm:inline">{isConnected ? 'live' : 'offline'}</span>
        </div>

        {actionLabel && (
          <Button variant="primary" size="sm" onClick={onPrimaryAction}>
            <Plus className="w-3.5 h-3.5" />
            <span>{actionLabel}</span>
          </Button>
        )}
      </div>
    </header>
  );
}
