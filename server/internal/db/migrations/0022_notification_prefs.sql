-- Migration 0022: Discord DM notification preferences (ADR 0162 §5,
-- #3320).
--
--   notification_prefs   one row per person who has saved the
--                        Notifications section: three switches, each
--                        off (0) by default, and when Discord last
--                        refused a DM to them.
--
--     turn          "Your turn": the game has waited on you for 2
--                   minutes while you are away.
--     table_ready   a table you sit at has filled or started while you
--                   are away.
--     table_filled  a table you host has filled while you are away.
--     dm_failed_at  the Unix millisecond time Discord answered 50007
--                   ("Cannot send messages to this user"), or NULL.
--                   While it is set nothing is sent; turning a switch
--                   on clears it.
--
-- Its own table rather than the user_settings body: ADR 0110 §4 keeps
-- that body opaque to the server, and the notifier has to read these.
-- No row means every switch off. Account deletion (ADR 0149) deletes
-- the row.
--
-- Every *_at column is Unix MILLISECONDS (UTC), matching 0002.
--
-- Additive: one new table, no backfill. Rolling back: an older binary
-- refuses this schema (ErrSchemaTooNew). The table can be dropped by
-- hand; see docs/environments.md.

CREATE TABLE notification_prefs (
    user_id      TEXT PRIMARY KEY REFERENCES users(id),
    turn         INTEGER NOT NULL DEFAULT 0 CHECK (turn IN (0, 1)),
    table_ready  INTEGER NOT NULL DEFAULT 0 CHECK (table_ready IN (0, 1)),
    table_filled INTEGER NOT NULL DEFAULT 0 CHECK (table_filled IN (0, 1)),
    dm_failed_at INTEGER,
    updated_at   INTEGER NOT NULL
);
