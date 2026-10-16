---
title: "Players can't draw more than one card each turn"
date: 2026-10-16
issues: [3927]
---
**Players can't draw more than one card each turn** (#3927, CR 121.2b / 121.3 / 614.17) — a permanent can limit how many cards players draw each turn, or stop them drawing at all. It is a rules gate, not a replacement: `Game.drawForbiddenLocked` (`game/draw_limit.go`) is asked for each individual draw before that draw's CR 614 window opens, so a draw the limit forbids is not performed. No replacement sees it (dredge, Underrealm Lich), no card moves, no `EventDrawCard` is emitted, nobody loses for an empty library, and the rest of the instruction goes on. The count is the cards actually drawn this turn (`Game.DrawnThisTurn`), so draws before the permanent arrived count and a replaced draw does not. The statics are `Spec.DrawLimits`, read from the battlefield through `CatalogAbilityKey`; the sandbox draw returns `ErrDrawRestricted`. **Cards:** Narset, Parter of Veils, Spirit of the Labyrinth, Leovold, Emissary of Trest, Maralen of the Mornsong and Mornsong Aria.
