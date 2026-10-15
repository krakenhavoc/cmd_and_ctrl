---
title: "A free play or cast from the top of the library during a resolution, else into your hand"
date: 2026-10-15
issues: [3801]
---
**A free play or cast from the top of the library during a resolution, else into your hand** (#3801, CR 608.2g, CR 305.2, CR 701.20; [ADR 0099](decisions/0099-discover.md) amendment 2026-10-15) — Ziatora's Envoy: "look at the top card of your library. You may play a land from the top of your library or cast a spell with mana value less than or equal to the damage dealt from the top of your library without paying its mana cost. If you don't, put that card into your hand."
- **The window.** `Game.OfferFreePlayFromTopOrHandForEffect` looks at (or reveals) the top card. A spell within the cap gets ADR 0099 §4's pass-closed free cast over the card in the library (`Zone: ZoneLibrary`, `LapseOnPass: LapseToHand`); a land the player can play now is played through `PlayLandDuringResolutionForEffect` under the new `play_or_hand` keyword. Anything else, and every decline, puts the card into the hand.
- **Hidden information.** Only the looker learns the card. `MayCastPrompt.Private` rides the existing `private_text` path and the view now clears `may_cast_card` for a private `may_cast`. A reveal is public.
- **Client and bot.** `play_or_hand` is worded as a play and opens no cast; the free-cast request also searches the viewer's library and casts with `from_zone: "library"`. The enumerator labels the accept "play it"; the heuristic's yes stands.
- **Cards:** Ziatora's Envoy (Full), Rashmi, Eternities Crafter (Full).
