---
title: "Proliferate is the player's choice"
date: 2026-10-07
issues: [2525]
---
**Proliferate is the player's choice** (#2525, CR 701.34a, [ADR 0013 amendment 2026-10-07](decisions/0013-replacement-effects.md)): "choose any number of permanents and/or players with counters on them" was an auto-pick for the proliferating player. It is now `PendingChoiceProliferate`, a card-set pick (floor zero) whose list names permanents and, through `PendingChoice.ChoosePlayers`, seats by player ID. It is queued after the keyword action's CR 614 window settles, once per time the settled count asks for, so "proliferate twice" is two choices. The engine's beneficial pick is `PendingChoice.ChooseSuggested` (wire `choose_suggested`): the client pre-selects it, the enumerator always offers it whole, and the heuristic answers with it. What follows "proliferate" on a card rides `Proliferate{Then}`. Nine caveats came off (Staff of Compleation, Karn's Bastion, Contagion Clasp, Evolution Sage, Flux Channeler, Inexorable Tide, Bloated Contaminator, Cankerbloom, Mutational Advantage) and Steady Progress is stamped Full.
