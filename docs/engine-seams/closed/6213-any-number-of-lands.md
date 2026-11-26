---
title: "Play any number of lands, and a trigger on each land after the first"
date: 2026-11-25
issues: [6213]
---
**Play any number of lands, and a trigger on each land after the first** (#6213, CR 305.2) — Fastbond's "You may play any number of lands on each of your turns" is `AdditionalLandPlays: game.UnlimitedLandPlays`, a sentinel that `EffectiveLandDropsLocked` treats as a ceiling: any grant of it makes the allowance that number and nothing sums past it, so the client's "no land plays left" never shows. "Whenever you play a land, if it wasn't the first land you played this turn" reads `Event.LandOrdinal` (`land_ordinal`), stamped on every played entry (`Played`, CR 305.1) with how many lands that player has played this turn counting this one; `effects.LandPlayedAfterTheFirst(ev, player)` is the predicate. A land put onto the battlefield is not a play and carries no ordinal. **Cards**: Fastbond (OLD b900) is built by its OLD batch.
