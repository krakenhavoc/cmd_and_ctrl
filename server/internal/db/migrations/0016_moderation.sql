-- Migration 0016: table_blocks and player_reports (#2984, ADR 0075
-- section 2.6): basic moderation for tables strangers share.
--
--   table_blocks     who may not take a seat at one table again, because
--                    the host (or an admin) removed them and chose to
--                    block them. One row per identity the removed seat
--                    could be recognised by, scoped to ONE table:
--                      kind = 'user'     subject = users(id)
--                      kind = 'discord'  subject = the Discord snowflake
--                      kind = 'name'     subject = the folded guest name
--                                        (moderation.Key), for a seat
--                                        with no account behind it
--                    Deleted with its game (ON DELETE CASCADE).
--
--   player_reports   a "report player" filed from a table: who reported
--                    whom, why, and the last few chat lines at that
--                    moment. Admin-read only. NOT keyed to the game by a
--                    foreign key: a report outlives a table an admin
--                    deletes, which is when it is most needed. Rows are
--                    removed 90 days after filing (lobby.ReportRetention)
--                    and the table is capped (lobby.ReportMaxRows).
--
-- Every *_at column is Unix MILLISECONDS (UTC), like 0002.
--
-- Additive: two new tables, no change to an existing one, no backfill.
-- Rolling back: an older binary refuses this schema (ErrSchemaTooNew).
-- The tables can be dropped by hand; see docs/environments.md.

CREATE TABLE table_blocks (
    game_id    TEXT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    kind       TEXT NOT NULL CHECK (kind IN ('user', 'discord', 'name')),
    subject    TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    PRIMARY KEY (game_id, kind, subject)
);

CREATE TABLE player_reports (
    id                 TEXT PRIMARY KEY,   -- a uuid
    created_at         INTEGER NOT NULL,
    game_id            TEXT NOT NULL,
    game_name          TEXT NOT NULL,
    reporter_player_id TEXT NOT NULL,
    reporter_user_id   TEXT,               -- users(id); NULL: a guest
    reporter_name      TEXT NOT NULL,
    reported_player_id TEXT NOT NULL,
    reported_user_id   TEXT,               -- users(id); NULL: a guest
    reported_name      TEXT NOT NULL,
    reason             TEXT NOT NULL,
    excerpt            TEXT NOT NULL       -- JSON array of the recent chat lines
);

CREATE INDEX player_reports_created_at ON player_reports (created_at);
CREATE INDEX player_reports_reporter_user ON player_reports (reporter_user_id);
CREATE INDEX player_reports_reported_user ON player_reports (reported_user_id);
