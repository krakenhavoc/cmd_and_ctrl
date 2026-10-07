---
title: "Damage that stays through cleanup"
date: 2026-10-07
issues: [2058]
---
**Damage that stays through cleanup** (#2058) — `Spec.DamageStaysThroughCleanup` declares the printed static "Damage isn't removed from this creature during cleanup steps". The CR 514.2 sweep (`sweepTurnEndLocked`) skips a permanent that has the ability, judged through `CatalogAbilityKey` so a permanent that lost all abilities (or this one) is cleaned as usual. The sweep now also clears damage on phased-out permanents, which CR 514.2 names and CR 702.26b says cannot keep the exemption. The kept damage is ordinary marked damage: regeneration and leaving the battlefield still clear it, and it counts for lethal damage on later turns. **Cards:** Ancient Adamantoise (Full).
