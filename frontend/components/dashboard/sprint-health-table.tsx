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
  return `${value.toFixed(1)} SP/day`;
}

export function SprintHealthTable({ rows }: { rows: ProjectSprintHealth[] }) {
  return (
    <div>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Project Name</TableHead>
            <TableHead>Sprint</TableHead>
            <TableHead>Committed SP</TableHead>
            <TableHead>Done SP</TableHead>
            <TableHead>Late-Add SP</TableHead>
            <TableHead>Days Remaining</TableHead>
            <TableHead>Required Velocity</TableHead>
            <TableHead>Achieved Velocity</TableHead>
            <TableHead>Early Warning (Status)</TableHead>
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
              <TableCell>{r.overdue ? 'Overdue' : `${r.daysRemaining} days`}</TableCell>
              <TableCell>{formatVelocity(r.requiredVelocity)}</TableCell>
              <TableCell>{formatVelocity(r.achievedVelocity)}</TableCell>
              <TableCell>
                <Badge variant={statusVariant[r.status]}>{statusLabel[r.status]}</Badge>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
      <div className="space-y-1 px-6 py-4 text-xs text-muted-foreground">
        <p>Required Velocity = (Committed SP − Done SP) / Days Remaining</p>
        <p>Achieved Velocity = Done SP / Days Elapsed</p>
        <p>
          Status: <span className="font-medium text-foreground">On Track</span> when achieved velocity meets or
          exceeds what&apos;s required, <span className="font-medium text-foreground">At Risk</span> at 60–99% of
          that pace, <span className="font-medium text-foreground">Critical</span> below 60% or once the sprint runs
          past its end date.
        </p>
      </div>
    </div>
  );
}
