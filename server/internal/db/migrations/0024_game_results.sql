-- Migration 0024: a table's opt-in to the Discord results feed (ADR
-- 0174 §1, #3322).
--
--   games.results   1 when the host has opted the table in to a results
--                   post in the community server's #game-results channel
--                   (POST /games "results", PUT /games/{id}/results), 0
--                   otherwise. The post is made once, when the game
--                   ends, so nothing else is stored: no message id.
--
-- Additive: one defaulted column, no backfill. It holds no personal
-- data. Rolling back: an older binary refuses this schema
-- (ErrSchemaTooNew). The column can be dropped by hand; see
-- docs/environments.md.

ALTER TABLE games ADD COLUMN results INTEGER NOT NULL DEFAULT 0
    CHECK (results IN (0, 1));
