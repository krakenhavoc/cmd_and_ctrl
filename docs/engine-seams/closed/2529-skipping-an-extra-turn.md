---
title: "Skipping an extra turn"
date: 2026-10-07
issues: [2529]
---
**Skipping an extra turn** (#2529, [ADR 0059 amendment 2026-10-07](decisions/0059-turn-machinery.md)): "if an opponent would begin an extra turn, that player skips that turn instead" (CR 614.1b, 614.10) is `effects.SkipOpponentsExtraTurns()`, a pure-cancel replacement over the new `RepEventExtraTurn`, which `popExtraTurnLocked` opens as a queued extra turn would begin (watch key `EventExtraTurnBegin`, an unemitted sentinel). A cancel is the skip: the turn never begins, the next queued turn or normal rotation follows, a delayed trigger bound to it is swept (CR 614.10a — an opponent's Final Fortune never loses them the game), and the turn does not count toward `TurnsBegun`. The window cannot pause (the seam has no resume), so it sets `mustSettleNow` and orders several skips in gather order, which is unobservable while every replacement of the kind is a cancel. The public log gains `extra_turn_skipped`. Trouble in Pairs is `full`. Still open: "skip your next turn" (Magosi, Lethal Vapors) and Savor the Moment's skipped untap step.
