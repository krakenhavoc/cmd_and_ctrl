---
title: "Maximum hand size"
date: 2026-10-03
issues: [2074]
---
**Maximum hand size** (#2074, [ADR 0113](decisions/0113-small-seams-for-the-s58-deck-requests.md) §3) — a permanent's static can now set a player's maximum hand size, change it, or reach players other than its controller: `Spec.HandSize` declares `game.HandSizeStatic{Players, Kind, N, When}`, and `Spec.NoMaxHandSize` folds into the same list. `Game.EffectiveMaxHandSizeLocked` applies every entry that reaches a player in CR 613.11 timestamp order, starting from seven: each battlefield permanent's statics at its own timestamp, read through `CatalogAbilityKey`, and the player's "for the rest of the game" grant at `Player.MaxHandSizeAt`, the time it resolved (CR 613.7b, an additive snapshot field). "No maximum" stays unbounded under a later change and gives way to a later set. The result is never below zero (CR 107.1b). The cleanup discard still asks only the active player, against their own maximum (CR 514.1). The seat panel shows a badge when a player's maximum is not seven. **Cards:** Jin-Gitaxias, Core Augur, Null Profusion and Price of Knowledge (Full).
