---
title: "A modal enters trigger's mode targets sized by the cast X"
date: 2026-11-18
issues: [5898]
---
**A modal enters trigger's mode targets sized by the cast X** (#5898, CR 107.3m, 603.3d, 608.2b) — "When Orcus enters, choose one — … • Return up to X target creature cards with total mana value X or less from your graveyard to the battlefield." The enters trigger sets `XFromCast`, so its item carries the cast X as `XValue`; the mode's clause is `CountFromX` with `UpToX` and `.WithTotal(effects.TotalManaValueAtMostX())` (`TargetTotal.MaxIsX`). `game.bindTriggerModeStepsCastX` rewrites the chosen modes' steps as the trigger goes on the stack: a `CountFromX` step takes up to X targets (at X = 0 it is dropped and takes none), and every X bound, the total included, is bound to the source's `Card.CastX`. The target walk's rebuild on restore runs the same function off the restored source, so an open walk keeps its X; on the stack, `itemAnnouncedClauses` binds the total from `XValue` for the CR 608.2b re-sum. Data only: no closure route, no snapshot change. `effects.Register` allows a total of X only on an "up to X" clause. **Cards:** Orcus, Prince of Undeath (Full).
