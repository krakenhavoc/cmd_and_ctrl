---
title: "Granted miracle"
date: 2026-10-30
issues: [4455]
---
**Granted miracle** (#4455): "Nonland cards in your hand have miracle {0}" (Molecule Man) is miracle (CR 702.94a) given to the cards in a hand by a permanent, in the printed shape of `game/miracle.go`. The static is a `GrantedAlternativeCost` under the key `granted-miracle` (`effects.MiracleZeroForNonlandCardsInYourHand`) with `RequiresGrant`: the offer is the price, claimable only under the hand `CastPermission` that a revealed trigger's resolution grants, so the offer, the view's picker and the bot's move list all vanish when the permanent leaves or loses its abilities. The reveal is `game.GrantedMiracleTrigger`, a row of the granting permanent grown from the offer in `buildDef` (the way the printed trigger is grown from a `Miracle` cost): it fires for the first card its controller drew this turn when that card is a nonland card still in their hand. A card that prints a miracle of its own has both abilities (CR 113.2c), each with its own reveal, key and price; `CastPermissionForClaimLocked` picks the hand permission that names the claimed key.
