---
title: "A cost modifier that prices ninjutsu abilities"
date: 2026-11-12
issues: [5672]
---
**A cost modifier that prices ninjutsu abilities** (#5672, CR 702.49a / 702.49d / 602.2b / 601.2f / 118.7a) — an activation cost modifier picks a keyword by a bit on the priced ability (exhaust, boast, equip, power-up), and ninjutsu had none: the keyword was a label, a cost and a zone list, so "Ninjutsu abilities you activate cost {1} less to activate" had nothing to match but label prose. `ActivatedAbilityShape.Ninjutsu` is that bit, set by `effects.Ninjutsu` and `effects.CommanderNinjutsu` (the variant, CR 702.49d), copied onto `game.AbilityCostSubject.Ninjutsu` when the activation is priced and read by `effects.ANinjutsuAbilityCost()`. It gates nothing; the keyword's rules stay in its cost, zones and effect.
- **Pricing.** No new path: a hand or command-zone activation was already priced through `AbilityManaCostAtXForEffect` with the source's zone, so the payment, the menu row and the bot enumerator all see the reduced cost. The reduction takes off generic mana only (CR 118.7a).
- **Granted abilities.** A ninjutsu ability a static gives to a card in hand (Satoru Umezawa) is built by the same constructor and carries the bit.
- **Guards.** `effects.Register` panics when a printed or granted row's label starts with the keyword without the bit, or sets the bit without the keyword.
- **Not persisted.** The bit is catalog data on the ability shape; nothing is added to the snapshot or the wire.
- **Cards:** Silver-Fur Master (Full).
