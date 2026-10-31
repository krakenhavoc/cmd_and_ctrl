---
title: "Flash while an opponent has cast a spell this turn"
date: 2026-10-29
issues: [4570]
---
**Flash while an opponent has cast a spell this turn** (#4570, CR 702.8a, CR 601.2i, [ADR 0066](decisions/0066-granted-cast-and-play-permissions.md)) — "As long as an opponent has cast a spell this turn, you may cast spells as though they had flash" (Captain Mar-Vell, Space-Born).
- **`CastTimingRule.OpponentCastASpellThisTurn`** is an additive, declaration-only bool beside Radagast's `FirstCreatureSpellThisTurn`. `castTimingVerdictLocked` skips the statement unless some seat other than the source's controller, still in the game, has a non-zero `SpellsCastThisTurn` tally. The condition is read live on every query, so a Mar-Vell that enters after the opponent's spell is covered at once, the window shuts as the turn ends or when Mar-Vell leaves, and nothing waits on the stack (a flash-granting trigger would do all three wrong). A stored grant clears the field on the way in. Card vocabulary: `effects.CastAsThoughFlashWhileAnOpponentHasCast`.
- **Snapshot:** one recorded shape line (`seats[].statics[].timing.opponentCastASpellThisTurn`, omitempty). The declared half is never stored, and the tally it reads already rides the snapshot.
- **Cards (1, Full):** Captain Mar-Vell, Space-Born.
