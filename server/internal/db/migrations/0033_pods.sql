-- Migration 0033: pods, groups of players for shared stats (ADR 0220 §9,
-- slice 2 of #4880).
--
--   pods                       one row per pod: its name (free text,
--                              filtered like every other name), its
--                              owner, and the invite link. Only the
--                              SHA-256 (hex) of the link's token is kept,
--                              never the token, with when the link stops
--                              working; both are NULL with no link.
--   pod_members                who is in which pod, and when they joined.
--                              The owner is always a member: the code
--                              holds it in the transaction of every
--                              change of owner. The member who joined
--                              first takes over when the owner leaves or
--                              deletes their account.
--   pod_invites                a pending invite to one person, and who
--                              sent it. No text. declined_at is set when
--                              the invitee chose Decline, which takes it
--                              off their list while the pod still sees
--                              it pending. An invite lasts 30 days
--                              (pods.InviteTTL), then it is deleted.
--   pod_suggestion_dismissals  a person's "Not now" on a suggested pod
--                              (slice 4 writes it), keyed by a SHA-256 of
--                              the set's sorted account ids. Kept 180
--                              days (pods.DismissTTL).
--
-- Deleting a pod deletes its members and invites by cascade. Account
-- deletion (ADR 0149) passes a pod the person owns to the member who
-- joined first, or deletes it if they were the last, then deletes every
-- member, invite (as invitee or inviter) and dismissal row of theirs.
--
-- Every *_at column is Unix MILLISECONDS (UTC), matching 0002.
--
-- Additive: four new tables, no backfill. Rolling back: an older
-- binary refuses this schema (ErrSchemaTooNew). The tables can be
-- dropped by hand; see docs/environments.md.

CREATE TABLE pods (
    id                TEXT PRIMARY KEY,
    name              TEXT NOT NULL,
    owner_id          TEXT NOT NULL REFERENCES users(id),
    invite_hash       TEXT UNIQUE,
    invite_expires_at INTEGER,
    created_at        INTEGER NOT NULL,
    updated_at        INTEGER NOT NULL,
    CHECK ((invite_hash IS NULL) = (invite_expires_at IS NULL))
);

CREATE TABLE pod_members (
    pod_id    TEXT NOT NULL REFERENCES pods(id) ON DELETE CASCADE,
    user_id   TEXT NOT NULL REFERENCES users(id),
    joined_at INTEGER NOT NULL,
    PRIMARY KEY (pod_id, user_id)
);
CREATE INDEX pod_members_user ON pod_members(user_id);

CREATE TABLE pod_invites (
    pod_id      TEXT NOT NULL REFERENCES pods(id) ON DELETE CASCADE,
    user_id     TEXT NOT NULL REFERENCES users(id),
    invited_by  TEXT NOT NULL REFERENCES users(id),
    created_at  INTEGER NOT NULL,
    declined_at INTEGER,
    PRIMARY KEY (pod_id, user_id)
);
CREATE INDEX pod_invites_user ON pod_invites(user_id);

CREATE TABLE pod_suggestion_dismissals (
    user_id      TEXT NOT NULL REFERENCES users(id),
    group_key    TEXT NOT NULL,
    dismissed_at INTEGER NOT NULL,
    PRIMARY KEY (user_id, group_key)
);
