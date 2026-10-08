---
title: "Counters that can't be removed"
date: 2026-10-08
issues: [1824]
---
**Counters that can't be removed** (#1824, [ADR 0058](decisions/0058-doesnt-untap.md) 2026-10-08 amendment) — `Spec.CounterRemovalLocks` declares a counter kind that can't be removed from a filtered set of permanents. It is a derived static read at the one removal choke point (`applyCounterByLocked`), so a stun counter on a locked permanent survives its untap step, a removal effect does nothing, and a counter-removal cost is refused up front (`counterKindsPaying`, `validateCounterRemovalLocked`) rather than paid for nothing. The lock lifts the moment its source leaves the battlefield. **Cards:** Fear of Sleep Paralysis (Full).
