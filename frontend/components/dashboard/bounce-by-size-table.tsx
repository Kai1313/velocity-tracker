import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import type { BounceBySize } from '@/lib/api';

export function BounceBySizeTable({ rows }: { rows: BounceBySize[] }) {
  const total = rows.reduce((sum, r) => sum + r.entries, 0);
  return (
    <div>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Story points</TableHead>
            <TableHead>Entries</TableHead>
            <TableHead>Entries bounced</TableHead>
            <TableHead>Total bounces</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.map((r) => (
            <TableRow key={r.band}>
              <TableCell className="font-medium">{r.band} SP</TableCell>
              <TableCell>{r.entries}</TableCell>
              <TableCell>{r.entries === 0 ? '—' : `${r.bouncedEntries} of ${r.entries}`}</TableCell>
              <TableCell>{r.entries === 0 ? '—' : r.totalBounces}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
      <div className="space-y-1 px-6 py-4 text-xs text-muted-foreground">
        <p>
          Counts every closed sprint that has bounce data ({total} sprint {total === 1 ? 'entry' : 'entries'} so far).
          Each row counts sprint entries, not tickets, so a ticket carried over twice appears three times.
        </p>
        <p>
          With only a few tracked sprints a band holds a handful of tickets — read these as hints about which ticket
          sizes to break down, not as conclusions.
        </p>
      </div>
    </div>
  );
}
