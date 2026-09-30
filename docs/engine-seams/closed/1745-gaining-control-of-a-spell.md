---
title: "Gaining control of a spell"
date: 2026-09-30
issues: [1745]
pr: 1813
---
**Gaining control of a spell** (#1745, [ADR 0104](decisions/0104-gaining-control-of-a-spell.md)) — a spell's control is layer 2 (CR 611.1, CR 613.1b), so a steal is a `setController` `ScopedEffect` pinned to the stack object (`AffectedObject.OnStack`/`Epoch`, `Duration.PinnedOnStack`). A stack step of the layer pass (`stackControlPassLocked`) materialises `StackItem.Controller` and the stack card's controller. At resolution the record is re-pinned to the permanent by its epoch, with `Card.BaseController` set to the caster (CR 110.2b, CR 400.7a). `GainControlOfSpellForEffect` and `ExchangeControlOfSpellAndPermanentForEffect` (CR 701.12, one timestamp); card-side `GainControlOfSpell` and `ExchangeControlOfSpellAnd` offer the new controller new targets. A departed player's control effects now end (CR 800.4a), which also fixed Act of Treason's reversion. Shipped on Invert Polarity, Aethersnatch, Commandeer (with `ExileFromHand` taking a count), Sudden Substitution and Perplexing Chimera. Still open: Chef's Kiss's random retarget (#1815).
