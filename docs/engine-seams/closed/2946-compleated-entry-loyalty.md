---
title: "Compleated entry loyalty"
date: 2026-10-14
issues: [2946]
---
**Compleated entry loyalty** (#2946, CR 702.150a / 306.5b, [ADR 0183](decisions/0183-compleated-entry-loyalty.md)) — `PaidCost.PhyrexianLife` counts the printed Phyrexian symbols, hybrid Phyrexian included, that a cast paid with life. A `{B}` paid with life under K'rrik's grant and fixed life costs are not counted. `CastProvenance.PhyrexianLife` carries the count onto the permanent (`Card.PhyrexianLifePaid`). The starting-loyalty placement is marked `ReplacementEvent.StartingLoyalty`, and `effects.Compleated()` is the card's own replacement on it: two fewer per symbol, and none at zero. It is ordered against a doubler by the controller under CR 616, so a Doubling Season gives 8 or 10. A planeswalker that was not cast, or was copied, enters with its printed loyalty. **Cards:** Vraska, Betrayal's Sting (Full; her −2 is The Irencrag's indefinite Treasure conversion, and her −9 places the poison difference through the player-counter window).
