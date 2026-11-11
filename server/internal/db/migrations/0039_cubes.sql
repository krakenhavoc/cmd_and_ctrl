-- Migration 0039: a signed-in person's saved cubes (ADR 0222 §8,
-- #5441).
--
-- A cube is a named list of cards a person saves so a draft can deal
-- its packs from it in place of a set's. It is the person's own: only
-- they can read it, change it or draft it.
--
--   cubes  one row per cube:
--     id          the cube's id
--     user_id     the person it belongs to
--     name        the name they gave it, which passed the name filter;
--                 one name once per person
--     singleton   1 when the list may hold one copy of a card (basic
--                 lands aside), which is the default; 0 for no limit
--     list        the card list as plain text, one "N Name" row a
--                 card, the form the deck importers read. Kept as text
--                 and checked against the catalog each time it is read,
--                 so a card that is built later stops blocking the cube
--     card_count  how many cards the list holds, copies counted
--     created_at, updated_at  Unix milliseconds
--
-- ADR 0149: the table is personal data. Account deletion deletes the
-- person's rows (accountdelete.Inventory, cubes.ForgetAccount).
--
-- Additive: one new table. Rolling back: an older binary refuses this
-- schema (ErrSchemaTooNew). The table can be dropped by hand; see
-- docs/environments.md.

CREATE TABLE cubes (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id),
    name       TEXT NOT NULL,
    singleton  INTEGER NOT NULL DEFAULT 1,
    list       TEXT NOT NULL,
    card_count INTEGER NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE UNIQUE INDEX cubes_by_owner_name ON cubes (user_id, name);
