---
title: "Search results held in place and split by another player"
date: 2026-11-25
issues: [6156]
pr: 6272
---
**Search results held in place and split by another player** (#6156, CR 701.20a, CR 701.23a) — `SearchLibrarySpec.HoldFound` is "search ... and reveal them" whose found cards stay where they were found, in the library or (with `AlsoGraveyard`) the graveyard, and are handed to `ThenTail` in the order taken, so the rest of the effect can have an opponent (a `RevealPick`) or the caster split them before anything moves. `SearchLibrarySpec.BothZones` makes an `AlsoGraveyard` search "your library **and** graveyard": both zones are named, so no zone question is asked and the library is always searched (the search event fires and an Opposition Agent sees it whatever is taken). A land put "under an opponent's control" is the existing `ZoneEntryOptions.Controller`. **Cards:** Ecological Appreciation, Verdant Mastery (Full).
