---
title: "If this mana ability has been activated four or more times this turn"
date: 2026-11-12
issues: [5745]
---
**If this mana ability has been activated four or more times this turn** (#5745, CR 400.7 / 605.3b / 603.7a / 603.7c) — Farrelite Priest's and Initiates of the Ebon Hand's "{1}: Add {W} (or {B}). If this ability has been activated four or more times this turn, sacrifice this creature at the beginning of the next end step." The count was already there: the per-turn half of the activation record (`Game.Activations.Turn`, read with `Game.ActivatedThisTurn(source, label)`) is keyed by the object and the printed ability, emptied on the turn advance, and written for every mana activation by both `ActivateManaAbility` and the auto-tapper's executor before anything is paid. What was missing was a rider that reads it.
- **The rider.** `effects.SacrificeAtNextEndStepIfActivatedRider(n, abilityLabel, triggerLabel)` is a mana ability's `Rider`. It runs after the mana is added (CR 605.3b), when the activation being resolved is already counted, so the fourth reads 4. From then on each activation creates its own delayed trigger at the beginning of the next end step (CR 603.7a); the first to resolve sacrifices the creature and the rest find nothing.
- **What "this ability" counts.** One ability of one object this turn. A creature that leaves the battlefield and returns is a new object with a new count (CR 400.7), and a waiting sacrifice does not reach it (CR 603.7c), because the trigger names the object it was scheduled from. A change of control keeps the count.
- **The auto-tapper** never plans such an ability: the rider is opaque to the planner, and it should not spend a creature the player did not choose to spend.
- **On disk.** One new delayed-trigger body key, `activation-count/sacrifice-this-creature`. No new snapshot field: the tally is the existing `activations.turn`. No closure field changes.

**Cards:** Farrelite Priest, Initiates of the Ebon Hand — Full. Dragon Whelp and Nalathni Dragon print the same sentence on a CR 602 ability, whose activations the same record counts.
