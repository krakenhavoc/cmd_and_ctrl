---
title: "\"You have hexproof\" only during your turn"
date: 2026-10-16
issues: [2013]
---
**"You have hexproof" only during your turn** (#2013, [ADR 0072](decisions/0072-protection.md) amendment of 2026-10-16) — a permanent's printed "<condition>, you have <keyword>" is `Spec.PlayerKeywordsIf`: the token plus a `While(g, source)` predicate, read through `game.CatalogPlayerKeywordsIf` by the same `playerAbilityTokensLocked` walk that reads `PlayerKeywords`. The condition is asked on every query with the granting permanent as it is now (CR 611.3a), so targeting, the picker, the damage built-in and the attachment check all follow it and nothing is stored, swept or snapshotted. `YouHaveDuringYourTurn` and `YouHaveWhile` build the entries, `Register` holds them to the `PlayerKeywords` vocabulary and requires a condition and a label, and `ThisHasDuringYourTurn` is the object half. **Cards:** Gruul Spellbreaker and Captain America, Super-Soldier (Full).
