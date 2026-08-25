# Velocity Tracker

A real app for tracking sprint velocity, planning accuracy, and late-add rate for a single software team. See [CONTEXT.md](CONTEXT.md) for the domain glossary (Ticket, SprintEntry, Project, Close Sprint, etc.) and [docs/adr/](docs/adr/) for the architectural decisions behind it. `README_velocity_tracker.md` is the original product sketch this app is built from.

## Stack

- **Frontend**: Next.js 15 / TypeScript, Tailwind CSS, [shadcn/ui](https://ui.shadcn.com) components, [next-themes](https://github.com/pacocoursey/next-themes) for dark mode (`frontend/`) — requires **Node 18.18+** (Next.js 15 will not run on Node 16)
- **Backend**: Go, stdlib `net/http` (Go 1.22+ routing patterns), raw `pgx` for Postgres access — no ORM, no router framework (`backend/`)
- **Database**: PostgreSQL, migrated via `golang-migrate` (embedded, runs automatically on backend startup)

## Running locally

```bash
docker compose up
```

This starts Postgres, the Go backend (`:8080`), and the Next.js frontend (`:3000`). Copy `.env.example` to `.env` first if you don't already have one — the defaults work as-is for local dev.

## Running the backend outside Docker

The backend can be run directly on the host (`go run ./cmd/server` from `backend/`) against the Postgres container:

```bash
docker compose up postgres
cd backend
DATABASE_URL="postgres://velo:velo@localhost:5432/velocity_tracker?sslmode=disable" go run ./cmd/server
```

Note the hostname difference: `.env`'s `DATABASE_URL` uses `postgres` (the Docker Compose service name), which only resolves *inside* the Compose network. Anything running on the host — including tests — needs `localhost` instead.

## Testing the backend

```bash
cd backend
go test ./...
```

Unit tests (service-layer business rules) run with no setup. Integration tests (repository layer, anything resting on DB constraints/triggers) need a real Postgres reachable via `TEST_DATABASE_URL`; they skip automatically if it's unset:

```bash
docker compose up postgres
cd backend
TEST_DATABASE_URL="postgres://velo:velo@localhost:5432/velocity_tracker?sslmode=disable" go test ./...
```

## Running the frontend

```bash
cd frontend
npm install
npm run dev
```

Needs Node 18.18+ (Next.js 15's minimum). If you have `nvm`, run `nvm use 18` (or newer) first. `BACKEND_INTERNAL_URL` (defaults to `http://localhost:8080`) tells the frontend where to reach the backend for its server-side data fetches; `NEXT_PUBLIC_BACKEND_URL` (also defaults to `http://localhost:8080`) is the browser-reachable equivalent, used by client-side mutations (the Admin UI's create/edit/delete calls) since those run in the browser, not on the Next.js server.

Note: `npx shadcn@latest add <component>` (mentioned below) requires Node 20+ and will crash on Node 18 — it's a separate requirement from the Node 18.18+ the app itself needs.

## API

All endpoints are JSON, `camelCase` in and out, no authentication in MVP. CRUD is provided for `User`, `Project`, `Ticket`, `Sprint`, and `SprintEntry` — see the handler packages under `backend/internal/handler/` for the exact routes.

**Dashboard** (read-only, computed server-side):
- `GET /dashboard/sprints` — workload/done story points per sprint, across all sprints
- `GET /dashboard/sprints/{id}` — workload/done story points per developer, within one sprint

See [ADR-0004](docs/adr/0004-dashboard-v1-simplified-metrics.md) for why these are simpler "workload vs. done" totals rather than the full Sprint Velocity/Planning Accuracy model from [CONTEXT.md](CONTEXT.md).

Not yet implemented: the Close Sprint auto-carry-over action (see [CONTEXT.md](CONTEXT.md#close-sprint)), and the Committed/Late-Add split, Planning Accuracy, and Late-Add Rate calculations.

## Dashboard UI

`/dashboard` shows all sprints with workload/done totals; clicking a sprint links to `/dashboard/{sprintId}` for its per-developer breakdown. Built with [shadcn/ui](https://ui.shadcn.com) (`Card`/`Table`/`Badge`) — component source lives in `frontend/components/ui/`, not an npm package, so it can be edited directly. To add more shadcn components later: `npx shadcn@latest add <component>` from `frontend/`.

Both dashboard pages share a header (`frontend/app/dashboard/layout.tsx`) with a theme toggle in the top-right corner. It cycles Light → Dark → System; System follows the OS preference and is the default on first visit, with the chosen theme persisted across visits. The header also links to `/admin/users` for data entry.

## Admin UI

`/admin` is where `User`, `Project`, `Ticket`, `Sprint`, and `SprintEntry` records actually get created — there's no other way to populate the app's data short of calling the API directly. `/admin/users`, `/admin/projects`, `/admin/tickets`, and `/admin/sprints` are implemented; `SprintEntry` is planned as follow-up work in the same pattern (it ties a `Ticket` and a `Sprint` together, so it needs both of the above to exist first — see [ADR-0001](docs/adr/0001-ticket-sprintentry-split.md)).

The Ticket form's Project/Assignee fields and the Sprint form's dates work as you'd expect from the API shapes: Assignee is nullable (an "Unassigned" option, not a required field), and dates round-trip between the `<input type="date">` UI (`YYYY-MM-DD`) and the backend's RFC3339 `time.Time` JSON via small converters in `app/admin/sprints/page.tsx`.

Each entity gets a list page (`Card`/`Table`, matching the dashboard's look) with dialog-based create/edit forms and a confirm step before delete. `frontend/components/admin/delete-confirm-button.tsx` is shared across entities; the create/edit dialogs are not, since their fields differ enough per entity that a shared abstraction isn't worth it yet.

_Current state_: `/admin` uses a left sidebar (`frontend/app/admin/layout.tsx`) linking between entity sections, with the theme toggle in its footer. This is a minimal, flat nav shell scoped to the MVP — it has no access control (consistent with the API having none — see below) and isn't meant to be the final admin IA once the app grows past five entities or gains real users/roles.
