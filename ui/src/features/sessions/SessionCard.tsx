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
      {/* Header */}
      <CardHeader>
        <div className="min-w-0 flex items-center gap-2.5">
          <span className="text-xs font-semibold text-zinc-200 truncate">{session.name}</span>
          <span className="text-[11px] font-mono text-zinc-500 bg-zinc-800/50 px-2 py-0.5 rounded border border-zinc-800">
            Profile: {profile?.name || session.profile_id}
          </span>
          <a
            href={session.notebook_url}
            target="_blank"
            rel="noreferrer"
            className="text-[11px] font-mono text-zinc-500 hover:text-zinc-300 flex items-center gap-1 transition-colors"
          >
            <span>Open</span>
            <ExternalLink className="w-3 h-3" />
          </a>
        </div>

        <Badge variant={statusVariant} dot>
          {session.kernel_state || 'unknown'}
        </Badge>
      </CardHeader>

      {/* Telemetry info */}
      <CardContent>
        <div className="flex flex-wrap items-center gap-4 text-xs font-mono text-zinc-400 py-2 px-3 bg-zinc-950/60 rounded border border-zinc-800/50">
          <div>
            <span className="text-zinc-500 mr-1.5">GPU:</span>
            <span>{session.gpu_name ? session.gpu_name : 'None'}</span>
            {session.gpu_memory_free_mb ? (
              <span className="text-zinc-500 ml-1">({session.gpu_memory_free_mb} MB free)</span>
            ) : null}
          </div>

          <div className="border-l border-zinc-800 pl-4">
            <span className="text-zinc-500 mr-1.5">RAM:</span>
            <span>{session.ram_free_gb ? `${session.ram_free_gb.toFixed(1)} GB free` : '—'}</span>
          </div>

          <div className="border-l border-zinc-800 pl-4">
            <span className="text-zinc-500 mr-1.5">Drive:</span>
            <span className={session.drive_mounted ? 'text-emerald-400' : 'text-zinc-500'}>
              {session.drive_mounted ? 'Mounted' : 'Unmounted'}
            </span>
          </div>
        </div>
      </CardContent>

      {/* Actions */}
      <CardFooter>
        <div className="flex items-center gap-2">
          <Button
            variant="secondary"
            size="sm"
            onClick={() => onInterrupt(session.id)}
            title="Send SIGINT signal to kernel"
          >
            <Square className="w-3 h-3 text-amber-400" />
            <span>Interrupt</span>
          </Button>

          <Button
            variant="secondary"
            size="sm"
            onClick={() => onRestart(session.id)}
            title="Restart kernel runtime"
          >
            <RotateCcw className="w-3 h-3" />
            <span>Restart</span>
          </Button>

          <Button
            variant="secondary"
            size="sm"
            onClick={() => onFocus(session.id)}
            title="Bring browser tab to front"
          >
            <Eye className="w-3 h-3" />
            <span>Focus Tab</span>
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
      </CardFooter>
    </Card>
  );
}
