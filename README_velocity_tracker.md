# Velocity Tracker

A simple project to track sprint velocity for a small software team using story points.

This tracker is designed for teams that work in 1-2 week sprints, estimate tasks with story points, and want a practical way to measure completed work per sprint and calculate average team velocity.

## Purpose

This project helps a team:

- Track completed story points in each sprint.
- Separate committed work from work added after sprint start.
- Handle carry-over tickets correctly.
- Measure sprint velocity and rolling average velocity.
- See planning accuracy and late-added work.

## Core Concepts

### Story Points

Story points represent relative effort or complexity.

Example scale:

- 1 = very small
- 2 = small
- 3 = medium
- 5 = large
- 8 = very large

### Sprint Velocity

Sprint velocity is the total story points of tickets that are fully done by the end of a sprint.

Important rules:

- Only count tickets that are fully done.
- Partially completed tickets count as 0 for that sprint.
- Carry-over tickets count in the sprint where they are actually completed.
- Tickets added after sprint start may still count toward completed velocity if they are finished before sprint end, but they should be tracked separately.

### Average Team Velocity

Average team velocity is the average completed points across several sprints.

Formula:

\[
\text{Average Velocity} = \frac{\text{Sprint 1 Velocity} + \text{Sprint 2 Velocity} + ... + \text{Sprint N Velocity}}{N}
\]

A practical baseline usually becomes useful after at least 3-5 sprints of data.

## Recommended Data Model

### Ticket Log

Each ticket can contain:

- Ticket ID
- Title
- Sprint
- Assignee
- Story Points
- Status
- Carry-over flag
- Added after sprint start flag
- Done date
- Notes

### Sprint Log

Each sprint can contain:

- Sprint name
- Start date
- End date
- Committed points at sprint start
- Added mid-sprint points
- Completed points
- Carry-over points not done
- Planning accuracy
- Late-add rate
- Notes

## Suggested Calculation Rules

### 1. Completed Points

At sprint end:

- Sum all story points for tickets marked Done in that sprint.
- This is the sprint velocity.

Example:

- Ticket A done = 3 points
- Ticket B done = 3 points
- Ticket C done = 2 points
- Sprint velocity = 8 points

### 2. Average Velocity

After multiple sprints:

- Add all sprint velocities.
- Divide by number of sprints.

Example:

- Sprint 1 = 8
- Sprint 2 = 12
- Sprint 3 = 10
- Average velocity = (8 + 12 + 10) / 3 = 10

### 3. Planning Accuracy

This shows how much of the originally planned work was actually completed.

Formula:

\[
\text{Planning Accuracy} = \frac{\text{Completed Planned Points}}{\text{Committed Points at Sprint Start}} \times 100
\]

### 4. Late-Add Rate

This shows how much work was added after the sprint already started.

Formula:

\[
\text{Late Add Rate} = \frac{\text{Added Mid-Sprint Points}}{\text{Committed Points at Sprint Start}} \times 100
\]

## How Carry-Over Should Work

If a ticket was started in the previous sprint but not finished:

- Do not count it in the previous sprint's velocity.
- Count it only in the sprint where it becomes Done.
- Keep the original story points unless the ticket was truly split or re-scoped.

Example:

- Sprint 1: Ticket X has 5 points, not finished -> counts as 0
- Sprint 2: Ticket X becomes Done -> counts as 5 in Sprint 2

## Suggested Features for the Project

For a simple app, these features are enough:

- Create and manage sprints.
- Add tickets to a sprint.
- Mark whether a ticket is carry-over.
- Mark whether a ticket was added after sprint start.
- Update ticket status to Done or Not Done.
- Show sprint summary totals.
- Show rolling average velocity across recent sprints.
- Show planning accuracy and late-add rate.

## Suggested UI Sections

### 1. Sprint Dashboard

Show:

- Sprint name
- Sprint date range
- Total committed points
- Total completed points
- Added mid-sprint points
- Carry-over not done
- Current sprint velocity

### 2. Ticket Table

Columns:

- Ticket ID
- Title
- Assignee
- Story Points
- Status
- Carry-over
- Added late
- Done date

### 3. Metrics Panel

Show:

- Current sprint velocity
- Last 3 sprint velocities
- Rolling average velocity
- Planning accuracy
- Late-add rate

## Suggested Logic for Implementation

A simple implementation approach:

1. Create a sprint record.
2. Save all planned tickets at sprint start.
3. Allow new tickets to be added later, but mark them as late-add.
4. At sprint end, calculate completed points from only Done tickets.
5. Save sprint totals into a sprint history table.
6. Recalculate rolling average from recent sprint history.

## Suggested Tech Direction

For a lightweight internal tool, a simple stack is enough:

- Frontend: Next.js or React
- Backend: Node.js/Express or Go
- Database: PostgreSQL or SQLite for a lightweight version
- Optional: chart library for velocity trend visualization

A minimal first version can even be:

- One page
- Local database or JSON seed
- CRUD for sprint and tickets
- Automatic metric calculation

## Example Sprint Summary Output

| Sprint | Committed | Added Mid-Sprint | Completed | Carry-Over Not Done | Velocity |
|---|---:|---:|---:|---:|---:|
| Sprint 1 | 15 | 3 | 8 | 2 | 8 |
| Sprint 2 | 12 | 4 | 10 | 1 | 10 |
| Sprint 3 | 14 | 2 | 12 | 0 | 12 |

Average velocity for the 3 sprints:

\[
\frac{8 + 10 + 12}{3} = 10
\]

## Practical Notes

- Keep sprint duration consistent if possible.
- Use one definition of Done for the whole team.
- Track completed work separately from commitment.
- Track mid-sprint additions separately so velocity does not become misleading.
- Expect the first few sprints to be noisy.
- Use rolling average, not a single sprint, for planning future work.

## MVP Checklist

- Sprint CRUD
- Ticket CRUD
- Story point input
- Done / Not Done status
- Carry-over flag
- Added-after-start flag
- Sprint summary calculation
- Average velocity calculation
- Basic history view

## Future Enhancements

- Velocity chart per sprint
- Per-developer contribution view
- Filters by team or project
- Export to CSV or Excel
- Jira import
- Forecast next sprint capacity using historical average

## Development Note

This project should optimize for clarity over process purity.

The main goal is not to enforce perfect Scrum. The main goal is to give the team a usable and honest view of:

- how many story points are actually being completed,
- how much work is being added after sprint start,
- and how reliable sprint planning is over time.
