-- Migration 0020: per-table Discord voice rooms (ADR 0159 §5, #3223).
--
--   games.voice_room        1 when the host has asked for a voice room
--                           (POST /games "voice_room", PUT
--                           /games/{id}/voice), 0 otherwise. The
--                           choice survives a restart; the room itself
--                           may not exist yet, or any more.
--   games.voice_channel_id  the Discord channel id of the table's live
--                           room, or NULL for none. The boot reconcile
--                           adopts it if the channel still exists in the
--                           category, and deletes every channel there
--                           that no table claims.
--
-- Additive: one defaulted column and one nullable column, no backfill.
-- Rolling back: an older binary refuses this schema (ErrSchemaTooNew).
-- Both columns can be dropped by hand; see docs/environments.md.

ALTER TABLE games ADD COLUMN voice_room INTEGER NOT NULL DEFAULT 0
    CHECK (voice_room IN (0, 1));
ALTER TABLE games ADD COLUMN voice_channel_id TEXT;
