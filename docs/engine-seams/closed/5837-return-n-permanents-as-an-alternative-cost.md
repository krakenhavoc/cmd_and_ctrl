---
title: "Returning N permanents to hand as an alternative cost"
date: 2026-11-19
issues: [5837]
---
**Returning N permanents to hand as an alternative cost** (#5837, CR 118.3, 118.9, 205.3i, 400.3, 601.2h) — Daze's `AlternativeCost.ReturnToHand` named exactly one permanent: `cardComponent` returned a count of 1 whatever the clause said, and the payer moved only the first ID. The count is now the clause's `Min`, one when unset, the reading a pitch and escape's exile already had, so `ReturnInstead(…, PermanentYouControl(…).WithCount(n, n), …)` asks for exactly n distinct permanents the caster controls, hides the offer from a board with fewer (CR 118.3) and returns every one of them to its owner's hand (CR 400.3) with the spell already on the stack (CR 601.2h). The view already shipped the clause's `min`/`max` on `pay_options`, and the bot enumerator already read `CardPaymentCount`, so neither changed. `effects.Register` refuses a return alternative cost with any count but a fixed one of at least one. No new stored state, snapshot field, effect key, closure route or wire field. **Cards** (3, all Full): Ensnare, Gush and Thwart.
