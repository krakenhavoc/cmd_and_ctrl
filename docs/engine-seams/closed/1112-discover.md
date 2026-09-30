---
title: "Discover"
date: 2026-09-30
issues: [1112]
---
**Discover** (#1112, [ADR 0099](decisions/0099-discover.md)): "discover N" (CR 701.57) is `effects.Discover{N}` / `DiscoverN(n)` over `Game.DiscoverThenForEffect`, callable from a spell, a trigger or an activated ability, with `Spec.Discovers` declaring it for cards/coverage. It shares cascade's exile-until walk and `may_cast` prompt (now worded per keyword, with `may_cast_keyword` and the branch labels on the wire). Accepting stamps a `{0}` grant with `TimingFlash` (CR 608.2g) and a `MaxSpellManaValue` cap judged against the face being cast; the grant closes on the discoverer's next priority pass (`CastPermission.LapseOnPass`, miracle's rule) and the card goes to hand. `EventDiscover` fires when the card is settled — after the discovered spell's `EventCast`, so "whenever you discover" resolves first — and `WheneverYouDiscover` watches it. Cascade moved onto the same cap, timing and pass-closed window. Primordial Gnawer and Tecutlan, the Searing Rift ship with it; the other discover cards are the card PR's.
