---
title: "Abilities that can't be countered"
date: 2026-10-16
issues: [2014]
---
**Abilities that can't be countered** (#2014, [ADR 0106 §4 amendment 2026-10-16](decisions/0106-five-small-seams-from-the-s50-rechecks.md)) — `game.CounterShieldStatic.Abilities` extends a battlefield "can't be countered" static from spells to activated and triggered abilities on the stack (CR 701.6a), built with `effects.SpellsAndAbilitiesCantBeCountered`. The ability half of the counter gate, `abilityCantBeCounteredLocked` (`game/cant_be_countered.go`), is asked by every counter verb for an ability item, so Stifle, Disallow, Tale's End and a declined ward or counter-unless-paid all stop at it. `Whose` is judged against the ability's controller, and "you cast" never covers an ability. A covered ability stays a legal target: the Stifle resolves and does nothing (CR 101.2). The stack chip, `cant_be_countered`, now shows on an ability item too. No snapshot change. **Cards:** Spider-Punk (Full).
