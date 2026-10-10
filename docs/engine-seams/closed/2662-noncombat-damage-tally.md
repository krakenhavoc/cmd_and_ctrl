---
title: "A per-turn tally of noncombat damage by source colour and controller"
date: 2026-10-09
issues: [2662]
---
**A per-turn tally of noncombat damage by source colour and controller** (#2662, CR 608.2h) — `TurnTally.NoncombatDamage` records every amount of noncombat damage dealt this turn, to a player or a permanent, with the source's colours and controller as they were when it dealt the damage: the damage event's `SourceLKI` snapshot, so a burn spell is read as it stood on the stack and a creature that has since died, changed colour or changed hands is read as it was. It is written at the one emitter of a settled damage event (`emitDealDamageLocked`), skips combat damage and damage the engine cannot attribute to a source, resets with the rest of the tally at the turn boundary, and is deep-copied by `cloneTurnTally`, so a snapshot, a clone and an undo carry it. `Game.NoncombatDamageThisTurnForEffect(controller, colour)` sums it. **Cards:** Ojer Axonil, Deepest Might // Temple of Power — Full (Temple of Power's "{2}{R}, {T}: Transform this land. Activate only if red sources you controlled dealt 4 or more noncombat damage this turn and only as a sorcery").
