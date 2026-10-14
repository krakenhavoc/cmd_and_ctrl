---
title: "Paradigm"
date: 2026-10-13
issues: [1866]
---
**Paradigm** (#1866, CR 702.192a, CR 603.7b, CR 707.12) — `game/paradigm.go`. A canonical keyword read off the resolving spell, so a deck-imported card works with no catalog entry; catalog cards declare it in `PrintedKeywords`.
- **The exile.** A resolved paradigm spell exiles itself as the last part of its own resolution, ahead of `routeStackCardToGraveyardLocked`'s other exits: it is the spell's instruction, not a replacement of its trip to the graveyard, so rebound, buyback and the Adventure have nothing to replace. A countered or fizzled one goes to the graveyard and creates nothing. A copy (`StackItem.IsCopy`) does neither.
- **The trigger and the record.** The first resolution of a name schedules one delayed trigger for the spell's controller: `At: StepPrecombatMain`, `ControllerTurnOnly`, `IndefiniteDuration()`, so it fires at each of their precombat main phases and stays queued (CR 603.7b). It carries the name in `Params.Name` and is itself the "first time this game" record: a later resolution of the same name by the same player finds it and schedules nothing. Body `paradigm/copy`, plain data, so a table holding one is a restore point.
- **The copy.** The body copies the exiled card through `CopyCardsAndOfferCastsForEffect` (#3258) and offers a `{0}` cast with flash timing that closes on the holder's next pass; an uncast copy ceases to exist (CR 704.5e).
- **Simplifications.** The copy is made from the card in exile (or a graveyard), not from a stored last-known spell, so a paradigm card that has gone to a hand or library offers nothing. A copy of a paradigm spell made by another effect (Twincast) does not start the record.
- **Cards:** Decorum Dissertation (new), Echocasting Symposium and Improvisation Capstone (caveats cleared), all Full.
