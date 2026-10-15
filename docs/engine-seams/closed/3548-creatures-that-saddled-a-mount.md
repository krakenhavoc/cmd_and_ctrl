---
title: "Creatures that saddled a Mount"
date: 2026-10-15
issues: [2704, 3548]
---

**Creatures that saddled a Mount** (#2704, CR 702.171c, [ADR 0186](decisions/0186-crewing-and-saddling-who-paid-and-what-they-count-for.md)) — `TurnTally.Crews` records every creature tapped to pay a crew or saddle cost, by the permanent's and the creature's `ObjectRef`, at payment. `Game.CreaturesThatSaddledForEffect` and `effects.SaddlersOfThis` read it, so the answer survives the Mount leaving. The last slice (#3548) ships Calamity, Galloping Inferno: two rounds of a `ChooseCardsPrompt` (exactly one) over the nonlegendary saddlers, each making a tapped and attacking token copy through `CreateTokensAttackingYourChoice` with a delayed sacrifice at the next end step. **Cards:** Rambling Possum, The Gitrog, Ravenous Ride, Fortune, Loyal Steed, Giant Beaver and Calamity, Galloping Inferno (all Full).
