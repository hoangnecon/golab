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
    <Card>
      {/* Header with Title and Status */}
      <div className="space-y-1.5">
        <div className="flex items-center justify-between gap-2">
          <h3 className="text-xs font-semibold text-zinc-100 truncate min-w-0 flex-1" title={session.name}>
            {session.name}
          </h3>
          <Badge variant={statusVariant} dot>
            {session.kernel_state || 'unknown'}
          </Badge>
        </div>

        <div className="flex flex-wrap items-center gap-2 text-[11px] font-mono text-zinc-500">
          <span className="bg-zinc-800/60 px-1.5 py-0.5 rounded text-zinc-400 border border-zinc-800 truncate max-w-[160px]">
            {profile?.name || session.profile_id}
          </span>
          <a
            href={session.notebook_url}
            target="_blank"
            rel="noreferrer"
            className="hover:text-zinc-300 flex items-center gap-1 transition-colors underline-offset-2 hover:underline"
          >
            <span>Open Colab</span>
            <ExternalLink className="w-3 h-3" />
          </a>
        </div>
      </div>

      {/* Telemetry Grid */}
      <div className="grid grid-cols-3 gap-2 bg-[#09090b] p-2.5 rounded-md border border-zinc-800/60 font-mono text-[11px]">
        <div className="min-w-0">
          <div className="text-[10px] text-zinc-500 uppercase tracking-wider mb-0.5">GPU</div>
          <div className="text-zinc-200 font-medium truncate" title={session.gpu_name || 'None'}>
            {session.gpu_name || 'None'}
          </div>
          <div className="text-[10px] text-zinc-500 truncate">
            {session.gpu_memory_free_mb ? `${session.gpu_memory_free_mb} MB free` : '—'}
          </div>
        </div>

        <div className="min-w-0 border-l border-zinc-800/80 pl-2">
          <div className="text-[10px] text-zinc-500 uppercase tracking-wider mb-0.5">RAM</div>
          <div className="text-zinc-200 font-medium truncate">
            {session.ram_free_gb ? `${session.ram_free_gb.toFixed(1)} GB` : '—'}
          </div>
          <div className="text-[10px] text-zinc-500">available</div>
        </div>

        <div className="min-w-0 border-l border-zinc-800/80 pl-2">
          <div className="text-[10px] text-zinc-500 uppercase tracking-wider mb-0.5">Drive</div>
          <div className={session.drive_mounted ? 'text-emerald-400 font-medium truncate' : 'text-zinc-500 font-medium truncate'}>
            {session.drive_mounted ? 'Mounted' : 'Unmounted'}
          </div>
          <div className="text-[10px] text-zinc-500 truncate">
            {session.drive_mounted ? '/content/drive' : 'offline'}
          </div>
        </div>
      </div>

      {/* Action Toolbar */}
      <div className="flex flex-wrap items-center justify-between gap-2 pt-1 border-t border-zinc-800/60">
        <div className="flex flex-wrap items-center gap-1.5">
          <Button
            variant="destructive"
            size="sm"
            onClick={() => onInterrupt(session.id)}
            title="Send SIGINT signal to kernel"
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
          title="Detach notebook from GoLab"
        >
          <Trash2 className="w-3.5 h-3.5 text-zinc-500 hover:text-red-400 transition-colors" />
        </Button>
      </div>
    </Card>
  );
}
