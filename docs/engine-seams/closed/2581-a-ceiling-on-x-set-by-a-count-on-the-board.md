---
title: "A ceiling on X set by a count on the board"
date: 2026-10-09
issues: [2581]
---
**A ceiling on X set by a count on the board** (#2581, CR 107.3a / 601.2b) — X could be bounded by a cost that charges it (pay X life, blight X) but not by a count the spell itself names, so "X can't be greater than the number of players in the game" would have cast for any X. `effects.Spec.XCeiling` declares the count, and `game.SpellXCeilingLocked` reads it once, as X is announced: the cast is refused above it, and a count that changes in response leaves the announced X alone (Open the Way's ruling). The legal-move enumerator folds it into the ceiling a cost already sets, so the bot and the MCP seat's open X range stop at it; the view stamps it per seat as `x_max`, the auto-tap preview reports it, and the client's X picker caps at it. **Cards:** Open the Way (`full`). **Still open:** Winter's Chill, which also needs a choice among paying {1}, {2} or nothing as it resolves (`pay-one-of-several-amounts`), and Shanna, Purifying Blade and Mortarion, Daemon Primarch, whose ceiling sits on a triggered "you may pay {X}".
