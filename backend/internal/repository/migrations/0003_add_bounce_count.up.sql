-- NULL means "bounces not tracked" (every entry that predates this feature),
-- which is different from 0 ("tracked, and QA never sent it back").
ALTER TABLE sprint_entry
  ADD COLUMN bounce_count INTEGER CHECK (bounce_count >= 0);

-- Tracking starts with whatever sprint is open right now; closed sprints are
-- locked history, so they stay NULL.
UPDATE sprint_entry SET bounce_count = 0
WHERE sprint_id IN (SELECT id FROM sprint WHERE status = 'Open');

-- Set after the backfill so only entries created from now on default to 0.
ALTER TABLE sprint_entry ALTER COLUMN bounce_count SET DEFAULT 0;
