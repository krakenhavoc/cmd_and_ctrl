---
title: "Promises about the next spell you cast"
date: 2026-10-07
issues: [1852]
---
**Promises about the next spell you cast** (#1852, [ADR 0106](decisions/0106-five-small-seams-from-the-s50-rechecks.md) amendment 2026-10-07) — one data payload, `PlayerStatic.NextSpell` (`game.NextSpellPromise`, `game/next_spell_promise.go`), held on the caster and spent at CR 601.2i by the first spell its filter matches, so a spell that is cast and then countered still used it. Its riders are read at three times: flash before the cast (folded into `castTimingVerdictLocked`, so the cast path, the bot enumerator and `castable_here` agree), a cost reduction or affinity while the cast is priced (one more CR 601.2f reduction in `activeCostModifiersLocked`), and an extra counter or uncounterability after it, as marks on the stack item (`StackItem.PromisedCounters`, read by the entry-counter pipeline, and the existing `CantBeCountered` marks). A rider outside those registers a `RegisterCastFollowUp` body. **Cards:** Savage Summoning, Quicken, Hardened Berserker, Kaza, Roil Chaser (Full); Saheeli, the Gifted (caveat: a token copy of an artifact with an on-enter hook skips it).
