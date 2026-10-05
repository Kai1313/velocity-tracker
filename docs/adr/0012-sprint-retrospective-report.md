# Sprint Retrospective Report computes Planning Accuracy and Late-Add Rate for closed sprints

CONTEXT.md has defined Planning Accuracy and Late-Add Rate since early on, but neither was ever computed anywhere in the app. `GET /dashboard/retrospective` and the `/dashboard/retrospective` page close that gap, answering "how reliably does this team commit" rather than Sprint Health's ([ADR-0010](0010-project-sprint-health.md)/[ADR-0011](0011-sprint-health-aggregates-across-projects.md)) "is the sprint in progress on pace right now."

## What it computes

For the most recently closed `limit` sprints (query param, default 3), oldest first:

- **Planning Accuracy** = Committed Completed Points ÷ Committed Points at Sprint Start × 100
- **Late-Add Rate** = Added Mid-Sprint Points ÷ Committed Points at Sprint Start × 100
- Both are `null` when the sprint had no committed points, rather than a divide-by-zero or a misleading 0%.

Three deliberate differences from Sprint Health's `SprintPoints` query:

- **Closed sprints only.** Both ratios are only final once a sprint's entries are locked; a still-`Open` sprint's "accuracy so far" is a moving target. Sprint Health already answers "how's the open sprint doing" — this is a different question.
- **Every project counted, regardless of its current status.** Sprint Health's `SprintPoints` filters to `Active` projects because it's a live pace signal. A retrospective is historical: if a project is archived after a sprint closes, that sprint's recorded commitment and completion shouldn't silently shrink because of an unrelated later decision. `ClosedSprintPoints` has no project-status filter at all.
- **Sprint-level only, not broken out per project** — same reasoning as ADR-0011: this is one team with shared capacity across projects, so a per-project split would fragment a number meant to answer a team-wide question.

The table also shows a ticket count beside each SP figure (`152 (24 tickets)`), counted over the same non-Cancelled sprint entries as the points, so carried-over entries count in both. The ratios stay SP-based.

"Most recently closed" is approximated by sprint ID order (creation sequence), the same convention every other dashboard query in this codebase already uses — there's no `closed_at` timestamp on `Sprint` to order by directly (closing is a status flip, see [ADR-0009](0009-sprint-reopening-is-an-escape-hatch.md)).

## Consequences

- `CONTEXT.md`'s Planning Accuracy/Late-Add Rate entries no longer say "not yet computed anywhere" — this is that computation.
- A sprint can be reopened and reclosed ([ADR-0009](0009-sprint-reopening-is-an-escape-hatch.md)); doing so doesn't change its position in "most recently closed" order, since that's still sprint-ID-based, not event-based. This matches existing behavior elsewhere and isn't a new inconsistency.
- `limit` is a plain query param with no upper bound. Fine at this app's scale (per [ADR-0006](0006-multi-team-readiness.md), a handful of sprints at a time); would want a cap if sprint history grows large enough for a huge `limit` to matter.
