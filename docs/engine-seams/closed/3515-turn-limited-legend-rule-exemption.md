---
title: "A legend-rule exemption that lasts one turn"
date: 2026-10-14
issues: [3515]
---
**A legend-rule exemption that lasts one turn** (#3515, [ADR 0032](decisions/0032-planeswalkers.md) amendment 2026-10-14, CR 704.5j) — `Spec.LegendRuleExemptions` (#2177) is a static read live off the battlefield with no duration, so a land declaring "the legend rule doesn't apply to permanents you control this turn" would have switched the rule off for its controller for the whole game. `ModLegendRuleDoesntApply` is now a stored `ScopedEffect` rules gate (reader `rule`, scope game, reads `Player`, swept at cleanup, CR 514.2), written by `effects.LegendRuleDoesntApplyThisTurn` and read by `legendRuleExemptControllersLocked` beside the printed statics, so the two compose. It covers only the named player's permanents; the first state-based action check after cleanup applies the rule again with the controller choosing. One new on-disk mod kind under schema v8, with one fixture. **Cards** (1): Hall of Echoes (Full).
