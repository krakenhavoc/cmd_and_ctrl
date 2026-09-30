---
title: "Delayed trigger on a named player's turn"
date: 2026-09-30
issues: [1538]
---
**Delayed trigger on a named player's turn** (#1538, CR 603.7 / 603.7d) — `game.DelayedTrigger.TurnOf` (and `effects.ScheduleDelayedTrigger.TurnOf`) names the one player whose turn the step must belong to, checked beside `ControllerTurnOnly` in `fireDelayedTriggersLocked`. It is the answer to "at the beginning of **that player's** next end step" without misstating CR 603.7d: the delayed ability is still controlled by whoever's effect created it, so it is their stack item and their APNAP slot, and only the moment it fires is the named player's. A trigger whose named player has left the game is dropped (CR 800.4a). The field is carried by the snapshot (`delayedTriggers[].turnOf`, additive within schema 7) and by clone. Proof card: The Eternal Wanderer's +1, which now returns the exiled card on its owner's end step (`exileTargetsThenReturnOnOwnersEndStep`, one delayed trigger per owner) and dropped its caveat. No seam row existed, because the gap earns one only when a second card waits on it.
