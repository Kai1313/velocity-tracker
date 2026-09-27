'use client';

import { useEffect, useState } from 'react';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select } from '@/components/ui/select';
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
import { listUsers, createUser, updateUser, deleteUser, type User, type Role } from '@/lib/api';

function UserFormDialog({
  user,
  onSaved,
}: {
  user?: User;
  onSaved: (user: User) => void;
}) {
  const isEdit = user !== undefined;
  const { open, setOpen, pending, error, setError, submit } = useFormDialogState();
  const [name, setName] = useState(user?.name ?? '');
  const [role, setRole] = useState<Role>(user?.role ?? 'Developer');

  useEffect(() => {
    if (open) {
      setName(user?.name ?? '');
      setRole(user?.role ?? 'Developer');
      setError(null);
    }
  }, [open, user, setError]);

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    submit(() => (isEdit ? updateUser(user.id, { name, role }) : createUser({ name, role })), onSaved);
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        {isEdit ? (
          <Button variant="ghost" size="sm">
            Edit
          </Button>
        ) : (
          <Button>Add user</Button>
        )}
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{isEdit ? 'Edit user' : 'Add user'}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="user-name">Name</Label>
            <Input id="user-name" value={name} onChange={(e) => setName(e.target.value)} required />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="user-role">Role</Label>
            <Select id="user-role" value={role} onChange={(e) => setRole(e.target.value as Role)}>
              <option value="Developer">Developer</option>
              <option value="Lead">Lead</option>
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

export default function UsersPage() {
  const { items: users, error, upsert, remove } = useEntityList(listUsers, 'Failed to load users');

  async function handleDelete(id: number) {
    await deleteUser(id);
    remove(id);
  }

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Users</h1>
          <p className="text-sm text-muted-foreground">Team members who can be assigned tickets.</p>
        </div>
        <UserFormDialog onSaved={upsert} />
      </div>

      {error && <p className="text-sm text-red-600 dark:text-red-400">{error}</p>}

      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Role</TableHead>
                <TableHead className="w-1" />
              </TableRow>
            </TableHeader>
            <TableBody>
              {users?.length === 0 && (
                <TableRow>
                  <TableCell colSpan={3} className="text-center text-muted-foreground">
                    No users yet.
                  </TableCell>
                </TableRow>
              )}
              {users?.map((u) => (
                <TableRow key={u.id}>
                  <TableCell className="font-medium">{u.name}</TableCell>
                  <TableCell className="text-muted-foreground">{u.role}</TableCell>
                  <TableCell className="flex justify-end gap-1">
                    <UserFormDialog user={u} onSaved={upsert} />
                    <DeleteConfirmButton entityLabel={u.name} onDelete={() => handleDelete(u.id)} />
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
