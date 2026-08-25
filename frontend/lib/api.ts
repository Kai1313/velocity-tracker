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

function backendUrl() {
  return process.env.BACKEND_INTERNAL_URL ?? 'http://localhost:8080';
}

async function getJSON<T>(path: string): Promise<T> {
  const res = await fetch(`${backendUrl()}${path}`, { cache: 'no-store' });
  if (!res.ok) {
    throw new Error(`${path} responded ${res.status}`);
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
