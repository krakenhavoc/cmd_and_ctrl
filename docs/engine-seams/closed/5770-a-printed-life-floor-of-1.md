---
title: "A printed life floor of 1"
date: 2026-11-13
issues: [5770]
---
**A printed life floor of 1** (#5770, CR 120.3a, 702.15b, 903.10a) — the permanent half of Angel's Grace's floor (#3172): "Damage that would reduce your life total to less than 1 reduces it to 1 instead" printed on a permanent. Declared as `effects.Spec.PlayerLifeFloorOne` (`game.CardDef.PlayerLifeFloorOne`) and read off the battlefield through `game.CatalogPlayerLifeFloorOne` and `CatalogAbilityKey` by the one reader, `playerLifeFloorLocked` (`game/life_floor.go`), the way Platinum Emperion's printed lock is read: the floor belongs to the permanent's controller, follows a change of control, ends when the permanent leaves or loses its abilities, and two copies compose. The one consumer is still the damage tail's cap, so the damage stays whole (lifelink, the commander tally and "dealt damage" triggers see all of it) and "lose life" effects and life payments are not floored. No new stored state, snapshot field, effect key, closure route or wire field.

**Cards:** Sustaining Spirit (Full).
