---
title: "Removing a chosen number of counters from among permanents"
date: 2026-10-14
issues: [3276]
---
**Removing a chosen number of counters from among permanents** (#3276, [ADR 0152](decisions/0152-more-than-meets-the-eye-convert-and-living-metal.md)) — a counter cost can be variable AND among: `CounterRemovalCost{Variable: true, Among: true}`. The activator announces the total (at least the floor) and the per-permanent split, validated as a set by the same `validateCounterRemovalLocked`, and the total is the paid-cost record (`ctx.CountersRemoved()`). `effects.Register` accepts the combination for a printed kind and a floor of one or more, and still refuses an any-kind variable split and a floor of zero, which no card prints. The bot enumerator offers the one largest payment (every counter of the kind on every matched permanent), and the client's many-pick counter modal already handled it. **Cards:** Jetfire, Ingenious Scientist // Jetfire, Air Guardian (Full).
