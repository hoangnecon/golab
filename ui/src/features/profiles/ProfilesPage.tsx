import React from 'react';
import { Play, Square, Trash2 } from 'lucide-react';
import { Profile } from '../../types';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell, TableEmpty } from '../../components/ui/Table';
import { Badge } from '../../components/ui/Badge';
import { Button } from '../../components/ui/Button';

interface ProfilesPageProps {
  profiles: Profile[];
  onLaunch: (id: string) => void;
  onStop: (id: string) => void;
  onDelete: (id: string) => void;
}

export function ProfilesPage({ profiles, onLaunch, onStop, onDelete }: ProfilesPageProps) {
  return (
    <div className="space-y-6">
      {/* Info card */}
      <div className="p-4 rounded-xl bg-[#121319] border border-[#1f2229] flex items-center justify-between gap-4">
        <div className="space-y-0.5">
          <h3 className="text-xs font-semibold text-zinc-200">Chrome Profile Sandbox Isolation</h3>
          <p className="text-[11px] text-zinc-400">
            Each profile runs with a dedicated user data folder and isolated CDP port (9222+), allowing simultaneous Colab Pro and Free tiers.
          </p>
        </div>
        <div className="text-xs font-mono text-zinc-400 bg-[#090a0d] px-3 py-1.5 rounded-lg border border-[#1e2025]">
          {profiles.length} Configured
        </div>
      </div>

      <Table>
        <TableHeader>
          <tr>
            <TableHead>Profile Name</TableHead>
            <TableHead>Google Account</TableHead>
            <TableHead>Tier</TableHead>
            <TableHead>CDP Port</TableHead>
            <TableHead>Process State</TableHead>
            <TableHead className="text-right">Actions</TableHead>
          </tr>
        </TableHeader>
        <TableBody>
          {profiles.length === 0 ? (
            <TableEmpty colSpan={6} message="No profiles configured" />
          ) : (
            profiles.map((p) => (
              <TableRow key={p.id}>
                <TableCell className="font-sans font-medium text-zinc-200">
                  {p.name}
                </TableCell>
                <TableCell className="text-zinc-400">
                  {p.account_email || '—'}
                </TableCell>
                <TableCell className="uppercase text-zinc-400">
                  <Badge variant="neutral">{p.tier}</Badge>
                </TableCell>
                <TableCell className="text-zinc-400 font-mono">
                  {p.cdp_port}
                </TableCell>
                <TableCell>
                  <Badge variant={p.is_running ? 'success' : 'neutral'} dot>
                    {p.is_running ? 'Active (Ready)' : 'Stopped'}
                  </Badge>
                </TableCell>
                <TableCell className="text-right">
                  <div className="flex items-center justify-end gap-2">
                    {!p.is_running ? (
                      <Button variant="secondary" size="sm" onClick={() => onLaunch(p.id)}>
                        <Play className="w-3 h-3 text-emerald-400" />
                        <span>Launch Chrome</span>
                      </Button>
                    ) : (
                      <Button variant="outline" size="sm" onClick={() => onStop(p.id)}>
                        <Square className="w-3 h-3 text-zinc-400" />
                        <span>Stop</span>
                      </Button>
                    )}
                    {p.id !== 'default' && (
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => onDelete(p.id)}
                        title="Delete profile"
                      >
                        <Trash2 className="w-3.5 h-3.5 text-zinc-500 hover:text-red-400" />
                      </Button>
                    )}
                  </div>
                </TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </div>
  );
}
