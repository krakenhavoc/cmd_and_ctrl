---
title: "Graveyard watchers distinguish separate death instructions"
date: 2026-10-13
issues: [3417]
---
**Graveyard watchers distinguish separate death instructions** (#3417, [ADR 0018 amendment](decisions/0018-triggers-on-the-stack.md)) — `Game.InSameSimultaneousExitForEffect` reads the engine's existing simultaneous-exit copies and requires both cards to belong to that actual exit. A whole-resolution `Event.Batch` can include several independent instructions and is not used for this question. **Card:** Nether Traitor (Full); its graveyard scope, owner/type checks, resolution-time payment and object-epoch guard are preserved. No event, protocol or snapshot field is added.
