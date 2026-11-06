-- Migration 0032: friends and blocks (ADR 0213 §2, §3 and §10, slice 3
-- of #3313).
--
--   friendships      one row per pair of friends, user_a < user_b, so a
--                    pair has exactly one row whichever side asked.
--                    Kept until either person removes it, or blocks the
--                    other.
--   friend_requests  a pending request from one person to another. No
--                    text: a request is the pair and a time. ignored_at
--                    is set when the receiver chose Ignore, which takes
--                    it off their list while the sender still sees it
--                    pending. A request lasts 30 days
--                    (social.RequestTTL), then it is deleted.
--   user_blocks      one person's block of another. It ends any
--                    friendship and request between the two, in either
--                    direction, and hides each from the other's lists.
--   social_prefs     one row per person who changed a social setting:
--                    who can see what they are doing on CMD & CTRL
--                    (presence, read from slice 4) and who can send
--                    them friend requests. No row means the defaults,
--                    'friends' and 'played_with'.
--
-- Every row names a person, and friendships, friend_requests and
-- user_blocks name two. Account deletion (ADR 0149) deletes every row
-- with the person on either side.
--
-- Every *_at column is Unix MILLISECONDS (UTC), matching 0002.
--
-- Additive: four new tables, no backfill. Rolling back: an older
-- binary refuses this schema (ErrSchemaTooNew). The tables can be
-- dropped by hand; see docs/environments.md.

CREATE TABLE friendships (
    user_a     TEXT NOT NULL REFERENCES users(id),
    user_b     TEXT NOT NULL REFERENCES users(id),
    created_at INTEGER NOT NULL,
    PRIMARY KEY (user_a, user_b),
    CHECK (user_a < user_b)
);
CREATE INDEX friendships_by_user_b ON friendships(user_b);

CREATE TABLE friend_requests (
    from_user  TEXT NOT NULL REFERENCES users(id),
    to_user    TEXT NOT NULL REFERENCES users(id),
    created_at INTEGER NOT NULL,
    ignored_at INTEGER,
    PRIMARY KEY (from_user, to_user),
    CHECK (from_user <> to_user)
);
CREATE INDEX friend_requests_by_to_user ON friend_requests(to_user);
CREATE INDEX friend_requests_by_created_at ON friend_requests(created_at);

CREATE TABLE user_blocks (
    user_id         TEXT NOT NULL REFERENCES users(id),
    blocked_user_id TEXT NOT NULL REFERENCES users(id),
    created_at      INTEGER NOT NULL,
    PRIMARY KEY (user_id, blocked_user_id),
    CHECK (user_id <> blocked_user_id)
);
CREATE INDEX user_blocks_by_blocked ON user_blocks(blocked_user_id);

CREATE TABLE social_prefs (
    user_id    TEXT PRIMARY KEY REFERENCES users(id),
    presence   TEXT NOT NULL DEFAULT 'friends'
               CHECK (presence IN ('friends', 'played_with', 'nobody')),
    requests   TEXT NOT NULL DEFAULT 'played_with'
               CHECK (requests IN ('played_with', 'nobody')),
    updated_at INTEGER NOT NULL
);
