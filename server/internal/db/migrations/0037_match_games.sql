-- Migration 0037: the games of a match (ADR 0222 §3, #5436).
--
-- A best-of-three is one table, so one games row, one game_stats row
-- and one result per seat: a match counts as one finished table. This
-- table keeps each game of it.
--
--   match_games  one row per finished game of a best-of-three:
--     game_id        the table (the match), as in game_stats
--     game_no        1, 2 or 3
--     outcome        'win' or 'draw'
--     winner_seat    the seat that won the game; NULL for a draw
--     cause          why the game ended (the game outcome's cause)
--     starting_seat  the seat that took the first turn
--     turns          the turn number the game ended on
--
-- A best-of-one writes no rows: its one game is the table, which
-- game_stats already describes.
--
-- The rows hold seat numbers and no name or user id. They are part of
-- the game record the other player keeps (ADR 0165 §5), deleted with
-- the game's statistics, and a seat's person is seats.user_id as for
-- every other statistics table.
--
-- Additive: one new table. Rolling back: an older binary refuses this
-- schema (ErrSchemaTooNew). The table can be dropped by hand; see
-- docs/environments.md.

CREATE TABLE match_games (
    game_id       TEXT NOT NULL REFERENCES game_stats(game_id) ON DELETE CASCADE,
    game_no       INTEGER NOT NULL,
    outcome       TEXT NOT NULL CHECK (outcome IN ('win', 'draw')),
    winner_seat   INTEGER,
    cause         TEXT NOT NULL,
    starting_seat INTEGER NOT NULL,
    turns         INTEGER NOT NULL,
    PRIMARY KEY (game_id, game_no)
);
