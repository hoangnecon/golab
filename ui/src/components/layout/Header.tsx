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
    sessions: 'Active Notebook Sessions',
    profiles: 'Google Accounts & CDP Profiles',
    activity: 'Agent Invocation Stream',
  };

  const actionLabels: Record<TabType, string | null> = {
    sessions: 'Attach Notebook',
    profiles: 'New Profile',
    activity: null,
  };

  const actionLabel = actionLabels[activeTab];

  return (
    <header className="h-14 px-6 border-b border-zinc-800/80 flex items-center justify-between bg-[#09090b]/40 backdrop-blur-sm sticky top-0 z-10 flex-shrink-0">
      <div className="flex items-center gap-3 min-w-0">
        {isCollapsed && (
          <button
            onClick={onToggleCollapse}
            title="Expand sidebar (Ctrl+B)"
            className="p-1 text-zinc-500 hover:text-zinc-300 hover:bg-zinc-900 rounded transition-colors mr-1"
          >
            <PanelLeftOpen className="w-4 h-4" />
          </button>
        )}
        <h1 className="text-xs font-medium text-zinc-200 truncate">{titles[activeTab]}</h1>

        {statusMessage && (
          <span className="text-[11px] font-mono text-zinc-400 bg-zinc-800/90 px-2 py-0.5 rounded border border-zinc-700/60 animate-in fade-in duration-100">
            {statusMessage}
          </span>
        )}
      </div>

      <div className="flex items-center gap-3">
        <div className="flex items-center gap-1.5 px-2 py-0.5 rounded bg-zinc-900 border border-zinc-800 text-[11px] font-mono text-zinc-400">
          <span className={`w-1.5 h-1.5 rounded-full ${isConnected ? 'bg-emerald-500' : 'bg-zinc-600'}`} />
          <span>{isConnected ? 'live' : 'offline'}</span>
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
