---
title: "Search a player's graveyard, hand, and library for all cards with a name"
date: 2026-10-16
issues: [4033]
---
**Search a player's graveyard, hand, and library for all cards with a name** (#4033, CR 701.23b, 400.2, [ADR 0015](decisions/0015-zone-browser-library-search.md) amendment 2026-10-16 (b)) — `SearchLibrarySpec.TakeAll` (`SearchLibrary.TakeAll` in the catalog), on an `AlsoGraveyardAndHand` search, is "all cards" rather than "any number": every match in the graveyard, a public zone, is found and exiled without being offered, and the searcher may still leave matches in the hand and the library, which are hidden. With no hidden match there is no prompt. **Cards:** Surgical Extraction, Extirpate, Eradicate, Scour, Splinter, Sowing Salt and Quash (new, Full).
