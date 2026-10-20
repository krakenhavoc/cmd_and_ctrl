---
title: "An activated ability's target chosen at random"
date: 2026-10-20
issues: [4051]
---
**An activated ability's target chosen at random** (#4051, CR 602.2b, CR 601.2c, [ADR 0019 amendment 2026-10-20](decisions/0019-structured-targeting.md)) — `TargetSpec.AtRandom` (`effects.ChosenAtRandom`, #1815) was honoured only by the trigger target walk. An activation announces its targets in one message, so the draw now happens in that announcement, which is where the Goblin Test Pilot ruling puts it ("The target is chosen at random as part of activating the ability").
- **The announcement draws it.** `activateCatalogAbilityLocked` calls `announceRandomTargetsLocked` (`game/random_target_announce.go`) after the payload's own refs are slotted: a ref aimed at a random clause is refused (`ErrInvalidParam`), and the clause gets one target drawn uniformly from its legal set by `drawStepTargetLocked`, the trigger walk's draw, on the "pick" stream keyed to the activator and the source (ADR 0054), so an undo replays it. With nothing to draw the activation is refused (CR 601.2c). The draw is carried on `ActivateAbilityParams.randomTargets`, as a random discard is (ADR 0109 §7), so a run re-made after a CR 903.9 question aims at the same target, and it is logged ("any target: Bob (at random)") once the ability is on the stack.
- **Nobody picks it.** `StepsChosenByController` leaves random steps out; the enumerator names no target for one and keeps only the prefixes it can draw for (`RandomStepFillableForEffect`); the view marks the clause `at_random` and the client's walk skips it, so an ability whose only clause is random is activated at once. `abilityClauseView` now carries `chosen_by_opponent` too, which a lone opponent's clause (Nova Pentacle) was missing.
- **Scope.** `effects.Register` allows the flag on an activated ability's lone clause only, the one printed shape; a second clause, a mode's clause or a spell's is refused.
- **Cards:** Goblin Test Pilot (Full).
