-- Migration 0011: user_playmats (ADR 0128).
--
--   user_playmats   one row per signed-in person who has set a
--                   playmat: the image behind their part of the board,
--                   which everyone at the table sees.
--     user_id       the owner.
--     file          the stored image's name under
--                   $CMDCTRL_DATA_DIR/playmats/ (random hex plus an
--                   allowlisted extension, internal/playmats).
--     wash          how strongly the image is darkened under the
--                   cards, a percentage from 30 to 90.
--     updated_at    Unix milliseconds of the last write.
--
-- Rolling back: an older binary refuses this schema (ErrSchemaTooNew).
-- The table is additive and can be dropped by hand; see
-- docs/environments.md.

CREATE TABLE user_playmats (
    user_id    TEXT PRIMARY KEY REFERENCES users(id),
    file       TEXT NOT NULL,
    wash       INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    -- The store checks both; these are the backstop.
    CHECK (wash BETWEEN 30 AND 90),
    CHECK (length(file) BETWEEN 1 AND 64)
);
