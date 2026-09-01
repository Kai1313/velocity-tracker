export type SprintSummary = {
  sprintId: number;
  sprintName: string;
  workloadPoints: number;
  donePoints: number;
  workloadTickets: number;
  doneTickets: number;
};

export type DeveloperSummary = {
  userId: number | null;
  name: string;
  workloadPoints: number;
  donePoints: number;
  workloadTickets: number;
  doneTickets: number;
};

export type Sprint = {
  id: number;
  name: string;
  startDate: string;
  endDate: string;
  status: 'Open' | 'Closed';
};

export type Role = 'Lead' | 'Developer';

export type User = {
  id: number;
  name: string;
  role: Role;
};

export type ProjectStatus = 'Active' | 'Archived';

export type Project = {
  id: number;
  name: string;
  status: ProjectStatus;
};

export type EntryStatus = 'Done' | 'NotDone' | 'Cancelled';

export type TicketDetail = {
  id: number;
  projectId: number;
  title: string;
  storyPoints: number;
  assigneeId: number | null;
  currentStatus: EntryStatus | null;
};

export type SprintEntry = {
  id: number;
  ticketId: number;
  sprintId: number;
  status: EntryStatus;
  addedAfterSprintStart: boolean;
  carriedFrom: number | null;
  pointsAtEntry: number;
  createdAt: string;
};

export type SprintEntryDetail = {
  entryId: number;
  ticketId: number;
  ticketTitle: string;
  projectName: string;
  assigneeName: string;
  status: EntryStatus;
  addedAfterSprintStart: boolean;
  pointsAtEntry: number;
  carriedFromSprintName: string | null;
};

export type SprintTicketBreakdown = {
  current: SprintEntryDetail[];
  carriedOver: SprintEntryDetail[];
};

function backendUrl() {
  return process.env.BACKEND_INTERNAL_URL ?? 'http://localhost:8080';
}

// Browser-side calls (from Client Components) can't reach BACKEND_INTERNAL_URL,
// which only resolves inside the Docker network or on the host running the
// Next.js server. NEXT_PUBLIC_BACKEND_URL is inlined at build time instead.
function publicBackendUrl() {
  return process.env.NEXT_PUBLIC_BACKEND_URL ?? 'http://localhost:8080';
}

export class ApiError extends Error {
  constructor(message: string, public status: number) {
    super(message);
  }
}

async function getJSON<T>(path: string): Promise<T> {
  const res = await fetch(`${backendUrl()}${path}`, { cache: 'no-store' });
  if (!res.ok) {
    throw new Error(`${path} responded ${res.status}`);
  }
  return (await res.json()) as T;
}

async function requestJSON<T>(method: string, path: string, body?: unknown): Promise<T> {
  const res = await fetch(`${publicBackendUrl()}${path}`, {
    method,
    headers: body === undefined ? undefined : { 'Content-Type': 'application/json' },
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  if (!res.ok) {
    const payload = await res.json().catch(() => null);
    throw new ApiError(payload?.error ?? `${path} responded ${res.status}`, res.status);
  }
  if (res.status === 204) {
    return undefined as T;
  }
  return (await res.json()) as T;
}

export function getSprintSummaries() {
  return getJSON<SprintSummary[]>('/dashboard/sprints');
}

export function getSprintDeveloperBreakdown(sprintId: number) {
  return getJSON<DeveloperSummary[]>(`/dashboard/sprints/${sprintId}`);
}

export function getSprintTicketBreakdown(sprintId: number) {
  return getJSON<SprintTicketBreakdown>(`/dashboard/sprints/${sprintId}/entries`);
}

export function getSprint(sprintId: number) {
  return getJSON<Sprint>(`/sprints/${sprintId}`);
}

export function listUsers() {
  return requestJSON<User[]>('GET', '/users');
}

export function createUser(input: { name: string; role: Role }) {
  return requestJSON<User>('POST', '/users', input);
}

export function updateUser(id: number, input: { name: string; role: Role }) {
  return requestJSON<User>('PUT', `/users/${id}`, input);
}

export function deleteUser(id: number) {
  return requestJSON<void>('DELETE', `/users/${id}`);
}

export function listProjects() {
  return requestJSON<Project[]>('GET', '/projects');
}

export function createProject(input: { name: string }) {
  return requestJSON<Project>('POST', '/projects', input);
}

export function updateProject(id: number, input: { name: string; status: ProjectStatus }) {
  return requestJSON<Project>('PUT', `/projects/${id}`, input);
}

export function deleteProject(id: number) {
  return requestJSON<void>('DELETE', `/projects/${id}`);
}

export function listTickets() {
  return requestJSON<TicketDetail[]>('GET', '/tickets');
}

export type TicketInput = { projectId: number; title: string; storyPoints: number; assigneeId: number | null };

export function createTicket(input: TicketInput) {
  return requestJSON<TicketDetail>('POST', '/tickets', input);
}

export function updateTicket(id: number, input: TicketInput) {
  return requestJSON<TicketDetail>('PUT', `/tickets/${id}`, input);
}

export function deleteTicket(id: number) {
  return requestJSON<void>('DELETE', `/tickets/${id}`);
}

export function listSprints() {
  return requestJSON<Sprint[]>('GET', '/sprints');
}

export function createSprint(input: { name: string; startDate: string; endDate: string }) {
  return requestJSON<Sprint>('POST', '/sprints', input);
}

export function updateSprint(
  id: number,
  input: { name: string; startDate: string; endDate: string; status: Sprint['status'] },
) {
  return requestJSON<Sprint>('PUT', `/sprints/${id}`, input);
}

export function deleteSprint(id: number) {
  return requestJSON<void>('DELETE', `/sprints/${id}`);
}

export type SprintEntryFilter = {
  sprintId?: number;
  projectId?: number;
  status?: EntryStatus;
  carriedOver?: boolean;
  search?: string;
};

export function listSprintEntries(filter: SprintEntryFilter = {}) {
  const params = new URLSearchParams();
  if (filter.sprintId != null) params.set('sprintId', String(filter.sprintId));
  if (filter.projectId != null) params.set('projectId', String(filter.projectId));
  if (filter.status) params.set('status', filter.status);
  if (filter.carriedOver) params.set('carriedOver', 'true');
  if (filter.search) params.set('search', filter.search);
  const qs = params.toString();
  return requestJSON<SprintEntry[]>('GET', `/sprint-entries${qs ? `?${qs}` : ''}`);
}

export type SprintEntryCreateInput = {
  ticketId: number;
  sprintId: number;
  status: EntryStatus;
  addedAfterSprintStart: boolean;
  carriedFrom: number | null;
  pointsAtEntry: number;
};

export type SprintEntryUpdateInput = Omit<SprintEntryCreateInput, 'ticketId' | 'sprintId'>;

export function createSprintEntry(input: SprintEntryCreateInput) {
  return requestJSON<SprintEntry>('POST', '/sprint-entries', input);
}

export function updateSprintEntry(id: number, input: SprintEntryUpdateInput) {
  return requestJSON<SprintEntry>('PUT', `/sprint-entries/${id}`, input);
}

export function deleteSprintEntry(id: number) {
  return requestJSON<void>('DELETE', `/sprint-entries/${id}`);
}
