---
title: "Marked damage survives a replaced destruction"
date: 2026-09-17
issues: [708, 816]
pr: 813
legacy_order: 134
---
**Marked damage survives a replaced destruction** (#708, #816) - `DamageMarked` and the CR 702.2c deathtouch flag are cleared by `clearBattlefieldDamage` from `MoveCard`'s battlefield-exit cleanup (`server/internal/game/zone.go`) - the landed outcome of EVERY exit, not just a destruction and not by the destroy entry points. #816 moved it there from `executeBattlefieldLeaveLocked`, which had left an exiled or bounced permanent carrying its damage into the new zone (CR 400.7: a new object), so it showed the number there and brought it back onto the battlefield when the card was replayed. A replacement that keeps the permanent in play (or that wants to read how much damage is on it) now sees the board as it is, which is the seam a regeneration shield or an "if it would be destroyed, instead" card needs. Regeneration shipped in #667 and removes the damage in its own replacement (CR 701.19a), through the same `clearBattlefieldDamage` helper.
