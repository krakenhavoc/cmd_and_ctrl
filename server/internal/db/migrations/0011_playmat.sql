-- Migration 0011: users.playmat_id (ADR 0128, playmats).
--
--   users.playmat_id  the uuid of the user's playmat image, the file
--                     <data dir>/playmats/<playmat_id>.jpg, or NULL for
--                     a user with none. One per account.
--
-- The image lives on disk, not in the database (the bugstore pattern),
-- so this column is only the pointer. Replacing or removing a playmat
-- deletes the old file; the server does that, not a trigger.
--
-- A plain ADD COLUMN with no default and no CHECK, like 0004, 0007 and
-- 0010. No backfill: every user row written before this reads NULL.
--
-- Rolling back: an older binary refuses this schema (ErrSchemaTooNew).
-- The column is additive and can be dropped by hand; see
-- docs/environments.md. The files under <data dir>/playmats are then
-- orphans and can be deleted.

ALTER TABLE users ADD COLUMN playmat_id TEXT;
