-- Migration 0007: games.outcome (ADR 0057 Decision 7, #1520).
--
-- winner_seat alone cannot tell a draw from an abandoned table: both
-- are NULL. `outcome` says which it was:
--
--   'win'   an ended game with a winner (winner_seat names the seat,
--           unless that player is no longer seated)
--   'draw'  an ended game nobody won (CR 104.4: every remaining
--           player lost at once)
--   NULL    unknown — every row written before this migration, a game
--           an admin closed, or one that has not ended
--
-- A plain ADD COLUMN with no default and no CHECK, like 0004's host
-- columns: existing rows read NULL, which is "unknown / abandoned",
-- and a value this binary does not know is carried, not refused.

ALTER TABLE games ADD COLUMN outcome TEXT;
