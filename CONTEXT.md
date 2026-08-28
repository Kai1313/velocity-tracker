# Velocity Tracker

A real app (not the throwaway `Team_Velocity_Tracker.example.xlsx` mockup), backed by PostgreSQL, for tracking sprint velocity, planning accuracy, and late-add rate for a single software team (one lead developer plus several developers). Multi-team support is an explicit Future Enhancement, not MVP scope — see [ADR-0006](docs/adr/0006-multi-team-readiness.md) for what that migration would actually require.

## Language

**Project**:
A lightweight grouping that every `Ticket` belongs to (e.g. "Website Redesign", "Internal Tooling"). Exists so tickets can be organized by what they're for. Has a `status` (Active / Archived) — archiving is how a project is retired without deleting it and orphaning its tickets. `Sprint` deliberately does *not* belong to a `Project`: this is a single team running one set of sprints, and a sprint's tickets can span multiple projects. See [ADR-0003](docs/adr/0003-project-ownership.md).

**Ticket**:
The stable record of a piece of work: id, title, story points, owning project. Attributes that don't change across sprints.
_Note_: `title` is not guaranteed unique across `Project`s — two projects can independently reuse the same numbering scheme (e.g. both landing on `2026_07_FEATURE_110`). Anywhere a ticket is identified by title alone, prefix it with its project name to stay unambiguous; see the [Admin UI section of README.md](README.md#admin-ui) for where this actually matters today.

**SprintEntry**:
One row per sprint that a Ticket appears in. Holds the per-sprint snapshot: status (Done / Not Done / Cancelled), whether it was added after sprint start, and — if carried over — a `carried-from` reference to the SprintEntry it continued from. A ticket that spans 3 sprints has 3 SprintEntry rows, one per sprint, so each sprint's history stays immutable once that sprint closes.
_Avoid_: treating Ticket and its sprint appearance as the same row — that was the xlsx mockup's mistake, and it's why the mockup couldn't compute carry-over or planning-accuracy history correctly.
_Validation_: `carried-from` is only accepted if it points to the *same ticket's* own `NotDone` entry in a sprint that's already `Closed` — anything else (a different ticket, a `Done`/`Cancelled` source, a still-`Open` source sprint) is rejected. See [ADR-0005](docs/adr/0005-carried-from-validation.md).

**User**:
A structured record for a team member (lead developer or developer) who can be assigned tickets. Replaces free-text assignee names. Carries a `role`, but no permissions are gated on it in MVP — it's descriptive only, reserved for future per-role views.

**Current Status** (of a Ticket):
Never stored directly on `Ticket`. Always computed by reading the `status` of that ticket's most recent `SprintEntry`. Avoids a denormalized field drifting out of sync if a past `SprintEntry` is corrected.

**Close Sprint**:
An explicit, manual action (MVP: performed by a user, not automatic on end date) that locks a sprint's `SprintEntry` rows against further edits and auto-creates carried-over `SprintEntry` rows in the next sprint for any ticket still Not Done. Before this action, a sprint is still "open" even past its end date. See [ADR-0001](docs/adr/0001-ticket-sprintentry-split.md) for why per-sprint history needs to be locked down at all.
_Current state_: the CRUD layer enforces the lock (a `Sprint` set to `Closed` rejects further `SprintEntry` edits) but does not yet implement the auto-carry-over creation — that's a separate follow-up feature, not part of basic entity CRUD. Until then, carrying a `NotDone` ticket forward is a manual step (create its next-sprint `SprintEntry` and set `carried-from`); the Sprint Entries admin page warns when a `NotDone` entry in a `Closed` sprint has no later entry yet, as a backstop against a missed carry-over. See [ADR-0005](docs/adr/0005-carried-from-validation.md).

**Sprint Velocity**:
Total story points across a sprint's **Committed Completed Points** and **Late-Add Completed Points**. Shown as a single number to the team, but always backed by the two-part split underneath.
_Avoid_: Completed Points (ambiguous — doesn't say whether late-adds are included)
_Not the same as_: the dashboard's **Workload**/**Done** figures (`GET /dashboard/sprints`). Those are a deliberately simpler v1 metric — total points committed vs. done, with no Committed/Late-Add split yet. See [ADR-0004](docs/adr/0004-dashboard-v1-simplified-metrics.md).

**Committed Completed Points**:
Story points completed in a sprint, counted only from tickets that were part of the sprint's original commitment (not added after sprint start).

**Late-Add Completed Points**:
Story points completed in a sprint, counted only from tickets added after the sprint started. Tracked separately from Committed Completed Points because Planning Accuracy must exclude them.

**Added Mid-Sprint Points**:
The total story points of tickets added after sprint start. This is a *derived* value — summed from each ticket's "added after sprint start" flag — never entered manually at the sprint level. A manually-typed aggregate with no link back to individual tickets would drift out of sync with reality.

**Cancelled** (ticket status):
A third ticket status alongside Done / Not Done. A ticket that's removed from scope mid-sprint and will never be finished. Excluded from all metrics (velocity, planning accuracy, carry-over) — without this status, an abandoned ticket would count against planning accuracy indefinitely and keep triggering carry-over logic forever. Applies only to the `SprintEntry` it's set on — does not retroactively exclude that ticket's earlier `Not Done` entries from past sprints. See [ADR-0001](docs/adr/0001-ticket-sprintentry-split.md#consequences).
_Avoid_: Not Done (reserved for "still open, still intended to be finished")
