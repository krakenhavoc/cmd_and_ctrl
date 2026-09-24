---
title: "\"You gain life equal to the damage dealt this way\""
date: 2026-09-17
issues: [807]
pr: 813
legacy_order: 131
---
**"You gain life equal to the damage dealt this way"** (#807) - the damage half of the entry above, and the same idiom: `Game.DealDamageToPlayerThenForEffect` / `Game.DealDamageToCreatureThenForEffect` for one target and `Game.DealDamageEachThenForEffect(source, targets, amount, then)` for "each opponent", on the `then` that now rides `damageTail` (`server/internal/game/damage_tail.go`). `then` gets the post-replacement amount, or 0 when the damage was prevented, Fogged or its target left. The batch routes each target as the kind of thing it is, so a mixed "each opponent and each creature they control" list is one call. Creeping Bloodsucker is on it; anything that says "equal to the damage dealt" can be written.
