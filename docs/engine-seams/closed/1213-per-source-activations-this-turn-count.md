---
title: "Per-source activations-this-turn count"
date: 2026-09-22
issues: [1213, 1181, 1183]
pr: 1220
legacy_order: 39
---
**Per-source activations-this-turn count** (#1213, closing what #1181 and #1183 built) — the row said `TurnTally` counts what a source's abilities RESOLVED or TRIGGERED this turn and nothing counts what was ACTIVATED. #1181's `Game.Activations` built the count for exhaust and gave it a per-turn scope with no reader; #1183 made mana abilities write it too. What was left was one `Condition` per card plus the return-to-hand cost both named cards also waited on, and both landed here. `effects.OncePerTurnActivation` now reads **`Game.ActivatedThisTurn(source, label)`** instead of `ResolvedThisTurn`, which is a correctness fix as well as an unblocking: CR 602.1b counts announcements, and the two tallies disagree in both directions on a printed line — two activations held on the stack at once both resolve later, so a resolution count let the second be announced (stronger than printed, the #259 direction), and an activation countered or fizzled never resolves, so a resolution count handed the player a second use of an ability they had already spent. Beledros Witherbloom, the one existing caller, is fixed by the same line. **2 cards:** Quirion Ranger and Wirewood Symbiote, both `full`. **Still ordinary card work rather than a seam:** boast (CR 702.142a, Broadside Bombardiers) is this condition plus "only if this creature attacked this turn", which `TurnTally` already answers.
