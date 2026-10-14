---
title: "Changing the life that damage costs"
date: 2026-10-13
issues: [3388]
---
**Changing the life that damage costs** (#3388, [ADR 0169](decisions/0169-the-life-that-damage-costs-runs-the-life-window.md), CR 120.3a, CR 614.1, CR 616.1) — `game/damage_tail.go`. Damage to a player used to lower the life total directly, so a "would lose life" replacement doubled a drain and a life payment but not a Bolt or combat damage.
- **The window.** Once a damage event has settled and been split into its results, `damageLifeLossThroughReplacementsLocked` runs the life it would cost through the CR 614 `RepEventLife` window as a new event. It settles inline (`mustSettleNow`), because the damage tail has no resume between the damage and its result. Gain replacements and the radiation clause cannot match it.
- **Order.** The lock ("your life total can't change") cancels the event as the window's preemptive built-in. The Angel's Grace floor caps the result after the window, the order the affected player would choose under CR 616.1.
- **The record.** A loss larger than the damage is a negative `Event.DamageNotLifeLoss`, so `DamageLifeLoss()` and every life-loss reader see the whole loss with no snapshot change. `Amount`, lifelink and the commander tally keep the damage.
- **Simplification.** Two different life-loss replacements on one damage event apply in gathered order with no CR 616 prompt. No card in the catalog makes that order observable.
- **Cards:** Bloodletter of Aclazotz (new, Full).
