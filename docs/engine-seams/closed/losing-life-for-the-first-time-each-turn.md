---
title: "Losing life for the first time each turn"
date: 2026-10-08
issues: [2540]
---
**Losing life for the first time each turn** (#2540) — `effects.WheneverYouLoseLifeForTheFirstTimeEachTurn` reads the turn tally. The tally listener runs ahead of the trigger harvester, so a trigger's `AppliesTo` sees `PlayerTurnTally.LifeLost` already including the loss it is asked about, and a loss is the turn's first exactly when the tally equals it. Several creatures dealing combat damage to one player are one event each, so the first sees only itself and the rest see the running total: the trigger fires once. A gain is not a loss, and damage that costs no life (infect, prevention, a locked life total) neither triggers nor uses up the turn's first. Cards: Gonti's Machinations and Vengeful Warchief. No snapshot change.
