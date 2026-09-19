-- Migration 0003: people (ADR 0051 decisions 2 and 5, S34 sub-PR 2).
--
-- A user is our own uuid. A Discord account is one identity attached
-- to it, so a second provider later is a new `provider` value and
-- nothing else. Every *_at column is Unix MILLISECONDS (UTC), the same
-- unit migration 0002 uses.
--
-- Then the three tables from 0002 that point at a person are rebuilt
-- so that pointer is a real foreign key:
--
--   games.created_by, invites.created_by, seats.user_id -> users(id)
--
-- seats.deck_id stays a plain column. `decks` arrives in sub-PR 5,
-- which rebuilds `seats` again to enforce it.
--
-- SQLite cannot add REFERENCES to an existing column, so each table is
-- rebuilt with the documented procedure
-- (https://sqlite.org/lang_altertable.html#otheralter): create the new
-- shape under a temporary name, copy every row, drop the old table,
-- rename the new one into place, recreate its indexes. The migration
-- runner has already turned foreign_keys OFF on this connection and
-- opened the transaction, and runs PRAGMA foreign_key_check before it
-- commits (see applyMigration in migrations.go for why both are
-- needed). Nothing in 0002 ever wrote created_by or user_id, so every
-- copied value is NULL and the check has nothing to find; if it did
-- find something, the migration rolls back and the boot stops.

CREATE TABLE users (
    id                      TEXT PRIMARY KEY,       -- uuid, minted by us
    display_name            TEXT NOT NULL,
    avatar_url              TEXT,                   -- same-origin /avatars/... path; NULL: none
    created_at              INTEGER NOT NULL,
    last_seen_at            INTEGER NOT NULL,
    -- Decision 6. Nothing reads this until sub-PR 7.
    sessions_invalid_before INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE identities (
    provider      TEXT NOT NULL,                    -- 'discord' today
    subject       TEXT NOT NULL,                    -- the provider's stable id (a Discord snowflake)
    user_id       TEXT NOT NULL REFERENCES users(id),
    display_name  TEXT NOT NULL,                    -- as the provider last reported it
    avatar_hash   TEXT,
    -- AES-256-GCM under CMDCTRL_IDENTITY_KEY (decision 5): a key-version
    -- byte, a 12-byte nonce, then ciphertext and tag. NULL when the key
    -- is not configured: the token is discarded, never stored in clear.
    refresh_token BLOB,
    scopes        TEXT NOT NULL,
    linked_at     INTEGER NOT NULL,
    refreshed_at  INTEGER,
    PRIMARY KEY (provider, subject),
    UNIQUE (user_id, provider)
);

-- games ------------------------------------------------------------

CREATE TABLE games_new (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    created_by  TEXT REFERENCES users(id),  -- NULL: admin-created
    state       TEXT NOT NULL,              -- lobby | active | ended
    created_at  INTEGER NOT NULL,
    started_at  INTEGER,
    ended_at    INTEGER,
    archived_at INTEGER,
    winner_seat INTEGER
);

INSERT INTO games_new (id, name, created_by, state, created_at, started_at, ended_at, archived_at, winner_seat)
    SELECT id, name, created_by, state, created_at, started_at, ended_at, archived_at, winner_seat FROM games;

DROP TABLE games;
ALTER TABLE games_new RENAME TO games;

-- seats ------------------------------------------------------------

CREATE TABLE seats_new (
    game_id            TEXT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    seat               INTEGER NOT NULL,
    player_id          TEXT NOT NULL,
    user_id            TEXT REFERENCES users(id),  -- NULL: guest or bot
    guest_name         TEXT,
    bot_tier           TEXT,
    deck_id            TEXT,                        -- decks(id) from sub-PR 5; unenforced until then
    deck_name          TEXT,
    pending_discord_id TEXT,
    PRIMARY KEY (game_id, seat)
);

INSERT INTO seats_new (game_id, seat, player_id, user_id, guest_name, bot_tier, deck_id, deck_name, pending_discord_id)
    SELECT game_id, seat, player_id, user_id, guest_name, bot_tier, deck_id, deck_name, pending_discord_id FROM seats;

DROP TABLE seats;
ALTER TABLE seats_new RENAME TO seats;

CREATE INDEX seats_user_id ON seats (user_id);
CREATE INDEX seats_pending_discord_id ON seats (pending_discord_id);

-- invites ----------------------------------------------------------

CREATE TABLE invites_new (
    token_hash BLOB PRIMARY KEY,
    game_id    TEXT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    kind       TEXT NOT NULL,
    created_by TEXT REFERENCES users(id),
    created_at INTEGER NOT NULL,
    expires_at INTEGER,
    revoked_at INTEGER
);

INSERT INTO invites_new (token_hash, game_id, kind, created_by, created_at, expires_at, revoked_at)
    SELECT token_hash, game_id, kind, created_by, created_at, expires_at, revoked_at FROM invites;

DROP TABLE invites;
ALTER TABLE invites_new RENAME TO invites;

CREATE INDEX invites_game_id ON invites (game_id);
