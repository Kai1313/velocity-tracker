CREATE TABLE project (
  id     SERIAL PRIMARY KEY,
  name   TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'Active' CHECK (status IN ('Active','Archived'))
);

ALTER TABLE ticket
  ADD COLUMN project_id INTEGER REFERENCES project(id);

ALTER TABLE ticket
  ALTER COLUMN project_id SET NOT NULL;

CREATE INDEX idx_ticket_project_id ON ticket(project_id);
