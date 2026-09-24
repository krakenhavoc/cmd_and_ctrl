---
title: "\"You gain life equal to the life lost this way\""
date: 2026-09-17
issues: [793]
pr: 806
legacy_order: 130
---
**"You gain life equal to the life lost this way"** (#793) - a life change can pause on a CR 616 ordering prompt, so the amount is taken from a continuation rather than read back off the life total: `Game.ChangePlayerLifeThenForEffect(source, player, delta, then)` for one player and `Game.LoseLifeEachThenForEffect(source, players, amount, then)` for the each-opponent drain, both in `effect_api.go` on the tail in `life_tail.go`. `then` gets the post-replacement amount, or 0 when the change was replaced away. Paying life is a separate entry point, `Game.PayLifeForEffect` - a cost (CR 118.3) still runs the window (CR 119.4) but settles without a prompt. Exsanguinate, Debt to the Deathless, Gray Merchant of Asphodel and Kokusho share one body on it; a card whose sentence is "target opponent loses N, you gain that much" writes the single form. Since #808 a player who leaves the game (including while their CR 616 prompt is open) is a zero-amount terminal outcome, so the rest of the drain still runs, and a payment the window cancels refuses the cost (CR 119.8).
