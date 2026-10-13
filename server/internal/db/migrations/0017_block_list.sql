-- Migration 0017: table_blocks gains entry_id and label, and a fourth
-- kind, 'spectator' (#3113, ADR 0075 section 2.6 amendment).
--
-- 0016 let a host remove-and-block a seat, with no way to see or undo
-- it. A visible block list needs two things a row did not carry:
--
--   entry_id   groups the rows ONE removal wrote. Removing a signed-in
--              person blocks their account id and their Discord id, two
--              rows; the host sees one entry and unblocks it as one.
--              An opaque uuid-shaped string; the route addresses it.
--   label      what the list shows for the entry: the account's display
--              name, or a guest's folded name (moderation.Key). Never a
--              Discord id. It is the name the person had at the table
--              when they were blocked; it does not follow a rename.
--
-- and the spectator kind:
--
--   kind = 'spectator'  subject = the random id minted into one
--              spectator session. It is the marker that kills THAT
--              session at this table (a signed session is stateless, so
--              removal has to be remembered somewhere that survives a
--              restart). It names no one, carries no label, and is not
--              listed; removing a spectator with "also block" writes
--              user / discord / name rows under the same entry_id, and
--              unblocking the entry deletes the marker too.
--
-- SQLite cannot widen a CHECK in place, so the table is rebuilt, as 0003
-- and 0005 rebuilt theirs. Nothing references table_blocks.
--
-- Backfill of the rows 0016 wrote: each gets its own entry_id; a
-- 'discord' row joins the entry of the 'user' row of the same table whose
-- account has that Discord id, so one removal still reads as one entry;
-- a 'user' row is labelled with the account's display name, a 'name' row
-- with its folded name, and a 'discord' row that joins nothing with a
-- plain "Discord account" (a snowflake is never shown).
--
-- Additive for a reader of the old shape; rolling back: an older binary
-- refuses this schema (ErrSchemaTooNew).

CREATE TABLE table_blocks_new (
    game_id    TEXT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    kind       TEXT NOT NULL CHECK (kind IN ('user', 'discord', 'name', 'spectator')),
    subject    TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    entry_id   TEXT NOT NULL,
    label      TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (game_id, kind, subject)
);

INSERT INTO table_blocks_new (game_id, kind, subject, created_at, entry_id, label)
    SELECT b.game_id, b.kind, b.subject, b.created_at,
           lower(hex(randomblob(16))),
           CASE b.kind
                WHEN 'name' THEN b.subject
                WHEN 'user' THEN COALESCE((SELECT u.display_name FROM users u WHERE u.id = b.subject), 'Signed-in player')
                ELSE COALESCE((SELECT u.display_name FROM identities i JOIN users u ON u.id = i.user_id
                                WHERE i.provider = 'discord' AND i.subject = b.subject), 'Discord account')
           END
      FROM table_blocks b;

DROP TABLE table_blocks;
ALTER TABLE table_blocks_new RENAME TO table_blocks;

UPDATE table_blocks
   SET entry_id = (SELECT n.entry_id
                     FROM table_blocks n
                     JOIN identities i ON i.user_id = n.subject AND i.provider = 'discord'
                    WHERE n.game_id = table_blocks.game_id AND n.kind = 'user'
                      AND i.subject = table_blocks.subject)
 WHERE kind = 'discord'
   AND EXISTS (SELECT 1
                 FROM table_blocks n
                 JOIN identities i ON i.user_id = n.subject AND i.provider = 'discord'
                WHERE n.game_id = table_blocks.game_id AND n.kind = 'user'
                  AND i.subject = table_blocks.subject);

CREATE INDEX table_blocks_entry ON table_blocks (game_id, entry_id);
