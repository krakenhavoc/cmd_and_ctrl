---
title: "Mana-spent-to-cast record on the stack item"
date: 2026-09-18
issues: [761]
pr: 958
legacy_order: 69
---
**Mana-spent-to-cast record on the stack item** (#761, [ADR 0068](decisions/0068-the-mana-spent-on-a-spell.md)) - `StackItem.Paid` is a `game.PaidCost{Mana, OnPaper, CountersRemoved, CountersAdded, LifePaid}`: `attemptSpend` and `SpendManaFor` return the tokens they spent, and every payment path records them. A permissive or `ForceCast` cast records `OnPaper` — "the engine waived this charge" — rather than an empty record, so "if no mana was spent" is never ambiguous, and every reader takes unknown as the weaker-than-printed answer. A copy of a spell records a real zero (CR 707.10). The solver gained one strategy switch, `SpendDistinctColors`, thrown by `Spec.WantsDistinctColors` and applied to the GENERIC half only, so it can never change whether a cost is payable. `EventManaSpent` and `StackItemView` carry the amount and the distinct colours. Readers: `ctx.ColorsSpent`, `ColorsSpentCount`, `ManaSpentOfColor`, `NoManaSpent`, `ManaSpentKnown`, `Paid`, and `Game.StackItemPaidForEffect` for a cast trigger. Shipped on Painful Truths (converge), Etched Oracle (sunburst), Slaying Fire (adamant) and Vexing Bauble ("if no mana was spent"). The row above is still open for SPEND RIDERS.
