---
title: "Mana spent only on costs that contain {C}"
date: 2026-11-25
issues: [6259]
---
**Mana spent only on costs that contain {C}** (#6259, CR 106.6) — Cultivator Drone's "Spend this mana only to cast a colorless spell, activate an ability of a colorless permanent, or pay a cost that contains {C}." `ManaSpendContext` gained a `CostHasC` bit beside `CostHasX`, set at the same sites from the same cost (`game.CostStringHasColorless`: a `{C}` or a colourless hybrid): `ManaSpendForCast` from the spell's mana cost, `ManaSpendForCastParams` from a claimed alternative cost, and the activation, the legal-move enumerator and the `?ability=` cost preview from the row's mana. `game.ManaRestrictColorlessCost` (`"cost:has-c"`) matches it for a cast or an activation; the mana may then pay any part of that cost. A payment that does not set the bit refuses the mana, the weaker direction. The snapshot gains the spend context's `CostHasC` key, additive under v9. **Cards:** Cultivator Drone is built by its OLD batch (b296).
