---
title: "Generic step-entry trigger event"
date: 2026-09-15
issues: [588]
pr: 593
legacy_order: 113
---
**Generic step-entry trigger event** (#588) - `EventStepBegan` carries a typed `Step` and the harvester watches it; `AtBeginningOfYourCombat`, `AtYourPostcombatMain`, `AtEndOfYourCombat`, `AtYourStep` / `AtEachStep` in `triggers_common.go`. The 15 cards that waited on it still need writing: Black Market Connections (#294); Seedborn Muse (#295); Unwinding Clock (#297); Ripples of Undeath (#297); Bender's Waterskin (#298); Mesmeric Orb (#302); Intruder Alarm (#308); Drumbellower (#310); +7 more.
