DROP TRIGGER IF EXISTS trg_single_open_sprintentry_per_ticket ON sprint_entry;
DROP FUNCTION IF EXISTS enforce_single_open_sprintentry_per_ticket();
DROP TABLE IF EXISTS sprint_entry;
DROP TABLE IF EXISTS sprint;
DROP TABLE IF EXISTS ticket;
DROP TABLE IF EXISTS users;
