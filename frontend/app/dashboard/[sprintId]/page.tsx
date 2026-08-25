import Link from 'next/link';
import { notFound } from 'next/navigation';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { getSprint, getSprintDeveloperBreakdown } from '@/lib/api';

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
    </main>
  );
}
