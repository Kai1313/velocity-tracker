# SprintEntry.carriedFrom is validated, but carry-over creation stays manual

Auditing real data surfaced that `carriedFrom` was, until now, an unchecked foreign key: `SprintEntryService.Create`/`Update` accepted any existing `SprintEntry` id, and the Admin UI's "Carried from" dropdown listed every entry in the system regardless of ticket, status, or sprint. Nothing enforced that a `carriedFrom` reference actually denoted a carry-over — the same ticket's own `NotDone` entry from a sprint that has already closed, per [CONTEXT.md](../../CONTEXT.md)'s definition of `SprintEntry`. One real entry happened to be linked correctly, but only because it was entered carefully by hand.

`SprintEntryService` now rejects a `carriedFrom` unless the referenced entry:

- belongs to the same ticket,
- has status `NotDone`, and
- belongs to a sprint whose status is `Closed`.

The Admin UI's "Carried from" dropdown is scoped to the same rules, and auto-selects the single matching candidate when there's exactly one (still changeable) — the common case is one prior `NotDone` entry per ticket. The Sprint Entries page also surfaces a standing warning listing any `NotDone` entry in a `Closed` sprint that has no later entry for that ticket yet, so a missed carry-over doesn't sit silently undetected.

This does **not** add automatic carry-over creation on Close Sprint. [CONTEXT.md](../../CONTEXT.md)'s "Close Sprint" entry already flags that as a separate follow-up feature, and current real-data volume (one carry-over across the whole dataset so far) doesn't justify it yet. Revisit automation if a typical sprint close leaves more than a handful of `NotDone` tickets, or if sprints start getting closed before the next one exists (auto-carry-over needs a target sprint to write into).

## Consequences

- Creating or editing a `SprintEntry` with an invalid `carriedFrom` now fails with a validation error instead of silently persisting a misleading link.
- The orphaned-carry-over warning is a read-only signal on the Sprint Entries page, not a blocker — it does not prevent closing a sprint or creating entries.
- If carry-over volume grows, the manual workflow (and this warning as its safety net) is the first thing to revisit before the volume makes missed carry-overs common.
