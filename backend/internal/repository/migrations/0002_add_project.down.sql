DROP INDEX IF EXISTS idx_ticket_project_id;
ALTER TABLE ticket DROP COLUMN IF EXISTS project_id;
DROP TABLE IF EXISTS project;
