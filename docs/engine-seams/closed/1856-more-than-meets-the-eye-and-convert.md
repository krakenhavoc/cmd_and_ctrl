---
title: "More Than Meets the Eye and convert"
date: 2026-10-13
issues: [1856]
---

**More Than Meets the Eye and convert** (#1856, CR 702.162 / 701.28 / 702.161, [ADR 0152](decisions/0152-more-than-meets-the-eye-convert-and-living-metal.md) PR 1) — `effects.MoreThanMeetsTheEye(cost)` is disturb's back-face offer with no zone binding, so a Transformers card can be cast converted from the hand, the command zone (with the commander tax) or anywhere a permission opens; it is on the stack back face up with its front face's mana value and enters back face up. `Game.ConvertPermanentForEffect` is the transform verb with the same guards and one body (`turnOverInPlaceLocked`), and its `EventTransform` carries `Converted`, so the log says "converted" and every "transforms into" reader sees it. CR 701.27f / 701.28e is now an engine rule: `Card.TurnOvers` counts in-place turn-overs, ability items and delayed triggers record it (`SourceTurnOvers`), and `effects.ConvertThis` and `TransformThis` ignore an instruction once it has moved. `effects.LivingMetal()` makes the back face an artifact creature on its controller's turn. **Cards:** Goldbug, Humanity's Ally // Goldbug, Scrappy Scout, Full. The bot's pricing of a converted cast is ADR 0152 PR 2.
