import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';

export type WorkloadDoneRow = {
  key: string | number;
  label: string;
  workloadPoints: number;
  donePoints: number;
  workloadTickets: number;
  doneTickets: number;
};

// The "Developer / Workload / Done / Tickets" table shape repeats across the
// sprint dashboard and the workload-breakdown page (current + carried-over).
export function WorkloadDoneTable({
  title,
  rows,
  emptyMessage = 'No tickets in this sprint yet.',
}: {
  title: string;
  rows: WorkloadDoneRow[];
  emptyMessage?: string;
}) {
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
            {rows.length === 0 && (
              <TableRow>
                <TableCell colSpan={4} className="text-center text-muted-foreground">
                  {emptyMessage}
                </TableCell>
              </TableRow>
            )}
            {rows.map((r) => (
              <TableRow key={r.key}>
                <TableCell className="font-medium">{r.label}</TableCell>
                <TableCell>{r.workloadPoints}</TableCell>
                <TableCell>
                  {r.donePoints}/{r.workloadPoints}
                </TableCell>
                <TableCell className="text-muted-foreground">
                  {r.doneTickets}/{r.workloadTickets} tickets
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}
