-- Migration 0002: games, seats and invites become rows (ADR 0051
-- decision 4, S34 sub-PR 3). Replaces <dataDir>/lobby/<id>.json; the
-- engine's own files (restore/, replays/, games/) stay on disk.
--
-- Every *_at column in this migration is Unix MILLISECONDS (UTC).
-- Seconds would tie two games created in the same second, and the
-- lobby listing orders by created_at.
--
-- Foreign keys that are NOT declared here, on purpose:
--
--   games.created_by, invites.created_by, seats.user_id -> users(id)
--     `users` arrives in S34 sub-PR 2.
--   seats.deck_id -> decks(id)
--     `decks` arrives in S34 sub-PR 5.
--
-- SQLite cannot add a REFERENCES clause to an existing column; it
-- needs a table rebuild. So these are plain TEXT columns for now. They
-- stay NULL until the owning sub-PR lands, and that sub-PR rebuilds
-- the table with the constraint if it wants one enforced. The
-- in-schema references below (seats/invites -> games) are real and
-- enforced (db.Open turns foreign_keys on).

CREATE TABLE games (
    id          TEXT PRIMARY KEY,           -- the engine's Game.ID (uuid)
    name        TEXT NOT NULL,
    created_by  TEXT,                       -- users(id) from sub-PR 2; NULL: admin-created
    state       TEXT NOT NULL,              -- lobby | active | ended
    created_at  INTEGER NOT NULL,
    started_at  INTEGER,
    ended_at    INTEGER,
    archived_at INTEGER,
    winner_seat INTEGER                     -- NULL until the engine reports one
);

CREATE TABLE seats (
    game_id            TEXT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    seat               INTEGER NOT NULL,
    player_id          TEXT NOT NULL,       -- the engine's Player.ID
    user_id            TEXT,                -- users(id) from sub-PR 2; NULL: guest or bot
    guest_name         TEXT,                -- the seat label
    bot_tier           TEXT,                -- NULL: human
    deck_id            TEXT,                -- decks(id) from sub-PR 5; NULL: placeholder or ad-hoc upload
    deck_name          TEXT,
    -- Migration side column (ADR 0051 "Migration" step 1/3): the
    -- Discord snowflake of a seat claimed through OAuth, held until a
    -- users row exists for it. Sub-PR 4 links it on first sign-in and
    -- clears it.
    pending_discord_id TEXT,
    PRIMARY KEY (game_id, seat)
);

-- "My games" (sub-PR 4) and tablemates (sub-PR 6) both start here.
CREATE INDEX seats_user_id ON seats (user_id);
CREATE INDEX seats_pending_discord_id ON seats (pending_discord_id);

CREATE TABLE invites (
    token_hash BLOB PRIMARY KEY,            -- sha256 of the 16 random bytes
    game_id    TEXT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    kind       TEXT NOT NULL,               -- player | spectator
    created_by TEXT,                        -- users(id) from sub-PR 2
    created_at INTEGER NOT NULL,
    expires_at INTEGER,                     -- NULL: never (the default)
    revoked_at INTEGER                      -- NULL: live
);

CREATE INDEX invites_game_id ON invites (game_id);
