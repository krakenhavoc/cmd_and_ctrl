-- Migration 0021: a table's post on the Discord #lfg board (ADR 0161
-- §6, #3319).
--
--   games.lfg             1 when the host has opted the table in to
--                         #lfg (POST /games "lfg", PUT
--                         /games/{id}/lfg), 0 otherwise. The choice
--                         survives a restart; the post itself may not
--                         exist yet, or any more.
--   games.lfg_message_id  the Discord message id of the table's live
--                         post, or NULL for none. The boot reconcile
--                         keeps it for a resident lobby table that is
--                         still opted in, and otherwise deletes the
--                         message and clears the column. It never
--                         touches a message no row names: #lfg is a
--                         shared channel.
--
-- Additive: one defaulted column and one nullable column, no backfill.
-- Neither holds personal data. Rolling back: an older binary refuses
-- this schema (ErrSchemaTooNew). Both columns can be dropped by hand;
-- see docs/environments.md.

ALTER TABLE games ADD COLUMN lfg INTEGER NOT NULL DEFAULT 0
    CHECK (lfg IN (0, 1));
ALTER TABLE games ADD COLUMN lfg_message_id TEXT;
