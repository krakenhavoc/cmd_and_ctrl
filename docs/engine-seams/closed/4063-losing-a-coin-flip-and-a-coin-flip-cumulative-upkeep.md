---
title: "Losing a coin flip, and a coin-flip cumulative upkeep"
date: 2026-10-19
issues: [4063]
---
**Losing a coin flip, and a coin-flip cumulative upkeep** (#4063, CR 705.2, CR 702.24a, [ADR 0054](decisions/0054-dice-rolls-and-coin-flips.md) amendment 2026-10-19) — "Whenever you lose a coin flip" is `effects.WheneverYouLoseACoinFlip`, matching an `EventFlipCoin` the source's controller flipped, with `Won` unset and a `Call` (only a called flip can be lost, so a face-only flip never matches): one trigger per lost coin, never for an opponent's flip or a table flip. "Cumulative upkeep—Flip a coin" is `CumulativeUpkeepPaying` over `effects.FlipCoinsPayment`, a payment that costs nothing to say yes to and whose flips are queued from the prompt's `OnPay`: one coin per age counter under one call, each a won or lost flip, and not stoppable once started (the 2006-07-15 ruling); no sacrifices the permanent. The coin is the game's keyed stream, so a restore taken after the flip replays the next one. **Cards:** Karplusan Minotaur (Full), whose lose ability's target is an opponent's choice (#2055).
