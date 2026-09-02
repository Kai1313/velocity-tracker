import Link from 'next/link';
import { notFound } from 'next/navigation';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { WorkloadDoneChart } from '@/components/dashboard/workload-done-chart';
import { getSprint, getSprintTicketBreakdown, type SprintEntryDetail } from '@/lib/api';

function workload(entries: SprintEntryDetail[]) {
  const counted = entries.filter((e) => e.status !== 'Cancelled');
  return {
    points: counted.reduce((sum, e) => sum + e.pointsAtEntry, 0),
    tickets: counted.length,
  };
}

type DeveloperStats = {
  category: string;
  workloadPoints: number;
  donePoints: number;
  workloadTickets: number;
  doneTickets: number;
};

// Per-developer workload/done stats within one subset (current or carried-over),
// matching the aggregation the "By developer" chart and table on
// /dashboard/{sprintId} use for the whole sprint. Cancelled entries are
// excluded, consistent with "workload" everywhere else in this app.
function developerStats(entries: SprintEntryDetail[]): DeveloperStats[] {
  const byDeveloper = new Map<string, DeveloperStats>();
  for (const e of entries) {
    if (e.status === 'Cancelled') continue;
    const bucket = byDeveloper.get(e.assigneeName) ?? {
      category: e.assigneeName,
      workloadPoints: 0,
      donePoints: 0,
      workloadTickets: 0,
      doneTickets: 0,
    };
    bucket.workloadPoints += e.pointsAtEntry;
    bucket.workloadTickets += 1;
    if (e.status === 'Done') {
      bucket.donePoints += e.pointsAtEntry;
      bucket.doneTickets += 1;
    }
    byDeveloper.set(e.assigneeName, bucket);
  }
  return Array.from(byDeveloper.values()).sort((a, b) =>
    a.category.localeCompare(b.category, undefined, { sensitivity: 'base' }),
  );
}

function DeveloperTable({ title, stats }: { title: string; stats: DeveloperStats[] }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
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
            {stats.map((s) => (
              <TableRow key={s.category}>
                <TableCell className="font-medium">{s.category}</TableCell>
                <TableCell>{s.workloadPoints}</TableCell>
                <TableCell>
                  {s.donePoints}/{s.workloadPoints}
                </TableCell>
                <TableCell className="text-muted-foreground">
                  {s.doneTickets}/{s.workloadTickets} tickets
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}

export default async function SprintEntriesPage({
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
  let ticketBreakdown;
  try {
    [sprint, ticketBreakdown] = await Promise.all([getSprint(id), getSprintTicketBreakdown(id)]);
  } catch {
    notFound();
  }

  const current = workload(ticketBreakdown.current);
  const carriedOver = workload(ticketBreakdown.carriedOver);
  const currentStats = developerStats(ticketBreakdown.current);
  const carriedOverStats = developerStats(ticketBreakdown.carriedOver);

  return (
    <main className="mx-auto max-w-5xl space-y-8 p-8">
      <div>
        <Link href={`/dashboard/${id}`} className="text-sm text-muted-foreground hover:underline">
          &larr; {sprint.name}
        </Link>
        <h1 className="mt-2 text-2xl font-semibold">{sprint.name} &middot; Workload breakdown</h1>
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Current sprint workload</CardTitle>
          </CardHeader>
          <CardContent className="text-3xl font-bold">
            {current.points} <span className="text-base font-normal text-muted-foreground">pts</span>
            <p className="mt-1 text-sm font-normal text-muted-foreground">
              {current.tickets} ticket{current.tickets === 1 ? '' : 's'}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Carry-over sprint workload</CardTitle>
          </CardHeader>
          <CardContent className="text-3xl font-bold">
            {carriedOver.points} <span className="text-base font-normal text-muted-foreground">pts</span>
            <p className="mt-1 text-sm font-normal text-muted-foreground">
              {carriedOver.tickets} ticket{carriedOver.tickets === 1 ? '' : 's'}
            </p>
          </CardContent>
        </Card>
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        {currentStats.length > 0 && (
          <Card>
            <CardHeader>
              <CardTitle>Current workload vs. done, per developer</CardTitle>
            </CardHeader>
            <CardContent>
              <WorkloadDoneChart
                data={currentStats.map((s) => ({ category: s.category, workload: s.workloadPoints, done: s.donePoints }))}
              />
            </CardContent>
          </Card>
        )}
        {carriedOverStats.length > 0 && (
          <Card>
            <CardHeader>
              <CardTitle>Carry-over workload vs. done, per developer</CardTitle>
            </CardHeader>
            <CardContent>
              <WorkloadDoneChart
                data={carriedOverStats.map((s) => ({ category: s.category, workload: s.workloadPoints, done: s.donePoints }))}
              />
            </CardContent>
          </Card>
        )}
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        {currentStats.length > 0 && <DeveloperTable title="Current — by developer" stats={currentStats} />}
        {carriedOverStats.length > 0 && <DeveloperTable title="Carry-over — by developer" stats={carriedOverStats} />}
      </div>
    </main>
  );
}
