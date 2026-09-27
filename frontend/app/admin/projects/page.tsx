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
import { useEntityList } from '@/lib/hooks/use-entity-list';
import { useFormDialogState } from '@/lib/hooks/use-form-dialog';
import { listProjects, createProject, updateProject, deleteProject, type Project, type ProjectStatus } from '@/lib/api';

function ProjectFormDialog({
  project,
  onSaved,
}: {
  project?: Project;
  onSaved: (project: Project) => void;
}) {
  const isEdit = project !== undefined;
  const { open, setOpen, pending, error, setError, submit } = useFormDialogState();
  const [name, setName] = useState(project?.name ?? '');
  const [status, setStatus] = useState<ProjectStatus>(project?.status ?? 'Active');

  useEffect(() => {
    if (open) {
      setName(project?.name ?? '');
      setStatus(project?.status ?? 'Active');
      setError(null);
    }
  }, [open, project, setError]);

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    submit(() => (isEdit ? updateProject(project.id, { name, status }) : createProject({ name })), onSaved);
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        {isEdit ? (
          <Button variant="ghost" size="sm">
            Edit
          </Button>
        ) : (
          <Button>Add project</Button>
        )}
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{isEdit ? 'Edit project' : 'Add project'}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="project-name">Name</Label>
            <Input id="project-name" value={name} onChange={(e) => setName(e.target.value)} required />
          </div>
          {isEdit && (
            <div className="space-y-1.5">
              <Label htmlFor="project-status">Status</Label>
              <Select
                id="project-status"
                value={status}
                onChange={(e) => setStatus(e.target.value as ProjectStatus)}
              >
                <option value="Active">Active</option>
                <option value="Archived">Archived</option>
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

export default function ProjectsPage() {
  const { items: projects, error, upsert, remove } = useEntityList(listProjects, 'Failed to load projects');

  async function handleDelete(id: number) {
    await deleteProject(id);
    remove(id);
  }

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Projects</h1>
          <p className="text-sm text-muted-foreground">What tickets are grouped under.</p>
        </div>
        <ProjectFormDialog onSaved={upsert} />
      </div>

      {error && <p className="text-sm text-red-600 dark:text-red-400">{error}</p>}

      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Status</TableHead>
                <TableHead className="w-1" />
              </TableRow>
            </TableHeader>
            <TableBody>
              {projects?.length === 0 && (
                <TableRow>
                  <TableCell colSpan={3} className="text-center text-muted-foreground">
                    No projects yet.
                  </TableCell>
                </TableRow>
              )}
              {projects?.map((p) => (
                <TableRow key={p.id}>
                  <TableCell className="font-medium">{p.name}</TableCell>
                  <TableCell>
                    <Badge variant={p.status === 'Active' ? 'success' : 'secondary'}>{p.status}</Badge>
                  </TableCell>
                  <TableCell className="flex justify-end gap-1">
                    <ProjectFormDialog project={p} onSaved={upsert} />
                    <DeleteConfirmButton entityLabel={p.name} onDelete={() => handleDelete(p.id)} />
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
