---
title: "Winning a coin flip, and Vehicle tokens with crew"
date: 2026-10-15
issues: [3767]
---
**Winning a coin flip, and Vehicle tokens with crew** (#3767, [ADR 0054 amendment 2026-10-15](decisions/0054-dice-rolls-and-coin-flips.md), CR 705.2, CR 301.7) — "Whenever you win a coin flip" is `effects.WheneverYouWinACoinFlip`, matching an `EventFlipCoin` whose `Won` is set and whose flipper (`Actor`) is the source's controller: one trigger per won coin, never for an opponent's flip, a face-only flip or a table flip. A bare "flip a coin" whose winning only matters to such a trigger is `effects.FlipACoin`, the called flip with no continuation. A Vehicle token is a token template whose crew ability is the printed Vehicle's `CrewCost` / `CrewEffect` registered under the token's key (`BlackjackToken`, `VehicleCrew1Token`). **Cards:** Setzer, Wandering Gambler (deck request #3655), Chance Encounter, Tavern Scoundrel, Mu Yanling, Wind Rider (Full). Still open: "whenever a player wins a coin flip" with "flip a coin until you lose a flip" (Okaun, Zndrsplt) and "whenever you lose a coin flip" (Karplusan Minotaur).
