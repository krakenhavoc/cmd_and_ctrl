-- Migration 0026: seats.stand_in, seats.stand_in_since (ADR 0177 §4,
-- #3337).
--
--   seats.stand_in        the bot tier playing a person's seat while
--                         they are away ('heuristic'), or NULL while the
--                         person plays it. The seat stays theirs:
--                         bot_tier stays NULL, so it is never read as a
--                         bot seat.
--   seats.stand_in_since  when that bot took the seat, Unix
--                         MILLISECONDS (UTC), matching 0002. NULL with
--                         stand_in.
--
-- A column, not the engine snapshot, because the bot drives the seat
-- from outside the game and the game itself does not change: a restart
-- brings the stand-in back with the table's other bots. The game's log
-- keeps the hand-over and the take-back as lines, so a later reader
-- (the statistics, #3370) can tell which games a bot played part of.
--
-- Neither column names a person: the seat's person is seats.user_id.
--
-- Additive: two nullable columns, no backfill. Rolling back: an older
-- binary refuses this schema (ErrSchemaTooNew). Drop them by hand; see
-- docs/environments.md.

ALTER TABLE seats ADD COLUMN stand_in TEXT;
ALTER TABLE seats ADD COLUMN stand_in_since INTEGER;
