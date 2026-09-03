'use client';

import { Suspense, useEffect, useRef, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';

import { Card, CardContent } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select } from '@/components/ui/select';
import { Combobox } from '@/components/ui/combobox';
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
  listProjects,
  type SprintEntry,
  type TicketDetail,
  type Sprint,
  type Project,
  type EntryStatus,
  type SprintEntryFilter,
} from '@/lib/api';

const NONE = 'none';
const ALL = 'all';

// Ticket titles (e.g. "2026_07_FEATURE_110") aren't unique across projects,
// so every ticket reference on this page is prefixed with its project name
// to stay unambiguous once two projects reuse the same ticket number.
function ticketLabel(id: number, tickets: TicketDetail[], projects: Project[]) {
  const t = tickets.find((t) => t.id === id);
  if (!t) return `Ticket #${id}`;
  const project = projects.find((p) => p.id === t.projectId)?.name ?? `Project #${t.projectId}`;
  return `${project} - ${t.title}`;
}

function EntryFormDialog({
  entry,
  entries,
  tickets,
  sprints,
  projects,
  onSaved,
}: {
  entry?: SprintEntry;
  entries: SprintEntry[];
  tickets: TicketDetail[];
  sprints: Sprint[];
  projects: Project[];
  onSaved: (entry: SprintEntry) => void;
}) {
  const isEdit = entry !== undefined;
  const [open, setOpen] = useState(false);
  const [ticketId, setTicketId] = useState<number | null>(entry?.ticketId ?? null);
  const [sprintId, setSprintId] = useState(entry?.sprintId ?? sprints[0]?.id ?? 0);
  const [status, setStatus] = useState<EntryStatus>(entry?.status ?? 'NotDone');
  const [addedAfterStart, setAddedAfterStart] = useState(entry?.addedAfterSprintStart ?? false);
  const [carriedFrom, setCarriedFrom] = useState<string>(entry?.carriedFrom != null ? String(entry.carriedFrom) : NONE);
  const [carriedFromTouched, setCarriedFromTouched] = useState(false);
  const [points, setPoints] = useState(entry?.pointsAtEntry ?? 1);
  const [pending, setPending] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const parentSprint = sprints.find((s) => s.id === (entry?.sprintId ?? sprintId));
  const locked = isEdit && parentSprint?.status === 'Closed';

  useEffect(() => {
    if (open) {
      setTicketId(entry?.ticketId ?? null);
      setSprintId(entry?.sprintId ?? sprints[0]?.id ?? 0);
      setStatus(entry?.status ?? 'NotDone');
      setAddedAfterStart(entry?.addedAfterSprintStart ?? false);
      setCarriedFrom(entry?.carriedFrom != null ? String(entry.carriedFrom) : NONE);
      setCarriedFromTouched(false);
      setPoints(entry?.pointsAtEntry ?? 1);
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
    if (!open) return;
    // Switching the Ticket dropdown (create mode) leaves a stale carriedFrom
    // pointing at the previous ticket's entry — it won't be in this ticket's
    // candidate list, so drop it and let the fill-in below reconsider.
    const stillValid = carriedFrom === NONE || carryCandidates.some((c) => String(c.id) === carriedFrom);
    if (!stillValid) {
      setCarriedFrom(NONE);
      setCarriedFromTouched(false);
      return;
    }
    if (carriedFromTouched || carriedFrom !== NONE) return;
    if (carryCandidates.length === 1) {
      setCarriedFrom(String(carryCandidates[0].id));
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, ticketId, carriedFromTouched, carriedFrom, carryCandidates.length]);

  function sprintLabel(id: number) {
    return sprints.find((s) => s.id === id)?.name ?? `Sprint #${id}`;
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!isEdit && ticketId === null) return;
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
              ticketId: ticketId as number,
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
              {ticketLabel(entry.ticketId, tickets, projects)} in {sprintLabel(entry.sprintId)}. Ticket and sprint
              can&apos;t be changed after creation.
            </DialogDescription>
          )}
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          {!isEdit && (
            <>
              <div className="space-y-1.5">
                <Label htmlFor="entry-ticket">Ticket</Label>
                <Combobox
                  id="entry-ticket"
                  value={ticketId != null ? String(ticketId) : null}
                  onChange={(v) => {
                    const id = Number(v);
                    setTicketId(id);
                    setPoints(tickets.find((t) => t.id === id)?.storyPoints ?? points);
                  }}
                  options={tickets.map((t) => ({ value: String(t.id), label: ticketLabel(t.id, tickets, projects) }))}
                  placeholder="Select a ticket"
                  searchPlaceholder="Search tickets…"
                  emptyMessage="No tickets found."
                />
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
                  {ticketLabel(c.ticketId, tickets, projects)} @ {sprintLabel(c.sprintId)} ({c.status})
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
            <Button type="submit" disabled={pending || locked || (!isEdit && ticketId === null)}>
              {pending ? 'Saving…' : 'Save'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

function SprintEntriesPageInner() {
  const router = useRouter();
  const searchParams = useSearchParams();

  // allEntries is unfiltered — the orphaned-carry-over check and the create/edit
  // dialog's "Carried from" candidates both need to see a ticket's entries across
  // every sprint, not just whichever one the table is currently filtered to.
  const [allEntries, setAllEntries] = useState<SprintEntry[] | null>(null);
  const [filteredEntries, setFilteredEntries] = useState<SprintEntry[] | null>(null);
  const [tickets, setTickets] = useState<TicketDetail[]>([]);
  const [sprints, setSprints] = useState<Sprint[]>([]);
  const [projects, setProjects] = useState<Project[]>([]);
  const [error, setError] = useState<string | null>(null);

  const sprintIdParam = searchParams.get('sprintId');
  const projectIdParam = searchParams.get('projectId');
  const statusParam = searchParams.get('status');
  const carriedOverParam = searchParams.get('carriedOver') === 'true';
  const searchParam = searchParams.get('search') ?? '';

  const [searchInput, setSearchInput] = useState(searchParam);

  useEffect(() => {
    setSearchInput(searchParam);
  }, [searchParam]);

  useEffect(() => {
    Promise.all([listSprintEntries(), listTickets(), listSprints(), listProjects()])
      .then(([e, t, s, p]) => {
        setAllEntries(e);
        setTickets(t);
        setSprints(s);
        setProjects(p);
      })
      .catch((err) => setError(err instanceof Error ? err.message : 'Failed to load sprint entries'));
  }, []);

  // Land on the currently-open sprint by default, since "entries for a specific
  // sprint" is the most common lookup — but only on the very first load with
  // nothing in the URL yet. This must apply once, not every time sprintId
  // becomes empty, or it would fight "Clear filters" and re-add itself.
  const appliedDefaultSprint = useRef(false);
  useEffect(() => {
    if (appliedDefaultSprint.current || sprints.length === 0) return;
    appliedDefaultSprint.current = true;
    if (sprintIdParam !== null) return;
    const open = sprints.find((s) => s.status === 'Open');
    if (!open) return;
    const params = new URLSearchParams(searchParams.toString());
    params.set('sprintId', String(open.id));
    router.replace(`?${params.toString()}`);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [sprints, sprintIdParam]);

  function currentFilter(): SprintEntryFilter {
    return {
      sprintId: sprintIdParam ? Number(sprintIdParam) : undefined,
      projectId: projectIdParam ? Number(projectIdParam) : undefined,
      status: (statusParam as EntryStatus) || undefined,
      carriedOver: carriedOverParam || undefined,
      search: searchParam || undefined,
    };
  }

  function refetchFiltered() {
    listSprintEntries(currentFilter())
      .then(setFilteredEntries)
      .catch((err) => setError(err instanceof Error ? err.message : 'Failed to load sprint entries'));
  }

  // eslint-disable-next-line react-hooks/exhaustive-deps
  useEffect(refetchFiltered, [sprintIdParam, projectIdParam, statusParam, carriedOverParam, searchParam]);

  function setParam(key: string, value: string | null) {
    const params = new URLSearchParams(searchParams.toString());
    if (value === null || value === '') {
      params.delete(key);
    } else {
      params.set(key, value);
    }
    router.replace(`?${params.toString()}`);
  }

  // Debounce the search box so typing doesn't hit the URL/API on every keystroke.
  useEffect(() => {
    const handle = setTimeout(() => {
      if (searchInput !== searchParam) setParam('search', searchInput || null);
    }, 300);
    return () => clearTimeout(handle);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [searchInput]);

  function upsert(entry: SprintEntry) {
    setAllEntries((prev) => {
      if (!prev) return [entry];
      const exists = prev.some((e) => e.id === entry.id);
      return exists ? prev.map((e) => (e.id === entry.id ? entry : e)) : [...prev, entry];
    });
    refetchFiltered();
  }

  async function handleDelete(id: number) {
    await deleteSprintEntry(id);
    setAllEntries((prev) => prev?.filter((e) => e.id !== id) ?? null);
    refetchFiltered();
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
  // Computed from allEntries (not the filtered table) since "is there a later
  // entry" is inherently a cross-sprint question.
  const orphanedCarryOvers = (allEntries ?? []).filter((e) => {
    if (e.status !== 'NotDone') return false;
    const sprint = sprints.find((s) => s.id === e.sprintId);
    if (!sprint || sprint.status !== 'Closed') return false;
    return !(allEntries ?? []).some((other) => {
      if (other.ticketId !== e.ticketId) return false;
      const otherSprint = sprints.find((s) => s.id === other.sprintId);
      return !!otherSprint && otherSprint.startDate > sprint.startDate;
    });
  });
  // The warning scopes to whichever sprint the table is currently filtered to,
  // rather than always listing every closed sprint's orphans at once.
  const visibleOrphans = sprintIdParam
    ? orphanedCarryOvers.filter((e) => e.sprintId === Number(sprintIdParam))
    : orphanedCarryOvers;

  const hasActiveFilters = Boolean(sprintIdParam || projectIdParam || statusParam || carriedOverParam || searchParam);

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
        <EntryFormDialog entries={allEntries ?? []} tickets={tickets} sprints={sprints} projects={projects} onSaved={upsert} />
      </div>

      {error && <p className="text-sm text-red-600 dark:text-red-400">{error}</p>}

      <Card>
        <CardContent className="flex flex-wrap items-end gap-3 p-4">
          <div className="space-y-1.5">
            <Label htmlFor="filter-sprint">Sprint</Label>
            <Select
              id="filter-sprint"
              value={sprintIdParam ?? ALL}
              onChange={(e) => setParam('sprintId', e.target.value === ALL ? null : e.target.value)}
            >
              <option value={ALL}>All sprints</option>
              {sprints.map((s) => (
                <option key={s.id} value={s.id}>
                  {s.name} ({s.status})
                </option>
              ))}
            </Select>
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="filter-project">Project</Label>
            <Select
              id="filter-project"
              value={projectIdParam ?? ALL}
              onChange={(e) => setParam('projectId', e.target.value === ALL ? null : e.target.value)}
            >
              <option value={ALL}>All projects</option>
              {projects.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.name}
                </option>
              ))}
            </Select>
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="filter-status">Status</Label>
            <Select
              id="filter-status"
              value={statusParam ?? ALL}
              onChange={(e) => setParam('status', e.target.value === ALL ? null : e.target.value)}
            >
              <option value={ALL}>All statuses</option>
              <option value="Done">Done</option>
              <option value="NotDone">NotDone</option>
              <option value="Cancelled">Cancelled</option>
            </Select>
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="filter-search">Search ticket</Label>
            <Input
              id="filter-search"
              placeholder="Ticket title…"
              value={searchInput}
              onChange={(e) => setSearchInput(e.target.value)}
              className="w-48"
            />
          </div>
          <label className="flex items-center gap-2 pb-2 text-sm">
            <input
              type="checkbox"
              checked={carriedOverParam}
              onChange={(e) => setParam('carriedOver', e.target.checked ? 'true' : null)}
              className="h-4 w-4 rounded border-border"
            />
            Carried-over only
          </label>
          <Button
            type="button"
            variant="ghost"
            size="sm"
            disabled={!hasActiveFilters}
            onClick={() => router.replace('?')}
          >
            Clear filters
          </Button>
        </CardContent>
      </Card>

      {visibleOrphans.length > 0 && (
        <div className="rounded-lg border border-amber-200 bg-amber-50 p-4 text-sm text-amber-900 dark:border-amber-900 dark:bg-amber-950 dark:text-amber-300">
          <p className="font-medium">
            {visibleOrphans.length} ticket{visibleOrphans.length > 1 ? 's' : ''} still NotDone in a closed sprint
            with no carry-over entry yet:
          </p>
          <ul className="mt-1.5 list-disc space-y-0.5 pl-5">
            {visibleOrphans.map((e) => (
              <li key={e.id}>
                {ticketLabel(e.ticketId, tickets, projects)} — {sprintName(e.sprintId)}
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
              {filteredEntries?.length === 0 && (
                <TableRow>
                  <TableCell colSpan={6} className="text-center text-muted-foreground">
                    {hasActiveFilters ? 'No entries match these filters.' : 'No sprint entries yet.'}
                  </TableCell>
                </TableRow>
              )}
              {filteredEntries?.map((e) => (
                <TableRow key={e.id}>
                  <TableCell className="font-medium">{ticketLabel(e.ticketId, tickets, projects)}</TableCell>
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
                    <EntryFormDialog
                      entry={e}
                      entries={allEntries ?? []}
                      tickets={tickets}
                      sprints={sprints}
                      projects={projects}
                      onSaved={upsert}
                    />
                    <DeleteConfirmButton
                      entityLabel={`${ticketLabel(e.ticketId, tickets, projects)} @ ${sprintName(e.sprintId)}`}
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

export default function SprintEntriesPage() {
  return (
    <Suspense fallback={<div className="mx-auto max-w-5xl p-6 text-sm text-muted-foreground">Loading…</div>}>
      <SprintEntriesPageInner />
    </Suspense>
  );
}
