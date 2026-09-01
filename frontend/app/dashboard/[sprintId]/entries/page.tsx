import Link from 'next/link';
import { notFound } from 'next/navigation';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { getSprint, getSprintTicketBreakdown, type SprintEntryDetail } from '@/lib/api';

function workload(entries: SprintEntryDetail[]) {
  const counted = entries.filter((e) => e.status !== 'Cancelled');
  return {
    points: counted.reduce((sum, e) => sum + e.pointsAtEntry, 0),
    tickets: counted.length,
  };
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
    </main>
  );
}
