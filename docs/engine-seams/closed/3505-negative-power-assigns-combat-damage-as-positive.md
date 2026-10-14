---
title: "Negative power assigns combat damage as positive"
date: 2026-10-14
issues: [3505]
---
**Negative power assigns combat damage as positive** (#3505, [ADR 0045](decisions/0045-combat-restrictions.md) amendment 2026-10-14) — `Characteristic.DamageAsPositivePower` is a per-creature layer 6 flag that `Card.combatDamagePower` reads: a creature with it assigns the absolute value of its unclamped power (`PowerForComparison`) instead of CR 510.1a's zero, while every other creature keeps the clamp. Trample, deathtouch, lifelink and commander damage all read that number. It sits beside `DamageByToughness` (Ghalta). The field is an additive snapshot key inside `Characteristic` (shape v8 recorded). **Cards:** Loot, the Anomaly (Full). Tarmogoyf was reviewed in the same slice and moves from Unreviewed to Full; the review found no engine gap.
