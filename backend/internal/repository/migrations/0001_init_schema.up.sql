CREATE TABLE users (
  id         SERIAL PRIMARY KEY,
  name       TEXT NOT NULL,
  role       TEXT NOT NULL CHECK (role IN ('Lead','Developer'))
);

CREATE TABLE ticket (
  id           SERIAL PRIMARY KEY,
  title        TEXT NOT NULL,
  story_points INTEGER NOT NULL CHECK (story_points > 0),
  assignee_id  INTEGER REFERENCES users(id)
);

CREATE TABLE sprint (
  id         SERIAL PRIMARY KEY,
  name       TEXT NOT NULL,
  start_date DATE NOT NULL,
  end_date   DATE NOT NULL,
  status     TEXT NOT NULL DEFAULT 'Open' CHECK (status IN ('Open','Closed'))
);

CREATE TABLE sprint_entry (
  id                        SERIAL PRIMARY KEY,
  ticket_id                 INTEGER NOT NULL REFERENCES ticket(id),
  sprint_id                 INTEGER NOT NULL REFERENCES sprint(id),
  status                    TEXT NOT NULL CHECK (status IN ('Done','NotDone','Cancelled')),
  added_after_sprint_start  BOOLEAN NOT NULL DEFAULT false,
  carried_from              INTEGER REFERENCES sprint_entry(id),
  points_at_entry           INTEGER NOT NULL CHECK (points_at_entry > 0),
  created_at                TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (ticket_id, sprint_id)
);

CREATE INDEX idx_sprint_entry_ticket_id ON sprint_entry(ticket_id);
CREATE INDEX idx_sprint_entry_sprint_id ON sprint_entry(sprint_id);

CREATE OR REPLACE FUNCTION enforce_single_open_sprintentry_per_ticket()
RETURNS TRIGGER AS $$
BEGIN
  IF (SELECT status FROM sprint WHERE id = NEW.sprint_id) = 'Open'
     AND EXISTS (
       SELECT 1 FROM sprint_entry se
       JOIN sprint s ON s.id = se.sprint_id
       WHERE se.ticket_id = NEW.ticket_id AND s.status = 'Open'
     )
  THEN
    RAISE EXCEPTION 'ticket % already has a SprintEntry in an open sprint', NEW.ticket_id;
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_single_open_sprintentry_per_ticket
BEFORE INSERT ON sprint_entry
FOR EACH ROW EXECUTE FUNCTION enforce_single_open_sprintentry_per_ticket();
