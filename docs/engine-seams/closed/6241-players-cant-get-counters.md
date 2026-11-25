---
title: "Players can't get counters"
date: 2026-11-25
issues: [6241]
---
**Players can't get counters** (#6241, CR 122.1, 101.2) — Solemnity's "Players can't get counters" and Suncleanser's "That player can't get counters for as long as this creature remains on the battlefield". A world rule the permanent carries on itself (`game.ExtraRestriction` `PlayersCantGetCounters`) and a `ModPlayerCantGetCounters` scoped record (`playerCantGetCounters`, an on-disk identity) naming one player or every player, for the duration the card states. Both are read by `playerCantGetCountersLocked` at the one player counter door, `counterPlacementRefusedLocked`, before the CR 614 window opens, so no replacement or trigger sees a refused placement. That covers every kind: poison, energy, experience, rad and tickets. Proliferate offers no such player, and counters already there stay. Card helpers: `effects.PlayersCantGetCounters` and `PlayerCantGetCountersWhileSourceRemains`. **Cards:** Suncleanser (b198) and Solemnity (b241) are built by their OLD batches.
