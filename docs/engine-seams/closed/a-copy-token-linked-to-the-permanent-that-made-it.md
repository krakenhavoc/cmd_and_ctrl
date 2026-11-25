---
title: "A copy token linked to the permanent that made it"
date: 2026-11-25
issues: [6173]
---
**A copy token linked to the permanent that made it** (#6173, CR 607.2) — already on the engine: `Card.LinkedPermanent` (the object an ability of this permanent affected, with its CR 400.7 epoch, carried by the snapshot) and `Game.LinkedPermanentForEffect`, which reads the link off the permanent as it last existed (CR 608.2h). The creating trigger sets the link from its token creation's keyed tail (`CreateTokenCopiesTogether.Then`), the permanent's own leave trigger exiles the linked token, and "when the token leaves the battlefield" is the permanent's own `EventLTB` trigger matching its link. No engine change. **Cards:** Dance of Many (Full).
