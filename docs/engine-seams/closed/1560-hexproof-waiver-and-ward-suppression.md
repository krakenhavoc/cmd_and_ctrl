---
title: "Hexproof waivers and ward that doesn't trigger"
date: 2026-09-27
issues: [1560]
---
**Hexproof waivers and ward that doesn't trigger** (#1560, CR 702.11 / CR 702.21, [ADR 0038 amendment 2026-09-27](decisions/0038-protection-style-keywords.md)). "Can be the targets of spells and abilities [you control] as though they didn't have hexproof" is now `Spec.HexproofBypasses`. It is read live at the targeting choke point, and only once hexproof alone would refuse the target, so shroud and protection are untouched. The announce gate and the CR 608.2b re-check share that reader. `YoursOnly` separates Nowhere to Run's "spells and abilities" from Kaya's "you control", and a `Player` half waives a player's hexproof. "Ward abilities of those creatures don't trigger" is `Spec.WardSuppressions`. It is read by `effects.WardGranted`, the trigger condition every ward in the catalog uses, and asked about the warded creature when the ability would trigger. **Nowhere to Run** and **Kaya, Bane of the Dead** now ship `full`. Detection Tower's turn-scoped waiver is still missing. Glaring Spotlight, which also needed a live-set "can't be blocked this turn", shipped with #1650.
