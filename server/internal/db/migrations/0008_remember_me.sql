-- Migration 0008: remember me (ADR 0110, S55). All of S55's schema in
-- one migration, so the feature PRs never contend for a number.
--
--   user_settings    one row per person: the synced subset of the
--                    client's settings, as a JSON object
--   table_setups     one row per person: the last setup they started a
--                    table with, built by the server
--   users.last_deck  the last deck a person seated, as JSON
--   decks.source_url the link a deck was imported from (owner answer 7)
--
-- Every *_at column is Unix MILLISECONDS (UTC), matching 0002 to 0007.
--
-- Rolling back: an older binary refuses this schema (ErrSchemaTooNew).
-- The two new columns and two new tables are additive, so they can be
-- undone by hand; see docs/environments.md.

-- user_settings ----------------------------------------------------

CREATE TABLE user_settings (
    user_id    TEXT PRIMARY KEY REFERENCES users(id),
    version    INTEGER NOT NULL,   -- the client's SETTINGS_VERSION that wrote it
    revision   INTEGER NOT NULL,   -- bumped on every write, for If-Match
    body       TEXT NOT NULL,      -- JSON object: the synced subset
    updated_at INTEGER NOT NULL,
    -- The 32 KiB cap lives in the store too; this is the backstop. It
    -- counts bytes, not characters, and requires a JSON object.
    CHECK (length(CAST(body AS BLOB)) <= 32768),
    CHECK (json_valid(body) AND json_type(body) = 'object')
);

-- table_setups -----------------------------------------------------

CREATE TABLE table_setups (
    user_id    TEXT PRIMARY KEY REFERENCES users(id),
    body       TEXT NOT NULL,      -- JSON, built by the server, never by a client
    game_id    TEXT,               -- the table it was captured from; deliberately NOT a
                                   -- foreign key, so deleting a game keeps the setup
    updated_at INTEGER NOT NULL,
    CHECK (json_valid(body))
);

-- users.last_deck / decks.source_url -------------------------------
--
-- Plain ADD COLUMNs, no default and no CHECK, like 0004 and 0007.
-- last_deck is JSON {"kind": "library" | "prebuilt", "id": "..."}: a
-- pre-built pick is not a decks row, so it is not a foreign key.

ALTER TABLE users ADD COLUMN last_deck TEXT;
ALTER TABLE decks ADD COLUMN source_url TEXT;

-- seats.deck_id: ON DELETE SET NULL --------------------------------
--
-- 0005 declared the foreign key with no ON DELETE, so deleting a deck a
-- seat once used would fail. Deleting a deck is now a feature (ADR 0110
-- section 6), and the seat keeps its deck_name for history. SQLite
-- cannot alter a foreign key in place, so `seats` is rebuilt a third
-- time, following 0003 and 0005 (create under a temporary name, copy
-- every row, drop, rename, recreate indexes; the runner has already
-- turned foreign_keys OFF and runs PRAGMA foreign_key_check before it
-- commits, see applyMigration). No other table references seats.

CREATE TABLE seats_new (
    game_id            TEXT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    seat               INTEGER NOT NULL,
    player_id          TEXT NOT NULL,
    user_id            TEXT REFERENCES users(id),
    guest_name         TEXT,
    bot_tier           TEXT,
    deck_id            TEXT REFERENCES decks(id) ON DELETE SET NULL,
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
