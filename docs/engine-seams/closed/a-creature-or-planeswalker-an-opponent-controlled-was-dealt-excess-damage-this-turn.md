---
title: "A creature or planeswalker an opponent controlled was dealt excess damage this turn"
date: 2026-11-12
issues: [5663]
---
**A creature or planeswalker an opponent controlled was dealt excess damage this turn** (#5663) — excess damage was rebuilt per event inside a trigger (`excess_damage.go`, noncombat damage to creatures only) and nothing kept it for the turn. `TurnTally.ExcessDamaged` now records, per controller as of the damage, that a creature was dealt more than lethal damage or a planeswalker more than its loyalty (CR 120.10), judged in `applyDamageToPermanentLocked` before the damage is marked: damage already marked counts toward lethal, so simultaneous sources are read together, and any amount greater than 1 from a source with deathtouch is excess (CR 120.4a). Combat or not, from any source. `Game.OpponentPermanentDealtExcessDamageThisTurn` reads it; `CondOpponentPermanentDealtExcessDamage` and `AtYourEndStepIfOpponentPermanentDealtExcessDamage` are the card-side intervening "if" (CR 603.4), with the footprint condition `opponent_permanent_dealt_excess_damage`. The snapshot field is additive. **Cards:** Rith, Liberated Primeval (Full).
