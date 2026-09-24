---
title: "Per-turn tally"
date: 2026-09-17
issues: [586]
pr: 768
legacy_order: 109
---
**Per-turn tally** (#586) - `Game.TurnTally` counts what happened this turn in one place, reset on turn advance: per player (`TurnTallyFor`: life gained and lost, cards drawn, creatures died, tokens created, permanents sacrificed, lands entered, attackers declared, combat damage to players), per player and subtype (`EnteredWithSubtypeThisTurn`, #743: the subtypes a permanent had as it entered, before statics apply, so Lilypad Village is not fooled by a Maskwood Nexus that arrives later), and per ability (`ResolvedThisTurn(source, label)`, `TriggeredThisTurn(source, label)`), with `EventsThisTurn()` for a filtered scan. It closes the trigger half of the old "Once-per-turn trigger / activation tally" row ("this ability triggers only once each turn" reads `TriggeredThisTurn`). Unblocked and still to be written: Morbid Opportunist (#295); Welcoming Vampire (#296); Terrasymbiosis (#300); Tocasia's Welcome (#300). Monument to Endurance (#301) also waits on modal triggers (#764). The activation half is still open: "Per-source activations-this-turn count" above.
