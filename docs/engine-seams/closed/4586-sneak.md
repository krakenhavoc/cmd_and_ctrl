---
title: "Sneak"
date: 2026-10-30
issues: [4586]
---
**Sneak** (#4586, [ADR 0025 amendment 2026-10-30](decisions/0025-alternative-costs.md), CR 702.190) — `effects.Sneak(cost)` is an alternative cost: the sneak mana plus `ReturnToHand` narrowed by ninjutsu's `UnblockedAttacker`, so the picker, the enumerator and the validator share one walk. `AlternativeCost.Sneak` times the offer in `CastTimingForOfferOpenLocked` (instant-speed for its claim, and only in its caster's own declare blockers step) and, on a permanent spell, enters it tapped and attacking what the returned creature was attacking: the cast reads the defender before the return onto `PaidCost.ReturnedAttacking`, and `applySneakEntryLocked` seeds `EntersTapped` / `EntersAttacking` onto the entry event (CR 506.3a: a noncreature is not attacking). No snapshot change. Cards (all Full): Elektra, Daughter of the Hand. TMNT (#4457) cards declare the same constructor.
