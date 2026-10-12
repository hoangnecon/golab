import React from 'react';
import { Sidebar, TabType } from './Sidebar';
import { Header } from './Header';
import { useSidebar } from '../../hooks/useSidebar';

interface AppLayoutProps {
  activeTab: TabType;
  onTabChange: (tab: TabType) => void;
  onPrimaryAction: () => void;
  statusMessage: string | null;
  isConnected: boolean;
  sessionCount: number;
  profileCount: number;
  activityCount: number;
  children: React.ReactNode;
}

export function AppLayout({
  activeTab,
  onTabChange,
  onPrimaryAction,
  statusMessage,
  isConnected,
  sessionCount,
  profileCount,
  activityCount,
  children,
}: AppLayoutProps) {
  const { isCollapsed, toggleSidebar } = useSidebar();

  return (
    <div className="flex h-screen bg-[#09090b] text-zinc-100 font-sans antialiased overflow-hidden">
      <Sidebar
        activeTab={activeTab}
        onTabChange={onTabChange}
        isCollapsed={isCollapsed}
        onToggleCollapse={toggleSidebar}
        sessionCount={sessionCount}
        profileCount={profileCount}
        activityCount={activityCount}
        isConnected={isConnected}
      />

      <div className="flex-1 flex flex-col min-w-0 bg-[#0c0c0e]">
        <Header
          activeTab={activeTab}
          isCollapsed={isCollapsed}
          onToggleCollapse={toggleSidebar}
          onPrimaryAction={onPrimaryAction}
          statusMessage={statusMessage}
          isConnected={isConnected}
        />

        <main className="flex-1 overflow-y-auto p-6">{children}</main>
      </div>
    </div>
  );
}
