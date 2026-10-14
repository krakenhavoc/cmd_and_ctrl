---
title: "A controller-only delayed mana trigger"
date: 2026-10-14
issues: [3504]
---
**A controller-only delayed mana trigger** (#3504, [ADR 0074](decisions/0074-triggered-mana-abilities.md) amendment 2026-10-14) — `DelayedTrigger.ManaTapControllerOnly` narrows a spell-made triggered mana ability (CR 605.1b) to the taps of the player who set it up, and `fireManaTriggersLocked` skips a production whose pool owner is not that player. `effects.TurnManaTrigger{ControllerOnly: true}` builds one: "until end of turn, whenever YOU tap a Mountain for mana, add an additional {R}". Unset, the record fires for every player's tap, which is what High Tide and Bubbling Muck print ("a player"). The field is an additive snapshot key (`manaTapControllerOnly`, shape v8 recorded). **Cards:** Pyre Rhymer // Molten Tide (Full).
