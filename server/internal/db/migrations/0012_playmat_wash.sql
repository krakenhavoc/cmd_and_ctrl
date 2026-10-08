-- Migration 0012: users.playmat_wash (ADR 0128 amendment, the
-- owner-set wash).
--
--   users.playmat_wash  how strongly the owner's playmat is darkened
--                       under the cards, a percentage from 30 to 90,
--                       or NULL for the default (58, the scrim ADR 0128
--                       shipped with). Everyone at the table sees the
--                       owner's choice.
--
-- A plain ADD COLUMN with no default and no CHECK, like 0004, 0007,
-- 0010 and 0011; the range is the service's (internal/playmat). No
-- backfill: every user row written before this reads NULL, the
-- default.
--
-- Rolling back: an older binary refuses this schema (ErrSchemaTooNew).
-- The column is additive and can be dropped by hand; see
-- docs/environments.md.

ALTER TABLE users ADD COLUMN playmat_wash INTEGER;
