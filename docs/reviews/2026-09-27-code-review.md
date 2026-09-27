# Code Review — 2026-09-27

Full-repo Standards review (no fixed diff point, no spec — see
[CODING_STANDARDS.md](../../CODING_STANDARDS.md) for rule definitions).
Spec axis skipped: no spec/issue existed for this pass.

Checked items are done; leave unchecked ones as-is until fixed so the record
stays accurate.

## Backend (Go)

- [x] CRUD handler pattern (decode → validate → call service → respond)
      repeated near-verbatim across `user_handler.go`, `project_handler.go`,
      `ticket_handler.go`, `sprint_handler.go`, `sprintentry_handler.go`.
      Same shape in `user_repository.go`, `project_repository.go`,
      `ticket_repository.go`, `sprint_repository.go` (Create/Get/List/
      Update/Delete → wrap error). — violates **B1**. Fixed: the handler
      half only — generic `handleCreate`/`handleGet`/`handleList`/
      `handleUpdate`/`handleDelete` helpers (`crud.go`) using Go generics,
      applied to all 5 handlers. Repository-layer duplication is
      deliberately left alone: a generic SQL CRUD layer is a separate,
      riskier change with less payoff (each repo's columns are a real
      domain difference, not pure boilerplate).
- [x] `sprintentry_service.go:82` — 6 positional params where the request
      struct already exists at the handler layer. — violates **B2**.
      Fixed: added `CreateSprintEntryInput`/`UpdateSprintEntryInput`, passed
      through from the handler instead of exploded params.
- [x] `sprintentry_handler.go:67-70` — casts a raw query string directly to
      `model.EntryStatus(v)` with no validation; an invalid value silently
      filters to zero rows instead of erroring. — violates **B3**.
      Fixed: `parseSprintEntryFilter` now rejects an invalid `status` value
      with `apperr.ErrValidation` (422), matching the existing enum-validation
      pattern in `validateEntryFields`.
- [x] `sprintentry_service.go:114-142` (`Update`) — never checks
      `carriedFrom != id` (self-reference). — violates **B4**.
      Fixed: `Update` now rejects a self-referencing `carriedFrom` before
      calling `validateCarriedFrom`.
- [x] `config.go:10-15` — `DatabaseURL` defaults to `""` with no validation
      in `Load()`; fails loud only later, via `RunMigrations`/`NewPostgres`
      in `main.go`. — violates **B5**.
      Fixed: `Load()` now returns an error when `DATABASE_URL` is unset;
      `main.go` fails fast on it.

Not flagged as issues (documented as intentional / already correct):
`cors.go:10`'s `Access-Control-Allow-Origin: *` (explicit no-auth MVP
choice — revisit if auth is added); dropped encode error in
`common.go:15` (standard Go/net-http practice); parameterized SQL
throughout (**B6**, no injection risk found); error mapping in `pgerr.go`
matches ADR-0008 exactly (**B7**); ADR-0005 and ADR-0007 both correctly
implemented — no contradictions found.

## Frontend (TypeScript / Next.js)

- [x] Five admin CRUD pages (`app/admin/{users,projects,sprints,tickets,
      sprint-entries}/page.tsx`) each reimplement an identical form-dialog +
      list-fetch + upsert + delete shape (e.g. `users/page.tsx:44-57` vs.
      `projects/page.tsx:45-58` vs. `sprints/page.tsx:58-72`). ~400
      duplicated lines; any change to error/pending-state UX currently means
      editing all 5 files. — violates **F1**. Fixed: `useEntityList` (list
      fetch/upsert/remove) applied to Users/Projects/Sprints/Tickets, and
      `useFormDialogState` (open/pending/error/submit) applied to all 5
      dialogs including Sprint Entries. Sprint Entries' page-level
      filtering/carry-over/orphan-detection logic is untouched — too
      bespoke to fit either hook without a leaky abstraction.
- [x] Three near-identical developer/workload/done tables
      (`dashboard/page.tsx:83-121`, `dashboard/[sprintId]/page.tsx:82-113`,
      `entries/page.tsx:55-96`), plus an ad-hoc find-by-id-or-fallback-label
      helper reimplemented in nearly every admin page. — violates **F2**.
      Fixed: `lookupLabel` utility replacing the repeated find-by-id
      helpers (tickets/sprint-entries pages); shared `WorkloadDoneTable`
      component replacing the duplicated Developer/Workload/Done/Tickets
      table on `dashboard/[sprintId]/page.tsx` and `entries/page.tsx` (the
      two literally-identical instances — `dashboard/page.tsx`'s table has
      a different shape, an extra Progress column and per-sprint not
      per-developer rows, so it was left as its own thing).
- [x] `lib/api.ts` — `getJSON` (115-121) throws a plain `Error` with no
      status, while `requestJSON` (123-137) throws a typed `ApiError`.
      — violates **F3**. Fixed: `getJSON` now throws `ApiError` with the
      response status, same as `requestJSON`.
- [x] `dashboard/[sprintId]/page.tsx:22-26` and `entries/page.tsx:111-115` —
      `catch { notFound(); }` renders a backend 500 identically to "sprint
      doesn't exist," a direct consequence of the **F3** gap above.
      — violates **F4**. Fixed: both now call `notFound()` only when the
      caught error is `ApiError` with `status === 404`, otherwise rethrow.
      Verified manually (backend stopped → HTTP 500 with an error digest,
      not the not-found page) and via a new Playwright regression test for
      the genuine-404 path.
- [x] `dashboard/page.tsx:48` — labels `workloadPoints` as "committed
      points," but CONTEXT.md's precise **Committed Points at Sprint Start**
      and ADR-0004 both describe this dashboard metric as a simpler v1
      figure without that split. — violates **P4**. Fixed: reworded to
      "workload" language matching the field's own name.

Not flagged as issues (below rule-of-three, or already correct): the
`sprintId`-parse-then-`notFound()` block duplicated across only 2 files;
`tickets`/`sprints`/`projects` traveling together as props through
`sprint-entries/page.tsx`. Sprint Health components
(`components/dashboard/sprint-health-*.tsx`) correctly match
CONTEXT.md/ADR-0010/ADR-0011 — no misuse found.

## Pre-existing e2e issues found during verification (out of scope, not fixed)

Discovered while running the full Playwright suite to verify the F3/F4 fix.
Neither relates to anything changed in this pass — both predate it and are
left for a separate fix:

- `dashboard.spec.ts:45` — the "clicking sprint workload…" test asserts
  `getByText('5', { exact: true })` against a card that renders
  `{points} <span>pts</span>`; the accessibility tree merges these into
  `"5 pts"`, so no element ever has the exact text `"5"`. Test-assertion bug,
  not a product bug.
- `sprint-entries.spec.ts:35` and `:61` — both use
  `page.locator('#entry-ticket').selectOption(...)`, but `#entry-ticket` is
  now a custom `Combobox` component (see commit `31ea4b9`, "implement
  Combobox component"), not a native `<select>`; `selectOption()` only works
  against native selects. The spec was never updated for that migration.
