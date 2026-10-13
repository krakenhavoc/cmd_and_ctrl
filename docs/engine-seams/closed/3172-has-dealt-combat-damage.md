---
title: "A permanent that hasn't dealt combat damage yet"
date: 2026-10-13
issues: [3172]
---
**A permanent that hasn't dealt combat damage yet** (#3172 slice cav-e, [ADR 0071 amendment 2026-10-13](decisions/0071-designations-that-switch-abilities-on.md), CR 120.3 / 400.7) — `Card.DealtCombatDamage` is set when combat damage a battlefield permanent was the source of actually lands (`emitDealDamageLocked`) and cleared when it leaves; `game.HasntDealtCombatDamage()` gates a printed static on its absence. Ruric Thar, Magecrusher has hexproof until it connects.
