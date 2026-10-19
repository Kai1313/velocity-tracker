'use client';

import { useState } from 'react';

import { Button } from '@/components/ui/button';
import { setSprintEntryBounceCount, type SprintEntry } from '@/lib/api';

// Shows an entry's bounce count with its earlier-sprint history, and, while
// the sprint is still editable, +/- buttons that save immediately. The
// buttons send the new absolute value, never a delta. Cancelled entries are
// excluded from every bounce metric, so they get no controls.
export function BounceStepper({
  entry,
  earlier,
  locked,
  onSaved,
}: {
  entry: SprintEntry;
  earlier: number;
  locked: boolean;
  onSaved: (entry: SprintEntry) => void;
}) {
  const [pending, setPending] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const count = entry.bounceCount;
  const earlierNote = earlier > 0 ? <span className="text-muted-foreground"> ({earlier} earlier)</span> : null;

  if (entry.status === 'Cancelled') {
    return <span className="text-muted-foreground">—</span>;
  }
  if (locked) {
    return (
      <span data-testid="bounce-count">
        {count ?? <span className="text-muted-foreground">—</span>}
        {earlierNote}
      </span>
    );
  }

  async function change(next: number) {
    setPending(true);
    setError(null);
    try {
      onSaved(await setSprintEntryBounceCount(entry.id, next));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save bounce count');
    } finally {
      setPending(false);
    }
  }

  return (
    <div>
      <div className="flex items-center gap-1">
        <Button
          type="button"
          variant="outline"
          size="sm"
          aria-label="Decrease bounce count"
          disabled={pending || !count}
          onClick={() => count && change(count - 1)}
        >
          −
        </Button>
        <span data-testid="bounce-count" className="min-w-6 text-center">
          {count ?? <span className="text-muted-foreground">—</span>}
        </span>
        <Button
          type="button"
          variant="outline"
          size="sm"
          aria-label="Increase bounce count"
          disabled={pending}
          onClick={() => change((count ?? 0) + 1)}
        >
          +
        </Button>
        {earlierNote}
      </div>
      {error && (
        <p role="alert" className="mt-1 text-xs text-destructive">
          {error}
        </p>
      )}
    </div>
  );
}
