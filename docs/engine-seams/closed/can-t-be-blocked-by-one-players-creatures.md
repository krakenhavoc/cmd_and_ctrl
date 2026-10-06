---
title: "Can't be blocked by one player's creatures"
date: 2026-10-05
issues: [2172]
---
**Can't be blocked by one player's creatures** (#2172, CR 509.1b) — a new scoped block-rule kind, `ModCantBeBlockedByPlayer`, stores the chosen player in `Mod.Player` and its refusal clause in `Mod.Text`. It is pinned to the target at resolution (CR 611.2c), swept at end of turn, and adapted by the block-rule walk into a pair rule, so the block validator and the enumerator both refuse a barred blocker; the blocker's controller is compared live. The kind is new vocabulary and reuses existing mod fields. **Cards:** The Black Gate (Full).
