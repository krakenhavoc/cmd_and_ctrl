---
title: "Getting energy, and energy paid or lost this turn"
date: 2026-10-08
issues: [1995]
---
**Getting energy, and energy paid or lost this turn** (#1995, [ADR 0129](decisions/0129-energy-getting-and-paying-it.md) §6, PR 5) — `PlayerTurnTally.EnergyPaidOrLost`, read through `Game.EnergyPaidOrLostThisTurn`, adds every negative energy delta that lands on a player: a payment through `payEnergyLocked` and a removal by an effect both count, and only what came off. It resets with the turn's tally and is an additive snapshot field in schema 7. "Whenever you get one or more {E}" is `effects.WheneverYouGetEnergy`, a trigger on `EventPlayerCounterPlaced` with a positive energy delta, one placement one trigger (CR 603.2c), and `EnergyGotten(item)` is "that much". `PaidOrLostEnergyThisTurn(n)` gates an activation and `CostsLessForEachEnergyPaidOrLost(n, label)` discounts a spell. **Cards** (6, all Full): Izzet Generatorium, Blaster Hulk, Aether Revolt, Brotherhood Scribe, Fabrication Module and Territorial Gorger. Closes `energy-paid-or-lost`.
