-- Migration 0038: a signed-in person's finished drafts (ADR 0222 §9,
-- #5440).
--
-- A draft lives in memory and is gone when it closes. What is kept of
-- it is one person's own look back at it: the set, the date, the cards
-- they picked and the deck they built, and how their matches went. A
-- guest's draft keeps nothing; only a signed-in seat writes a row.
--
--   draft_history  one row per person per draft:
--     draft_id    the draft's id (drafts have no table of their own)
--     user_id     the person
--     draft_name  the draft's name, which passed the name filter
--     set_code    the set drafted ("blb"), and set_name its name
--     mode        'pick_one' or 'pick_two'
--     seats       the size of the draft
--     best_of     the length of its matches, 1 or 3
--     seat        the person's seat number
--     drafted_at  when the row was written, Unix milliseconds
--     picks       JSON: the cards the person picked, in the order taken,
--                 each with whether their deck played it
--     basics      JSON: the basic lands their deck added, by name
--
--   draft_history_matches  one row per round the person played:
--     round             1 to 3
--     opponent_kind     'person', 'guest', 'bot' or 'bye'
--     opponent_user_id  the opponent's account when it was a signed-in
--                       person; NULL otherwise, and set NULL when that
--                       person deletes their account
--     result            'win', 'loss' or 'draw'; '' for a bye
--     games_won, games_lost  the games of the match
--
-- No name of another person is stored: an opponent is shown by the
-- name their account has when the page is read.
--
-- ADR 0149: both tables are personal data. Account deletion deletes
-- the person's rows and clears opponent_user_id wherever it names them
-- (accountdelete.Inventory, drafthistory.ForgetAccount).
--
-- Additive: two new tables. Rolling back: an older binary refuses this
-- schema (ErrSchemaTooNew). The tables can be dropped by hand; see
-- docs/environments.md.

CREATE TABLE draft_history (
    draft_id   TEXT NOT NULL,
    user_id    TEXT NOT NULL REFERENCES users(id),
    draft_name TEXT NOT NULL,
    set_code   TEXT NOT NULL,
    set_name   TEXT NOT NULL,
    mode       TEXT NOT NULL,
    seats      INTEGER NOT NULL,
    best_of    INTEGER NOT NULL,
    seat       INTEGER NOT NULL,
    drafted_at INTEGER NOT NULL,
    picks      TEXT NOT NULL,
    basics     TEXT NOT NULL,
    PRIMARY KEY (draft_id, user_id)
);

CREATE INDEX draft_history_by_user ON draft_history (user_id, drafted_at DESC);

CREATE TABLE draft_history_matches (
    draft_id         TEXT NOT NULL,
    user_id          TEXT NOT NULL,
    round            INTEGER NOT NULL,
    opponent_kind    TEXT NOT NULL CHECK (opponent_kind IN ('person', 'guest', 'bot', 'bye')),
    opponent_user_id TEXT,
    result           TEXT NOT NULL CHECK (result IN ('win', 'loss', 'draw', '')),
    games_won        INTEGER NOT NULL,
    games_lost       INTEGER NOT NULL,
    PRIMARY KEY (draft_id, user_id, round),
    FOREIGN KEY (draft_id, user_id) REFERENCES draft_history (draft_id, user_id) ON DELETE CASCADE
);

CREATE INDEX draft_history_matches_by_opponent ON draft_history_matches (opponent_user_id);
