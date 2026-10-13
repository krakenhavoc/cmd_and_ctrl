---
title: "The six remaining villainous-choice cards"
date: 2026-10-13
issues: [3143]
---
**The six remaining villainous-choice cards** (#3143, CR 701.55, [ADR 0018 amendment 2026-10-13 (second)](decisions/0018-triggers-on-the-stack.md)) — Missy, Hunted by The Family, Genesis of the Daleks, The Master, Gallifrey's End, Midnight Crusader Shuttle and Ensnared by the Mara each used the keyword action and needed one more piece.
- **A chain of choices.** Hunted by The Family faces one choice per target, in APNAP order of the targets' controllers, each body starting the next once its last question is answered (`Facing.Remaining`).
- **The power that died.** `TurnTally.DiedPowerBySubtype` / `Game.DiedPowerWithSubtypeThisTurn`: the total last-known power of creatures of a type that died this turn.
- **Into combat without being declared.** `Game.PutOnBattlefieldAttackingForEffect` puts a creature already in play into an attack (CR 506.3c) after a same-effect control change.
- **A tied "most life" opponent.** `effects.OpponentsWithTheMostLife` and `ChoosePlayer.Only`.
- **Cards (6):** all Full. The pieces the issue thought were missing (a graveyard put with a listed face-down body, a no-duration set-characteristics effect, the free cast from exile) already existed.
