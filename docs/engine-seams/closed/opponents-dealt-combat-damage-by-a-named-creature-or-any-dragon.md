---
title: "Opponents dealt combat damage by a named creature or any creature of a type"
date: 2026-10-15
issues: [3786]
---
**Opponents dealt combat damage by a named creature or any creature of a type** (#3786, CR 510.2, CR 400.7, CR 608.2h) — the combat-damage tally is keyed by the dealer's controller, so it answered "a Dragon you control" and not "a Dragon". `Game.PlayersDealtCombatDamageThisTurnByAnySubtype` walks that cell once per seat and unions the victims (a changeling counts for any creature type), and `Game.ObjectRefDealtCombatDamageToPlayerThisTurn` reads the per-object damage record by a stamped `ObjectRef`, so a resolving ability still counts the damage its own source dealt after the source died or was exiled in response, while a creature that came back as a new object does not. Estinien Varlineau counts the distinct opponents in the union, on the existing `AtYourSecondMainPhase` trigger. **Cards:** Estinien Varlineau (Full).
