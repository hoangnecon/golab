import React from 'react';
import { Plus, Layers } from 'lucide-react';
import { Session, Profile } from '../../types';
import { SessionCard } from './SessionCard';
import { Button } from '../../components/ui/Button';

interface SessionsPageProps {
  sessions: Session[];
  profiles: Profile[];
  onAttachClick: () => void;
  onInterrupt: (id: string) => void;
  onRestart: (id: string) => void;
  onFocus: (id: string) => void;
  onDetach: (id: string) => void;
}

export function SessionsPage({
  sessions,
  profiles,
  onAttachClick,
  onInterrupt,
  onRestart,
  onFocus,
  onDetach,
}: SessionsPageProps) {
  if (sessions.length === 0) {
    return (
      <div className="border border-dashed border-zinc-800 rounded-lg p-12 text-center max-w-lg mx-auto mt-12 flex flex-col items-center justify-center">
        <div className="w-10 h-10 rounded-full bg-zinc-900 border border-zinc-800 flex items-center justify-center text-zinc-500 mb-3">
          <Layers className="w-5 h-5" />
        </div>
        <h3 className="text-sm font-medium text-zinc-300">No active sessions</h3>
        <p className="text-xs text-zinc-500 mt-1 mb-5 max-w-sm">
          Attach a Colab notebook URL to route kernel commands and inspect execution.
        </p>
        <Button variant="primary" size="sm" onClick={onAttachClick}>
          <Plus className="w-3.5 h-3.5" />
          <span>Attach Notebook</span>
        </Button>
      </div>
    );
  }

  const busyCount = sessions.filter(
    (s) => s.kernel_state === 'busy' || s.kernel_state === 'executing'
  ).length;

  return (
    <div className="space-y-6">
      {/* Metric KPI Overview */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
        <div className="p-3.5 rounded-xl bg-[#121319] border border-[#1f2229]">
          <div className="text-[11px] font-mono text-zinc-500 uppercase tracking-wider">Sessions</div>
          <div className="text-xl font-semibold text-zinc-100 mt-0.5">{sessions.length}</div>
        </div>

        <div className="p-3.5 rounded-xl bg-[#121319] border border-[#1f2229]">
          <div className="text-[11px] font-mono text-zinc-500 uppercase tracking-wider">Busy Kernels</div>
          <div className="text-xl font-semibold text-amber-400 mt-0.5">{busyCount}</div>
        </div>

        <div className="p-3.5 rounded-xl bg-[#121319] border border-[#1f2229]">
          <div className="text-[11px] font-mono text-zinc-500 uppercase tracking-wider">Profiles</div>
          <div className="text-xl font-semibold text-zinc-100 mt-0.5">{profiles.length}</div>
        </div>

        <div className="p-3.5 rounded-xl bg-[#121319] border border-[#1f2229]">
          <div className="text-[11px] font-mono text-zinc-500 uppercase tracking-wider">Direct Tunnel</div>
          <div className="text-xl font-semibold text-emerald-400 mt-0.5">3453 Ready</div>
        </div>
      </div>

      {/* Session Cards Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {sessions.map((sess) => {
          const prof = profiles.find((p) => p.id === sess.profile_id);
          return (
            <SessionCard
              key={sess.id}
              session={sess}
              profile={prof}
              onInterrupt={onInterrupt}
              onRestart={onRestart}
              onFocus={onFocus}
              onDetach={onDetach}
            />
          );
        })}
      </div>
    </div>
  );
}
