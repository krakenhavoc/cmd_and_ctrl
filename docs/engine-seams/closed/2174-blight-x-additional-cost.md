---
title: "Blight X as an additional cost"
date: 2026-10-06
issues: [2174]
---
**Blight X as an additional cost** (#2174, CR 701.68a, CR 107.3a) — `AdditionalCost.BlightX` (built with `effects.BlightXCost()`) is "As an additional cost to cast this spell, blight X": the caster announces X on `CastSpellParams.XValue`, the one slot Toxic Deluge's pay-X-life uses, names the one creature on `blight_ids`, and the X -1/-1 counters land at CR 601.2h with the spell already on the stack, through the same CR 614 window every blight uses (`blightLocked`), so a creature that dies of them dies before the spell resolves and the cost stays paid. The ceiling printed on the card, the greatest toughness among the caster's creatures, is part of the component and read from one place, `Game.BlightXCeilingForEffect`, by the announce validator, the view (`additional_cost.blight_x_max`, with `demands_x` so the X prompt opens, and `blight_options`) and the bot enumerator, which announces at the ceiling and never offers X=0 or a cast with no creature. `Register` refuses blight X in an optional cost, an either/or branch, or beside pay-X-life or an optional blight. **Cards:** Soul Immolation (Full).
