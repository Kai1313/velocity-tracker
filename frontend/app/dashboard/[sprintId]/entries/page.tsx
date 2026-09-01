import Link from 'next/link';
import { notFound } from 'next/navigation';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { WorkloadDoneChart, type WorkloadDoneDatum } from '@/components/dashboard/workload-done-chart';
import { getSprint, getSprintTicketBreakdown, type SprintEntryDetail } from '@/lib/api';

function workload(entries: SprintEntryDetail[]) {
  const counted = entries.filter((e) => e.status !== 'Cancelled');
  return {
    points: counted.reduce((sum, e) => sum + e.pointsAtEntry, 0),
    tickets: counted.length,
  };
}

// Per-developer Workload/Done pair within one subset (current or carried-over),
// matching the aggregation the developer breakdown chart on /dashboard/{sprintId}
// uses for the whole sprint. Cancelled entries are excluded, consistent with
// "workload" everywhere else in this app.
function developerBreakdown(entries: SprintEntryDetail[]): WorkloadDoneDatum[] {
  const byDeveloper = new Map<string, { workload: number; done: number }>();
  for (const e of entries) {
    if (e.status === 'Cancelled') continue;
    const bucket = byDeveloper.get(e.assigneeName) ?? { workload: 0, done: 0 };
    bucket.workload += e.pointsAtEntry;
    if (e.status === 'Done') bucket.done += e.pointsAtEntry;
    byDeveloper.set(e.assigneeName, bucket);
  }
  return Array.from(byDeveloper, ([category, points]) => ({ category, ...points })).sort((a, b) =>
    a.category.localeCompare(b.category, undefined, { sensitivity: 'base' }),
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
  const currentBreakdown = developerBreakdown(ticketBreakdown.current);
  const carriedOverBreakdown = developerBreakdown(ticketBreakdown.carriedOver);

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
        {currentBreakdown.length > 0 && (
          <Card>
            <CardHeader>
              <CardTitle>Current workload vs. done, per developer</CardTitle>
            </CardHeader>
            <CardContent>
              <WorkloadDoneChart data={currentBreakdown} />
            </CardContent>
          </Card>
        )}
        {carriedOverBreakdown.length > 0 && (
          <Card>
            <CardHeader>
              <CardTitle>Carry-over workload vs. done, per developer</CardTitle>
            </CardHeader>
            <CardContent>
              <WorkloadDoneChart data={carriedOverBreakdown} />
            </CardContent>
          </Card>
        )}
      </div>
    </main>
  );
}
