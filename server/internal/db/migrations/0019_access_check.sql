-- Migration 0019: the sign-up gate's last check (ADR 0156 §1 item 7,
-- S71, #3179).
--
--   users.access_checked_at  the Unix millisecond time Discord last
--                            answered whether this person holds the
--                            community server's admitting role
--                            (Elder Dragon Highlander on the C2
--                            server), or NULL
--                            for never (the gate was off, or the
--                            person is an exempt admin).
--   users.access_result      that answer: 'admitted', 'not_member' or
--                            'no_role'; NULL for never ("unchecked" in
--                            the admin view). An outage writes nothing,
--                            so it never overwrites a real answer.
--
-- A refused sign-in writes nothing at all (ADR 0156 §1 item 4), so a
-- refused result is only ever recorded for someone who already had an
-- account, by the re-check at renewal.
--
-- Every *_at column is Unix MILLISECONDS (UTC), matching 0002.
--
-- Additive: two nullable columns, no backfill. Rolling back: an older
-- binary refuses this schema (ErrSchemaTooNew). Both columns can be
-- dropped by hand; see docs/environments.md.

ALTER TABLE users ADD COLUMN access_checked_at INTEGER;
ALTER TABLE users ADD COLUMN access_result TEXT
    CHECK (access_result IN ('admitted', 'not_member', 'no_role'));
