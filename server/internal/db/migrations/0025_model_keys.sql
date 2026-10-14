-- Migration 0025: a person's own model key for their AI seats (ADR 0182,
-- #3460).
--
--   model_keys   one row per signed-in person who has saved a key.
--
--     provider        an id from modelkeys.Providers: "anthropic",
--                     "openai", "openrouter", ... or "custom".
--     base_url        the https endpoint of a "custom" provider, and ''
--                     for every listed one (their address is in the code).
--     model           the model id the person's seats ask for by default.
--     sealed_key      the API key, sealed with AES-256-GCM under
--                     CMDCTRL_IDENTITY_KEY (ADR 0051 decision 5): version
--                     byte, nonce, ciphertext, tag. The person's id and the
--                     provider are bound in as associated data. Never
--                     stored in the clear; with no sealing key the feature
--                     is off and nothing is written.
--     last4           the key's last four characters, so the Settings page
--                     can say which key is saved without holding it.
--     games_per_day   the owner's limit: how many games a day their seats
--                     may play on the key (UTC days).
--     calls_per_game  the owner's limit: how many model calls one seat may
--                     make in one game.
--     day, day_games  the UTC date the count is for and how many games
--                     have used the key on it.
--     created_at, updated_at, last_used_at
--     last_error, last_error_at
--                     the last failure, as one of the server's own fixed
--                     sentences (never the provider's text, which can echo
--                     part of a key), so the owner is told on Settings.
--
--   seats.bot_key_user_id, seats.bot_key_model
--                   a bot seat that runs on its owner's key: whose key
--                   (users.id), and the model it asks for. NULL for every
--                   other seat. A column, not the engine snapshot, so a
--                   restart brings the seat back on the same key. Not a
--                   foreign key, so the hand rollback can drop it (SQLite
--                   cannot drop a column that is one).
--
-- Account deletion (ADR 0149) deletes the person's model_keys row. A
-- seat's bot_key_user_id then points at the tombstone, like seats.user_id:
-- it carries no identity and has no key, so a restored seat plays the
-- heuristic rather than sitting idle.
--
-- Every *_at column is Unix MILLISECONDS (UTC), matching 0002.
--
-- Additive: one new table and two nullable columns, no backfill. Rolling
-- back: an older binary refuses this schema (ErrSchemaTooNew). Drop them
-- by hand; see docs/environments.md.

CREATE TABLE model_keys (
    user_id        TEXT PRIMARY KEY REFERENCES users(id),
    provider       TEXT NOT NULL CHECK (provider <> ''),
    base_url       TEXT NOT NULL DEFAULT '',
    model          TEXT NOT NULL CHECK (model <> ''),
    sealed_key     BLOB NOT NULL CHECK (length(sealed_key) > 29),
    last4          TEXT NOT NULL,
    games_per_day  INTEGER NOT NULL CHECK (games_per_day BETWEEN 1 AND 100),
    calls_per_game INTEGER NOT NULL CHECK (calls_per_game BETWEEN 1 AND 5000),
    day            TEXT,
    day_games      INTEGER NOT NULL DEFAULT 0 CHECK (day_games >= 0),
    created_at     INTEGER NOT NULL,
    updated_at     INTEGER NOT NULL,
    last_used_at   INTEGER,
    last_error     TEXT,
    last_error_at  INTEGER
);

ALTER TABLE seats ADD COLUMN bot_key_user_id TEXT;
ALTER TABLE seats ADD COLUMN bot_key_model TEXT;
