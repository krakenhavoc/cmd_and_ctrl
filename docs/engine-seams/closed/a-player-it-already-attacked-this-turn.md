---
title: "A player it already attacked this turn"
date: 2026-10-05
issues: [2171]
---
**A player it already attacked this turn** (#2171, CR 508.1c) — `AttackTargetRestriction` gains a `NotAlreadyAttackedThisTurn` clause that reads `TurnTally.Attacks` for the attacker's current object epoch and refuses a player it has already been declared attacking this turn. A flickered creature is a new object (CR 400.7) and may attack that player again; planeswalkers and battles are never refused by the clause. Both declaration verbs, the CR 508.1d search and the enumerator's per-attacker target list share the predicate, so only the players left are offered. The snapshot gains one additive field. **Cards:** Bloodthirster (Full).
