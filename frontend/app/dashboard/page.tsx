import Link from 'next/link';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { getSprintSummaries, type SprintSummary } from '@/lib/api';

function completionBadge(summary: SprintSummary) {
  if (summary.workloadPoints === 0) {
    return <Badge variant="secondary">No workload</Badge>;
  }
  const pct = Math.round((summary.donePoints / summary.workloadPoints) * 100);
  const variant = pct >= 100 ? 'success' : pct > 0 ? 'warning' : 'secondary';
  return <Badge variant={variant}>{pct}%</Badge>;
}

export default async function DashboardPage() {
  const summaries = await getSprintSummaries();

  const totalWorkload = summaries.reduce((sum, s) => sum + s.workloadPoints, 0);
  const totalDone = summaries.reduce((sum, s) => sum + s.donePoints, 0);

  return (
    <main className="mx-auto max-w-5xl space-y-8 p-8">
      <div>
        <h1 className="text-2xl font-semibold">Sprint Dashboard</h1>
        <p className="text-sm text-muted-foreground">
          Workload and completed story points, per sprint. Click a sprint for its per-developer breakdown.
        </p>
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle>Sprints tracked</CardTitle>
          </CardHeader>
          <CardContent className="text-3xl font-bold">{summaries.length}</CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Total workload (pts)</CardTitle>
          </CardHeader>
          <CardContent className="text-3xl font-bold">{totalWorkload}</CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Total done (pts)</CardTitle>
          </CardHeader>
          <CardContent className="text-3xl font-bold">
            {totalDone}
            <span className="ml-2 text-base font-normal text-muted-foreground">/ {totalWorkload}</span>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Sprint</TableHead>
                <TableHead>Workload (pts)</TableHead>
                <TableHead>Done (pts)</TableHead>
                <TableHead>Tickets</TableHead>
                <TableHead>Progress</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {summaries.length === 0 && (
                <TableRow>
                  <TableCell colSpan={5} className="text-center text-muted-foreground">
                    No sprints yet.
                  </TableCell>
                </TableRow>
              )}
              {summaries.map((s) => (
                <TableRow key={s.sprintId}>
                  <TableCell className="font-medium">
                    <Link href={`/dashboard/${s.sprintId}`} className="hover:underline">
                      {s.sprintName}
                    </Link>
                  </TableCell>
                  <TableCell>{s.workloadPoints}</TableCell>
                  <TableCell>
                    {s.donePoints}/{s.workloadPoints}
                  </TableCell>
                  <TableCell className="text-muted-foreground">
                    {s.doneTickets}/{s.workloadTickets} tickets
                  </TableCell>
                  <TableCell>{completionBadge(s)}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </main>
  );
}
