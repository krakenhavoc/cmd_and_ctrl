---
title: "The next time you would lose the game this turn, instead"
date: 2026-11-25
issues: [6244]
---
**The next time you would lose the game this turn, instead** (#6244, CR 614.1a, 614.5, 704.7, 119.5) — Stunning Reversal's "The next time you would lose the game this turn, instead draw seven cards and your life total becomes 1" is `Game.ReplaceNextLossThisTurnForEffect(player, source, draw, life, label)`: one `ScopedEffect` (`ScopeGame`, `UntilEndOfTurn`) with a `ModNextLossInstead` mod (`nextLossInstead`) naming the player, the cards drawn (`Amount`) and the life total set (`Power`). It is read at the losing door The Golden Throne's static uses (`replaceLossLocked`), ahead of any permanent's replacement, for a state-based loss (one record replaces every loss a single pass deals, CR 704.7) and for an effect loss; a concession is never replaced. The first loss uses it up before the draw, so a loss the seven cards or a later event cause that turn stands, and an unused record ends at the cleanup step. `EventGameLossReplaced` names the spell. An older binary refuses a restore point that carries one. **Cards**: Stunning Reversal (OLD b205) is built by its OLD batch.
