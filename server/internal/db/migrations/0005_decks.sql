-- Migration 0005: the deck library (ADR 0051 decision 7, S34 sub-PR 5).
--
-- A deck belongs to a person and is stored as the text they pasted,
-- not the resolved card list: the catalog changes weekly, and a deck
-- stored as resolved cards would drift silently as oracle ids move
-- between Scryfall dumps (ADR 0044 decision 7). The pasted text is
-- re-parsed through the same parse -> resolve -> validate pipeline an
-- upload takes, at the moment it is seated, and a card that no longer
-- resolves surfaces as an ordinary validation error.
--
-- Every *_at column is Unix MILLISECONDS (UTC), matching 0002/0003.
--
-- Then `seats` is rebuilt a second time, following 0003's procedure
-- (create under a temporary name, copy every row, drop, rename,
-- recreate indexes; the migration runner has already turned
-- foreign_keys OFF and runs PRAGMA foreign_key_check before it
-- commits — see applyMigration in migrations.go), so that
-- seats.deck_id becomes a real foreign key into decks(id). It has
-- been a plain TEXT column since migration 0002 waiting for `decks`
-- to exist; nothing has ever written a value into it (see 0002's and
-- 0003's comments), so there is nothing for the rebuild's
-- foreign_key_check to trip over.

CREATE TABLE decks (
    id            TEXT PRIMARY KEY,       -- uuid, minted by us
    owner_id      TEXT NOT NULL REFERENCES users(id),
    name          TEXT NOT NULL,
    source_format TEXT NOT NULL,          -- moxfield | text
    source_text   TEXT NOT NULL,          -- what the player pasted
    commanders    TEXT NOT NULL,          -- json array of names, for listing
    card_count    INTEGER NOT NULL,
    created_at    INTEGER NOT NULL,
    updated_at    INTEGER NOT NULL
);

-- GET /me/decks and the update rule (same owner, same name) both key
-- off owner_id.
CREATE INDEX decks_owner_id ON decks (owner_id);

-- seats ------------------------------------------------------------

CREATE TABLE seats_new (
    game_id            TEXT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    seat               INTEGER NOT NULL,
    player_id          TEXT NOT NULL,
    user_id            TEXT REFERENCES users(id),
    guest_name         TEXT,
    bot_tier           TEXT,
    deck_id            TEXT REFERENCES decks(id),
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
