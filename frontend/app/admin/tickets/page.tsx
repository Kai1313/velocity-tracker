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
  DialogFooter,
} from '@/components/ui/dialog';
import { DeleteConfirmButton } from '@/components/admin/delete-confirm-button';
import {
  listTickets,
  createTicket,
  updateTicket,
  deleteTicket,
  listProjects,
  listUsers,
  type TicketDetail,
  type Project,
  type User,
  type EntryStatus,
} from '@/lib/api';

const UNASSIGNED = 'unassigned';
const ALL = 'all';
const NO_STATUS = 'none';

const statusVariant: Record<EntryStatus, 'success' | 'warning' | 'secondary'> = {
  Done: 'success',
  NotDone: 'warning',
  Cancelled: 'secondary',
};

function StatusBadge({ status }: { status: EntryStatus | null }) {
  if (status === null) return <Badge variant="outline">No status</Badge>;
  return <Badge variant={statusVariant[status]}>{status}</Badge>;
}

function TicketFormDialog({
  ticket,
  projects,
  users,
  onSaved,
}: {
  ticket?: TicketDetail;
  projects: Project[];
  users: User[];
  onSaved: (ticket: TicketDetail) => void;
}) {
  const isEdit = ticket !== undefined;
  const [open, setOpen] = useState(false);
  const [projectId, setProjectId] = useState(ticket?.projectId ?? projects[0]?.id ?? 0);
  const [title, setTitle] = useState(ticket?.title ?? '');
  const [storyPoints, setStoryPoints] = useState(ticket?.storyPoints ?? 1);
  const [assigneeId, setAssigneeId] = useState<string>(
    ticket?.assigneeId != null ? String(ticket.assigneeId) : UNASSIGNED,
  );
  const [pending, setPending] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (open) {
      setProjectId(ticket?.projectId ?? projects[0]?.id ?? 0);
      setTitle(ticket?.title ?? '');
      setStoryPoints(ticket?.storyPoints ?? 1);
      setAssigneeId(ticket?.assigneeId != null ? String(ticket.assigneeId) : UNASSIGNED);
      setError(null);
    }
  }, [open, ticket, projects]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setPending(true);
    setError(null);
    const input = {
      projectId,
      title,
      storyPoints,
      assigneeId: assigneeId === UNASSIGNED ? null : Number(assigneeId),
    };
    try {
      const saved = isEdit ? await updateTicket(ticket.id, input) : await createTicket(input);
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
          <Button variant="ghost" size="sm">
            Edit
          </Button>
        ) : (
          <Button disabled={projects.length === 0}>Add ticket</Button>
        )}
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{isEdit ? 'Edit ticket' : 'Add ticket'}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="ticket-project">Project</Label>
            <Select
              id="ticket-project"
              value={projectId}
              onChange={(e) => setProjectId(Number(e.target.value))}
              required
            >
              {projects.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.name}
                </option>
              ))}
            </Select>
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="ticket-title">Title</Label>
            <Input id="ticket-title" value={title} onChange={(e) => setTitle(e.target.value)} required />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="ticket-points">Story points</Label>
            <Input
              id="ticket-points"
              type="number"
              min={1}
              value={storyPoints}
              onChange={(e) => setStoryPoints(Number(e.target.value))}
              required
            />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="ticket-assignee">Assignee</Label>
            <Select id="ticket-assignee" value={assigneeId} onChange={(e) => setAssigneeId(e.target.value)}>
              <option value={UNASSIGNED}>Unassigned</option>
              {users.map((u) => (
                <option key={u.id} value={u.id}>
                  {u.name}
                </option>
              ))}
            </Select>
          </div>
          {error && <p className="text-sm text-red-600 dark:text-red-400">{error}</p>}
          <DialogFooter>
            <Button type="submit" disabled={pending}>
              {pending ? 'Saving…' : 'Save'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

export default function TicketsPage() {
  const [tickets, setTickets] = useState<TicketDetail[] | null>(null);
  const [projects, setProjects] = useState<Project[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [error, setError] = useState<string | null>(null);

  const [projectFilter, setProjectFilter] = useState(ALL);
  const [assigneeFilter, setAssigneeFilter] = useState(ALL);
  const [statusFilter, setStatusFilter] = useState(ALL);
  const [search, setSearch] = useState('');

  useEffect(() => {
    Promise.all([listTickets(), listProjects(), listUsers()])
      .then(([t, p, u]) => {
        setTickets(t);
        setProjects(p);
        setUsers(u);
      })
      .catch((err) => setError(err instanceof Error ? err.message : 'Failed to load tickets'));
  }, []);

  function upsert(ticket: TicketDetail) {
    setTickets((prev) => {
      if (!prev) return [ticket];
      const exists = prev.some((t) => t.id === ticket.id);
      return exists ? prev.map((t) => (t.id === ticket.id ? ticket : t)) : [...prev, ticket];
    });
  }

  async function handleDelete(id: number) {
    await deleteTicket(id);
    setTickets((prev) => prev?.filter((t) => t.id !== id) ?? null);
  }

  function projectName(id: number) {
    return projects.find((p) => p.id === id)?.name ?? `#${id}`;
  }

  function assigneeName(id: number | null) {
    if (id === null) return 'Unassigned';
    return users.find((u) => u.id === id)?.name ?? `#${id}`;
  }

  const hasActiveFilters = Boolean(
    projectFilter !== ALL || assigneeFilter !== ALL || statusFilter !== ALL || search,
  );

  const filteredTickets =
    tickets === null
      ? null
      : tickets.filter((t) => {
          if (projectFilter !== ALL && String(t.projectId) !== projectFilter) return false;
          if (assigneeFilter !== ALL) {
            const matches = assigneeFilter === UNASSIGNED ? t.assigneeId === null : String(t.assigneeId) === assigneeFilter;
            if (!matches) return false;
          }
          if (statusFilter !== ALL) {
            const matches = statusFilter === NO_STATUS ? t.currentStatus === null : t.currentStatus === statusFilter;
            if (!matches) return false;
          }
          if (search && !t.title.toLowerCase().includes(search.toLowerCase())) return false;
          return true;
        });

  function clearFilters() {
    setProjectFilter(ALL);
    setAssigneeFilter(ALL);
    setStatusFilter(ALL);
    setSearch('');
  }

  return (
    <div className="mx-auto max-w-4xl space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Tickets</h1>
          <p className="text-sm text-muted-foreground">
            {projects.length === 0
              ? 'Add a project first — tickets must belong to one.'
              : 'Stable work records: id, title, story points, owning project.'}
          </p>
        </div>
        <TicketFormDialog projects={projects} users={users} onSaved={upsert} />
      </div>

      {error && <p className="text-sm text-red-600 dark:text-red-400">{error}</p>}

      <Card>
        <CardContent className="flex flex-wrap items-end gap-3 p-4">
          <div className="space-y-1.5">
            <Label htmlFor="filter-project">Project</Label>
            <Select id="filter-project" value={projectFilter} onChange={(e) => setProjectFilter(e.target.value)}>
              <option value={ALL}>All projects</option>
              {projects.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.name}
                </option>
              ))}
            </Select>
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="filter-assignee">Assignee</Label>
            <Select id="filter-assignee" value={assigneeFilter} onChange={(e) => setAssigneeFilter(e.target.value)}>
              <option value={ALL}>All assignees</option>
              <option value={UNASSIGNED}>Unassigned</option>
              {users.map((u) => (
                <option key={u.id} value={u.id}>
                  {u.name}
                </option>
              ))}
            </Select>
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="filter-status">Status</Label>
            <Select id="filter-status" value={statusFilter} onChange={(e) => setStatusFilter(e.target.value)}>
              <option value={ALL}>All statuses</option>
              <option value="Done">Done</option>
              <option value="NotDone">NotDone</option>
              <option value="Cancelled">Cancelled</option>
              <option value={NO_STATUS}>No status</option>
            </Select>
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="filter-search">Search title</Label>
            <Input
              id="filter-search"
              placeholder="Ticket title…"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-48"
            />
          </div>
          <Button type="button" variant="ghost" size="sm" disabled={!hasActiveFilters} onClick={clearFilters}>
            Clear filters
          </Button>
        </CardContent>
      </Card>

      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Title</TableHead>
                <TableHead>Project</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Points</TableHead>
                <TableHead>Assignee</TableHead>
                <TableHead className="w-1" />
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredTickets?.length === 0 && (
                <TableRow>
                  <TableCell colSpan={6} className="text-center text-muted-foreground">
                    {hasActiveFilters ? 'No tickets match these filters.' : 'No tickets yet.'}
                  </TableCell>
                </TableRow>
              )}
              {filteredTickets?.map((t) => (
                <TableRow key={t.id}>
                  <TableCell className="font-medium">{t.title}</TableCell>
                  <TableCell className="text-muted-foreground">{projectName(t.projectId)}</TableCell>
                  <TableCell>
                    <StatusBadge status={t.currentStatus} />
                  </TableCell>
                  <TableCell>{t.storyPoints}</TableCell>
                  <TableCell className="text-muted-foreground">{assigneeName(t.assigneeId)}</TableCell>
                  <TableCell className="flex justify-end gap-1">
                    <TicketFormDialog ticket={t} projects={projects} users={users} onSaved={upsert} />
                    <DeleteConfirmButton entityLabel={t.title} onDelete={() => handleDelete(t.id)} />
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
