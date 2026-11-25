---
title: "A trigger when a marked player loses the game"
date: 2026-11-25
issues: [6123]
---
**A trigger when a marked player loses the game** (#6123, CR 603.7a / 603.7d / 800.4a / 707.2 / 608.2h / 122.1) — Miss Highwater's "they … get a contract counter. For as long as they have a contract counter, when they lose the game, for each artifact and creature they controlled, create a token that's a copy of it." had three missing pieces.
- **The counter.** `game.CounterContract` ("contract") is a player counter with no rule of its own, beside poison, energy, experience, rad and ticket; the seat view shows it with the other player counters.
- **The trigger.** `effects.WhenPlayerLosesWithCounter{Player, Counter, Label, Body}` schedules a delayed triggered ability on `EventPlayerEliminated` (emitted before CR 800.4a takes the player's objects), controlled by the scheduling ability's controller and owing nothing to its source, so it still triggers after Miss Highwater has left. Its condition reads the counter on the player as they leave ("for as long as they have a contract counter"); it fires once and lasts as long as the game.
- **What they controlled.** Leaving the game is not a zone change, so no last-known record of the departing player's permanents survived it. `leaveGameLocked` now writes `Player.DepartedPermanents` before the event: each permanent they controlled as a token copy would copy it (its copiable values, CR 707.2) and with the card types it had (CR 608.2h). `game.DepartedPermanentsForEffect` reads it, `effects.CreateTokenCopiesOfDeparted` makes one token copy of each match as one creation (CR 608.2f), and `effects.TokenCopiesOfDepartedArtifactsAndCreaturesBody` is Miss Highwater's body. Carried in the snapshot, additive.
- **Curse of Vengeance** needed none of this: an Aura enchanting a player sees `EventPlayerEliminated` while it is still attached, and reads its spite counters off itself, live or as it last existed (`PermanentForEffect` on the source object), which an engine test pins.
- **Cards** (0 here): Miss Highwater is the GAP3 row's and Curse of Vengeance OLD batch 257's, neither pushed yet.
