---
title: "Playing an exiled card until you exile another with the same source"
date: 2026-10-08
issues: [2539]
---
**Playing an exiled card until you exile another with the same source** (#2539, ADR 0066 amendment of 2026-10-08) — `game.UntilSourceExilesAnother` is a new duration kind. The permission names its holder and the source object, and it ends only when `Game.ExileTopUntilAnotherForEffect`, the one helper that exiles "with this", marks it ended on the holder's next exile by the same source object. An exile that exiled nothing closes nothing, a new object of the same card never matches (CR 400.7), the newest card stays playable after the source leaves the battlefield, and an ability already on the stack when the source left still closes the old window. An older binary refuses a file that carries the kind. Cards: Unstable Amulet, Furious Rise and Superior Foes of Spider-Man. Snapshot: two new duration keys, `SourceEpoch` and `Ended`, additive within v7.
