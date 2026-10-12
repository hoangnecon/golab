import React from 'react';
import { ActivityLog } from '../../types';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell, TableEmpty } from '../../components/ui/Table';
import { Badge } from '../../components/ui/Badge';

interface ActivityPageProps {
  activity: ActivityLog[];
}

export function ActivityPage({ activity }: ActivityPageProps) {
  const totalInvocations = activity.length;
  const errorCount = activity.filter((a) => a.status !== 'success').length;
  const avgLatency =
    totalInvocations > 0
      ? Math.round(activity.reduce((acc, a) => acc + (a.latency_ms || 0), 0) / totalInvocations)
      : 0;

  return (
    <div className="space-y-6">
      {/* Metric KPI Overview */}
      <div className="grid grid-cols-3 gap-3">
        <div className="p-3.5 rounded-xl bg-[#121319] border border-[#1f2229]">
          <div className="text-[11px] font-mono text-zinc-500 uppercase tracking-wider">Total Calls</div>
          <div className="text-xl font-semibold text-zinc-100 mt-0.5">{totalInvocations}</div>
        </div>

        <div className="p-3.5 rounded-xl bg-[#121319] border border-[#1f2229]">
          <div className="text-[11px] font-mono text-zinc-500 uppercase tracking-wider">Avg Latency</div>
          <div className="text-xl font-semibold text-emerald-400 mt-0.5">{avgLatency} ms</div>
        </div>

        <div className="p-3.5 rounded-xl bg-[#121319] border border-[#1f2229]">
          <div className="text-[11px] font-mono text-zinc-500 uppercase tracking-wider">Errors</div>
          <div className="text-xl font-semibold text-red-400 mt-0.5">{errorCount}</div>
        </div>
      </div>

      <Table>
        <TableHeader>
          <tr>
            <TableHead>Timestamp</TableHead>
            <TableHead>Agent MCP Tool</TableHead>
            <TableHead>Execution Latency</TableHead>
            <TableHead>Result Status</TableHead>
            <TableHead>Execution Payload / Details</TableHead>
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
                    {log.status.toUpperCase()}
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
