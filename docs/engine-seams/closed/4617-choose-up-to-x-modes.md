---
title: "\"Choose up to X\" modes where X is the spell's X"
date: 2026-10-30
issues: [4617]
---
**"Choose up to X" modes where X is the spell's X** (#4617, CR 107.3m, CR 700.2b, CR 700.2d, [ADR 0065 amendment 2026-10-30](decisions/0065-modal-and-multi-target-clauses.md)) — "When The Ruinous Wrecking Crew enters, choose up to X —": a mode count that is the X paid for the spell that became the permanent.
- **`ModeSpec.UpToCastX`**, built with `effects.ChooseUpToX(...)`. `modeBoundsLocked` answers `0..min(X, bullets)` from `ModeCountQuery.X`, which the trigger's `mode_pick` reads off the entering object's `CastProvenance.X` as it goes on the stack. The prompt carries the bound and names the X ("Choose up to X (X is 2)"). At X = 0, or for a permanent that was not cast, nothing is asked and the trigger is removed.
- **An empty answer removes the ability.** Answering a "choose up to" prompt with no bullet used to ask the same question again (Ertai Resurrected, Hullbreaker Horror, Ravager Wurm); it now removes the trigger, and the client's picker reads an absent `mode_min` as 0 and offers "Choose none".
- `effects.Register` keeps it to an enters trigger, and refuses it on a granted one (CR 107.3j).
- **Cards (1, Full):** The Ruinous Wrecking Crew.
