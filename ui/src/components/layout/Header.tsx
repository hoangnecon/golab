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
  const titles: Record<TabType, { title: string; subtitle: string }> = {
    sessions: {
      title: 'Notebook Sessions',
      subtitle: 'Active Colab tabs routed via direct tunnel',
    },
    profiles: {
      title: 'Chrome Profiles',
      subtitle: 'Multi-account Chrome instances with isolated CDP ports',
    },
    activity: {
      title: 'Agent Activity Stream',
      subtitle: 'Real-time telemetry and MCP tool invocation logs',
    },
  };

  const actionLabels: Record<TabType, string | null> = {
    sessions: 'Attach Notebook',
    profiles: 'Add Profile',
    activity: null,
  };

  const { title, subtitle } = titles[activeTab];
  const actionLabel = actionLabels[activeTab];

  return (
    <header className="h-14 px-4 sm:px-6 border-b border-[#1e2025] flex items-center justify-between bg-[#0e1014]/90 backdrop-blur-md sticky top-0 z-20 flex-shrink-0 gap-3">
      {/* Left side: Breadcrumb & Title */}
      <div className="flex items-center gap-3 min-w-0">
        {isCollapsed && (
          <button
            onClick={onToggleCollapse}
            title="Expand sidebar (Ctrl+B)"
            className="p-1.5 text-zinc-400 hover:text-white hover:bg-zinc-800 rounded-md transition-colors flex-shrink-0"
          >
            <PanelLeftOpen className="w-4 h-4" />
          </button>
        )}

        <div className="flex flex-col min-w-0">
          <div className="flex items-center gap-2">
            <span className="text-xs font-semibold text-zinc-100 truncate">{title}</span>
            {statusMessage && (
              <span className="hidden md:inline-block text-[11px] font-mono text-emerald-400 bg-emerald-950/40 px-2 py-0.5 rounded border border-emerald-800/40 truncate animate-in fade-in">
                {statusMessage}
              </span>
            )}
          </div>
          <span className="hidden sm:inline-block text-[11px] text-zinc-500 truncate">{subtitle}</span>
        </div>
      </div>

      {/* Right side: Connection + Action */}
      <div className="flex items-center gap-2.5 flex-shrink-0">
        <div className="flex items-center gap-2 px-2.5 py-1 rounded-md bg-[#14161b] border border-[#22252c] text-[11px] font-mono text-zinc-400">
          <span className={`w-1.5 h-1.5 rounded-full flex-shrink-0 ${isConnected ? 'bg-emerald-400 shadow-[0_0_8px_rgba(52,211,153,0.5)]' : 'bg-zinc-600'}`} />
          <span className="hidden sm:inline">{isConnected ? 'CONNECTED' : 'DISCONNECTED'}</span>
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
