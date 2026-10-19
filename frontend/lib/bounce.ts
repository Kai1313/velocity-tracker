// Bounce figures are only meaningful with their raw counts beside them: with a
// handful of tickets per sprint, "1.4" alone hides whether it is 7 bounces on
// 5 tickets or 1 on 1. `tracked` is how many tickets had bounce data at all;
// 0 means the sprint predates bounce tracking, which is not the same as "0".
export function formatBounceAverage(total: number, tracked: number): string | null {
  if (tracked === 0) return null;
  const avg = (total / tracked).toFixed(1);
  return `${total} ${total === 1 ? 'bounce' : 'bounces'} / ${tracked} ${tracked === 1 ? 'ticket' : 'tickets'} = ${avg}`;
}

// An entry's bounces from its ticket's entries in earlier sprints (sprint ID
// order, matching the backend's stuck-ticket query), skipping Cancelled and
// untracked entries. A carried-over entry's own count restarts at 0, so this
// is what keeps the ticket's history visible.
export function earlierBounces(
  entry: { ticketId: number; sprintId: number },
  allEntries: { ticketId: number; sprintId: number; status: string; bounceCount: number | null }[],
): number {
  return allEntries
    .filter((o) => o.ticketId === entry.ticketId && o.sprintId < entry.sprintId && o.status !== 'Cancelled')
    .reduce((sum, o) => sum + (o.bounceCount ?? 0), 0);
}
