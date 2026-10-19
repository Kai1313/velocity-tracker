import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import type { StuckTicketList } from '@/lib/api';

export function StuckTicketsCard({ list }: { list: StuckTicketList }) {
  const rows = list.tickets;
  return (
    <Card>
      <CardHeader className="space-y-1.5">
        <CardTitle>Stuck tickets</CardTitle>
        <p className="text-sm font-normal text-muted-foreground">
          Not-done tickets in an open sprint that QA has sent back {list.minBounces} or more times in total, counting earlier
          sprints. A carried-over ticket starts the new sprint at 0 bounces but keeps its history here.
        </p>
      </CardHeader>
      <CardContent className="overflow-x-auto p-0 pb-6">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Sprint</TableHead>
              <TableHead>Ticket</TableHead>
              <TableHead>Assignee</TableHead>
              <TableHead>SP</TableHead>
              <TableHead>Bounces</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {rows.length === 0 && (
              <TableRow>
                <TableCell colSpan={5} className="text-center text-muted-foreground">
                  No stuck tickets.
                </TableCell>
              </TableRow>
            )}
            {rows.map((r) => (
              <TableRow key={r.entryId}>
                <TableCell className="text-muted-foreground">{r.sprintName}</TableCell>
                <TableCell className="font-medium">
                  {r.projectName} - {r.ticketTitle}
                </TableCell>
                <TableCell className="text-muted-foreground">{r.assigneeName}</TableCell>
                <TableCell>{r.pointsAtEntry}</TableCell>
                <TableCell>
                  {r.totalBounces}
                  <span className="text-muted-foreground">
                    {' '}
                    ({r.bounceCount} this sprint, {r.earlierBounces} earlier)
                  </span>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}
