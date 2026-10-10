---
title: "Power and toughness set to a count"
date: 2026-10-09
issues: [2569]
---
**Power and toughness set to a count** (#2569, CR 608.2h, [ADR 0032 amendment of 2026-10-09](decisions/0032-planeswalkers.md)) — Gideon, Champion of Justice's "becomes a Human Soldier creature with power and toughness each equal to the number of loyalty counters on him" reads the count once, as the ability resolves, and he keeps that size until end of turn (owner decision: locked, not live). A resolved effect determines game information once, when it is applied (CR 608.2h), so no new Mod was needed: `gideonAnimation.SizeFromLoyalty` makes `effects.animateGideon` read the source's loyalty counters at resolution and write that number into the existing layer 7b `SetBasePowerMod` / `SetBaseToughnessMod`. Damage that can't be prevented or a proliferate later in the turn changes his loyalty, not his size, and he is still a planeswalker, so his last loyalty counter still takes him (CR 704.5i). **Cards:** Gideon, Champion of Justice (Full): the +1 counts the target opponent's creatures as it resolves, the 0 is the becoming above with the #2046 damage shield, and the −15 exiles every other permanent.
