---
title: "Doesn't untap during the next two untap steps"
date: 2026-10-08
issues: [2029]
---
**Doesn't untap during the next two untap steps** (#2029, CR 502.3, [ADR 0058](decisions/0058-doesnt-untap.md) 2026-10-08 amendment) — the one-shot untap marker counts steps: `UntapSkip.Extra` is the steps it still skips after the next, `Game.SkipNextUntapStepsForEffect` records it, and each matching untap step spends one. Overlapping effects raise the count and never stack, so a second exert still adds no second marker (CR 701.43b). Keyed to "its controller", the count follows whoever controls the permanent at each step. `effects.DoesntUntapNextUntapStep` takes `Steps`. **Telekinesis** is new and `full`.
