---
title: "Conspire"
date: 2026-11-05
issues: [4690]
---
**Conspire** (#4690, CR 702.78a / 601.2f / 707.10c / 707.10f) — conspire had no code. It is now casualty's shape with a tap cost: `effects.Conspire()` is the optional cost keyed `conspire` carrying #3868's `TapOthersCost` (two untapped creatures you control), flagged `SharesColorWithSource` so each must share a colour with the spell being cast, judged on the spell's and the creatures' current colours by the one candidate walk the view, the bot and the validator share. The creatures ride `teamwork_ids` and are tapped with the spell on the stack; the copy is the engine's own "when you cast this spell" trigger (`game/conspire.go`, keyed body `conspire/copy`), read off the paid record and copied from last-known information, so a copy of a permanent spell becomes a token (CR 707.10f) and the copy may take new targets. A permanent grants it through `SpellsYouCastHave(label, Noncreature(), Conspire())`.
- **The wire and the bot.** No new field: `optional_costs[i].tap_others_label` / `tap_others_options` (now asked with the spell as the cost's source) and the replicate tap picker; the enumerator names the two cheapest-to-tap matching creatures. No snapshot change.
- **Still open.** CR 702.78b's separate instances (one conspire per spell), and a granted conspire on a spell whose own cost already taps permanents on `teamwork_ids` (New Blood, teamwork, escalate's taps, a replicate cost that taps), which is not offered.

**Cards:** Raiding Schemes, with a caveat.
