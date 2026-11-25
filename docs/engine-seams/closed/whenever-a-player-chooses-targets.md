---
title: "Whenever a player chooses one or more targets"
date: 2026-11-25
issues: [6228]
---
**Whenever a player chooses one or more targets** (#6228, CR 601.2c, 602.2b, 603.3d, 707.10c, 115.7a) — Psychic Battle. `game.EventTargetsChosen` is emitted once per announcement, after its `EventBecomesTarget` fan-out, from a cast, an activation, a triggered ability's placement, a copy and a change of targets (`emitTargetsChosenLocked`, `game/targets_chosen.go`), gated on a battlefield watcher like `EventTriggerPlaced`. A change of targets is stamped with the changing card's name (`Event.RetargetSourceName`), so "changing targets this way doesn't trigger abilities of permanents named Psychic Battle" is a filter. `game.RetargetChangeAll` is CR 115.7a's "change the target or targets": every target to another legal target, staged slot by slot and committed once, or none of them (`game/retarget_change_all.go`). Card helper: `effects.TopCardContestChangesTargets(name)`. **Cards:** Psychic Battle (b665) is built by its OLD batch.
