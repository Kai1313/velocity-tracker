# Velocity Tracker

A real app for tracking sprint velocity, planning accuracy, and late-add rate for a single software team. See [CONTEXT.md](CONTEXT.md) for the domain glossary (Ticket, SprintEntry, Project, Close Sprint, etc.) and [docs/adr/](docs/adr/) for the architectural decisions behind it. `README_velocity_tracker.md` is the original product sketch this app is built from.

## Stack

- **Frontend**: Next.js / TypeScript (`frontend/`)
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

## API

All endpoints are JSON, `camelCase` in and out, no authentication in MVP. CRUD is provided for `User`, `Project`, `Ticket`, `Sprint`, and `SprintEntry` — see the handler packages under `backend/internal/handler/` for the exact routes.

Not yet implemented: the Close Sprint auto-carry-over action (see [CONTEXT.md](CONTEXT.md#close-sprint)), and all velocity/planning-accuracy/late-add-rate reporting — both are deferred follow-ups on top of this CRUD layer.
