import Link from 'next/link';
import { notFound } from 'next/navigation';

import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { WorkloadDoneChart } from '@/components/dashboard/workload-done-chart';
import { getSprint, getSprintDeveloperBreakdown, getSprintTicketBreakdown, type EntryStatus, type SprintEntryDetail } from '@/lib/api';

const statusVariant: Record<EntryStatus, 'success' | 'warning' | 'secondary'> = {
  Done: 'success',
  NotDone: 'warning',
  Cancelled: 'secondary',
};

function TicketTable({
  title,
  entries,
  showCarriedFrom,
  emptyMessage,
}: {
  title: string;
  entries: SprintEntryDetail[];
  showCarriedFrom: boolean;
  emptyMessage: string;
}) {
  const totalPoints = entries.reduce((sum, e) => sum + e.pointsAtEntry, 0);
  const columnCount = showCarriedFrom ? 5 : 4;

  return (
    <Card>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
        <p className="text-sm text-muted-foreground">
          {entries.length} ticket{entries.length === 1 ? '' : 's'} &middot; {totalPoints} pts
        </p>
      </CardHeader>
      <CardContent className="p-0">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Ticket</TableHead>
              <TableHead>Assignee</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Points</TableHead>
              {showCarriedFrom && <TableHead>Carried from</TableHead>}
            </TableRow>
          </TableHeader>
          <TableBody>
            {entries.length === 0 && (
              <TableRow>
                <TableCell colSpan={columnCount} className="text-center text-muted-foreground">
                  {emptyMessage}
                </TableCell>
              </TableRow>
            )}
            {entries.map((e) => (
              <TableRow key={e.entryId}>
                <TableCell className="font-medium">
                  {e.projectName} - {e.ticketTitle}
                </TableCell>
                <TableCell>{e.assigneeName}</TableCell>
                <TableCell>
                  <div className="flex flex-wrap gap-1">
                    <Badge variant={statusVariant[e.status]}>{e.status}</Badge>
                    {e.addedAfterSprintStart && <Badge variant="secondary">Late add</Badge>}
                  </div>
                </TableCell>
                <TableCell>{e.pointsAtEntry}</TableCell>
                {showCarriedFrom && <TableCell className="text-muted-foreground">{e.carriedFromSprintName}</TableCell>}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}

export default async function SprintDashboardPage({
  params,
}: {
  params: Promise<{ sprintId: string }>;
}) {
  const { sprintId } = await params;
  const id = Number(sprintId);
  if (!Number.isInteger(id)) {
    notFound();
  }

  let sprint;
  let breakdown;
  let ticketBreakdown;
  try {
    [sprint, breakdown, ticketBreakdown] = await Promise.all([
      getSprint(id),
      getSprintDeveloperBreakdown(id),
      getSprintTicketBreakdown(id),
    ]);
  } catch {
    notFound();
  }

  const totalWorkload = breakdown.reduce((sum, d) => sum + d.workloadPoints, 0);
  const totalDone = breakdown.reduce((sum, d) => sum + d.donePoints, 0);

  return (
    <main className="mx-auto max-w-5xl space-y-8 p-8">
      <div>
        <Link href="/dashboard" className="text-sm text-muted-foreground hover:underline">
          &larr; All sprints
        </Link>
        <h1 className="mt-2 text-2xl font-semibold">{sprint.name}</h1>
        <p className="text-sm text-muted-foreground">
          {new Date(sprint.startDate).toLocaleDateString()} &ndash; {new Date(sprint.endDate).toLocaleDateString()} &middot;{' '}
          {sprint.status}
        </p>
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Sprint workload (pts)</CardTitle>
          </CardHeader>
          <CardContent className="text-3xl font-bold">{totalWorkload}</CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Sprint done (pts)</CardTitle>
          </CardHeader>
          <CardContent className="text-3xl font-bold">
            {totalDone}
            <span className="ml-2 text-base font-normal text-muted-foreground">/ {totalWorkload}</span>
          </CardContent>
        </Card>
      </div>

      {breakdown.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Workload vs. done, per developer</CardTitle>
          </CardHeader>
          <CardContent>
            <WorkloadDoneChart
              data={breakdown.map((d) => ({ category: d.name, workload: d.workloadPoints, done: d.donePoints }))}
            />
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <CardTitle>By developer</CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Developer</TableHead>
                <TableHead>Workload (pts)</TableHead>
                <TableHead>Done (pts)</TableHead>
                <TableHead>Tickets</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {breakdown.length === 0 && (
                <TableRow>
                  <TableCell colSpan={4} className="text-center text-muted-foreground">
                    No tickets in this sprint yet.
                  </TableCell>
                </TableRow>
              )}
              {breakdown.map((d) => (
                <TableRow key={d.userId ?? 'unassigned'}>
                  <TableCell className="font-medium">{d.name}</TableCell>
                  <TableCell>{d.workloadPoints}</TableCell>
                  <TableCell>
                    {d.donePoints}/{d.workloadPoints}
                  </TableCell>
                  <TableCell className="text-muted-foreground">
                    {d.doneTickets}/{d.workloadTickets} tickets
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <TicketTable
        title="Current sprint tickets"
        entries={ticketBreakdown.current}
        showCarriedFrom={false}
        emptyMessage="No tickets currently planned for this sprint."
      />

      <TicketTable
        title="Carried over tickets"
        entries={ticketBreakdown.carriedOver}
        showCarriedFrom
        emptyMessage="No carried-over tickets in this sprint."
      />
    </main>
  );
}
