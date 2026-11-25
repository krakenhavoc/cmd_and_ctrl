---
title: "Commander tax paid in life"
date: 2026-11-25
issues: [6151]
---
**Commander tax paid in life** (#6151, CR 903.8, CR 119.4) — `Spec.CommanderTaxInLife` (`CardDef.CommanderTaxInLife`, read through `game.CommanderTaxInLifeFor`) is "Rather than pay {2} for each previous time you've cast this spell from the command zone this game, pay 2 life that many times": the same count of previous casts (`Player.CommanderCasts`) charges 2 life each instead of {2}. `commanderTaxLocked` is the one reader: the pricer adds no tax mana for such a card, and CastSpell checks the life against the caster's life total with every other life cost of the cast (refusing the cast when it can't be paid) and pays it with them while the spell is on the stack. The tax stays an additional cost, so it applies under an alternative cost too (CR 118.9d). The command zone's `cast_prices` carries it as `life`, broken out as `commander_tax_life` for the hover text, and the bot enumerator states it in the move's life cost and never offers the cast at the cost of the seat's last life. **Cards:** Liesa, Shroud of Dusk (Full).
