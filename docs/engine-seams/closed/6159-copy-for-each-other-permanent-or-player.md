---
title: "Copy a spell or ability for each other permanent or player it could target"
date: 2026-11-25
issues: [6159]
---
**Copy a spell or ability for each other permanent or player it could target** (#6159, CR 707.10d) — `effects.CopyForEachOtherPermanentOrPlayer` copies a spell or ability whose single target is a permanent or a player (`effects.ItemTargetsOnlyASinglePermanentOrPlayer`) once for each other permanent or player it could target. The copies are controlled by the effect's controller, and each targets a different one of those permanents or players. The copies go on the stack in the controller's order, asked one at a time with a keyed option pick (`option-pick/copy-for-each-could-target-order`) whose options show the permanents as cards and the players as seats; the card-order prompt can't hold a player. The prompt is plain data, so the table stays a restore point. Underneath are `CopySpellRetargetedEachForEffect` and `CopyAbilityRetargetedEachForEffect` (Mirrorwing Dragon, Agrus Kos), which now also move a player target onto a permanent and back. **Cards:** Radiant Performer (Full). Ink-Treader Nephilim is left to its batch.
