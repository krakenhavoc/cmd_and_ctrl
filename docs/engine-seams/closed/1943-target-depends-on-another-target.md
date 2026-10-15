---
title: "A target that depends on another target"
date: 2026-10-15
issues: [1943]
---
**A target that depends on another target** (#1943, CR 601.2c, CR 608.2b, CR 608.2h, [ADR 0065](decisions/0065-modal-and-multi-target-clauses.md) amendments of 2026-10-11 and 2026-10-15) closes the "A target that depends on another target" row: no card waits on it now.
- **The relation.** `TargetSpec.Relation` binds a single later pick to a required single earlier pick in the same mode occurrence. Announcement, resolution, trigger prompts, copy and retarget prompts, the bot enumerator and the client picker (through the wire's `relation.pairs`) share one judge, `Game.TargetRelationMatchesForEffect`.
- **Three relationships.** `InGraveyardOfTargetsController` (Goblin Welder), `AttachedToTarget` (Fiery Annihilation) and `ControlledByTargetsController` ("another target creature that player controls"). The last reads an anchor that has left the battlefield by its last-known controller, so the later target stays legal and the spell still resolves.
- **Cards:** 3 (Goblin Welder, Fiery Annihilation, Breaking of the Fellowship).
