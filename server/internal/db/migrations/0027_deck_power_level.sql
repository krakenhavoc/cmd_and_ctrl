-- Migration 0027: decks.power_level (#3671).
--
--   decks.power_level  the deck's estimated power on the Commander
--                      Bracket scale, 1 to 5 in half steps (3.5 is a
--                      strong 3), or NULL while unrated. An admin sets
--                      it by hand (PUT /admin/decks/{id}/power); a save
--                      that changes the deck's list clears it, because
--                      the rating described the old list.
--
-- It is about the deck, not the person, and goes with the deck row when
-- the deck or the account is deleted.
--
-- Additive: one nullable column, no backfill. Rolling back: an older
-- binary refuses this schema (ErrSchemaTooNew). Drop it by hand; see
-- docs/environments.md.

ALTER TABLE decks ADD COLUMN power_level REAL
    CHECK (power_level IS NULL OR (power_level >= 1 AND power_level <= 5));
