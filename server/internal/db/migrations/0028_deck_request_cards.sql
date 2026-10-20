-- Migration 0028: the cards each deck request listed (#4135).
--
-- ADR 0095 §3 keeps one deck_requests row per deck: the issue that
-- tracks it. It does not say which cards that issue lists, so a deck
-- update (#2036) had to guess: it treated whatever the old version was
-- missing as already requested. A player who declined to request at
-- one update and then updated again was offered cards an earlier issue
-- already listed, and a comment could list a card the issue had
-- already checked off.
--
-- deck_request_cards is one row per card a request listed, under the
-- deck key it was listed for:
--
--   deck_key      "moxfield:<id>", "archidekt:<id>" or "list:<hash>",
--                 as deck_requests.deck_key. When a saved deck's key
--                 changes in an update (a pasted list's key is its
--                 content), the new key inherits the old key's rows,
--                 so the deck's history follows it whether or not the
--                 player asked for anything at that update.
--   card_key      the card's Scryfall oracle id, or "name:<lower-case
--                 name>" when the index could not resolve it.
--   card_name     the card's name as the issue printed it.
--   issue_number  the GitHub issue that listed it.
--
-- The rows name a deck and a card, never a person: the requester is in
-- deck_request_asks, which account deletion removes. Like
-- deck_requests, these rows are not about a person and are kept.
--
-- Additive: one new table, no backfill. A request filed before this
-- migration has no rows, and the update flow falls back to the older
-- "what the old version was missing" rule for it. Rolling back: an
-- older binary refuses this schema (ErrSchemaTooNew). Drop the table
-- by hand; see docs/environments.md.

CREATE TABLE deck_request_cards (
    deck_key     TEXT NOT NULL,
    card_key     TEXT NOT NULL,
    card_name    TEXT NOT NULL,
    issue_number INTEGER NOT NULL,
    PRIMARY KEY (deck_key, card_key)
);
