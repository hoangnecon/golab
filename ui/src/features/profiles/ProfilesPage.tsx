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
    <div className="space-y-4">
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
                    {p.is_running ? 'Running' : 'Stopped'}
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
