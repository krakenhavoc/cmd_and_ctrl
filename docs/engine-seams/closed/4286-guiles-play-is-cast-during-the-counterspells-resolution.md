---
title: "Guile's play is cast during the counterspell's resolution"
date: 2026-10-24
issues: [4286]
---
**Guile's play is cast during the counterspell's resolution** (#4286, CR 608.2g; [ADR 0066](decisions/0066-granted-cast-and-play-permissions.md)'s 2026-10-24 note on the #4250 part 2 amendment) — "instead exile that spell and you may play that card without paying its mana cost" no longer leaves a pass-closed `FreeCastFromExilePermission` for the next priority. Once the countered spell has landed in exile, the replaced counter asks the countering player the `resolution_cast` prompt over it (`Game.OfferResolutionCastForEffect`, keyword `free_cast`, decline "Leave it in exile"), so the cast is made before the counterspell finishes. Every counter the engine has runs inside an open resolution (a hard counter, a declined counter-unless-pays, a ward, and the destination counters Hinder, Devious Cover-Up, Remand and Memory Lapse); a counter made with no resolution open keeps the pass-closed grant. Free, cast-only, X is 0 and a commander's CR 903.9a check are unchanged. No wire or snapshot change.
