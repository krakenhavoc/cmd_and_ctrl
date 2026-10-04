---
title: "Alternative costs for every spell you cast"
date: 2026-10-04
issues: [2163]
---
**Alternative costs for every spell you cast** (#2163, CR 118.9, [ADR 0118](decisions/0118-strict-payment-by-default-and-alternative-costs-for-every-spell.md) §3) — `Spec.GrantedAlternativeCosts` declares a static that offers its controller one more alternative cost for each spell they cast. The offers are derived from the battlefield on every cast query (`grantedAlternativeCostsLocked`), so nothing is stored and an offer lasts exactly as long as its source is under its controller; a permanent that has lost its abilities grants nothing. They sit last in `CastOffersForLocked`, claimable only where the printed mana cost could be paid (CR 118.9a), and reach the cost picker through `card.alternative_costs`. Helpers: `PayWUBRGForSpellsYouCast` (`granted-wubrg`) and `CastFromHandWithoutPayingManaCost` (`granted-free`, hand only). **Cards:** Fist of Suns, Jodah, Archmage Eternal and Omniscience in full, and Leyline of Mutation with the other Leylines' caveat that it can't begin the game on the battlefield (CR 103.6a). Hunting Velociraptor's granted prowl stays on `ability-cost-modification`.
