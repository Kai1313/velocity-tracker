# Sprint Health is a project-level early-warning signal, thresholds are provisional

This automates a report the team already maintained by hand — a running "current sprint health by project" table (referred to in the code as "Chart 42C") comparing how fast a project needs to move against how fast it's actually moving. `GET /dashboard/project-sprint-health` and the `Sprint Health by Project` card on `/dashboard` replace that manual copy with a computed one, sourced from the same data as [ADR-0004](0004-dashboard-v1-simplified-metrics.md)'s workload/done totals.

This is a separate decision from ADR-0004, not a revision of it: ADR-0004 scoped which totals the dashboard ships (workload/done vs. the full Sprint Velocity split); this ADR is about turning those totals into a per-project pace comparison, at a coarser grain (project, not sprint or developer) than anything else on the dashboard.

## What it computes

For every `Active` project with entries in a currently `Open` sprint, one row per (project, sprint) pair:

- **Required Velocity** = (Committed SP − Done SP) ÷ Days Remaining — the pace still needed to finish on time. Null once the sprint is overdue (days remaining ≤ 0), rather than an inflated or infinite number — the point is triggering the alarm, not precision.
- **Achieved Velocity** = Done SP ÷ Days Elapsed. Null on the sprint's first day, since 0 SP/day on day one reads as "the team achieved nothing," which is misleading rather than informative.
- Both velocities use **committed-only** points (excluding Late-Add), so a team finishing late-added tickets while original commitments stall doesn't read as "on pace" just because more points got marked Done. Late-add points are surfaced separately on each row instead, as a visible signal rather than one baked into the ratio.
- Days are calendar days from each row's own sprint dates. A project with tickets split across multiple concurrently-open sprints gets one independent row per sprint rather than a merged countdown — the one-open-sprint-per-*ticket* database trigger doesn't limit how many sprints a project's tickets collectively touch.

**Status**, from achieved ÷ required velocity:
- `OnTrack` — ratio ≥ 1.0, nothing committed left to do, or still the sprint's first day (pace not yet measurable)
- `AtRisk` — ratio between 0.6 and 1.0
- `Critical` — ratio below 0.6, or the sprint is overdue

## Consequences

- **The 1.0/0.6 cutoffs are a placeholder, not a validated decision.** Nothing in the data backs "0.6" specifically — it's a reasonable-sounding line for shipping v1. Revisit it once there's enough real sprint history to check whether teams flagged `AtRisk` actually recover and teams flagged `Critical` actually don't.
- Because it reuses ADR-0004's committed/done totals, extending this to include the Committed/Late-Add split (if that's ever built per ADR-0004) is a query change to `ProjectSprintPoints`, not a redesign of the health calculation.
- Scoped to `Active` projects and `Open` sprints only, matching how the manual report was used — a completed or archived project/sprint has nothing left to warn about.
