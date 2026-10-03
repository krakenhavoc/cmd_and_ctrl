-- Migration 0009: admin mode (ADR 0112 §2, S57). The sprint's only
-- schema change.
--
--   users.admin_mode_at  the Unix millisecond time this person switched
--                        admin mode on, or 0 for off (player mode).
--
-- An allowlisted person (CMDCTRL_DISCORD_ADMIN_USER_IDS) is an admin
-- only while admin mode is on, and it lapses 12 hours after this time
-- (users.AdminModeTTL). The default is 0, so every existing row starts
-- in player mode, the owner's included: admin rights are something a
-- person turns on.
--
-- Every *_at column is Unix MILLISECONDS (UTC), matching 0002 to 0008.
--
-- Rolling back: an older binary refuses this schema (ErrSchemaTooNew).
-- The column is additive and can be dropped by hand; see
-- docs/environments.md.

ALTER TABLE users ADD COLUMN admin_mode_at INTEGER NOT NULL DEFAULT 0;
