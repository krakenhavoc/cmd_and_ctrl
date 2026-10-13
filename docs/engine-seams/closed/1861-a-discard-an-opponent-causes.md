---
title: "A discard an opponent causes"
date: 2026-10-13
issues: [1861, 3168]
---
**A discard an opponent causes** (#1861, CR 113.8, CR 614.12, [ADR 0153](decisions/0153-a-discard-an-opponent-causes.md)), in `game/discard_entry.go` and `game/move_cause.go`.
- **The cause.** A discard's replacement event carries `Source` and `DiscardCauseController`. The controller is the controller of the spell or ability on the stack, not of the source card, and it is nil for a cost or a cleanup discard. `DiscardCausedByOpponentOf` is the predicate the triggered cards use.
- **Onto the battlefield.** A discard whose replacement settles on the battlefield goes through `discardOntoBattlefieldLocked`. It enters from the hand under its owner's control, gets its own replacement window (which can pause, and undo works across the pause), and receives any planned counters (Dodecapod). The log shows `EventETB` and then `EventDiscardCard` with the zone the card really reached, so the card still counts as discarded. The rest of the discard batch waits on a paused entry. (#3167)
- **Cards** (#3257): Obstinate Baloth, Loxodon Smiter, Nullhide Ferox, Wilt-Leaf Liege, Dodecapod and Nephalia Academy, plus ten triggered cards. Nine of those carry the owner-decision-6 caveat, because they trigger from the graveyard only. Painful Quandary's "unless" stays an effect discard, so a Baloth enters.
- **Bot.** `CardView.discard_by_opponent` lets the heuristic pitch these cards first when an opponent forces a discard.
