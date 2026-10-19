# ADR 0192 — A deck names its printing, and the default avoids gimmick reprints

**Status:** Accepted · 2026-10-14 · S74 — Table truth: the log, the board and the art say what happened.
**Issue:** [#3605](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3605).
**Numbering:** checked on 2026-10-14. `origin/develop` ends at 0191, and no other remote head (`cost-ledger`, `docs/issue-audit`, `feat/750-conditional-block-restrictions`, `feat/playmats`, `fix/2947-resolution-copy-casts`, `fix/2950-darkness-replacement-continuation`, `fix/3159-3160-attachments`, `fix/caddy-reload-admin-off`, `wip/836-one-click-default`, `main`) holds a higher one. The issue's claim comment says "ADR 0192 reserved", and no open issue's claim names it.
**Builds on:** [ADR 0110](0110-remember-me.md) (a saved deck is its source text, re-parsed on every seat), [ADR 0158](0158-signed-in-pages-tables-with-history-decks-and-the-catalog.md) §2 (the Decks page draws each commander), [ADR 0010](0010-card-effect-catalog.md) (the catalog is keyed by oracle id, so a printing never changes what a card does).

---

## Context

A player's commander, Raffine, Scheming Seer, rendered with the MSCHF Secret Lair art. Their Archidekt deck holds the Streets of New Capenna printing. Two causes:

1. **Imports drop the printing.** `deck.Entry` was a name, a count and board flags. Archidekt sends the Scryfall `uid`, the `edition.editioncode` and the `collectorNumber`; Moxfield sends `scryfall_id`, `set` and `cn`; a text list can say `1 Raffine, Scheming Seer (SNC) 213`. All three were read and thrown away.
2. **The default is "the last record in the dump".** `cards.Index` keeps one printing per name (`byName`) and per oracle id (`byOracle`) and kept the last playable record it saw, on the guess that the dump lists newer printings later. For Raffine that is `slz` #93, a Secret Lair box released 2026-09-02. Any card with a recent Secret Lair, box, funny or promo reprint gets that art for every deck that names no printing.

Every part of the game already follows the printing: `deck.Resolve` returns a `cards.Card`, the deck importer stamps `game.Card.ScryfallID` from its `ID`, and the card's image, artist credit and face art all hang off that id. So there is nothing to add to the game state, the snapshot or the wire. The work is to choose the right `cards.Card` in `Resolve`.

## Decision

### 1. The default printing is ranked, not "last seen"

When two playable records compete for one name or oracle id, `preferIncoming` compares `printingRank` field by field, in this order, and the better one wins:

1. an English printing over any other language;
2. the set type's tier: `expansion`, `core`, `commander` and `draft_innovation` first; then `masters`, `eternal`, `starter`, `duel_deck`, `from_the_vault`, `premium_deck`, `planechase`, `archenemy` and `spellbook`; then any set type not named here; last `box` (Secret Lair), `promo`, `funny`, `treasure_chest`, `arsenal` and `alchemy`;
3. a paper, regular-sized printing over a digital-only or oversized one;
4. a non-promo printing over a promo;
5. a regular frame over a showcase, extended-art, etched, inverted, textless, shattered-glass, full-art or borderless one;
6. the newer `released_at`.

A tie keeps the old rule, that the later record wins, so a fixture or a card with no dates behaves as it did. A card that only has gimmick printings still resolves to one of them: the ranking orders printings of a card and never excludes one. `isPlayablePrint` is unchanged and still comes first, so legality, the oracle-id join and token or art-series exclusion are untouched.

The ranking reads six new `Card` fields (`released_at`, `promo`, `full_art`, `digital`, `oversized`, `frame_effects`), all `omitempty`. They ride `GET /cards/{id}` and nothing else; no game or protocol type reads them.

### 2. A deck's printing is carried on `Entry` and honoured in `Resolve`

`Entry` gains `PrintingID`, `SetCode` and `CollectorNo`, filled by each importer: Archidekt `uid` / `editioncode` / `collectorNumber`, Moxfield `scryfall_id` / `set` / `cn`, and a text row's trailing `(SET) number`. `Resolve` finds the card by name as before, then asks `Index.FindPrintingOf` for the named printing and uses it when found. Lookup is by printing id first, then by set code and collector number, which the index now keeps in a `bySetNo` map.

`FindPrintingOf` answers only for a printing that is a **playable print of the same card** (same oracle id, falling back to the same name). A wrong, stale or hostile id in a deck can therefore cost a player a picture and never change which card they hold. A printing id the index knows but cannot use does not fall through to the set and number, which would only name some other card. A printing the index lacks, a set with no number, and a malformed id all read as "named none", and the deck resolves to the default.

**Saved decks need no schema change.** A saved deck is its source text (ADR 0110): a pasted list or Moxfield export is stored verbatim and re-parsed, so it carries its printing already. A deck fetched from a link is stored as `List.Text()`; that now writes `1 Name (SET) number` for every card that has both, and groups copies by printing rather than by name so two printings of a basic land survive a save. A deck saved before this change has no suffix and resolves to the default, the same as a deck whose source named nothing. Games in progress keep the `ScryfallID` already stamped on their cards, so snapshots and restore points are untouched.

### 3. The Decks page follows the deck

The library row stores commander names only. `libraryDeckInfo` now reads the printing off the stored source (`libraryCommanderPrintings`) and passes it to `libraryCommanders`, so the Decks grid draws the commander the table will show. A source that no longer parses falls back to the default art.

## Consequences

- With no printing named, Raffine renders snc #213. Checked against the 2026-10-05 dump by `TestRealDumpRaffineDefaultsToSNC`, which skips unless `CMDCTRL_SCRYFALL_DUMP` is set.
- A card whose newest ordinary printing is old now shows that printing instead of a later Secret Lair or promo reprint. That is the intent.
- The deck checker, coverage report and catalog are keyed by oracle id and name, and are unaffected.
- The original scope deferred an in-app printing picker (added in the amendment below) and printings for tokens (ADR 0078 owns those). A deck from a source that names no printing gets the default; a player who wants a specific art pastes a list with `(SET) number`.


## Amendment — saved-deck printing picker (2026-10-18, #4124)

The owner requested an in-app picker after a saved deck reported nine
reskinned titles as unknown. The Decks grid now offers **Choose printings**:
select a saved card row, browse full-card previews, filter by set, number
or alternate title, then explicitly save or cancel. Choices apply to every
copy in that source row on its next seating; ongoing games are untouched.

This extends decision 2 using the same stored `(SET) number` suffix. It
adds no preference table or snapshot field. Ownership is checked before
reads and writes, all selected printings must match the card's oracle
identity, and a source revision prevents a stale picker overwriting a
changed list. An atomic source update preserves metadata and the power
rating because only art changed. Unresolved names and all board flags
survive the rewrite. Canonical names and Scryfall `flavor_name` aliases
both resolve, and an alias selects its corresponding reskinned printing.
Canonical names take precedence on a title collision.
