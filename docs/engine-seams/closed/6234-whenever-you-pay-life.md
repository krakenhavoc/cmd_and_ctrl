---
title: "Whenever you pay life"
date: 2026-11-25
issues: [6234]
---
**Whenever you pay life** (#6234, CR 119.4) — every life payment, whether a spell's or ability's cost, Phyrexian mana, ward, a mana ability's "Pay 1 life" or a "pay N life" choice, goes through `payLifeAsCostLocked`. Since #5948 its `EventChangeLife` already carries `LifePayment` (`life_payment`), so the engine needed no new event. Life lost to damage or to an effect carries no flag. The trigger matcher is `effects.YouPayLife`: the source's controller paid life, and "that many" is `-ev.Amount`, the life that actually moved after any replacement. Engine tests pin Phyrexian mana and the activated-cost door against damage and an effect's loss. **Cards**: Font of Agonies (OLD b177) is built by its OLD batch.
