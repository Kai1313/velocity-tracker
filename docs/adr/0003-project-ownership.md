# Project is a lightweight ownership tag on Ticket, not on Sprint

We added a `Project` entity so tickets can be organized by what they belong to. The question was where it attaches: to `Ticket`, to `Sprint`, or both.

We chose `Ticket.project_id` (mandatory, `NOT NULL`) and left `Sprint` project-agnostic. This is one team running one continuous set of sprints — see [ADR-0002](0002-go-backend.md) and [README](../../README_velocity_tracker.md) for the single-team framing — and that team's sprints routinely mix work from more than one project. Forcing a `Sprint` to belong to a single `Project` would either fragment sprints per project (contradicting "one team, one sprint cadence") or leave the field meaningless for mixed sprints.

`Project` also carries a `status` (`Active` / `Archived`) rather than being deletable outright once referenced. This follows the same history-preservation stance as [ADR-0001](0001-ticket-sprintentry-split.md): a `Ticket` row must always be able to resolve its `project_id`, so a `Project` with existing tickets is archived, never deleted.

## Consequences

- Every `Ticket` must be created with a valid, existing `project_id` — there is no "uncategorized" bucket.
- Sprint-level and velocity-level reporting cannot filter or group by project without joining through `Ticket`; `Sprint` itself carries no project information.
- Deleting a `Project` is only possible while it has zero tickets; retiring a project with history means setting it `Archived` instead.
