import Link from 'next/link';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { SprintRetrospectiveTable } from '@/components/dashboard/sprint-retrospective-table';
import { getSprintRetrospective } from '@/lib/api';

const WINDOW_OPTIONS = [3, 6, 9];
const DEFAULT_LIMIT = 3;

export default async function RetrospectivePage({
  searchParams,
}: {
  searchParams: Promise<{ limit?: string }>;
}) {
  const { limit: limitParam } = await searchParams;
  const parsed = limitParam ? Number(limitParam) : DEFAULT_LIMIT;
  const limit = Number.isInteger(parsed) && parsed > 0 ? parsed : DEFAULT_LIMIT;

  const rows = await getSprintRetrospective(limit);

  return (
    <main className="mx-auto max-w-5xl space-y-8 p-8">
      <div>
        <Link href="/dashboard" className="text-sm text-muted-foreground hover:underline">
          &larr; Sprint dashboard
        </Link>
        <h1 className="mt-2 text-2xl font-semibold">Sprint Retrospective</h1>
        <p className="text-sm text-muted-foreground">
          Commitment reliability for the team&apos;s last {limit} closed sprints — how much of each sprint&apos;s
          original plan got done, and how much got added after it started.
        </p>
      </div>

      <div className="flex items-center gap-2 text-sm">
        <span className="text-muted-foreground">Show last:</span>
        {WINDOW_OPTIONS.map((n) => (
          <Link
            key={n}
            href={`/dashboard/retrospective?limit=${n}`}
            className={
              n === limit
                ? 'rounded-md bg-primary px-2.5 py-1 text-primary-foreground'
                : 'rounded-md px-2.5 py-1 text-muted-foreground hover:bg-muted'
            }
          >
            {n} sprints
          </Link>
        ))}
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Planning Accuracy &amp; Late-Add Rate</CardTitle>
        </CardHeader>
        <CardContent className="overflow-x-auto p-0 pb-6">
          <SprintRetrospectiveTable rows={rows} />
        </CardContent>
      </Card>
    </main>
  );
}
