-- Migration 0014: bot_model_budget (ADR 0033, amendment 2026-10-12,
-- #2980): the hosted-model spend guard's counters.
--
--   bot_model_budget  one row per counter. key is "day:<YYYY-MM-DD>"
--                     or "month:<YYYY-MM>" (UTC) for tokens spent on
--                     hosted model calls in that period, or "disabled"
--                     (1 when an admin switched model calls off).
--                     updated_at is Unix MILLISECONDS (UTC), matching
--                     0002 to 0013. The writer deletes the rows of
--                     every other period on each save, so the table
--                     holds one day, one month and the switch.
--
-- Rolling back: an older binary refuses this schema (ErrSchemaTooNew).
-- The table is additive and standalone; drop it by hand after
-- restoring the old binary. See docs/environments.md.

CREATE TABLE bot_model_budget (
    key        TEXT PRIMARY KEY,
    value      INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);
