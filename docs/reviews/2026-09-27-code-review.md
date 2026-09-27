# Code Review — 2026-09-27

Full-repo Standards review (no fixed diff point, no spec — see
[CODING_STANDARDS.md](../../CODING_STANDARDS.md) for rule definitions).
Spec axis skipped: no spec/issue existed for this pass.

Checked items are done; leave unchecked ones as-is until fixed so the record
stays accurate.

## Backend (Go)

- [ ] CRUD handler pattern (decode → validate → call service → respond)
      repeated near-verbatim across `user_handler.go`, `project_handler.go`,
      `ticket_handler.go`, `sprint_handler.go`, `sprintentry_handler.go`.
      Same shape in `user_repository.go`, `project_repository.go`,
      `ticket_repository.go`, `sprint_repository.go` (Create/Get/List/
      Update/Delete → wrap error). — violates **B1**
- [ ] `sprintentry_service.go:82` — 6 positional params where the request
      struct already exists at the handler layer. — violates **B2**
- [ ] `sprintentry_handler.go:67-70` — casts a raw query string directly to
      `model.EntryStatus(v)` with no validation; an invalid value silently
      filters to zero rows instead of erroring. — violates **B3**
- [ ] `sprintentry_service.go:114-142` (`Update`) — never checks
      `carriedFrom != id` (self-reference). — violates **B4**
- [ ] `config.go:10-15` — `DatabaseURL` defaults to `""` with no validation
      in `Load()`; fails loud only later, via `RunMigrations`/`NewPostgres`
      in `main.go`. — violates **B5**

Not flagged as issues (documented as intentional / already correct):
`cors.go:10`'s `Access-Control-Allow-Origin: *` (explicit no-auth MVP
choice — revisit if auth is added); dropped encode error in
`common.go:15` (standard Go/net-http practice); parameterized SQL
throughout (**B6**, no injection risk found); error mapping in `pgerr.go`
matches ADR-0008 exactly (**B7**); ADR-0005 and ADR-0007 both correctly
implemented — no contradictions found.

## Frontend (TypeScript / Next.js)

- [ ] Five admin CRUD pages (`app/admin/{users,projects,sprints,tickets,
      sprint-entries}/page.tsx`) each reimplement an identical form-dialog +
      list-fetch + upsert + delete shape (e.g. `users/page.tsx:44-57` vs.
      `projects/page.tsx:45-58` vs. `sprints/page.tsx:58-72`). ~400
      duplicated lines; any change to error/pending-state UX currently means
      editing all 5 files. — violates **F1**
- [ ] Three near-identical developer/workload/done tables
      (`dashboard/page.tsx:83-121`, `dashboard/[sprintId]/page.tsx:82-113`,
      `entries/page.tsx:55-96`), plus an ad-hoc find-by-id-or-fallback-label
      helper reimplemented in nearly every admin page. — violates **F2**
- [ ] `lib/api.ts` — `getJSON` (115-121) throws a plain `Error` with no
      status, while `requestJSON` (123-137) throws a typed `ApiError`.
      — violates **F3**
- [ ] `dashboard/[sprintId]/page.tsx:22-26` and `entries/page.tsx:111-115` —
      `catch { notFound(); }` renders a backend 500 identically to "sprint
      doesn't exist," a direct consequence of the **F3** gap above.
      — violates **F4**
- [ ] `dashboard/page.tsx:48` — labels `workloadPoints` as "committed
      points," but CONTEXT.md's precise **Committed Points at Sprint Start**
      and ADR-0004 both describe this dashboard metric as a simpler v1
      figure without that split. — violates **P4**

Not flagged as issues (below rule-of-three, or already correct): the
`sprintId`-parse-then-`notFound()` block duplicated across only 2 files;
`tickets`/`sprints`/`projects` traveling together as props through
`sprint-entries/page.tsx`. Sprint Health components
(`components/dashboard/sprint-health-*.tsx`) correctly match
CONTEXT.md/ADR-0010/ADR-0011 — no misuse found.
