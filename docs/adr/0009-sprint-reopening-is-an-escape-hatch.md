# Reopening a Closed sprint is a deliberate, unrestricted escape hatch

`SprintService.Update` (`backend/internal/service/sprint_service.go`) accepts a status change to either `Open` or `Closed` with no transition guard — a `Closed` sprint can be set back to `Open` via a plain `PUT`, at any time, by anyone (there's no auth in MVP, see [ADR-0006](0006-multi-team-readiness.md)). This is intentional, not an oversight: the working case is a `SprintEntry` that got registered under the wrong sprint (e.g. a ticket meant for Sprint 8 was mistakenly logged against Sprint 7) and needs correcting after the fact. `SprintEntryService.Update` locks edits once the parent sprint is `Closed`, so reopening the sprint is the only way to fix a misfiled entry without going around the API by hand.

This is asymmetric with `SprintEntry` itself on purpose: per [CONTEXT.md](../../CONTEXT.md)'s "Close Sprint" entry, a *sprint's own record* (name, dates, status) is not the immutable history that `SprintEntry` rows are — it's closer to a status flag that gates entry edits, and flags are meant to be flippable when a mistake needs fixing.

## Consequences

- No confirmation, audit trail, or restriction exists around reopening today — reopening and re-closing a sprint leaves no record that it happened. Acceptable for the current single-admin, no-auth MVP; worth revisiting (e.g. an audit log, or restricting reopen to some role) if multi-user access control ever lands (see [ADR-0006](0006-multi-team-readiness.md)).
- Reopening does not undo anything a `Closed` state may imply elsewhere in the app beyond unlocking `SprintEntry` edits — there is no auto-carry-over to unwind, since auto-carry-over creation isn't implemented yet (see [ADR-0005](0005-carried-from-validation.md)).
- If a future change makes Close Sprint auto-create carry-over `SprintEntry` rows in the next sprint, reopening the source sprint afterward will need a decision about what happens to those already-created rows — out of scope here, since auto-carry-over doesn't exist yet.
