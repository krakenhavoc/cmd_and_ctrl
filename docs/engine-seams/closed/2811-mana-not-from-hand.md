---
title: "Mana that can't be spent to cast spells from your hand"
date: 2026-10-09
issues: [2811]
---
**Mana that can't be spent to cast spells from your hand** (#2811, CR 601.2a, [ADR 0040](decisions/0040-mana-pipeline.md)) — `game.ManaSpendContext` carries `CastFrom`, the zone a spell is being cast from, set by `ManaSpendForCastFrom` and `ManaSpendForCastParams` (the zone read off `CastSpellParams.FromZone`, an empty one being the hand). Every cast payment threads it: `applyCastCostLocked`, the auto-tapper, the mana-spend riders, the legal-move enumerator and the cast preview, so the view, the bot and the engine agree. The new negative tag `game.ManaRestrictNotFromHand` admits any activation or unlock and a cast from a known zone other than the hand, and refuses a cast from hand or from a zone the caller did not name (weaker than printed, never stronger). **Cards:** Heartwood Crafter loses its caveat (Full); Karolina Dean, Runaway — Full. **Not built:** Vhal, Candlekeep Researcher (a mana ability whose amount is its toughness has no declared shape) and Mm'menon, the Right Hand (also needs an artifact cast-permission filter).
