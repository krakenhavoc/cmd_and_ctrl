-- Migration 0023: game statistics (ADR 0165 §3, #3364).
--
--   game_stats       one row per ended game: the definitions' version,
--                    where the record came from (live | replay), whether
--                    the event log it was derived from was complete, the
--                    turn and round counts and the starting seat.
--   seat_stats       one row per seat of that game: result, place,
--                    elimination, life and damage totals, spells, draws,
--                    attacks, tokens, mulligans and commander names.
--   seat_card_stats  one row per seat and card name it cast: how often,
--                    and on which turn first.
--
-- No names of people and no user ids: a seat's person is seats.user_id,
-- through the join on (game_id, seat). Every row is derived from the
-- game and can be thrown away and computed again (gamestats.Version).
-- All three go with their game (ON DELETE CASCADE).
--
-- Additive: three new tables, no backfill (cmd/statsbackfill, #3365).
-- Rolling back: an older binary refuses this schema (ErrSchemaTooNew).
-- The tables can be dropped by hand; see docs/environments.md.

CREATE TABLE game_stats (
    game_id       TEXT PRIMARY KEY REFERENCES games(id) ON DELETE CASCADE,
    stats_version INTEGER NOT NULL,
    source        TEXT NOT NULL CHECK (source IN ('live', 'replay')),
    log_complete  INTEGER NOT NULL CHECK (log_complete IN (0, 1)),
    turns         INTEGER NOT NULL,
    rounds        INTEGER NOT NULL,
    starting_seat INTEGER,
    computed_at   INTEGER NOT NULL
);

CREATE TABLE seat_stats (
    game_id                  TEXT NOT NULL REFERENCES game_stats(game_id) ON DELETE CASCADE,
    seat                     INTEGER NOT NULL,
    result                   TEXT NOT NULL CHECK (result IN ('win', 'loss', 'draw', 'none')),
    place                    INTEGER,
    eliminated_turn          INTEGER,
    conceded                 INTEGER NOT NULL CHECK (conceded IN (0, 1)),
    final_life               INTEGER NOT NULL,
    life_gained              INTEGER NOT NULL,
    life_lost                INTEGER NOT NULL,
    damage_to_players        INTEGER NOT NULL,
    combat_damage_to_players INTEGER NOT NULL,
    commander_damage_dealt   INTEGER NOT NULL,
    commander_damage_taken   INTEGER NOT NULL,
    spells_cast              INTEGER NOT NULL,
    cards_drawn              INTEGER NOT NULL,
    attacks                  INTEGER NOT NULL,
    tokens_created           INTEGER NOT NULL,
    mulligans                INTEGER NOT NULL,
    commanders               TEXT NOT NULL,  -- json array of names
    PRIMARY KEY (game_id, seat)
);

CREATE TABLE seat_card_stats (
    game_id         TEXT NOT NULL,
    seat            INTEGER NOT NULL,
    name            TEXT NOT NULL,
    scryfall_id     TEXT,
    casts           INTEGER NOT NULL,
    first_cast_turn INTEGER NOT NULL,
    PRIMARY KEY (game_id, seat, name),
    FOREIGN KEY (game_id, seat) REFERENCES seat_stats(game_id, seat) ON DELETE CASCADE
);

CREATE INDEX seat_card_stats_name ON seat_card_stats (name);
