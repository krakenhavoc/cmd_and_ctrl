---
title: "Unattaching an Equipment as a cost, and a division read from it"
date: 2026-11-02
issues: [4799]
---
**Unattaching an Equipment as a cost, and a division read from it** (#4799, CR 601.2b, 601.2d, 602.2b, 701.3d, [ADR 0020](decisions/0020-activated-abilities.md) amendment 2026-11-02) — `game.AbilityCost.UnattachEquipment` is "Unattach an Equipment from <this>": the activator names, in `unattach_ids`, an Equipment attached to the source, whoever controls it (CR 301.5d), and the payment is `UnattachForEffect`, so the Equipment stays on the battlefield and `EventUnattach` fires. With nothing attached the ability can't be activated (CR 118.3). `DivideAmountArgs.Unattached` hands the pick to a divided amount rule, so "damage equal to that Equipment's mana value divided as you choose" is sized from the Equipment the same announcement names (`effects.DivideUnattachedManaValue`). The row ships `unattach_label` / `unattach_options` and the divided clause `divide.by_unattach` (the amount per candidate); the client asks for the Equipment before the targets, and the bot's enumerator offers one row per attached Equipment. **Cards:** Captain America, First Avenger — Full.
