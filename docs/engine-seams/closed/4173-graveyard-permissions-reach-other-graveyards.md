---
title: "Graveyard permissions that reach other players' graveyards"
date: 2026-10-19
issues: [4173, 4177]
---
**Graveyard permissions that reach other players' graveyards** (#4173, #4177, [ADR 0066](decisions/0066-granted-cast-and-play-permissions.md) amendment of 2026-10-19) — under Shaman's Trance a permission over "your graveyard" (Underworld Breach's escape, Yawgmoth's Will, Lurrus, Muldrotha, a flashback grant) now opens the cards in every graveyard. `CastPermission.GraveyardsAreYours` is set on the Trance's own stored grants; while one is live `permissionReachesPileLocked` (now a `Game` method) lets the holder's permissions over their own pile reach any pile, so the cast path, the enumerator and the view agree. A derived permission's offer (Breach's escape) is now claimable and listed beside a stored plain permission over the same card (`CastPermissionForClaimLocked`, `otherGrantedOffersLocked`). No new stored record: a resolved Trance adds only the `graveyardsAreYours` key to its permissions (fixture `v9/shamans_trance.json` stays a subset). **Cards:** Shaman's Trance, now Full.
