# Sprint Health aggregates across all projects, not per project

The original spec author flagged that [ADR-0010](0010-project-sprint-health.md)'s per-`(project, sprint)` grain distorts the metric: developers on this team work across multiple `Project`s within the same sprint, so their capacity is shared, not partitioned. A project showing "behind pace" in isolation may just be the one a developer spent less time on that week — the per-project number reads as a signal about that project when it's really a signal about how time was allocated across all of them.

This is not a multi-team feature. [ADR-0006](0006-multi-team-readiness.md) already confirms the app has exactly one team; "Team/Product" language from the original request describes that single team's shared capacity, not a new grouping entity. Introducing a `Team` table with a many-`Project`-to-one-`Team` relationship would model a dimension that, today, only ever has one member — schema and migration cost for nothing a query change doesn't already buy.

## What changed

- `ProjectSprintPoints`/`ProjectSprintHealth` (model, repository, service) are renamed to `SprintPoints`/`SprintHealth` and drop `ProjectID`/`ProjectName` entirely. The repository query groups only by `sprint.id` (still joining `project` to filter `status = 'Active'`), summing committed/done/late-add points across every active project's entries in that sprint.
- `GET /dashboard/project-sprint-health` is renamed to `GET /dashboard/sprint-health`.
- The dashboard's "Sprint Health by Project" card/table drops the Project Name column and its title, becoming "Sprint Health" — one row per currently `Open` sprint (typically 1-2, per ADR-0006's usage snapshot) instead of one row per active project.
- No project-level breakdown is kept anywhere in this feature, not even as a collapsed/secondary view — the complaint was that the per-project number actively misleads, not merely that it's redundant. A project-level diagnostic view, if ever needed, is a separate future feature built on its own terms.

## Consequences

- The velocity formulas and `OnTrack`/`AtRisk`/`Critical` thresholds from ADR-0010 are unchanged — only the grouping key moved from project to sprint.
- If this app ever grows a second real team, `SprintPoints`/`SprintHealth` will need a `team_id` dimension reintroduced (this time backed by a real `Team` entity, per ADR-0006's readiness assessment) rather than reverting to per-project rows.
