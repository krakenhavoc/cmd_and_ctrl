---
title: "Delve, the payment"
date: 2026-09-30
issues: [1732]
---
**Delve, the payment** (#1732, [ADR 0100](decisions/0100-delve-either-or-and-variable-sacrifice-costs.md) sub-PR 1) — a spell with `Spec.Delve` may exile cards from its caster's graveyard to pay the generic mana in its total cost (CR 702.66a). The budget is `CastPrice.DelveBudget`, priced after the cost modifiers and the convoke taps and without the colours a "spend as though any colour" grant folded in, so the validator, the view, the auto-tap preview and the bot all read one number. The cards are exiled at CR 601.2h with the spell on the stack, a delved commander is asked about CR 903.9 first, and `PaidCost.Delved` records the objects that landed in exile. Still open: the readers of the cards "exiled with" a permanent (Murktide Regent, Soulflayer, CR 607.2q) and Teval's granted delve — see the Delve row.
