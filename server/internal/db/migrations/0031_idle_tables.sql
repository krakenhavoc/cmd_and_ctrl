-- Migration 0031: closing idle tables (ADR 0211 §1, §4, #4562).
--
--   games.last_active_at  when the table was last used: the newest
--                         lobby touch or human move the idle sweep saw,
--                         Unix MILLISECONDS (UTC), matching 0002. NULL
--                         until the sweep first writes it, which is
--                         every row that exists when this ships; such a
--                         table's clock starts when it is next loaded.
--   games.closed_reason   why the server closed the table: idle_lobby
--                         (an unstarted table nobody touched),
--                         idle_game (a started game nobody moved in) or
--                         idle_bots (a game left to its bots). NULL for
--                         an open table and for an admin's archive;
--                         unarchiving clears it.
--
-- The ADR names this migration 0029; that number went to game
-- statistics version 2 first, so it is 0031.
--
-- Additive: two nullable columns, no backfill. Neither names a person.
-- Rolling back: an older binary refuses this schema (ErrSchemaTooNew).
-- The columns can be dropped by hand; see docs/environments.md.

ALTER TABLE games ADD COLUMN last_active_at INTEGER;
ALTER TABLE games ADD COLUMN closed_reason TEXT
    CHECK (closed_reason IN ('idle_lobby', 'idle_game', 'idle_bots'));
