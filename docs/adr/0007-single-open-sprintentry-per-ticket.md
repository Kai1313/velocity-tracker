# A ticket can have at most one SprintEntry in an open sprint, enforced at the database layer

A ticket being actively planned in two open sprints at once doesn't make sense — per [CONTEXT.md](../../CONTEXT.md), "Current Status" is read from a ticket's *most recent* `SprintEntry`, and workload/done totals assume each ticket has one live entry at a time. This invariant is real, but it was never written down anywhere: not in `SprintEntryService.Create`/`Update` (`backend/internal/service/sprintentry_service.go`), not in CONTEXT.md's `SprintEntry` entry, not in any prior ADR. The only place it actually exists is a `BEFORE INSERT` trigger in the initial schema migration (`backend/internal/repository/migrations/0001_init_schema.up.sql`):

```sql
CREATE TRIGGER trg_single_open_sprintentry_per_ticket
BEFORE INSERT ON sprint_entry
FOR EACH ROW EXECUTE FUNCTION enforce_single_open_sprintentry_per_ticket();
```

`enforce_single_open_sprintentry_per_ticket()` raises an exception if the ticket already has a `SprintEntry` in another sprint whose status is `Open`. `pgerr.go` (see [ADR-0008](0008-validation-and-error-mapping-boundary.md)) maps that `RAISE EXCEPTION` onto `apperr.ErrConflict`, so a violation reaches the API as a 409, not a 400 or a 500.

This ADR doesn't change the behavior — it records the decision that already exists in code, so it's discoverable without reading migration SQL.

## Consequences

- The invariant is enforced only on `INSERT`, not `UPDATE` — moving an existing `SprintEntry` to point at a different `Open` sprint via `PUT` is not currently checked by this trigger. Not a problem today (`ticketId`/`sprintId` are fixed at creation per the Admin UI, see [README.md](../../README.md#admin-ui)), but worth re-checking if that constraint ever loosens.
- If this rule needs to change (e.g. to support a ticket intentionally split across two concurrent sprints), the change belongs in this trigger, not in the Go service layer — the service layer currently assumes the database will catch this and doesn't duplicate the check.
