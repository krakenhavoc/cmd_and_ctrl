-- Migration 0006: deck requests (ADR 0095 §3, #1631).
--
-- A deck request files a GitHub issue listing the cards a deck needs
-- the engine to learn. deck_requests is one row per deck: the issue
-- currently tracking it. A request for a deck whose issue is still
-- open comments on that issue instead of filing another; a request
-- for a deck whose issue was closed files a new one and repoints the
-- row, which is why issue_number is not unique history but "the
-- current issue".
--
-- deck_request_asks is one row per ask that reached GitHub (an issue
-- filed, or a comment added). It is the rate limit — three asks per
-- requester per rolling 24 hours, counted here so it survives a
-- deploy — and the "never comment twice for the same person" check.
-- requester is "discord:<snowflake>": the bot and the site share the
-- key, so switching surfaces does not reset the limit. The snowflake
-- is stored for that and is never published.
--
-- deck_key is "moxfield:<id>" or "archidekt:<id>". Neither table
-- references users: the bot files for Discord members who may never
-- have signed in here.
--
-- Every time column is Unix MILLISECONDS (UTC), matching 0002/0003.

CREATE TABLE deck_requests (
    deck_key     TEXT PRIMARY KEY,
    issue_number INTEGER NOT NULL,
    issue_url    TEXT NOT NULL,
    created_at   INTEGER NOT NULL
);

CREATE TABLE deck_request_asks (
    deck_key     TEXT NOT NULL,
    requester    TEXT NOT NULL,
    at           INTEGER NOT NULL
);

-- The rate limit: one requester's asks in a time window.
CREATE INDEX deck_request_asks_requester_at ON deck_request_asks (requester, at);
