import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import type { SprintRetrospective } from '@/lib/api';

function formatPercent(value: number | null) {
  if (value === null) return '—';
  return `${value.toFixed(0)}%`;
}

// Lighter than the cell's own muted text (/70) so the count stays secondary
// even in the Late-Add column, where the SP number is already muted.
function TicketCount({ count }: { count: number }) {
  return <span className="text-muted-foreground/70"> ({count} {count === 1 ? 'ticket' : 'tickets'})</span>;
}

export function SprintRetrospectiveTable({ rows }: { rows: SprintRetrospective[] }) {
  return (
    <div>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Sprint</TableHead>
            <TableHead>Committed SP</TableHead>
            <TableHead>Committed Done SP</TableHead>
            <TableHead>Late-Add SP</TableHead>
            <TableHead>Planning Accuracy</TableHead>
            <TableHead>Late-Add Rate</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.length === 0 && (
            <TableRow>
              <TableCell colSpan={6} className="text-center text-muted-foreground">
                No closed sprints with work in them yet.
              </TableCell>
            </TableRow>
          )}
          {rows.map((r) => (
            <TableRow key={r.sprintId}>
              <TableCell className="font-medium">{r.sprintName}</TableCell>
              <TableCell>
                {r.committedPoints}
                <TicketCount count={r.committedTickets} />
              </TableCell>
              <TableCell>
                {r.committedDonePoints}
                <TicketCount count={r.committedDoneTickets} />
              </TableCell>
              <TableCell className="text-muted-foreground">
                {r.lateAddTickets > 0 ? (
                  <>
                    +{r.lateAddPoints}
                    <TicketCount count={r.lateAddTickets} />
                  </>
                ) : (
                  '—'
                )}
              </TableCell>
              <TableCell>{formatPercent(r.planningAccuracy)}</TableCell>
              <TableCell>{formatPercent(r.lateAddRate)}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
      <div className="space-y-1 px-6 py-4 text-xs text-muted-foreground">
        <p>Planning Accuracy = Committed Done SP ÷ Committed SP × 100 — how much of the original commitment got done.</p>
        <p>Late-Add Rate = Late-Add SP ÷ Committed SP × 100 — how much work got added after the sprint started, relative to what was committed.</p>
        <p>Counts every project regardless of its current status, so a later archive doesn&apos;t reshape a past sprint&apos;s numbers.</p>
      </div>
    </div>
  );
}
