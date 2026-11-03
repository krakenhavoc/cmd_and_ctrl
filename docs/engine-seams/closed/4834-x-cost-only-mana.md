---
title: "Mana spent only on costs that contain {X}"
date: 2026-11-03
issues: [4834]
---
**Mana spent only on costs that contain {X}** (#4834, CR 106.6, CR 107.3) — "Spend this mana only on costs that contain {X}" (Elementalist's Palette). The restricted mana pays the whole of such a cost, not only its X part, so it is a property of the payment rather than of the object paid for: `ManaSpendContext` gained a `CostHasX` bit beside `Equip`, set by `ManaSpendForCast` from the spell's mana cost (by `ManaSpendForCastParams` from the alternative cost the cast claims), and from the {X} symbols of the row's mana (`AbilityCost.XSlots`) by the activation, the legal-move enumerator and the `?ability=` cost preview. `game.ManaRestrictXCost` (`"cost:has-x"`) matches it for a cast or an activation. A payment that does not set the bit refuses the mana, the weaker direction. The snapshot gains the spend context's `CostHasX` key (`pendingChoices[].assist.spend`), additive under v9. **Cards:** Elementalist's Palette (Full).
