---
title: "Roll n dice and ignore the lowest"
date: 2026-10-19
issues: [4174]
---
**Roll n dice and ignore the lowest** (#4174, CR 706.6, [ADR 0054](decisions/0054-dice-rolls-and-coin-flips.md) amendment 2026-10-19) — `Game.RollDiceIgnoringLowestForEffect` draws all n dice from the keyed roll stream (so an undo or a restore replays it), drops the lowest k, and emits one `EventRollDie` per kept die, so an ignored roll triggers nothing and no effect applies to it. Tied lowest results are the same number and ask nothing. `effects.rollDiceIgnoringLowest` is the catalog helper. **Cards:** Berserker's Frenzy (Full; the cast window is also #4180's, "before combat or during combat before blockers are declared").
