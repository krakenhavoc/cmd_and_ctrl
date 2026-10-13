---
title: "Conditional flash on X (per-offer X ceiling on AlternativeCost)"
date: 2026-10-13
issues: [3225]
---
**Conditional flash on X (per-offer X ceiling on AlternativeCost)** (#3225, CR 601.3c) — "You may cast this spell as though it had flash if X is 3 or less" could not be said: the flat grant opens instant speed for every X, and the alternative cost that opens it for one claim (`AsThoughFlash`) bounded nothing, so shipping either would have let Silver Scrutiny draw ten cards at instant speed. `AlternativeCost.XCeiling` is the condition, carried on the offer and built by `effects.FlashIfXAtMost`; `game.SpellXCeilingForOfferLocked` folds it with a card's printed `Spec.XCeiling` and is read by the announce gate (X above it under the claim is refused), the enumerator's `printedXCeiling` (no instant-speed move above it), the view (`alternative_costs[].x_max`) and the auto-tap preview, and the client's X picker caps at the claimed offer's `x_max`. The sorcery-speed cast for the printed cost takes any X. **Cards:** Silver Scrutiny (`full`). **Still open:** a granted offer (`GrantedAlternativeCost`) cannot carry a ceiling; no card needs one yet.
