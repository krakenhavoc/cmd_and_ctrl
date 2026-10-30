---
title: "A Hero or noncreature spell, and flash if cast using teamwork"
date: 2026-10-30
issues: [4455]
---
**A Hero or noncreature spell, and flash if cast using teamwork** (#4455, CR 601.3c, CR 607.2a, CR 601.3f) — two small additive seams for two Marvel cards.
- **`PermissionFilter.NoncreatureOrCreatureType`** is the OR that `CreatureType` (an AND) could not say: a card qualifies when it is not a creature or has the named creature type. It is judged on the face being cast like every spell clause and carried through `ThroughTextOf`. Scarlet Witch, Chaotic Avenger stamps ONE `ScopeCards` permission (`{0}`, pass-closed) over every card still exiled with her (`Card.ExiledWith`, face down via `ExileHiddenLinkedForEffect`), so the free cast reaches the cards of earlier hits and is one cast.
- **`AdditionalCost.AsThoughFlash`** on an OPTIONAL cost is "you may cast this spell as though it had flash if it's cast using" it. `game.OptionalCostsOpenCastWindow` reads the announced costs and `CastTimingForCastOpenLocked` joins them to the offer's own flash (so a per-player restriction still closes it). CastSpell, the enumerator (a closed window drops only the sets that do not announce the cost) and the view's exile and graveyard `castable_here` (the teamwork must be payable) read the same function. `effects.TeamworkOpeningFlash(n)` is the card vocabulary, and `Register` refuses the flag on a mandatory cost. No protocol change.
- **Snapshot:** one recorded shape line per path that reaches a `PermissionFilter` (`...filter.noncreatureOrCreatureType`, omitempty) in `v9.txt`; `AdditionalCost` is catalog data and is not persisted.
- **Cards (2, Full):** Scarlet Witch, Chaotic Avenger; Quantum Reduction.
