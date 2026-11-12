---
title: "Granting jump-start to two-colour graveyard cards"
date: 2026-11-12
issues: [5645]
---
**Granting jump-start to two-colour graveyard cards** (#5645, CR 702.133a, CR 105.2, [ADR 0066 amendment 2026-11-12](decisions/0066-granted-cast-and-play-permissions.md)) — "Each instant and sorcery card in your graveyard that's exactly two colors has jump-start" (Niv-Mizzet, Supreme).
- **`CastPermission.DiscardCardInAddition`** is jump-start's additional cost on a granted offer: the card's printed mana cost stays owed and one card from the holder's hand is discarded as a cost, named in `alt_cost_ids`. It is `DiscardLandCard` (retrace) without the land clause; `DiscardCard` is the other reading, "rather than paying its mana cost". The grant's key is the new constant `game.AltCostKeyJumpStart` (`"jump-start"`, the spelling the printed keyword already claims), and `ExileOnResolution` is the exile as the spell leaves the stack.
- **`PermissionFilter.ColorCountExactly`** is an additive colour-count clause: a card qualifies when it has exactly N colours, counted off its effective colours in `PermissionFilter.Matches`, so a hybrid card is every colour of its symbols (CR 202.2d) and a split card in a graveyard is the colours of both halves (CR 709.4b). Zero constrains nothing.
- **Snapshot:** `discardCardInAddition` and `filter.colorCountExactly` (both omitempty) under every path that reaches a `CastPermission`, recorded in `v9.txt`. Niv-Mizzet's permission is derived off the battlefield and never stored.
- **Cards (1, Full):** Niv-Mizzet, Supreme.
