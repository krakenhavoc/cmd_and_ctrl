---
title: "Snow and legendary landwalk"
date: 2026-11-25
issues: [6143]
---
**Snow and legendary landwalk** (#6143, CR 702.14c) — "snow landwalk", "legendary landwalk" and "snow <basic land type>walk" (Zombie Musher, Livonya Silone, Ayumi, Rime Dryad, Legions of Lim-Dûl) are canonical keyword tokens in `canonicalKeywords` and rows of `landwalkTokens` (game/landwalk.go), each a land supertype the defending player's land must have, plus a land subtype for the snow <type>walk forms, read through the effective view as nonbasic landwalk is. "Snow landwalk of the chosen type" (Barbarian Guides) can name any CR 205.3i land type, so it is a parameterised, grant-only token, `game.SnowLandwalkToken(landType)` ("snow gatewalk"), parsed by the same block check. `game.LandwalkTokens()` lists every landwalk token the engine reads, for "loses all landwalk abilities" (Hammerheim). Text-changing effects respell the land type in "snow swampwalk", and the bot's landwalk estimate reads the supertypes. **Cards**: built by their OLD batches.
