# Coding Standards

Rules for this repo, derived from its actual patterns (not a generic style guide).
Each rule is tagged **(hard)** — a real violation, flag it as a defect — or
**(judgement call)** — a Fowler-style smell heuristic: worth raising, not
worth blocking on. A future code review should cite the rule ID (e.g. `B1`,
`F3`) when it finds a breach.

Nothing here duplicates what tooling already enforces — this repo currently
has no lint config (`.eslintrc`, `.golangci.yml`) in either stack, so these
rules aren't backed by automation yet.

## Principles

- **P1 (hard)** — No hardcoded secrets or API keys. Use the existing
  `.env`/`.env.example` pattern.
- **P2 (judgement call)** — Extract shared logic once it repeats a 4th time
  nearly verbatim (rule of three, generously applied). Don't abstract on the
  first or second duplication — that's speculative generality in reverse.
- **P3 (hard)** — Fail loud. Never swallow an error in a way that erases a
  distinction the caller needs (e.g. "the resource doesn't exist" vs. "the
  request failed") — surface the real failure mode.
- **P4 (judgement call)** — When labeling something with a term defined in
  [CONTEXT.md](CONTEXT.md), only use that exact term if the value matches its
  precise definition there. If it's a looser or simplified metric, name it
  after its own field, not the precise domain term.

## Backend (Go)

- **B1 (judgement call)** — Once a CRUD handler (decode → validate → call
  service → respond) or repository (Create/Get/List/Update/Delete → wrap
  error) pattern repeats a 4th+ time nearly verbatim across entities, extract
  a shared generic helper rather than adding another copy.
- **B2 (judgement call)** — Pass the request/DTO struct that already exists
  at the handler layer through to the service layer, rather than exploding it
  into positional parameters.
- **B3 (hard)** — Validate enum-typed query/path parameters against their
  allowed values before using them in a query. An invalid value must error,
  not silently filter to zero rows.
- **B4 (hard)** — When an entity can reference itself by ID (e.g. a
  "carried from" pointer), validate against self-reference before persisting.
- **B5 (judgement call)** — Validate required configuration (e.g. a database
  URL) eagerly in `Load()`. Don't let a missing value surface later as a
  lower-level connection error.
- **B6 (hard)** — All SQL must be parameterized. No string-built queries
  incorporating request input.
- **B7 (hard)** — Map repository/driver errors to domain sentinel errors at
  the repository boundary (per [ADR-0008](docs/adr/0008-validation-and-error-mapping-boundary.md)).
  Don't leak raw driver errors to callers.

## Frontend (TypeScript / Next.js)

- **F1 (judgement call)** — Once an admin CRUD page (form dialog + list
  fetch + upsert + delete) repeats a 4th+ time nearly verbatim, extract a
  shared hook/component rather than adding another copy.
- **F2 (judgement call)** — Once the same table shape, or the same
  id-to-label lookup, repeats 3+ times across pages, extract one shared
  table component / one `lookupLabel(list, id, fallback)` utility instead of
  reimplementing it per page.
- **F3 (hard)** — Every error thrown from the API layer must carry a
  status/type (e.g. `ApiError`), never a plain `Error`. Callers rely on this
  to distinguish failure modes.
- **F4 (hard)** — Don't collapse every fetch failure into `notFound()`. Only
  call it for an actual absent-resource response; other failures must
  surface distinctly.

## Preserving what's already correct

These aren't new rules, but call out patterns already in place that any
change should keep intact:

- Sprint Health computation matches [ADR-0010](docs/adr/0010-project-sprint-health.md) /
  [ADR-0011](docs/adr/0011-sprint-health-aggregates-across-projects.md) exactly — keep it that way.
- `carried-from` validation matches [ADR-0005](docs/adr/0005-carried-from-validation.md).
- The single-open-`SprintEntry`-per-ticket invariant ([ADR-0007](docs/adr/0007-single-open-sprintentry-per-ticket.md))
  is deliberately enforced only by a DB trigger, not duplicated in Go — don't
  add a service-side check for it.
