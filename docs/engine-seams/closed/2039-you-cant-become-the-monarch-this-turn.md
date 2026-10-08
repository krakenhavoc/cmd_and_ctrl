---
title: "\"You can't become the monarch this turn\""
date: 2026-10-08
issues: [2039]
---
**"You can't become the monarch this turn"** (#2039, [ADR 0096](decisions/0096-the-monarch-from-a-card-effect.md) amendment 2026-10-08) — `becomeMonarchLocked`, the monarch's one write, had no gate, so nothing could stop a player becoming the monarch. `ModCantBecomeMonarch` is now a stored `ScopedEffect` rules gate (reader `rule`, scope game, reads `Player`, swept at cleanup, CR 514.2), written by `effects.CantBecomeTheMonarchThisTurn` and read at that one write, so the CR 725.2 combat-damage steal, a card's "you become the monarch" and CR 725.4's hand-on all obey it at once. The current monarch simply stays the monarch when the player named is barred (Jared's rulings), and the hand-on skips a barred player. The sandbox's manual `SetMonarch` bypasses the gate. **Cards** (1): Jared Carthalion, True Heir (Full), whose prevention line is ADR 0108 §8 with a "while you're the monarch" condition.
