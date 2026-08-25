// Thin fetch wrappers around the real backend API, used to set up and tear
// down fixture data around the UI flows under test. Deliberately separate
// from frontend/lib/api.ts (that module reads Next.js env vars meant for
// the app's own runtime, not a Node test process).

const API_BASE = process.env.E2E_API_BASE_URL ?? 'http://localhost:8080';

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    method,
    headers: body === undefined ? undefined : { 'Content-Type': 'application/json' },
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  if (!res.ok) {
    throw new Error(`${method} ${path} responded ${res.status}: ${await res.text()}`);
  }
  if (res.status === 204) return undefined as T;
  return (await res.json()) as T;
}

// Suffix fixture names with this so parallel test runs (or a stale prior
// run's leftovers) never collide with each other or with real dev data.
export function uniqueName(prefix: string): string {
  return `${prefix} ${Date.now()}-${Math.floor(Math.random() * 10000)}`;
}

export type ApiProject = { id: number; name: string; status: 'Active' | 'Archived' };
export type ApiUser = { id: number; name: string; role: 'Lead' | 'Developer' };
export type ApiTicket = { id: number; projectId: number; title: string; storyPoints: number; assigneeId: number | null };
export type ApiSprint = { id: number; name: string; startDate: string; endDate: string; status: 'Open' | 'Closed' };
export type ApiSprintEntry = { id: number; ticketId: number; sprintId: number };

export const api = {
  createProject: (name: string) => request<ApiProject>('POST', '/projects', { name }),
  deleteProject: (id: number) => request<void>('DELETE', `/projects/${id}`),

  createUser: (name: string, role: ApiUser['role'] = 'Developer') =>
    request<ApiUser>('POST', '/users', { name, role }),
  deleteUser: (id: number) => request<void>('DELETE', `/users/${id}`),

  createTicket: (projectId: number, title: string, storyPoints = 3) =>
    request<ApiTicket>('POST', '/tickets', { projectId, title, storyPoints, assigneeId: null }),
  deleteTicket: (id: number) => request<void>('DELETE', `/tickets/${id}`),
  listTickets: () => request<ApiTicket[]>('GET', '/tickets'),

  createSprint: (name: string, startDate: string, endDate: string) =>
    request<ApiSprint>('POST', '/sprints', { name, startDate, endDate }),
  closeSprint: (sprint: ApiSprint) =>
    // The backend rejects unknown fields, so only send what PUT accepts —
    // spreading the whole ApiSprint (which includes `id`) would 400.
    request<ApiSprint>('PUT', `/sprints/${sprint.id}`, {
      name: sprint.name,
      startDate: sprint.startDate,
      endDate: sprint.endDate,
      status: 'Closed',
    }),
  deleteSprint: (id: number) => request<void>('DELETE', `/sprints/${id}`),

  createSprintEntry: (ticketId: number, sprintId: number, pointsAtEntry: number) =>
    request<ApiSprintEntry>('POST', '/sprint-entries', {
      ticketId,
      sprintId,
      status: 'NotDone',
      addedAfterSprintStart: false,
      carriedFrom: null,
      pointsAtEntry,
    }),
  deleteSprintEntry: (id: number) => request<void>('DELETE', `/sprint-entries/${id}`),
  listSprintEntries: () => request<ApiSprintEntry[]>('GET', '/sprint-entries'),
};

// Deletes any sprint-entries referencing the given ticket/sprint ids first,
// so the subsequent ticket/sprint deletes don't fail on the FK constraint.
export async function deleteEntriesReferencing(ticketIds: number[], sprintIds: number[]) {
  const entries = await api.listSprintEntries();
  for (const e of entries) {
    if (ticketIds.includes(e.ticketId) || sprintIds.includes(e.sprintId)) {
      await api.deleteSprintEntry(e.id);
    }
  }
}

// Deletes any tickets still referencing the given project ids — needed when
// a test fails partway through and its own UI-driven ticket delete never ran.
export async function deleteTicketsReferencing(projectIds: number[]) {
  const tickets = await api.listTickets();
  for (const t of tickets) {
    if (projectIds.includes(t.projectId)) {
      await api.deleteTicket(t.id);
    }
  }
}
