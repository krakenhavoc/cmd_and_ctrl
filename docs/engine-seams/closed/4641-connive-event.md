---
title: "A connives event"
date: 2026-10-30
issues: [4641]
---
**A connives event** (#4641, CR 701.50) — `game.EventConnived` is emitted once per conniving permanent, by `Game.ConnivedForEffect(creature, controller)` (`game/connive.go`), when the connive is complete (CR 701.50f). Details are in the 2026-10-30 amendment to [ADR 0025](../../decisions/0025-alternative-costs.md).
- **The event.** `CardID` is the permanent, `Actor` its controller as it connived (its last controller if it has left, CR 701.50b). Emitted after the discard and the counter on every branch: an empty hand and a departed permanent included. Several permanents conniving at once (CR 701.50c) connive in turn, so it is one event each. Connive 0 emits nothing (CR 701.50e). It is silent in the log, because the draw, the discard and the counter are already lines.
- **Emitters.** `ConniveEach`, `ConniveX`, Ledger Shredder and Cyclonus, the Saboteur, each from their own discard continuation.
- **Card side** (`cards/effects/connive.go`): `ACreatureYouControlConnived` and `WheneverACreatureYouControlConnives`.
- **Cards:** Iron Monger, Sadistic Tycoon, Full.
