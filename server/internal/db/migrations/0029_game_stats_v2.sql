-- Migration 0029: game statistics, version 2 (ADR 0217 §1, #4880).
--
--   game_stats         gains frames_complete (the replay was read whole),
--                      win_kind, win_turn and combo_rule (§3).
--   seat_stats         gains turn_position, deck_hash, out_kind and
--                      out_by_seat, the turn, land-drop, screw, flood and
--                      colour-screw counts (§4), the attack share's sums,
--                      commander casts and life lost to combat.
--   seat_turn_stats    one row per seat and own turn: land drops, the
--                      hand and mana at the start of main, attacks.
--   seat_target_stats  one row per seat and opponent seat: attacks,
--                      damage, combat damage and commander damage. It
--                      names a seat number, never a person.
--   seat_card_seen     one row per seat and card name that was ever in
--                      its hand, cast or not. Its "cast" column is quoted
--                      everywhere: CAST is an SQL keyword.
--
-- As in 0023: no names of people and no user ids. A seat's person is
-- seats.user_id, through the join on (game_id, seat). Every row is
-- derived from the game and its replay and can be thrown away and
-- computed again (gamestats.Version, cmd/statsbackfill). The new tables
-- go with their seat line, and so with their game (ON DELETE CASCADE).
--
-- Additive: new nullable columns and new tables. A version-1 record
-- reads NULL in every new column until the backfill computes it again.
-- Rolling back: an older binary refuses this schema (ErrSchemaTooNew);
-- the tables and columns can be dropped by hand, see docs/environments.md.

ALTER TABLE game_stats ADD COLUMN frames_complete INTEGER CHECK (frames_complete IN (0, 1));
ALTER TABLE game_stats ADD COLUMN win_kind TEXT CHECK (win_kind IN
    ('combo', 'concession', 'win_effect', 'poison', 'commander_damage', 'combat', 'noncombat', 'decking'));
ALTER TABLE game_stats ADD COLUMN win_turn INTEGER;
ALTER TABLE game_stats ADD COLUMN combo_rule TEXT CHECK (combo_rule IN ('loop', 'repeats', 'actions', 'combats'));

ALTER TABLE seat_stats ADD COLUMN turn_position INTEGER;
ALTER TABLE seat_stats ADD COLUMN deck_hash TEXT;
ALTER TABLE seat_stats ADD COLUMN out_kind TEXT CHECK (out_kind IN
    ('combo', 'concession', 'win_effect', 'poison', 'commander_damage', 'combat', 'noncombat', 'decking'));
ALTER TABLE seat_stats ADD COLUMN out_by_seat INTEGER;
ALTER TABLE seat_stats ADD COLUMN own_turns INTEGER;
ALTER TABLE seat_stats ADD COLUMN land_drops_hit INTEGER;
ALTER TABLE seat_stats ADD COLUMN land_drops_missed INTEGER;
ALTER TABLE seat_stats ADD COLUMN screw_turns INTEGER;
ALTER TABLE seat_stats ADD COLUMN flood_turns INTEGER;
ALTER TABLE seat_stats ADD COLUMN colour_screw_turns INTEGER;
ALTER TABLE seat_stats ADD COLUMN screw_game INTEGER CHECK (screw_game IN (0, 1));
ALTER TABLE seat_stats ADD COLUMN flood_game INTEGER CHECK (flood_game IN (0, 1));
ALTER TABLE seat_stats ADD COLUMN colour_screw_game INTEGER CHECK (colour_screw_game IN (0, 1));
ALTER TABLE seat_stats ADD COLUMN creatures_eligible INTEGER;
ALTER TABLE seat_stats ADD COLUMN creatures_attacked INTEGER;
ALTER TABLE seat_stats ADD COLUMN commander_casts INTEGER;
ALTER TABLE seat_stats ADD COLUMN commander_first_cast_turn INTEGER;
ALTER TABLE seat_stats ADD COLUMN life_lost_to_combat INTEGER;

CREATE TABLE seat_turn_stats (
    game_id           TEXT NOT NULL,
    seat              INTEGER NOT NULL,
    own_turn          INTEGER NOT NULL,
    turn              INTEGER NOT NULL,
    land_played       INTEGER NOT NULL,
    land_drops        INTEGER NOT NULL,
    lands_controlled  INTEGER NOT NULL,
    mana_sources      INTEGER NOT NULL,
    hand_size         INTEGER NOT NULL,
    lands_in_hand     INTEGER NOT NULL,
    spells_in_hand    INTEGER NOT NULL,
    missed            INTEGER NOT NULL CHECK (missed IN (0, 1)),
    screw             INTEGER NOT NULL CHECK (screw IN (0, 1)),
    flood             INTEGER NOT NULL CHECK (flood IN (0, 1)),
    colour_screw      INTEGER NOT NULL CHECK (colour_screw IN (0, 1)),
    attackers         INTEGER NOT NULL,
    eligible          INTEGER NOT NULL,
    damage_to_players INTEGER NOT NULL,
    spells_cast       INTEGER NOT NULL,
    PRIMARY KEY (game_id, seat, own_turn),
    FOREIGN KEY (game_id, seat) REFERENCES seat_stats(game_id, seat) ON DELETE CASCADE
) WITHOUT ROWID;

CREATE TABLE seat_target_stats (
    game_id          TEXT NOT NULL,
    seat             INTEGER NOT NULL,
    target_seat      INTEGER NOT NULL,
    attacks          INTEGER NOT NULL,
    damage           INTEGER NOT NULL,
    combat_damage    INTEGER NOT NULL,
    commander_damage INTEGER NOT NULL,
    PRIMARY KEY (game_id, seat, target_seat),
    FOREIGN KEY (game_id, seat) REFERENCES seat_stats(game_id, seat) ON DELETE CASCADE
) WITHOUT ROWID;

CREATE TABLE seat_card_seen (
    game_id          TEXT NOT NULL,
    seat             INTEGER NOT NULL,
    name             TEXT NOT NULL,
    copies           INTEGER NOT NULL,
    in_opening_hand  INTEGER NOT NULL CHECK (in_opening_hand IN (0, 1)),
    first_seen_turn  INTEGER NOT NULL,
    held_turns       INTEGER NOT NULL,
    affordable_turns INTEGER NOT NULL,
    "cast"           INTEGER NOT NULL CHECK ("cast" IN (0, 1)),
    PRIMARY KEY (game_id, seat, name),
    FOREIGN KEY (game_id, seat) REFERENCES seat_stats(game_id, seat) ON DELETE CASCADE
) WITHOUT ROWID;
