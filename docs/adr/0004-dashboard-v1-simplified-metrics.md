# Dashboard v1 ships workload/done points, not the full Sprint Velocity split

[CONTEXT.md](../../CONTEXT.md) already defines a precise metric model: Sprint Velocity splits into Committed Completed Points and Late-Add Completed Points, plus Planning Accuracy and Late-Add Rate. The first dashboard does not implement that split.

Instead, it ships two simpler numbers per sprint and per developer:

- **Workload**: sum of `points_at_entry` across all of a sprint's `SprintEntry` rows.
- **Done**: sum of `points_at_entry` where that entry's status is `Done`.

Both exclude `Cancelled` entries, per the existing rule in CONTEXT.md that Cancelled tickets are excluded from all metrics.

This was a deliberate scope cut, not an oversight: the Committed/Late-Add split only matters once "added after sprint start" tickets are common enough in the data to be worth separating out, and Planning Accuracy/Late-Add Rate are derived from that same split. Shipping workload/done first answers the immediate question ("how much is done vs. planned, per sprint and per developer") without waiting on the more nuanced breakdown.

## Consequences

- The dashboard's numbers are not directly comparable to a future "Sprint Velocity" figure once the Committed/Late-Add split is added — workload here includes late-adds undifferentiated from committed work.
- `GET /dashboard/sprints` and `GET /dashboard/sprints/{id}` compute these sums in SQL (joining `sprint_entry` to `sprint`/`ticket`/`users`), not in the frontend, so adding the split later is a query change, not an API redesign.
- Planning Accuracy and Late-Add Rate remain unimplemented; a v2 dashboard iteration can add them once the split above exists.
