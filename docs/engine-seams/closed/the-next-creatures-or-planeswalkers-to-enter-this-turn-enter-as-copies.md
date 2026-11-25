---
title: "The next creatures or planeswalkers to enter this turn enter as copies"
date: 2026-11-25
issues: [6160]
pr: 6254
---
**The next creatures or planeswalkers to enter this turn enter as copies** (#6160, CR 614.1c, CR 707.2) — `game.ModNextEntersAsCopies`, a one-shot, turn-scoped scoped replacement registered by `Game.NextCreaturesOrPlaneswalkersEnterAsCopiesThisTurnForEffect`. The chosen creature is pinned in `Mod.Objects[0]`, and its copiable values are read as the permanents enter, from last-known information once it has left (CR 608.2h). The first entry of a creature or planeswalker spends it: every member of that one simultaneous entry (`entry_batch.go`) copies, recorded in `Objects[1:]` with `SpentBatch`, and nothing after it does. The spent record is dropped at the next event batch. A planeswalker that copies a creature enters with no loyalty. Additive: a new mod kind on existing fields, checked at registration and on restore. **Cards:** Mystic Reflection (Full).
