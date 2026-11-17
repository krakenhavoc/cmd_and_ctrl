---
title: "Sacrificing a creature in addition on a cast permission"
date: 2026-11-12
issues: [5670]
---
**Sacrificing a creature in addition on a cast permission** (#5670, CR 118.8, CR 118.3, CR 601.2h) — "Once during each of your turns, you may cast an instant or sorcery spell from your graveyard by sacrificing a creature in addition to paying its other costs" (Maestros Ascendancy). A granted permission could ask for a land (`SacrificeLandInAddition`, Exploration Broodship) or a nonland permanent (`SacrificeNonlandPermanentInAddition`, Into the Pit) but not a creature.
- **`CastPermission.SacrificeCreatureInAddition`** is the third sibling: the synthesised offer keeps the card's printed mana cost and charges one creature the caster controls, sacrificed as a cost (`AlternativeCost.Sacrifice`) and named in `alt_cost_ids`. The creature is judged by its types on the battlefield, so a land creature or an animated artifact pays and a creature card that is not a creature there does not. A caster with no creature cannot claim the offer. Read only with an `AltCostKey`, as the siblings are.
- **The rest of the card is existing data:** one use per turn per granting object is `PerType: []string{game.PermissionTypePermanent}` (Johann, Apprentice Sorcerer), "your turns" is `Timing: game.TimingYourTurnOnly`, and "if a spell cast this way would be put into your graveyard, exile it instead" is `ExileInsteadOfGraveyard` (Torrential Gearhulk).
- **Snapshot:** `sacrificeCreatureInAddition` (omitempty) under every path that reaches a `CastPermission`, recorded in `v9.txt`. The Ascendancy's permission is derived off the battlefield and never stored; the exile clause and the spent use ride the stack item and the turn tally, which a restore already carries.
- **Cards (1, Full):** Maestros Ascendancy.
