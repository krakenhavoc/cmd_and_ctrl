-- Migration 0013: user_playmats (ADR 0128 §11, playmats v2: up to
-- three saved mats per account).
--
--   user_playmats  one row per saved playmat: (user_id, slot) is the
--                  key, slot is 1 to 3, playmat_id is the uuid of the
--                  image file <data dir>/playmats/<playmat_id>.jpg,
--                  width and height are the stored image's pixels
--                  (NULL on a row backfilled below; the service reads
--                  the file's header then), created_at is Unix
--                  MILLISECONDS (UTC), matching 0002 to 0012.
--
-- users.playmat_id stays, and now means "the ACTIVE playmat": the one
-- the table shows. NULL is "none", which is legal for a person who has
-- saved mats. The invariant, kept by every write in
-- internal/playmat: a non-NULL users.playmat_id equals the playmat_id
-- of exactly one user_playmats row of the same user. SQLite cannot add
-- a foreign key to an existing column, so it is the service's, not the
-- schema's, and it is tested.
--
-- Backfill: every existing non-NULL users.playmat_id becomes that
-- person's slot 1, so the pointer matches a saved row from the first
-- moment. An empty-string pointer (no code path writes one) is cleared
-- so it cannot be an unmatched pointer either. width and height stay
-- NULL: the files are on disk and the service reads their headers.
--
-- Rolling back: an older binary refuses this schema (ErrSchemaTooNew).
-- The table is additive and can be dropped by hand; users.playmat_id is
-- untouched and keeps pointing at the active mat, so the older binary
-- finds the playmat it knew. See docs/environments.md. Files of the
-- other two slots are then orphans and can be deleted.

CREATE TABLE user_playmats (
    user_id    TEXT NOT NULL REFERENCES users(id),
    slot       INTEGER NOT NULL CHECK (slot BETWEEN 1 AND 3),
    playmat_id TEXT NOT NULL,
    width      INTEGER,
    height     INTEGER,
    created_at INTEGER NOT NULL,
    PRIMARY KEY (user_id, slot)
);

-- One image belongs to one slot; a replacement mints a new uuid.
CREATE UNIQUE INDEX user_playmats_playmat_id ON user_playmats (playmat_id);

UPDATE users SET playmat_id = NULL WHERE playmat_id = '';

INSERT INTO user_playmats (user_id, slot, playmat_id, created_at)
SELECT id, 1, playmat_id, CAST(strftime('%s', 'now') AS INTEGER) * 1000
FROM users
WHERE playmat_id IS NOT NULL;
