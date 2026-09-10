import { Badge } from '@/components/ui/badge';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import type { ProjectSprintHealth, SprintHealthStatus } from '@/lib/api';

const statusLabel: Record<SprintHealthStatus, string> = {
  OnTrack: 'On Track',
  AtRisk: 'At Risk',
  Critical: 'Critical',
};

const statusVariant: Record<SprintHealthStatus, 'success' | 'warning' | 'destructive'> = {
  OnTrack: 'success',
  AtRisk: 'warning',
  Critical: 'destructive',
};

function formatVelocity(value: number | null) {
  if (value === null) return '—';
  return `${value.toFixed(1)} SP/hari`;
}

export function SprintHealthTable({ rows }: { rows: ProjectSprintHealth[] }) {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Project Name</TableHead>
          <TableHead>Sprint</TableHead>
          <TableHead>Committed SP</TableHead>
          <TableHead>Done SP</TableHead>
          <TableHead>Late-Add SP</TableHead>
          <TableHead>Sisa Waktu</TableHead>
          <TableHead>Velocity Dibutuhkan</TableHead>
          <TableHead>Velocity Sejauh Ini</TableHead>
          <TableHead>Sinyal Dini (Status)</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {rows.length === 0 && (
          <TableRow>
            <TableCell colSpan={9} className="text-center text-muted-foreground">
              No projects with work in an open sprint.
            </TableCell>
          </TableRow>
        )}
        {rows.map((r) => (
          <TableRow key={`${r.projectId}-${r.sprintId}`}>
            <TableCell className="font-medium">{r.projectName}</TableCell>
            <TableCell className="text-muted-foreground">{r.sprintName}</TableCell>
            <TableCell>{r.committedPoints}</TableCell>
            <TableCell>{r.donePoints}</TableCell>
            <TableCell className="text-muted-foreground">
              {r.lateAddPoints > 0 ? `+${r.lateAddPoints}` : '—'}
            </TableCell>
            <TableCell>{r.overdue ? 'Overdue' : `${r.daysRemaining} hari`}</TableCell>
            <TableCell>{formatVelocity(r.requiredVelocity)}</TableCell>
            <TableCell>{formatVelocity(r.achievedVelocity)}</TableCell>
            <TableCell>
              <Badge variant={statusVariant[r.status]}>{statusLabel[r.status]}</Badge>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
