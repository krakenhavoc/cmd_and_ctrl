---
title: "Every graveyard arrival opens the CR 614 window"
date: 2026-09-18
issues: [931]
pr: 964
legacy_order: 77
---
**Every graveyard arrival opens the CR 614 window** (#931, [ADR 0013 §5q](decisions/0013-replacement-effects.md)): the last two movers that put a card into a graveyard without asking now go through the shared exit primitive — a library SEARCH with `Dest: game.ZoneGraveyard` (`executeSearchTakeLocked`, sixth route template `searchRoute(player, dest)`; the hand destination rides the same take, so a tutored commander gets CR 903.9 there too) and SURVEIL's graveyard leg (`ResolveSurveil`, on `millRoute`, keeping the `EventMill` it has emitted per binned card since S22). Both join the one batch body (`routeAllThenLocked`), so both can PAUSE: a search reports the cards that ARRIVED (CR 400.7) and finishes its `EventSearchLibrary`, its shuffle and its `Then` from the continuation; surveil's `EventSurveil` (`Amount` = how many really reached a graveyard) and the rest of its effect do the same. Cards declare the sentence through `effects.GraveyardBecomesExile{OpponentsOnly}.Build()` (`cards/effects/graveyard_replacements.go`), which watches both exit kinds. **2 cards shipped off #383's skip list: Rest in Peace** (caveat: a dying commander sent to the command zone is not exiled) **and Leyline of the Void** (same caveat, plus no "begin the game with it on the battlefield"). Still waiting: **Dauthi Voidwalker**, which needs this plus a void counter on the exiled card and permission to cast it.
