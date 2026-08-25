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
import { listSprints, createSprint, updateSprint, deleteSprint, type Sprint } from '@/lib/api';

// HTML date inputs use YYYY-MM-DD; the backend's Go time.Time fields marshal as RFC3339.
function toDateInputValue(iso: string) {
  return iso.slice(0, 10);
}

function toRFC3339(dateInputValue: string) {
  return new Date(`${dateInputValue}T00:00:00Z`).toISOString();
}

function SprintFormDialog({
  sprint,
  onSaved,
}: {
  sprint?: Sprint;
  onSaved: (sprint: Sprint) => void;
}) {
  const isEdit = sprint !== undefined;
  const [open, setOpen] = useState(false);
  const [name, setName] = useState(sprint?.name ?? '');
  const [startDate, setStartDate] = useState(sprint ? toDateInputValue(sprint.startDate) : '');
  const [endDate, setEndDate] = useState(sprint ? toDateInputValue(sprint.endDate) : '');
  const [status, setStatus] = useState<Sprint['status']>(sprint?.status ?? 'Open');
  const [pending, setPending] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (open) {
      setName(sprint?.name ?? '');
      setStartDate(sprint ? toDateInputValue(sprint.startDate) : '');
      setEndDate(sprint ? toDateInputValue(sprint.endDate) : '');
      setStatus(sprint?.status ?? 'Open');
      setError(null);
    }
  }, [open, sprint]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setPending(true);
    setError(null);
    try {
      const input = { name, startDate: toRFC3339(startDate), endDate: toRFC3339(endDate) };
      const saved = isEdit ? await updateSprint(sprint.id, { ...input, status }) : await createSprint(input);
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
          <Button>Add sprint</Button>
        )}
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{isEdit ? 'Edit sprint' : 'Add sprint'}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="sprint-name">Name</Label>
            <Input id="sprint-name" value={name} onChange={(e) => setName(e.target.value)} required />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <Label htmlFor="sprint-start">Start date</Label>
              <Input
                id="sprint-start"
                type="date"
                value={startDate}
                onChange={(e) => setStartDate(e.target.value)}
                required
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="sprint-end">End date</Label>
              <Input
                id="sprint-end"
                type="date"
                value={endDate}
                onChange={(e) => setEndDate(e.target.value)}
                required
              />
            </div>
          </div>
          {isEdit && (
            <div className="space-y-1.5">
              <Label htmlFor="sprint-status">Status</Label>
              <Select id="sprint-status" value={status} onChange={(e) => setStatus(e.target.value as Sprint['status'])}>
                <option value="Open">Open</option>
                <option value="Closed">Closed</option>
              </Select>
            </div>
          )}
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

export default function SprintsPage() {
  const [sprints, setSprints] = useState<Sprint[] | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    listSprints()
      .then(setSprints)
      .catch((err) => setError(err instanceof Error ? err.message : 'Failed to load sprints'));
  }, []);

  function upsert(sprint: Sprint) {
    setSprints((prev) => {
      if (!prev) return [sprint];
      const exists = prev.some((s) => s.id === sprint.id);
      return exists ? prev.map((s) => (s.id === sprint.id ? sprint : s)) : [...prev, sprint];
    });
  }

  async function handleDelete(id: number) {
    await deleteSprint(id);
    setSprints((prev) => prev?.filter((s) => s.id !== id) ?? null);
  }

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Sprints</h1>
          <p className="text-sm text-muted-foreground">New sprints start Open; closing is explicit.</p>
        </div>
        <SprintFormDialog onSaved={upsert} />
      </div>

      {error && <p className="text-sm text-red-600 dark:text-red-400">{error}</p>}

      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Dates</TableHead>
                <TableHead>Status</TableHead>
                <TableHead className="w-1" />
              </TableRow>
            </TableHeader>
            <TableBody>
              {sprints?.length === 0 && (
                <TableRow>
                  <TableCell colSpan={4} className="text-center text-muted-foreground">
                    No sprints yet.
                  </TableCell>
                </TableRow>
              )}
              {sprints?.map((s) => (
                <TableRow key={s.id}>
                  <TableCell className="font-medium">{s.name}</TableCell>
                  <TableCell className="text-muted-foreground">
                    {toDateInputValue(s.startDate)} &ndash; {toDateInputValue(s.endDate)}
                  </TableCell>
                  <TableCell>
                    <Badge variant={s.status === 'Open' ? 'success' : 'secondary'}>{s.status}</Badge>
                  </TableCell>
                  <TableCell className="flex justify-end gap-1">
                    <SprintFormDialog sprint={s} onSaved={upsert} />
                    <DeleteConfirmButton entityLabel={s.name} onDelete={() => handleDelete(s.id)} />
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
