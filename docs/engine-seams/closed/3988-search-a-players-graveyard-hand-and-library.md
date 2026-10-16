---
title: "Search a player's graveyard, hand, and library by name, and play one of the exiled cards"
date: 2026-10-16
issues: [3988]
---
**Search a player's graveyard, hand, and library by name, and play one of the exiled cards** (#3988, CR 701.23, 611.2b, 611.2c, 305.2, [ADR 0015](decisions/0015-zone-browser-library-search.md) amendment of 2026-10-16) — `SearchLibrarySpec.AlsoGraveyardAndHand` (`SearchLibrary.AlsoGraveyardAndHand` in the catalog), with `LibraryOwner`, searches the named player's graveyard, hand and library in one prompt. The searcher looks at the whole hand. The hand candidates are marked in `PendingChoice.SearchHandCards` (wire `search_hand`, additive snapshot field `searchHandCards`), and `search_all_zones` says every zone was searched. Because the clause names all three zones, the library is always searched and shuffled, whatever was taken. A permission with a count (`CastsLeft`) is spent by a land played through it as well as by a cast (the land-play half came with #3955), so "play one of the exiled cards" is one use, land or spell. **Cards:** Kotose, the Silent Spider (new, Full).
