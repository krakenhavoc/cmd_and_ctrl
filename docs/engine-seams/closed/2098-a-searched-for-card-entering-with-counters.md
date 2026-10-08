---
title: "A searched-for card entering with counters"
date: 2026-10-08
issues: [2098]
---
**A searched-for card entering with counters** (#2098, CR 614.1c / 122.6, [ADR 0113](decisions/0113-small-seams-for-the-s58-deck-requests.md) third amendment of 2026-10-08) — `SearchLibrarySpec.EntersWithCounters` (mirrored on the `effects.SearchLibrary` primitive) is copied onto the entry event by `searchEnterBattlefieldLocked`, the way a token's counters are (#762). The counters go through the CR 614 counter pipeline and are on the permanent before `EventETB` fires, so a Doubling Season doubles them and an enters trigger reading counters sees them; a paused entry keeps them because a resume reads `ev.EntersWithCounters`. No wire, snapshot or bot change. **Card:** Neoform (Full).
