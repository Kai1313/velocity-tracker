# Multi-team readiness: the schema is prepared, auth is not

This is a readiness assessment, not a decision to build multi-team support now — [CONTEXT.md](../../CONTEXT.md) already scopes that as an explicit Future Enhancement, not MVP. This ADR records what a future migration would actually cost, so that judgment doesn't have to be re-derived from scratch when the question comes up again.

The team's real usage as of this writing (13 `Project`s reading as client engagements, all routing through the same 2 `Sprint`s) confirms the app is still growing in project *count*, not team *count* — the dimension [ADR-0003](0003-project-ownership.md) was built for. Multi-team is not a current need. This ADR is about what happens if that changes later.

## What's already in place

[ADR-0003](0003-project-ownership.md) deliberately kept `Sprint` independent of `Project` — "a sprint's tickets can span multiple projects" — for a completely different reason (one team, many clients) than multi-team support. That decision happens to leave `Sprint` as the natural seam for a future `Team` boundary: a `team_id` on `Sprint` (and likely `User`) would scope the team-relevant data without touching `Project` or `Ticket` at all.

The service/repository layering (interfaces like `SprintEntryRepository`, `SprintLookup`) keeps SQL isolated from handlers, so adding a `WHERE team_id = ?` to the currently-unscoped list/dashboard queries (`GET /sprint-entries`, `GET /dashboard/sprints`, etc.) is mechanical and bounded — roughly 7 known repository files, not a rewrite. Existing single-team data backfills as "Team 1" with no breakage.

## What isn't already in place

There is currently **no authentication or authorization anywhere in the app** — CORS is wide open by design (see `backend/internal/handler/cors.go`), and no request carries a caller identity. `User.role` exists but "no permissions are gated on it in MVP" (CONTEXT.md).

This matters because multi-team support without access control isn't multi-team — it's a shared data pool with a cosmetic filter that anyone can bypass by calling the API directly with a different `team_id`. Unlike the schema changes above, auth is a cross-cutting change (middleware, session/token handling, threading a caller identity into every repository call) that has to exist *before* team boundaries mean anything real. It does not get cheaper by deferring it, and it is not something the current architecture buys for free the way the `Sprint`/`Project` decoupling does.

## Consequences

- If multi-team is scoped later, expect the data-model and query changes to be cheap (additive columns, bounded query changes) but expect auth to be a full, separate effort budgeted on its own — not a byproduct of the schema migration.
- Building real auth *before* multi-team is needed is not wasted work if it happens anyway (per-role views are already anticipated in `CONTEXT.md`'s `User` entry) — so auth is worth prioritizing on its own merits, independent of whether multi-team ever ships.
