---
title: "Target count set by the multikicker count"
date: 2026-11-25
issues: [6150]
pr: 6273
---
**Target count set by the multikicker count** (#6150, CR 601.2c, 702.33c-d) — "Choose any target, then choose another target for each time this spell was kicked" is `TargetSpec.CountFromKicks`, declared with `effects.CountedByKicks(...)`. `TargetSpecUnderOptionalCosts` binds it to 1 + the kicks announced with the optional costs (on a copy, Min = Max), wherever an announcement is judged: the announce gate, the CR 608.2b re-check and the move enumerator. The wire field is `count_from_kicks`; the client picker asks for 1 + `CastChoices.kicks`. `Register` refuses the flag without a kicker or on a multi-clause statement. **Cards:** Comet Storm, Strength of the Tajuru (Full).
