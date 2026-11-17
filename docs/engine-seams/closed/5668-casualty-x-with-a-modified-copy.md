---
title: "Casualty X with a modified copy"
date: 2026-11-12
issues: [5668]
---
**Casualty X with a modified copy** (#5668, CR 702.153a / 107.3a / 707.9b / 608.3f / 306.5b / 704.5i) — casualty was a fixed "creature with power N or greater" and its trigger copied with no exception. `effects.CasualtyX()` is the optional cost keyed `casualty` with `AdditionalCost.CasualtyX` set: one creature with power 0 or greater, sacrificed with the spell on the stack. X is the sacrificed creature's power as it last existed on the battlefield (CR 608.2h), the greatest X that creature can pay for; announcing a smaller X for the same creature is not offered. The engine's casualty trigger (`game/casualty.go`, the same keyed body `casualty/copy`) carries the sacrificed object on `Params.Object`, read off `PaidCost.SacrificedObjects`, and makes its first copy with the printed exceptions: not legendary, starting loyalty X. The copy of the permanent spell becomes a token as it resolves.
- **Starting loyalty 0 is a number.** `PrintedValues.SetStartingLoyalty` sets the copiable starting loyalty and marks it known (`Card.StartingLoyaltyKnown`), so the CR 306.5b stamp does not fall back to the catalog's printed loyalty: a copy made by sacrificing a creature with power 0 enters with no loyalty counters and CR 704.5i puts it into the graveyard.
- **Held to its shape.** `Register` allows casualty X only as the card's own optional cost built by `CasualtyX()`, on a card with no other sacrifice among its additional costs (X is read off the first permanent the cast sacrificed), and `SpellsYouCastHave` refuses to grant it.
- **Snapshot.** Additive: `startingLoyaltyKnown` on a card and `StartingLoyaltyKnown` in its copiable values. No new effect key for the seam, no wire change: the cost is offered by the optional-cost picker like any casualty.

**Cards:** Ob Nixilis, the Adversary (Full).
