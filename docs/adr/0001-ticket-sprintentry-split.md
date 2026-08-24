# Ticket and SprintEntry are separate entities

We considered modeling a ticket as a single row that gets edited in place as it moves between sprints (e.g. updating its `Sprint` field when carried over), but rejected it because it destroys per-sprint history: once a row is overwritten, the fact that it was planned-but-not-done in an earlier sprint is gone, and metrics like planning accuracy and carry-over-not-done can no longer be computed reliably after the fact.

Instead, `Ticket` holds stable attributes (id, title, story points), and `SprintEntry` holds one immutable-once-closed row per sprint a ticket appears in (status, added-after-start flag, and a `carried-from` reference to the prior sprint's entry when carried over). A ticket's "current status" is never stored directly — it's always computed from its latest `SprintEntry` — so there's a single source of truth per sprint and no risk of a denormalized field drifting out of sync.

## Consequences

Cancelling a ticket does **not** retroactively change earlier `SprintEntry` rows for that ticket. If a ticket was Not Done in Sprint 1 (a real planning miss at the time) and is later Cancelled in Sprint 3, Sprint 1's entry still counts as a planning miss — only the Cancelled entry itself is excluded from metrics. This follows directly from immutability: each sprint's numbers must reflect what was actually true when that sprint closed, otherwise a team could "clean up" past planning-accuracy history by cancelling old tickets after the fact.
