import React from 'react';
import { ExternalLink, Square, RotateCcw, Eye, Trash2 } from 'lucide-react';
import { Session, Profile } from '../../types';
import { Card, CardHeader, CardContent, CardFooter } from '../../components/ui/Card';
import { Badge } from '../../components/ui/Badge';
import { Button } from '../../components/ui/Button';

interface SessionCardProps {
  session: Session;
  profile?: Profile;
  onInterrupt: (id: string) => void;
  onRestart: (id: string) => void;
  onFocus: (id: string) => void;
  onDetach: (id: string) => void;
}

export function SessionCard({
  session,
  profile,
  onInterrupt,
  onRestart,
  onFocus,
  onDetach,
}: SessionCardProps) {
  const isBusy = session.kernel_state === 'busy' || session.kernel_state === 'executing';
  const statusVariant = isBusy
    ? 'warning'
    : session.kernel_state === 'idle'
    ? 'success'
    : 'neutral';

  return (
    <div className="bg-[#121319] border border-[#1f2229] hover:border-[#2a2f3a] rounded-xl p-4 sm:p-5 flex flex-col gap-4 transition-all shadow-sm">
      {/* Header with Title, Status & Profile */}
      <div className="space-y-2">
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0 flex-1">
            <h3 className="text-sm font-semibold text-zinc-100 truncate" title={session.name}>
              {session.name}
            </h3>
            <div className="flex items-center gap-2 mt-1 text-[11px] font-mono text-zinc-400">
              <span className="px-2 py-0.5 rounded bg-[#181a22] border border-[#252833] text-zinc-300">
                {profile?.name || session.profile_id}
              </span>
              <a
                href={session.notebook_url}
                target="_blank"
                rel="noreferrer"
                className="hover:text-white flex items-center gap-1 transition-colors underline-offset-2 hover:underline text-zinc-500 hover:text-zinc-300"
              >
                <span>Google Colab</span>
                <ExternalLink className="w-3 h-3" />
              </a>
            </div>
          </div>

          <Badge variant={statusVariant} dot>
            {(session.kernel_state || 'unknown').toUpperCase()}
          </Badge>
        </div>
      </div>

      {/* Telemetry Hardware Matrix */}
      <div className="grid grid-cols-3 gap-2 bg-[#0a0b0f] p-3 rounded-lg border border-[#1a1c24] font-mono">
        <div className="min-w-0">
          <div className="text-[10px] text-zinc-500 uppercase tracking-wider mb-1 font-sans">GPU</div>
          <div className="text-xs text-zinc-200 font-semibold truncate" title={session.gpu_name || 'None'}>
            {session.gpu_name || 'None'}
          </div>
          <div className="text-[10px] text-zinc-500 truncate mt-0.5">
            {session.gpu_memory_free_mb ? `${session.gpu_memory_free_mb} MB free` : 'CPU Only'}
          </div>
        </div>

        <div className="min-w-0 border-l border-[#1a1c24] pl-3">
          <div className="text-[10px] text-zinc-500 uppercase tracking-wider mb-1 font-sans">RAM</div>
          <div className="text-xs text-zinc-200 font-semibold truncate">
            {session.ram_free_gb ? `${session.ram_free_gb.toFixed(1)} GB` : '—'}
          </div>
          <div className="text-[10px] text-zinc-500 truncate mt-0.5">Available</div>
        </div>

        <div className="min-w-0 border-l border-[#1a1c24] pl-3">
          <div className="text-[10px] text-zinc-500 uppercase tracking-wider mb-1 font-sans">Drive</div>
          <div className={session.drive_mounted ? 'text-xs text-emerald-400 font-semibold truncate' : 'text-xs text-zinc-500 font-semibold truncate'}>
            {session.drive_mounted ? 'Mounted' : 'Offline'}
          </div>
          <div className="text-[10px] text-zinc-500 truncate mt-0.5">
            {session.drive_mounted ? '/content/drive' : 'Unmounted'}
          </div>
        </div>
      </div>

      {/* Action Toolbar */}
      <div className="flex items-center justify-between gap-2 pt-2 border-t border-[#1a1c24]">
        <div className="flex items-center gap-2">
          <Button
            variant="destructive"
            size="sm"
            onClick={() => onInterrupt(session.id)}
            title="Send SIGINT signal to interrupt cell"
          >
            <Square className="w-3 h-3 text-red-300" />
            <span>Interrupt</span>
          </Button>

          <Button
            variant="secondary"
            size="sm"
            onClick={() => onRestart(session.id)}
            title="Restart kernel runtime"
          >
            <RotateCcw className="w-3 h-3 text-zinc-400" />
            <span>Restart</span>
          </Button>

          <Button
            variant="secondary"
            size="sm"
            onClick={() => onFocus(session.id)}
            title="Bring browser tab to front"
          >
            <Eye className="w-3 h-3 text-zinc-400" />
            <span>Focus</span>
          </Button>
        </div>

        <Button
          variant="ghost"
          size="icon"
          onClick={() => onDetach(session.id)}
          title="Detach session"
        >
          <Trash2 className="w-3.5 h-3.5 text-zinc-500 hover:text-red-400 transition-colors" />
        </Button>
      </div>
    </div>
  );
}
