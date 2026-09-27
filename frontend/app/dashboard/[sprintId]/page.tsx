import Link from 'next/link';
import { notFound } from 'next/navigation';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { WorkloadDoneChart } from '@/components/dashboard/workload-done-chart';
import { WorkloadDoneTable } from '@/components/dashboard/workload-done-table';
import { ApiError, getSprint, getSprintDeveloperBreakdown } from '@/lib/api';

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
  try {
    [sprint, breakdown] = await Promise.all([getSprint(id), getSprintDeveloperBreakdown(id)]);
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) {
      notFound();
    }
    throw err;
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
        <Link href={`/dashboard/${id}/entries`} className="block">
          <Card className="transition-colors hover:bg-accent/50">
            <CardHeader>
              <CardTitle>Sprint workload (pts)</CardTitle>
            </CardHeader>
            <CardContent className="text-3xl font-bold">{totalWorkload}</CardContent>
          </Card>
        </Link>
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

      <WorkloadDoneTable
        title="By developer"
        rows={breakdown.map((d) => ({
          key: d.userId ?? 'unassigned',
          label: d.name,
          workloadPoints: d.workloadPoints,
          donePoints: d.donePoints,
          workloadTickets: d.workloadTickets,
          doneTickets: d.doneTickets,
        }))}
      />
    </main>
  );
}
