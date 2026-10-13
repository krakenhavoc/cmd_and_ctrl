-- Migration 0018: the pre-built deck izzet-aggro is now mono-red-aggro
-- (#3071). The deck was rebuilt as mono-red under Purphoros, and its ID
-- followed. Data only: no table or column changes.
--
-- Two places store a pre-built deck ID (migration 0008):
--
--   users.last_deck     {"kind": "prebuilt", "id": "<deck id>"}
--   table_setups.body   {"bots": [{"tier", "deck_id", "name"}, ...], ...}
--
-- Both are rewritten with the JSON functions, so a bot NAME that
-- happens to contain the old ID is left alone. Every other field and
-- the bots' order are kept. decks.legacyIDs still resolves the old ID
-- for what this cannot reach: a guest's browser, an engine snapshot.

UPDATE users
   SET last_deck = json_set(last_deck, '$.id', 'mono-red-aggro')
 WHERE last_deck IS NOT NULL
   AND json_valid(last_deck)
   AND json_extract(last_deck, '$.kind') = 'prebuilt'
   AND json_extract(last_deck, '$.id') = 'izzet-aggro';

UPDATE table_setups
   SET body = json_set(body, '$.bots', (
       SELECT json_group_array(json(
                  CASE WHEN json_extract(b.value, '$.deck_id') = 'izzet-aggro'
                       THEN json_set(b.value, '$.deck_id', 'mono-red-aggro')
                       ELSE b.value
                  END))
         FROM (SELECT value FROM json_each(table_setups.body, '$.bots') ORDER BY key) AS b))
 WHERE json_type(body, '$.bots') = 'array'
   AND EXISTS (SELECT 1 FROM json_each(table_setups.body, '$.bots')
                WHERE json_extract(value, '$.deck_id') = 'izzet-aggro');
