-- Migration 0030: the advanced-stats setting (ADR 0217 §2 and owner
-- answer 1, slice 3 of #4880).
--
--   stats_prefs      one row per person who has a stats setting: the
--                    advanced-stats switch, who set it, and the columns
--                    later slices of ADR 0217 fill in.
--
--     advanced         NULL, 0 or 1. NULL means "follow the default".
--     advanced_set_by  'self' or 'admin' when advanced is set, NULL
--                      when it is not. An admin's switch is the same
--                      column, and the person can still switch it off.
--     share_pods       slice 7: whether the person shares their stats
--                      with their pods. 0 until then.
--     share_link_hash  slice 6: the SHA-256 (hex) of the person's share
--                      link token, or NULL. The token is never stored.
--     share_link_at    slice 6: when that link was made, or NULL.
--     updated_at       when the row last changed.
--
--   stats_settings   at most one row (id = 1): the server's default for
--                    a person whose advanced is NULL, set by an admin
--                    (PUT /admin/stats/advanced-default). No row means
--                    gamestats.AdvancedDefault (off). It is about the
--                    server, not a person.
--
-- Its own table rather than the user_settings body: ADR 0110 §4 keeps
-- that body opaque to the server, and GET /me/stats has to read this.
-- No row means "follow the default". Account deletion (ADR 0149)
-- deletes the row.
--
-- Every *_at column is Unix MILLISECONDS (UTC), matching 0002.
--
-- Additive: two new tables, no backfill. Rolling back: an older binary
-- refuses this schema (ErrSchemaTooNew). The tables can be dropped by
-- hand; see docs/environments.md.

CREATE TABLE stats_prefs (
    user_id         TEXT PRIMARY KEY REFERENCES users(id),
    advanced        INTEGER CHECK (advanced IN (0, 1)),
    advanced_set_by TEXT CHECK (advanced_set_by IN ('self', 'admin')),
    share_pods      INTEGER NOT NULL DEFAULT 0 CHECK (share_pods IN (0, 1)),
    share_link_hash TEXT UNIQUE,
    share_link_at   INTEGER,
    updated_at      INTEGER NOT NULL,
    CHECK ((advanced IS NULL) = (advanced_set_by IS NULL))
);

CREATE TABLE stats_settings (
    id               INTEGER PRIMARY KEY CHECK (id = 1),
    advanced_default INTEGER NOT NULL CHECK (advanced_default IN (0, 1)),
    updated_at       INTEGER NOT NULL
);
