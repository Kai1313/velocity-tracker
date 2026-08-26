'use client';

import { useEffect, useState } from 'react';

import { Card, CardContent } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select } from '@/components/ui/select';
import { Badge } from '@/components/ui/badge';
import {
  Dialog,
  DialogTrigger,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog';
import { DeleteConfirmButton } from '@/components/admin/delete-confirm-button';
import {
  listSprintEntries,
  createSprintEntry,
  updateSprintEntry,
  deleteSprintEntry,
  listTickets,
  listSprints,
  type SprintEntry,
  type TicketDetail,
  type Sprint,
  type EntryStatus,
} from '@/lib/api';

const NONE = 'none';

function EntryFormDialog({
  entry,
  entries,
  tickets,
  sprints,
  onSaved,
}: {
  entry?: SprintEntry;
  entries: SprintEntry[];
  tickets: TicketDetail[];
  sprints: Sprint[];
  onSaved: (entry: SprintEntry) => void;
}) {
  const isEdit = entry !== undefined;
  const [open, setOpen] = useState(false);
  const [ticketId, setTicketId] = useState(entry?.ticketId ?? tickets[0]?.id ?? 0);
  const [sprintId, setSprintId] = useState(entry?.sprintId ?? sprints[0]?.id ?? 0);
  const [status, setStatus] = useState<EntryStatus>(entry?.status ?? 'NotDone');
  const [addedAfterStart, setAddedAfterStart] = useState(entry?.addedAfterSprintStart ?? false);
  const [carriedFrom, setCarriedFrom] = useState<string>(entry?.carriedFrom != null ? String(entry.carriedFrom) : NONE);
  const [carriedFromTouched, setCarriedFromTouched] = useState(false);
  const [points, setPoints] = useState(entry?.pointsAtEntry ?? tickets[0]?.storyPoints ?? 1);
  const [pending, setPending] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const parentSprint = sprints.find((s) => s.id === (entry?.sprintId ?? sprintId));
  const locked = isEdit && parentSprint?.status === 'Closed';

  useEffect(() => {
    if (open) {
      setTicketId(entry?.ticketId ?? tickets[0]?.id ?? 0);
      setSprintId(entry?.sprintId ?? sprints[0]?.id ?? 0);
      setStatus(entry?.status ?? 'NotDone');
      setAddedAfterStart(entry?.addedAfterSprintStart ?? false);
      setCarriedFrom(entry?.carriedFrom != null ? String(entry.carriedFrom) : NONE);
      setCarriedFromTouched(false);
      setPoints(entry?.pointsAtEntry ?? tickets[0]?.storyPoints ?? 1);
      setError(null);
    }
  }, [open, entry, tickets, sprints]);

  // Only entries for the same ticket, still NotDone, from a sprint that's
  // already Closed are legitimate carry-over sources — anything else is
  // either a different ticket's history or not actually "carried" yet.
  const carryCandidates = entries.filter(
    (e) =>
      e.id !== entry?.id &&
      e.ticketId === ticketId &&
      e.status === 'NotDone' &&
      sprints.find((s) => s.id === e.sprintId)?.status === 'Closed',
  );

  useEffect(() => {
    if (!open || carriedFromTouched || carriedFrom !== NONE) return;
    if (carryCandidates.length === 1) {
      setCarriedFrom(String(carryCandidates[0].id));
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, ticketId, carriedFromTouched, carriedFrom, carryCandidates.length]);

  function ticketLabel(id: number) {
    return tickets.find((t) => t.id === id)?.title ?? `Ticket #${id}`;
  }
  function sprintLabel(id: number) {
    return sprints.find((s) => s.id === id)?.name ?? `Sprint #${id}`;
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setPending(true);
    setError(null);
    const carriedFromValue = carriedFrom === NONE ? null : Number(carriedFrom);
    try {
      const saved =
        isEdit && entry
          ? await updateSprintEntry(entry.id, {
              status,
              addedAfterSprintStart: addedAfterStart,
              carriedFrom: carriedFromValue,
              pointsAtEntry: points,
            })
          : await createSprintEntry({
              ticketId,
              sprintId,
              status,
              addedAfterSprintStart: addedAfterStart,
              carriedFrom: carriedFromValue,
              pointsAtEntry: points,
            });
      onSaved(saved);
      setOpen(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Save failed');
    } finally {
      setPending(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        {isEdit ? (
          <Button variant="ghost" size="sm" disabled={locked} title={locked ? 'Sprint is closed — entry is locked' : undefined}>
            Edit
          </Button>
        ) : (
          <Button disabled={tickets.length === 0 || sprints.length === 0}>Add entry</Button>
        )}
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{isEdit ? 'Edit sprint entry' : 'Add sprint entry'}</DialogTitle>
          {isEdit && entry && (
            <DialogDescription>
              {ticketLabel(entry.ticketId)} in {sprintLabel(entry.sprintId)}. Ticket and sprint can&apos;t be changed
              after creation.
            </DialogDescription>
          )}
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          {!isEdit && (
            <>
              <div className="space-y-1.5">
                <Label htmlFor="entry-ticket">Ticket</Label>
                <Select
                  id="entry-ticket"
                  value={ticketId}
                  onChange={(e) => {
                    const id = Number(e.target.value);
                    setTicketId(id);
                    setPoints(tickets.find((t) => t.id === id)?.storyPoints ?? points);
                  }}
                  required
                >
                  {tickets.map((t) => (
                    <option key={t.id} value={t.id}>
                      {t.title}
                    </option>
                  ))}
                </Select>
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="entry-sprint">Sprint</Label>
                <Select id="entry-sprint" value={sprintId} onChange={(e) => setSprintId(Number(e.target.value))} required>
                  {sprints.map((s) => (
                    <option key={s.id} value={s.id}>
                      {s.name} ({s.status})
                    </option>
                  ))}
                </Select>
              </div>
            </>
          )}
          <div className="space-y-1.5">
            <Label htmlFor="entry-status">Status</Label>
            <Select id="entry-status" value={status} onChange={(e) => setStatus(e.target.value as EntryStatus)} disabled={locked}>
              <option value="Done">Done</option>
              <option value="NotDone">NotDone</option>
              <option value="Cancelled">Cancelled</option>
            </Select>
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="entry-points">Points at entry</Label>
            <Input
              id="entry-points"
              type="number"
              min={1}
              value={points}
              onChange={(e) => setPoints(Number(e.target.value))}
              disabled={locked}
              required
            />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="entry-carried-from">Carried from</Label>
            <Select
              id="entry-carried-from"
              value={carriedFrom}
              onChange={(e) => {
                setCarriedFrom(e.target.value);
                setCarriedFromTouched(true);
              }}
              disabled={locked}
            >
              <option value={NONE}>None</option>
              {carryCandidates.map((c) => (
                <option key={c.id} value={c.id}>
                  {ticketLabel(c.ticketId)} @ {sprintLabel(c.sprintId)} ({c.status})
                </option>
              ))}
            </Select>
          </div>
          <label className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={addedAfterStart}
              onChange={(e) => setAddedAfterStart(e.target.checked)}
              disabled={locked}
              className="h-4 w-4 rounded border-border"
            />
            Added after sprint start
          </label>
          {locked && (
            <p className="text-sm text-muted-foreground">
              This sprint is closed — its entries are locked history and can&apos;t be edited.
            </p>
          )}
          {error && <p className="text-sm text-red-600 dark:text-red-400">{error}</p>}
          <DialogFooter>
            <Button type="submit" disabled={pending || locked}>
              {pending ? 'Saving…' : 'Save'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

export default function SprintEntriesPage() {
  const [entries, setEntries] = useState<SprintEntry[] | null>(null);
  const [tickets, setTickets] = useState<TicketDetail[]>([]);
  const [sprints, setSprints] = useState<Sprint[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    Promise.all([listSprintEntries(), listTickets(), listSprints()])
      .then(([e, t, s]) => {
        setEntries(e);
        setTickets(t);
        setSprints(s);
      })
      .catch((err) => setError(err instanceof Error ? err.message : 'Failed to load sprint entries'));
  }, []);

  function upsert(entry: SprintEntry) {
    setEntries((prev) => {
      if (!prev) return [entry];
      const exists = prev.some((e) => e.id === entry.id);
      return exists ? prev.map((e) => (e.id === entry.id ? entry : e)) : [...prev, entry];
    });
  }

  async function handleDelete(id: number) {
    await deleteSprintEntry(id);
    setEntries((prev) => prev?.filter((e) => e.id !== id) ?? null);
  }

  function ticketTitle(id: number) {
    return tickets.find((t) => t.id === id)?.title ?? `#${id}`;
  }
  function sprintName(id: number) {
    return sprints.find((s) => s.id === id)?.name ?? `#${id}`;
  }
  function sprintClosed(id: number) {
    return sprints.find((s) => s.id === id)?.status === 'Closed';
  }

  const statusVariant: Record<EntryStatus, 'success' | 'warning' | 'secondary'> = {
    Done: 'success',
    NotDone: 'warning',
    Cancelled: 'secondary',
  };

  // A NotDone entry whose sprint has closed should have gotten a carry-over
  // entry in a later sprint by now. One still missing means either the
  // carry-over was never created, or was created without linking carriedFrom
  // back to this entry — both silently break carry-over/planning metrics.
  const orphanedCarryOvers = (entries ?? []).filter((e) => {
    if (e.status !== 'NotDone') return false;
    const sprint = sprints.find((s) => s.id === e.sprintId);
    if (!sprint || sprint.status !== 'Closed') return false;
    return !(entries ?? []).some((other) => {
      if (other.ticketId !== e.ticketId) return false;
      const otherSprint = sprints.find((s) => s.id === other.sprintId);
      return !!otherSprint && otherSprint.startDate > sprint.startDate;
    });
  });

  return (
    <div className="mx-auto max-w-5xl space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Sprint Entries</h1>
          <p className="text-sm text-muted-foreground">
            {tickets.length === 0 || sprints.length === 0
              ? 'Add a ticket and a sprint first — an entry ties one to the other.'
              : "One row per sprint a ticket appears in. Closing a sprint locks its entries against further edits."}
          </p>
        </div>
        <EntryFormDialog entries={entries ?? []} tickets={tickets} sprints={sprints} onSaved={upsert} />
      </div>

      {error && <p className="text-sm text-red-600 dark:text-red-400">{error}</p>}

      {orphanedCarryOvers.length > 0 && (
        <div className="rounded-lg border border-amber-200 bg-amber-50 p-4 text-sm text-amber-900 dark:border-amber-900 dark:bg-amber-950 dark:text-amber-300">
          <p className="font-medium">
            {orphanedCarryOvers.length} ticket{orphanedCarryOvers.length > 1 ? 's' : ''} still NotDone in a closed
            sprint with no carry-over entry yet:
          </p>
          <ul className="mt-1.5 list-disc space-y-0.5 pl-5">
            {orphanedCarryOvers.map((e) => (
              <li key={e.id}>
                {ticketTitle(e.ticketId)} — {sprintName(e.sprintId)}
              </li>
            ))}
          </ul>
        </div>
      )}

      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Ticket</TableHead>
                <TableHead>Sprint</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Points</TableHead>
                <TableHead />
                <TableHead className="w-1" />
              </TableRow>
            </TableHeader>
            <TableBody>
              {entries?.length === 0 && (
                <TableRow>
                  <TableCell colSpan={6} className="text-center text-muted-foreground">
                    No sprint entries yet.
                  </TableCell>
                </TableRow>
              )}
              {entries?.map((e) => (
                <TableRow key={e.id}>
                  <TableCell className="font-medium">{ticketTitle(e.ticketId)}</TableCell>
                  <TableCell className="text-muted-foreground">{sprintName(e.sprintId)}</TableCell>
                  <TableCell>
                    <Badge variant={statusVariant[e.status]}>{e.status}</Badge>
                  </TableCell>
                  <TableCell>{e.pointsAtEntry}</TableCell>
                  <TableCell>
                    {e.addedAfterSprintStart && <Badge variant="secondary">Late add</Badge>}
                    {sprintClosed(e.sprintId) && <Badge variant="outline">Locked</Badge>}
                  </TableCell>
                  <TableCell className="flex justify-end gap-1">
                    <EntryFormDialog entry={e} entries={entries ?? []} tickets={tickets} sprints={sprints} onSaved={upsert} />
                    <DeleteConfirmButton
                      entityLabel={`${ticketTitle(e.ticketId)} @ ${sprintName(e.sprintId)}`}
                      onDelete={() => handleDelete(e.id)}
                    />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}
