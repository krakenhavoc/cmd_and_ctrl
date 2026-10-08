---
title: "Can't be blocked by creatures with low power, this turn"
date: 2026-10-08
issues: [2600]
---
**Can't be blocked by creatures with low power, this turn** (#2600, CR 509.1b) — a new scoped block-rule kind, `ModCantBeBlockedByPower`, reads the power ceiling from `Mod.Amount` and its refusal clause from `Mod.Text`. It is pinned to the attacker at resolution (CR 611.2c), swept at end of turn, and adapted by the block-rule walk into a pair rule, so the block validator and the enumerator both read it. The blocker's power is compared live as blockers are declared, so a pump or shrink made after the effect moves it in or out of the barred set. The card helper is `effects.CantBeBlockedThisTurnByPower`; `effects.ExertedGetsAndCantBeBlockedByPower` registers Rhonas's Stalwart's pump and rule as one record. Snapshot: the kind is new vocabulary and reuses existing fields.
