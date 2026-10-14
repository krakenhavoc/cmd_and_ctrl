---
title: "Life instructions after a replaced move"
date: 2026-10-14
issues: [2950]
---
**Life instructions after a replaced move** (#2950, CR 608.2c / 614.6 / 616.1g, [ADR 0178](decisions/0178-life-instructions-after-a-replaced-move.md)) — "instead exile it and you gain 2 life" is one modified event in printed order. A move replacement declares the gain with `Game.ChangeLifeAfterReplacedMoveForEffect` rather than changing life inside `Replace`. The move's finisher lands the card, then runs each declared change through the life pipeline, and only then the move's own continuation (a destroying spell's "then draw a card"), which waits if a change asks a CR 616 question. The rest of the move rides the life event's tail as an immutable value, so undo into the open prompt replays it. The change survives a later redirect or cancel, and a cost move settles it without asking. **Cards:** The Darkness Crystal (Full; its gain used to land before the exile, and the destroy's continuation ran while the life question was open).
