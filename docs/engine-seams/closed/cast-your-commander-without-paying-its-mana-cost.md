---
title: "Cast your commander from the command zone without paying its mana cost"
date: 2026-11-25
issues: [6242]
---
**Cast your commander from the command zone without paying its mana cost** (#6242, CR 903.8, 118.9, 608.2g) — Geode Golem's "you may cast your commander from the command zone without paying its mana cost". `game.FreeCastFromCommandPermission` is a stored, pass-closed ScopeCards grant over the caster's own commanders in the command zone (`game/free_cast_command.go`), priced like a hand grant (`AltCostKeyGrantedCost`) and offered during the resolution through the `resolution_cast` prompt, which now accepts command-zone candidates. The command zone is otherwise still closed to permissions (`castPermissionLocked` reads only these stored grants there). The commander tax is layered over the free cost as before, so a second free cast costs {2}, and the cast bumps `Player.CommanderCasts`. Card helper: `effects.OfferFreeCommanderCast(ctx, name)`. **Cards:** Geode Golem (b195) is built by its OLD batch.
