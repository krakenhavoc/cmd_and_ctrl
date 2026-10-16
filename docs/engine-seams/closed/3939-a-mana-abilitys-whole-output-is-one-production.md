---
title: "A mana ability's whole output is one production"
date: 2026-10-16
issues: [3939]
---
**A mana ability's whole output is one production** (#3939, [ADR 0013 amendment 2026-10-16](decisions/0013-replacement-effects.md)) — the CR 614 window on mana production (`RepEventProduceMana`, #1222) opened once per mana slot, so Ancient Tomb's `{C}{C}` was two one-mana events and Damping Sphere's "if a land is tapped for two or more mana, it produces {C} instead" never saw two. Now an activation's whole output is one event (CR 106.12b). `ActivateManaAbility` and the auto-tap executor produce a source's mana in one call. An activation with colours to pick asks one `mana_pick` per pick, carrying the colours it knows (`PendingChoice.ManaSettled`, `ManaPicksLeft`), and produces everything on the last answer, so a filter land's two picks are one production and fire the triggered mana abilities once. The planner prices a multi-slot source as one production when that disagrees with the per-slot price (`priceWholeOutputLocked`), and declines a source with picks in that case. The ledger books what actually arrived (`bookProduced`). **Damping Sphere** ships `full`.
