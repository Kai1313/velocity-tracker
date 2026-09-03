# Validation lives at the database boundary, not in a frontend schema library

Two related conventions govern how bad input is caught and reported, both undocumented until now beyond inline code comments.

**Postgres error codes are mapped by operation, not just by code.** `wrapMutationErr` (`backend/internal/repository/pgerr.go`) turns a `ForeignKeyViolation` into `apperr.ErrValidation` on insert/update ("you pointed at something that doesn't exist") but into `apperr.ErrConflict` on delete ("something still points at this"). The same Postgres error code means two different things to an API caller depending on which direction the reference was broken from — a validation problem to fix on the request versus a conflict to resolve before the delete can proceed. `UniqueViolation` maps to `ErrConflict`; `CheckViolation`/`NotNullViolation` map to `ErrValidation`; a custom trigger's `RaiseException` (see [ADR-0007](0007-single-open-sprintentry-per-ticket.md)) also maps to `ErrConflict`, since it's signaling a state conflict, not malformed input.

**The frontend has no client-side validation library.** No `zod`, `react-hook-form`, `yup`, or equivalent appears in `frontend/package.json`. Every Admin UI form submits to the backend and surfaces whatever `ApiError` comes back (`frontend/lib/api.ts`) rather than re-implementing the backend's validation rules in JavaScript first. Combined with the backend having no ORM (raw `pgx`, see [ADR-0002](0002-go-backend.md)), this keeps exactly one place — the database's constraints plus `pgerr.go`'s mapping — as the source of truth for what's valid, instead of two implementations that could drift apart.

## Consequences

- A new constraint (check, unique, FK, or trigger) added to the schema is automatically enforced and correctly classified (validation vs. conflict) with no Go or TypeScript changes required, as long as it fits one of the cases `wrapMutationErr` already handles.
- Forms give no inline "this field is invalid" feedback before submit — every validation error is a round trip to the backend. Acceptable for the current admin-only, no-auth MVP scale; worth revisiting if form complexity or user base grows enough that round-trip latency on every mistake becomes a real usability cost.
- Adding a genuinely new *kind* of constraint (one `wrapMutationErr`'s existing `switch` doesn't cover) silently falls through to the raw Postgres error via `wrapReadErr`, not a clean `ErrValidation`/`ErrConflict` — a case worth checking whenever a new migration adds an unfamiliar constraint type.
