---
title: "\"Can't attack you during their next turn\" on a player"
date: 2026-10-08
issues: [2109]
---
**"Can't attack you during their next turn" on a player** (#2109, [ADR 0063 amendment 2026-10-08](decisions/0063-durations-and-control.md), CR 508.1c / 611.2) — `Game.GrantCantAttackPlayerForEffect` appends a `PlayerStatic` to the restricted player whose `CantAttack` payload (`CantAttackGrant{Protected, FromTurnsBegun}`) names the protected player and the turn the window opens at, with an `UntilEndOfYourNextTurnDuration` duration. `playerCantAttackRefusalLocked`, read inside `canAttackTargetWithLocked` (testing the window and the duration itself), refuses the protected player and any permanent they control, so both declaration verbs, the CR 508.1d search and the enumerator follow. Additive snapshot data (`seats[].statics[].cantAttack`). The Second Doctor ships `full`.
