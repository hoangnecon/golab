import React from 'react';
import { ActivityLog } from '../../types';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell, TableEmpty } from '../../components/ui/Table';
import { Badge } from '../../components/ui/Badge';

interface ActivityPageProps {
  activity: ActivityLog[];
}

export function ActivityPage({ activity }: ActivityPageProps) {
  return (
    <div className="space-y-4">
      <Table>
        <TableHeader>
          <tr>
            <TableHead>Timestamp</TableHead>
            <TableHead>Agent MCP Tool</TableHead>
            <TableHead>Latency</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Details</TableHead>
          </tr>
        </TableHeader>
        <TableBody>
          {activity.length === 0 ? (
            <TableEmpty colSpan={5} message="No agent invocations recorded yet" />
          ) : (
            activity.map((log) => (
              <TableRow key={log.id}>
                <TableCell className="text-zinc-500 font-mono text-[11px]">
                  {new Date(log.created_at).toLocaleTimeString()}
                </TableCell>
                <TableCell className="font-mono font-medium text-zinc-200">
                  {log.tool_name}
                </TableCell>
                <TableCell className="text-zinc-400 font-mono text-[11px]">
                  {log.latency_ms}ms
                </TableCell>
                <TableCell>
                  <Badge variant={log.status === 'success' ? 'success' : 'error'}>
                    {log.status}
                  </Badge>
                </TableCell>
                <TableCell className="text-zinc-500 font-mono text-[11px] truncate max-w-sm">
                  {log.error_message || 'OK'}
                </TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </div>
  );
}
