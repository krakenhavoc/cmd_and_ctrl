---
title: "Can't be countered by spells of a colour"
date: 2026-11-25
issues: [6263]
---
**Can't be countered by spells of a colour** (#6263, CR 701.6a, 702.16b, 611.2c) — Autumn's Veil's "Spells you control can't be countered by blue or black spells this turn, and creatures you control can't be the targets of blue or black spells this turn." The counter-shield turn grant gains `CounterShieldGrant.ByColors` ("UB"), and the gate asks it of the spell resolving as the counter is made (`counteringSpellSharesColorLocked`): a colourless or red spell, and any ability, still counter. The targeting clause is a live `ScopeYourCreatures` record with `ModCantBeTargetedBySpellsOfColors`, read by `canBeTargetedBy` for spells only, so a blue spell already aimed at a creature loses that target at resolution (CR 608.2b). The card primitives are `effects.SpellsCantBeCounteredByColorsThisTurn` and `CantBeTargetedBySpellsOfColorsThisTurn`. See ADR 0106's 2026-11-25 amendment. **Cards:** Autumn's Veil is built by its OLD batch (b452).
