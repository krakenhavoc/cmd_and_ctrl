-- Migration 0034: scheduled tables and their RSVPs (ADR 0215 §1, §3,
-- §12 and its privacy section; slice 2 of #3321).
--
--   scheduled_tables  one row per plan: a table booked for a start time.
--                     It is a plan, not a table, until its doors open 15
--                     minutes before the start; then game_id names the
--                     lobby table the server made from it. The name is
--                     the only free text, filtered like every table
--                     name. starts_at is a UTC instant and zone the
--                     host's IANA zone, kept for slice 8's series.
--                     options is the create options as JSON (voice
--                     room, #lfg, results, Planechase, the last setup,
--                     fill with bots). moves counts the times the host
--                     moved it (at most 3). discord_event_id is slice
--                     4's mirror in Discord, NULL until then.
--   scheduled_rsvps   who has a seat or is on the waitlist, keyed by
--                     Discord id, which every account has; user_id is
--                     the account, NULL for a person with no account
--                     (slice 5's Interested). The order is rsvp_at, ties
--                     by discord_id. state is the person's place:
--                       seat     before doors-open, one of the first N
--                       waiting  on the waitlist
--                       held     after doors-open, a seat held for them
--                                until held_until
--                       seated   they took it
--                       lapsed   the hold ran out, or the game started
--                                without them
--                       removed  the host removed them from this
--                                occurrence; they cannot reserve again
--                     The host has a row of their own (source 'host'):
--                     they hold seat 1 like everyone else.
--
-- A plan and its RSVPs are kept until 24 hours after the start, or 24
-- hours after a cancel (schedule.Keep), then deleted; deleting a plan
-- deletes its RSVPs by cascade. Account deletion (ADR 0149) deletes the
-- plans a person hosts and every RSVP of theirs, by user id and by
-- Discord id.
--
-- Every *_at column is Unix MILLISECONDS (UTC), matching 0002.
--
-- Additive: two new tables, no backfill. Rolling back: an older binary
-- refuses this schema (ErrSchemaTooNew). The tables can be dropped by
-- hand; see docs/environments.md.

CREATE TABLE scheduled_tables (
    id               TEXT PRIMARY KEY,
    host_user_id     TEXT NOT NULL REFERENCES users(id),
    name             TEXT NOT NULL CHECK (length(name) BETWEEN 1 AND 80),
    starts_at        INTEGER NOT NULL,
    length_minutes   INTEGER NOT NULL CHECK (length_minutes IN (120, 180, 240)),
    zone             TEXT NOT NULL,
    players          INTEGER NOT NULL CHECK (players BETWEEN 2 AND 4),
    options          TEXT NOT NULL DEFAULT '{}',
    status           TEXT NOT NULL CHECK (status IN ('scheduled', 'open', 'settled', 'cancelled', 'missed')),
    moves            INTEGER NOT NULL DEFAULT 0 CHECK (moves BETWEEN 0 AND 3),
    discord_event_id TEXT,
    game_id          TEXT REFERENCES games(id) ON DELETE SET NULL,
    created_at       INTEGER NOT NULL,
    updated_at       INTEGER NOT NULL
);
CREATE INDEX scheduled_tables_host ON scheduled_tables(host_user_id, status);
CREATE INDEX scheduled_tables_status ON scheduled_tables(status, starts_at);

CREATE TABLE scheduled_rsvps (
    schedule_id TEXT NOT NULL REFERENCES scheduled_tables(id) ON DELETE CASCADE,
    discord_id  TEXT NOT NULL CHECK (discord_id <> ''),
    user_id     TEXT REFERENCES users(id),
    source      TEXT NOT NULL CHECK (source IN ('host', 'site', 'discord')),
    state       TEXT NOT NULL CHECK (state IN ('seat', 'waiting', 'held', 'seated', 'lapsed', 'removed')),
    rsvp_at     INTEGER NOT NULL,
    held_until  INTEGER,
    updated_at  INTEGER NOT NULL,
    PRIMARY KEY (schedule_id, discord_id),
    CHECK ((state = 'held') = (held_until IS NOT NULL))
);
CREATE INDEX scheduled_rsvps_user ON scheduled_rsvps(user_id);
CREATE INDEX scheduled_rsvps_discord ON scheduled_rsvps(discord_id);
