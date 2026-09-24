---
title: "\"Exile this <permanent>\" as a CR 602 cost, paid from the battlefield"
date: 2026-09-24
issues: [1404]
pr: 1416
legacy_order: 11
---
**"Exile this <permanent>" as a CR 602 cost, paid from the battlefield** (#1404, CR 602.2b / 406 / 903.9, [ADR 0020 amendment 2026-09-24](decisions/0020-activated-abilities.md) Decisions 45-46). This seam had no row of its own. #1381 found it and pinned Perpetual Timepiece in `knownOracleMismatches`. `AbilityCost.ExileSelf` now follows the ability's zone. `game.ExileSelfZoneSupported` (graveyard or battlefield) is the one predicate both `effects.Register` and `validateExileSelfCostLocked` read. From the battlefield the source leaves through the exit primitive's battlefield arm. Leaves-the-battlefield watchers fire; dies and sacrifice watchers do not. The source is gone before the ability resolves, and its leaves-triggers go on the stack above the ability. A commander's owner is asked about the command zone before anything is paid, through #1397's ask-first gate (ADR 0013 §5af), which already lists the source whenever the cost has `ExileSelf`. `AbilityAutoTapExclusions` excludes the exiled source. Cards: Perpetual Timepiece (caveat and pin removed), Hanged Executioner, Nyx Weaver, Feldon's Cane. Still open: a cost that exiles this AND other permanents (Mechtitan Core, craft).
