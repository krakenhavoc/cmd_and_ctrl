---
title: "A replacement for would connive"
date: 2026-10-30
issues: [4589]
---
**A replacement for would connive** (#4589, [ADR 0013 amendment 2026-10-30 (#4589)](decisions/0013-replacement-effects.md), CR 701.50a, 614.1a, 614.5) — "If a creature you control would connive, instead you draw a card, then that creature connives" (Leader, Super-Genius). Connive is now an event: `Game.ConniveForEffect` opens a CR 614 window on `KeywordActionConnive` (the count is the N of connive N; the permanent is the event's `Source`) and runs the caller's keyed tail with the settled count and the extra draws (`ReplacementEvent.ConniveExtraDraws`). `ConniveEach`, `ConniveX`, Ledger Shredder and Cyclonus, the Saboteur all go through it. `effects.DrawACardThenConniveInstead` is the card-side replacement; the replaced connive keeps its N ("that creature connives" refers back to the replaced connive, CR 701.50d). A window waiting on an ordering prompt is a restore point. **Snapshot:** additive in v9 (`conniveExtraDraws`). **Cards:** Leader, Super-Genius (Full).
