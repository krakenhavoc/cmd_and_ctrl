---
title: "A granted ability that lasts until a card is cast from exile"
date: 2026-10-14
issues: [3520]
---
**A granted ability that lasts until a card is cast from exile** (#3520, [ADR 0063](decisions/0063-durations-and-control.md) amendment of 2026-10-14) — `game.UntilCardCast` is the ninth `DurationKind`. It names the exile object (`Duration.Source` plus `SourceEpoch`) and ends on an event: `CastSpell` marks `Ended` on every scoped effect naming the card it is casting out of exile and sweeps them, so a granted ability leaves the layer pass as the spell goes on the stack, after its mana has paid for the cast. A card leaving exile uncast ends nothing, and a new object of the same card never matches. `Game.UntilCardCastDuration(card)` builds it for `GrantAbilitiesFor`. The cost half: `ExileSelfZoneSupported` now includes the hand, so an activated ability that functions from the hand can have "Exile this card from your hand" as a cost. Snapshot: `scopedEffects[].duration.Kind` takes the new value; no new key. **Cards:** Emrakul, the Exigent Doom (the caveat removed, now Full).
