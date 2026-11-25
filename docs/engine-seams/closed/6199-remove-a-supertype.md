---
title: "Remove a supertype"
date: 2026-11-25
issues: [6199]
---
**Remove a supertype** (#6199, CR 613.1d, 613.7, 205.4g) — Melting's "All lands are no longer snow" and Arcum's Weathervane's "Target snow land is no longer snow". A new layer-4 mod kind, `ModRemoveSupertypes` (`removeSupertypes`, an on-disk identity), built by `game.RemoveSupertypesMod`, removes the supertypes and keeps every other type, subtype and supertype. It applies in timestamp order with `ModAddSupertypes`, so a later "becomes snow" puts snow back and an earlier one does not. A land that is no longer snow is no snow source, so it can't pay `{S}` (CR 107.4h). The printed static is `effects.AllLandsAreNoLongerSupertype("Snow")`, through the exported `Characteristic.RemoveSupertypes`. Restore and registration refuse a mod that names no supertype or something that isn't one. **Cards:** Melting (b848) and Arcum's Weathervane (b833) are built by their OLD batches.
