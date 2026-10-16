---
title: "Mana abilities any player may activate"
date: 2026-10-16
issues: [1864]
---
**Mana abilities any player may activate** (#1864, [ADR 0106 §1 amendment 2026-10-16](decisions/0106-five-small-seams-from-the-s50-rechecks.md)) — `ManaAbility.AnyPlayer` is the CR 605 list's "Any player may activate this ability", read through one gate, `game.MayActivateMana`, by the mana activation path, the legal-move enumerator and the view. The activator pays the cost (CR 602.1a), the mana goes to the activator's pool (CR 106.4), and "their turn" in the row's condition is the activator's (CR 109.5). Every other seat gets its own copy of the row, stamped with that seat as the activator; the digest lists it under `mana_abilities`, the client opens another player's permanent for it, and the log says "Bob activated Alice's Mana Cache". The auto-tapper does not spend it: the activator floats the mana from the board. **Cards:** Mana Cache (Full).
