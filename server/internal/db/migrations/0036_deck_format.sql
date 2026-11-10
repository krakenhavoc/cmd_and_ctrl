-- Migration 0036: the format a saved deck was built for (ADR 0222 §2,
-- #2352).
--
--   decks.format  the format the deck is checked against and offered
--                 at: 'commander' or 'standard'. Every deck saved
--                 before formats existed is a Commander deck, which is
--                 the column's default, so no row is rewritten.
--
-- Set when a deck is saved from a table (the table's format). It says
-- nothing new about the person: it describes a list they already
-- saved, and account deletion (ADR 0149) deletes it with the deck's
-- row.
--
-- Additive: one column with a default, no backfill. Rolling back: an
-- older binary refuses this schema (ErrSchemaTooNew). The column can be
-- dropped by hand; see docs/environments.md.

ALTER TABLE decks ADD COLUMN format TEXT NOT NULL DEFAULT 'commander';
