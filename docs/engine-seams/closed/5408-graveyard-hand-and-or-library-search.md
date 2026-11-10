---
title: "Search your graveyard, hand, and/or library"
date: 2026-11-09
issues: [5408]
---
**Search your graveyard, hand, and/or library** (#5408, CR 701.23a, 701.24a) — `SearchLibrarySpec.AlsoGraveyard` with `AlsoHand` (the same two flags on `effects.SearchLibrary`) is the three-zone and/or search. The zone question, asked before any card is shown, offers all three zones: `search_other_zone` gains a third wire value, `"graveyard_and_hand"`, and `ChooseSearchZones` takes any non-empty subset of `"library"`, `"graveyard"` and `"hand"`, each named on its own. Only the chosen zones' matches are offered, and the library is shuffled and the search announced if and only if it was chosen, whatever is taken, nothing included. The enumerator offers all seven subsets with all three the always-legal answer; the heuristic searches every zone. The client's sheet makes all three the primary, with library only, graveyard and hand, graveyard only and hand only as secondaries (`L.searchEveryZone`, `L.searchGraveyardAndHand`). No snapshot change: both flags were already on the keyed search frame. **Cards now Full:** Say Its Name.
