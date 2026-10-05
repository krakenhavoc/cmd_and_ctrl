---
title: "Mana you don't lose as steps and phases end"
date: 2026-10-05
issues: [2166]
---
**Mana you don't lose as steps and phases end** (#2166, [ADR 0040](decisions/0040-mana-pipeline.md) amendment 2026-10-05) — the step-boundary sweep (`sweepManaPoolLocked`, `game/mana_keep.go`) keeps or converts each unspent mana from three sources: a static over a player's pool derived from the battlefield (`Spec.ManaPool`: keep all, keep some colours, or Kruphix's becomes-colourless, which keeps the mana's restrictions), a granted player statement with a duration (`PlayerStatic.KeepManaColors`), and a per-mana mark (`ManaRiderKeepUntilEndOfTurn`, `effects.KeepManaUntilEndOfTurn()`) that expires when the cleanup step begins. The mark rides the token's rider slot, so it is additive in the snapshot and carried by every mint route. **Cards:** Upwelling, Kruphix, God of Horizons, Karn, Legacy Reforged, The Last Agni Kai and Savage Ventmaw (Full); Leyline Tyrant (caveat: its dies trigger). **Not built:** Omnath, Locus of Mana, whose +1/+1 per unspent green mana needs a layer 7 read of the pool.
