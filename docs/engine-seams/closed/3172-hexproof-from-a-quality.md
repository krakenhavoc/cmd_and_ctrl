---
title: "Hexproof from a quality"
date: 2026-10-13
issues: [3172]
---
**Hexproof from a quality** (#3172 slice cav-e, [ADR 0038 amendment 2026-10-13](decisions/0038-protection-style-keywords.md), CR 702.11e) — "hexproof from black", "from monocolored", "from each color", "from planeswalkers" are the token `hexproof from <quality>`, parsed only by `game.ParseHexproofFromQuality` (`game/hexproof_from.go`), which reuses protection's quality grammar and adds the two colour-count qualities. `canBeTargetedBy` and the player half refuse a source an opponent controls that has the quality; the hexproof waiver waives it and "can't have hexproof" strips it. Granted by `GrantToAttached`, `GrantKeywordUntilEOT` and `GainPlayerKeyword`, and stamped by the deck importer from the oracle line. Cards: Eradicator Valkyrie, Breaker of Creation, Dragonfire Blade, Veil of Summer, Soulflayer.
