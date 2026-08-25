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
