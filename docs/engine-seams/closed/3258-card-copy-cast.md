---
title: "Copying a card and casting the copy"
date: 2026-10-13
issues: [3258]
---
**Copying a card and casting the copy** (#3258, CR 707.12, CR 704.5e, CR 707.10, [ADR 0066](decisions/0066-granted-cast-and-play-permissions.md) amendment 2026-10-13) — `game/card_copy.go` is the rules side, `effects/copy_and_cast.go` the card side. The general form of the prepare spell's copy (`game/prepare.go`), for any number of cards at once.
- **The copy.** `Game.CopyCardsAndOfferCastsForEffect(CardCopyOffer, cards)` makes a copy of each named card in exile or a graveyard, in the zone the card is in, from its copiable values (CR 707.2), marked `Card.CardCopy`. Unlike a prepare copy it keeps every copiable value, so its oracle ID is the original's and the catalog resolves it unchanged. Created, not moved: no zone-move event announces it.
- **The offer.** One ADR 0066 `ScopeCards` permission per copy, so "any number of the copies" is no cap, casting one never touches another, and the order the player casts them in is the order they go on the stack. The window is the existing pass-closed one (`TimingFlash`, `LapseStaysInExile`). `CardCopyOffer.Cost` is `{0}` for "without paying its mana cost" (it also locks X at 0) or empty for "cast the copy" at its printed cost; `CardCopyOffer.FollowUp` is a registered cast follow-up for "if you do".
- **The sweep.** `cardCopySweepSBALocked` is CR 704.5e: a `CardCopy` in any zone no stored permission offers to anyone ceases to exist. `closePassWindowsLocked` runs it as a pass lapses a window, so the copy goes with the pass. A copy that is cast is the spell on the stack (`StackItem.IsCopy`) and ceases on resolution; a permanent copy becomes a token.
- **Arcane Bombardment's trigger** needed `CastTally.InstantSorcery`, the count of instant and sorcery spells cast this turn, bumped beside `Artifact` as each spell becomes cast.
- **Snapshot.** `cardCopy` on the card and `instantSorcery` on the turn tally, additive in v8.
- **Cards:** Mizzix's Mastery (Full, overload included) and Arcane Bombardment (Full).
- **Not built:** a total-mana-value budget (Uldaros Theorix, `budgeted-free-casts`), a resolution-time cast with an "if you do" that can read the cast's outcome inline, a card copy for an Aura-enchanted card (Spellweaver Volute, #3157) and paradigm's recurring offer (#1866). Kaervek, the Punisher (#2947) needs the same call at its printed cost with a follow-up, and waits on the crime-detection fix (#2944).
