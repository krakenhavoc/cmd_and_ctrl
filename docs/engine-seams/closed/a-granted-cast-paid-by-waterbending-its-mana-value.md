---
title: "A granted cast paid by waterbending its mana value"
date: 2026-10-16
issues: [3989]
---
**A granted cast paid by waterbending its mana value** (#3989, [ADR 0066](decisions/0066-granted-cast-and-play-permissions.md) amendment 2026-10-16) — `CastPermission.WaterbendManaValue` prices a granted cast at "waterbend {N}", N the mana value of the face being cast with {X} at 0 (CR 107.3b). The {N} is the alternative cost (`CastCostFor`), so taxes and reductions apply to it (CR 118.9d), and `TapPermanentsCostForGrantLocked` hands the cast a waterbend tap cost marked `InCost`, so tapped artifacts and creatures pay only that {N} (CR 701.67b). The validator, the pricer, the view's tap picker and the enumerator read the same two answers. **Cards:** Hama, the Bloodbender (with a caveat: a spell cast this way cannot also use convoke or improvise).
