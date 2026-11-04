---
title: "Granted sneak for creature cards in your graveyard"
date: 2026-11-04
issues: [4680]
---
**Granted sneak for creature cards in your graveyard** (#4680, [ADR 0025 amendment 2026-11-04](decisions/0025-alternative-costs.md), CR 702.190, CR 113.6e; follows [sneak](engine-seams/closed/4586-sneak.md)) — "Creature cards in your graveyard have sneak {3}{B}. You may cast creature spells from your graveyard using their sneak abilities." is `effects.CreatureCardsInYourGraveyardHaveSneak(cost, label)`, Goblin Formula's shape: one standing `CastPermission` over the graveyard under `granted-sneak`, whose `SneakCost` `AlternativeCostFor` turns into a sneak offer (the mana, `game.SneakReturnSpec()` and the `Sneak` bit for the window and the tapped-and-attacking entry). Rule 4 of `validateCastPathLocked` lets that permission open the zone only to a sneak claim of a creature spell, the granted offer or the card's own, which `CastOffersForLocked` lists beside it. `CastPermission.SneakCost` is additive in snapshot v9. **Cards:** Ninja Teen — Full.
