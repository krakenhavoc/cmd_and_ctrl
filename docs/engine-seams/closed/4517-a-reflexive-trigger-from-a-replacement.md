---
title: "A reflexive trigger from a replacement"
date: 2026-10-29
issues: [4517]
---
**A reflexive trigger from a replacement** (#4517, [ADR 0178](decisions/0178-life-instructions-after-a-replaced-move.md) amendment of 2026-10-29, CR 603.12, CR 614.6, CR 616.1) — "If a creature an opponent controls would die, exile it instead. When you do, …". The replacement declares the trigger on the move it modifies (`Game.QueueReflexiveTriggerAfterReplacedMoveForEffect`), as it declares a life change or an extra turn, and the move's finisher creates it once the card has landed where the replacement sent it, through the same dispatch a resolving effect's reflexive trigger takes. It goes on the stack the next time a player would receive priority, under the replacement's controller.
- **The condition.** No trigger when the move is cancelled or a later replacement sends the card elsewhere, and none when another replacement (Rest in Peace) was applied first, because this one never applied. It does not trigger on the exile event, which would be stronger than printed. One trigger per replaced move, so a wipe gives one each.
- **Snapshot:** additive in v9. A paused move with a declared trigger writes `replacementResume.event.reflexiveAfterMove` under its own key, `replacement-resume/move-reflexive`; restore checks the body key. **Closure ratchet:** unchanged.
- **Cards (1, Full):** Head of the Hunt. Kalitas, Traitor of Ghet makes its token as part of the modified event, not as a trigger, and waits on #4522.
- **Tests:** `game/move_reflexive_test.go` (landing, one per move, a redirected or cancelled move, an unreplaced death, malformed declarations, a paused move restored and answered both ways), refusal cases in `replacement_resume_keyed_test.go`, and `cards/effects/head_of_the_hunt_test.go` (flash, one creature, a state-based death, a wipe of three, your own creature, Head of the Hunt gone first, Rest in Peace chosen either way, a restore with the trigger pending).
- **Back-outs:** without the finisher's call, eight tests fail; without the landed-zone check, the redirect and the restored "yes" fail; without the key override, the paused-move restore fails; without the restore mirror, the mirror round trip and the paused-move restore fail.
- **Still open:** a wipe that destroys the replacement's source loses it for the later legs (#4521).
